package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/glb/nw-api-gogin/internal/api"
	"github.com/glb/nw-api-gogin/internal/auth"
	"github.com/glb/nw-api-gogin/internal/catalog"
	"github.com/glb/nw-api-gogin/internal/db"
	httpmw "github.com/glb/nw-api-gogin/internal/http/middleware"
	"github.com/glb/nw-api-gogin/pkg/logger"
	"github.com/glb/nw-api-gogin/pkg/metrics"
)

func main() {
	gin.SetMode(envOrDefault("GIN_MODE", gin.ReleaseMode))

	appLog, err := logger.New(envOrDefault("LOG_LEVEL", "info"))
	if err != nil {
		panic(err)
	}
	defer logger.Sync(appLog)

	dbCfg := db.LoadConfig()
	gormDB, err := db.Connect(dbCfg)
	if err != nil {
		appLog.Fatal("failed to connect database", zap.Error(err))
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		appLog.Fatal("failed to access sql db", zap.Error(err))
	}
	defer sqlDB.Close()

	repo := catalog.NewRepository(gormDB)
	catalogService := catalog.NewService(repo, appLog.Named("catalog"))

	tokenSecret := strings.TrimSpace(os.Getenv("TOKEN_SECRET"))
	if tokenSecret == "" {
		appLog.Warn("TOKEN_SECRET not set, using development default")
		tokenSecret = "development-secret"
	}

	tokenKeyID := strings.TrimSpace(os.Getenv("TOKEN_KEY_ID"))
	if tokenKeyID == "" {
		tokenKeyID = "primary"
	}

	keyManager, err := auth.NewHMACKeyManager([]byte(tokenSecret), tokenKeyID)
	if err != nil {
		appLog.Fatal("failed to setup token key manager", zap.Error(err))
	}

	tokenTTL := time.Hour
	if rawTTL := strings.TrimSpace(os.Getenv("TOKEN_TTL")); rawTTL != "" {
		if d, parseErr := time.ParseDuration(rawTTL); parseErr != nil {
			appLog.Warn("invalid TOKEN_TTL, falling back to default", zap.String("value", rawTTL), zap.Error(parseErr))
		} else {
			tokenTTL = d
		}
	}

	audience := []string{"northwind-api"}
	if rawAudience := strings.TrimSpace(os.Getenv("TOKEN_AUDIENCE")); rawAudience != "" {
		parts := strings.Split(rawAudience, ",")
		parsed := make([]string, 0, len(parts))
		for _, part := range parts {
			if value := strings.TrimSpace(part); value != "" {
				parsed = append(parsed, value)
			}
		}
		if len(parsed) > 0 {
			audience = parsed
		}
	}

	authUsername := envOrDefault("AUTH_ADMIN_USERNAME", "admin")
	authPassword := envOrDefault("AUTH_ADMIN_PASSWORD", "changeit")
	if authUsername == "admin" && authPassword == "changeit" {
		appLog.Warn("using default admin credentials; set AUTH_ADMIN_USERNAME and AUTH_ADMIN_PASSWORD for production")
	}

	authenticator, err := auth.NewStaticAuthenticator(map[string]struct {
		Password  string
		Principal auth.Principal
	}{
		authUsername: {
			Password: authPassword,
			Principal: auth.Principal{
				Subject: authUsername,
				Scopes:  []string{"admin", "manager", "viewer"},
			},
		},
	})
	if err != nil {
		appLog.Fatal("failed to initialize authenticator", zap.Error(err))
	}

	tokenService, err := auth.NewService(auth.Config{
		Issuer:         envOrDefault("TOKEN_ISSUER", "northwind-api"),
		Audience:       audience,
		AccessTokenTTL: tokenTTL,
	}, authenticator, keyManager)
	if err != nil {
		appLog.Fatal("failed to initialize token service", zap.Error(err))
	}

	httpMetrics := metrics.NewHTTPMetrics(nil)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(httpmw.RequestID())
	r.Use(httpMetrics.Middleware())
	r.Use(httpmw.Logging(appLog))

	swagger, err := api.GetSwagger()
	if err != nil {
		appLog.Fatal("failed to load swagger", zap.Error(err))
	}
	swagger.Servers = nil

	r.Use(ginmiddleware.OapiRequestValidator(swagger))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			appLog.Warn("readiness check failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	h := api.NewHandler(catalogService, tokenService)
	api.RegisterHandlersWithOptions(r, h, api.GinServerOptions{
		Middlewares: []api.MiddlewareFunc{
			api.AuthMiddleware(tokenService),
		},
		ErrorHandler: func(c *gin.Context, handlerErr error, statusCode int) {
			traceID := httpmw.RequestIDFromContext(c.Request.Context())
			resp := api.ErrorResponse{Code: "invalid_request", Message: handlerErr.Error()}
			if traceID != "" {
				resp.TraceId = &traceID
			}
			c.JSON(statusCode, resp)
		},
	})

	addr := serverAddress()
	appLog.Info("starting http server", zap.String("address", addr))
	if err := r.Run(addr); err != nil {
		appLog.Fatal("server shutdown with error", zap.Error(err))
	}
}

func envOrDefault(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

func serverAddress() string {
	if addr := strings.TrimSpace(os.Getenv("SERVER_ADDR")); addr != "" {
		return addr
	}
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}
	if strings.HasPrefix(port, ":") {
		return port
	}
	return ":" + port
}
