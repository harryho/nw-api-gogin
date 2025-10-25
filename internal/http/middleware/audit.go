package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PrincipalExtractor extracts subject and scopes for audit logging.
type PrincipalExtractor func(*gin.Context) (string, []string)

// Audit emits structured logs for mutating requests.
func Audit(log *zap.Logger, extract PrincipalExtractor) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}
	if extract == nil {
		extract = func(*gin.Context) (string, []string) { return "", nil }
	}
	mutating := map[string]struct{}{
		http.MethodPost:   {},
		http.MethodPut:    {},
		http.MethodPatch:  {},
		http.MethodDelete: {},
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		if _, ok := mutating[c.Request.Method]; !ok {
			return
		}

		subject, scopes := extract(c)
		entry := log.With(
			zap.String("method", c.Request.Method),
			zap.String("path", fallbackPath(c)),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
		if subject != "" {
			entry = entry.With(zap.String("subject", subject))
		}
		if len(scopes) > 0 {
			entry = entry.With(zap.Strings("scopes", scopes))
		}
		if reqID := RequestIDFromContext(c.Request.Context()); reqID != "" {
			entry = entry.With(zap.String("request_id", reqID))
		}
		if len(c.Errors) > 0 {
			entry = entry.With(zap.String("errors", c.Errors.String()))
		}

		entry.Info("audit event")
	}
}

func fallbackPath(c *gin.Context) string {
	if p := c.FullPath(); p != "" {
		return p
	}
	return c.Request.URL.Path
}
