package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewHTTPMetricsRegistersCollectors(t *testing.T) {
	registry := prometheus.NewRegistry()

	m := NewHTTPMetrics(registry)

	if m.RequestsTotal == nil || m.RequestDuration == nil {
		t.Fatalf("expected metrics vectors to be initialized")
	}

	if err := registry.Register(m.RequestsTotal); err == nil {
		t.Fatalf("expected counter to already be registered")
	}
	if err := registry.Register(m.RequestDuration); err == nil {
		t.Fatalf("expected histogram to already be registered")
	}
}

func TestHTTPMetricsMiddlewareRecordsMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registry := prometheus.NewRegistry()
	m := NewHTTPMetrics(registry)

	router := gin.New()
	router.Use(m.Middleware())
	router.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "world")
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}

	labels := prometheus.Labels{"method": "GET", "path": "/hello", "status": "200"}
	if count := testutil.ToFloat64(m.RequestsTotal.With(labels)); count != 1 {
		t.Fatalf("expected request counter to increment, got %f", count)
	}

	if samples := testutil.CollectAndCount(m.RequestDuration, "http_request_duration_seconds"); samples == 0 {
		t.Fatalf("expected histogram to record observations")
	}
}

func TestFormatStatus(t *testing.T) {
	if got := formatStatus(204); got != "204" {
		t.Fatalf("expected string representation of status, got %q", got)
	}
}
