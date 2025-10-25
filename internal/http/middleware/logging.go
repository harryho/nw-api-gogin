package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Logging(log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start)
		reqLog := log.With(
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		if reqID := RequestIDFromContext(c.Request.Context()); reqID != "" {
			reqLog = reqLog.With(zap.String("request_id", reqID))
		}

		if len(c.Errors) > 0 {
			reqLog.Error("request completed with errors", zap.String("errors", c.Errors.String()))
			return
		}

		reqLog.Info("request completed")
	}
}
