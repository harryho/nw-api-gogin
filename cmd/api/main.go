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

	h := api.NewHandler(catalogService)
	api.RegisterHandlersWithOptions(r, h, api.GinServerOptions{
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
