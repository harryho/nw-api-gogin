package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestLoggingMiddlewareEmitsInfoLog(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.InfoLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(RequestID())
	router.Use(Logging(logger))
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(RequestIDHeader, "req-123")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected a single log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zap.InfoLevel {
		t.Fatalf("expected info level, got %s", entry.Level)
	}
	if entry.Message != "request completed" {
		t.Fatalf("unexpected log message: %q", entry.Message)
	}

	contextFields := entry.ContextMap()
	if contextFields["request_id"] != "req-123" {
		t.Fatalf("expected request_id field, got %v", contextFields["request_id"])
	}
	if contextFields["path"] != "/ping" {
		t.Fatalf("expected path field \"/ping\", got %v", contextFields["path"])
	}
}

func TestLoggingMiddlewareEmitsErrorLogWhenErrorsPresent(t *testing.T) {
	gin.SetMode(gin.TestMode)

	core, logs := observer.New(zap.DebugLevel)
	logger := zap.New(core)

	router := gin.New()
	router.Use(Logging(logger))
	router.GET("/fail", func(c *gin.Context) {
		_ = c.Error(errors.New("boom"))
		c.Status(http.StatusTeapot)
	})

	req := httptest.NewRequest(http.MethodGet, "/fail", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTeapot {
		t.Fatalf("expected status 418, got %d", resp.Code)
	}

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected a single log entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != zap.ErrorLevel {
		t.Fatalf("expected error level, got %s", entry.Level)
	}
	if entry.Message != "request completed with errors" {
		t.Fatalf("unexpected log message: %q", entry.Message)
	}
}

func TestLoggingMiddlewareHandlesNilLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(Logging(nil))
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
}
