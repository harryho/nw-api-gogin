package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddlewareGeneratesIdentifierWhenMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var ctxID, storedID, headerID string

	router := gin.New()
	router.Use(RequestID())
	router.GET("/ping", func(c *gin.Context) {
		ctxID = RequestIDFromContext(c.Request.Context())
		value, _ := c.Get(RequestIDHeader)
		storedID, _ = value.(string)
		headerID = c.Writer.Header().Get(RequestIDHeader)
		c.String(http.StatusOK, "pong")
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	if ctxID == "" {
		t.Fatalf("expected middleware to populate context request id")
	}
	if ctxID != storedID {
		t.Fatalf("expected context and stored IDs to match: %q vs %q", ctxID, storedID)
	}
	if headerID != ctxID {
		t.Fatalf("expected response header to match context request id")
	}
}

func TestRequestIDMiddlewarePreservesExistingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const existing = "existing-id"
	router := gin.New()
	router.Use(RequestID())
	router.GET("/ping", func(c *gin.Context) {
		if got := RequestIDFromContext(c.Request.Context()); got != existing {
			t.Fatalf("expected request id %q, got %q", existing, got)
		}
		stored, _ := c.Get(RequestIDHeader)
		storedID, _ := stored.(string)
		if storedID != existing {
			t.Fatalf("expected stored id %q, got %q", existing, storedID)
		}
		if header := c.Writer.Header().Get(RequestIDHeader); header != existing {
			t.Fatalf("expected header id %q, got %q", existing, header)
		}
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(RequestIDHeader, existing)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.Code)
	}
}

func TestRequestIDFromContextHandlesMissingValues(t *testing.T) {
	if RequestIDFromContext(nil) != "" {
		t.Fatalf("expected empty id for nil context")
	}
	if RequestIDFromContext(context.Background()) != "" {
		t.Fatalf("expected empty id when context lacks value")
	}
}
