package app

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/harryho/nw-api-gogin/internal/api"
	"github.com/harryho/nw-api-gogin/internal/auth"
	"github.com/harryho/nw-api-gogin/internal/catalog"
	"github.com/harryho/nw-api-gogin/internal/db"
	httpmw "github.com/harryho/nw-api-gogin/internal/http/middleware"
	"github.com/harryho/nw-api-gogin/pkg/logger"
	"github.com/harryho/nw-api-gogin/pkg/metrics"
	"github.com/harryho/nw-api-gogin/pkg/telemetry"
)

// Application represents the fully configured HTTP application.
type Application struct {
	Engine *gin.Engine
	Logger *zap.Logger

	shutdownMu    sync.Mutex
	shutdownFuncs []func(context.Context)
	closed        bool
}

// New constructs a fully wired Gin engine along with its supporting defaults based on environment configuration.
func New(ctx context.Context) (*Application, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	gin.SetMode(envOrDefault("GIN_MODE", gin.ReleaseMode))

	appLog, err := logger.New(envOrDefault("LOG_LEVEL", "info"))
	if err != nil {
		return nil, err
	}

	application := &Application{Logger: appLog}
	application.addShutdown(func(context.Context) { logger.Sync(appLog) })

	telemetryShutdown, err := telemetry.Setup(ctx, telemetry.Config{
		Endpoint:       strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		Headers:        parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")),
		Insecure:       strings.EqualFold(strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_INSECURE")), "true"),
		ServiceName:    envOrDefault("SERVICE_NAME", "northwind-api"),
		ServiceVersion: envOrDefault("SERVICE_VERSION", "development"),
		Environment:    envOrDefault("APP_ENV", "local"),
	}, appLog.Named("telemetry"))
	if err != nil {
		appLog.Warn("failed to configure telemetry", zap.Error(err))
	} else {
		application.addShutdown(func(shutdownCtx context.Context) {
			if shutdownCtx == nil {
				shutdownCtx = context.Background()
			}
			timeoutCtx, cancel := context.WithTimeout(shutdownCtx, 5*time.Second)
			defer cancel()
			if shutdownErr := telemetryShutdown(timeoutCtx); shutdownErr != nil {
				appLog.Warn("failed to shutdown telemetry", zap.Error(shutdownErr))
			}
		})
	}

	dbCfg := db.LoadConfig()
	gormDB, err := db.Connect(dbCfg)
	if err != nil {
		application.Shutdown(context.Background())
		return nil, err
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		application.Shutdown(context.Background())
		return nil, err
	}
	application.addShutdown(func(context.Context) {
		if closeErr := sqlDB.Close(); closeErr != nil {
			appLog.Warn("failed to close database", zap.Error(closeErr))
		}
	})

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
		application.Shutdown(context.Background())
		return nil, err
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

	authHash, err := resolveAuthAdminHash(authPassword)
	if err != nil {
		application.Shutdown(context.Background())
		return nil, err
	}

	authenticator, err := auth.NewStaticAuthenticator(map[string]struct {
		PasswordHash []byte
		Principal    auth.Principal
	}{
		authUsername: {
			PasswordHash: authHash,
			Principal: auth.Principal{
				Subject: authUsername,
				Scopes:  []string{"admin", "manager", "viewer"},
			},
		},
	})
	if err != nil {
		application.Shutdown(context.Background())
		return nil, err
	}

	tokenService, err := auth.NewService(auth.Config{
		Issuer:         envOrDefault("TOKEN_ISSUER", "northwind-api"),
		Audience:       audience,
		AccessTokenTTL: tokenTTL,
	}, authenticator, keyManager)
	if err != nil {
		application.Shutdown(context.Background())
		return nil, err
	}

	httpMetrics := metrics.NewHTTPMetrics(nil)
	rateLimitRPS := envAsFloat("RATE_LIMIT_RPS", 25)
	rateLimitBurst := envAsInt("RATE_LIMIT_BURST", 50)

	securityCfg := httpmw.DefaultSecurityConfig()
	if strings.EqualFold(envOrDefault("DISABLE_HSTS", "false"), "true") {
		securityCfg.StrictTransportSecurity = ""
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(httpmw.RequestID())
	router.Use(otelgin.Middleware(envOrDefault("SERVICE_NAME", "northwind-api")))
	router.Use(httpmw.SecurityHeaders(securityCfg))
	router.Use(httpmw.RateLimit(appLog.Named("ratelimit"), httpmw.RateLimitConfig{
		RequestsPerSecond: rateLimitRPS,
		Burst:             rateLimitBurst,
		TTL:               15 * time.Minute,
	}))
	router.Use(httpMetrics.Middleware())
	router.Use(httpmw.Logging(appLog))
	router.Use(httpmw.Audit(appLog.Named("audit"), func(c *gin.Context) (string, []string) {
		if principal, ok := api.PrincipalFromContext(c); ok {
			return principal.Subject, principal.Scopes
		}
		return "", nil
	}))

	swagger, err := api.GetSwagger()
	if err != nil {
		application.Shutdown(context.Background())
		return nil, err
	}
	swagger.Servers = nil

	router.Use(ginmiddleware.OapiRequestValidator(swagger))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			appLog.Warn("readiness check failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	handler := api.NewHandler(catalogService, tokenService)
	api.RegisterHandlersWithOptions(router, handler, api.GinServerOptions{
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

	application.Engine = router
	return application, nil
}

// Shutdown attempts to gracefully shut down all managed resources.
func (a *Application) Shutdown(ctx context.Context) {
	if a == nil {
		return
	}
	a.shutdownMu.Lock()
	if a.closed {
		a.shutdownMu.Unlock()
		return
	}
	a.closed = true
	funcs := make([]func(context.Context), len(a.shutdownFuncs))
	copy(funcs, a.shutdownFuncs)
	a.shutdownMu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	for i := len(funcs) - 1; i >= 0; i-- {
		fn := funcs[i]
		if fn == nil {
			continue
		}
		safeCall(ctx, fn)
	}
}

// ServerAddress resolves the listen address from the environment with the
// same behaviour as the CLI binary uses.
func ServerAddress() string {
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

func (a *Application) addShutdown(fn func(context.Context)) {
	if fn == nil {
		return
	}
	a.shutdownFuncs = append(a.shutdownFuncs, fn)
}

func safeCall(ctx context.Context, fn func(context.Context)) {
	defer func() {
		_ = recover()
	}()
	fn(ctx)
}

// resolveAuthAdminHash returns the bcrypt hash for the seeded admin user.
// If AUTH_ADMIN_PASSWORD_HASH is set, it is used directly (for prod pre-hashed
// rotation). Otherwise the plaintext AUTH_ADMIN_PASSWORD is hashed at startup.
func resolveAuthAdminHash(plaintext string) ([]byte, error) {
	if h := os.Getenv("AUTH_ADMIN_PASSWORD_HASH"); h != "" {
		return []byte(h), nil
	}
	return bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
}

func envOrDefault(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return fallback
}

func envAsFloat(key string, fallback float64) float64 {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envAsInt(key string, fallback int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return parsed
}

func parseHeaders(raw string) map[string]string {
	result := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		p := strings.TrimSpace(pair)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" || value == "" {
			continue
		}
		result[key] = value
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
