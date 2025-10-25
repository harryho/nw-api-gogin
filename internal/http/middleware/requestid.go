package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		c.Writer.Header().Set(RequestIDHeader, reqID)
		ctx := context.WithValue(c.Request.Context(), requestIDKey{}, reqID)
		c.Request = c.Request.WithContext(ctx)
		c.Set(RequestIDHeader, reqID)
		c.Next()
	}
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		return id
	}
	return ""
}
