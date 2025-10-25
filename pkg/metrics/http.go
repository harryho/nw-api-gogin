package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type HTTPMetrics struct {
	RequestsTotal   *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	registerer      prometheus.Registerer
}

func NewHTTPMetrics(registerer prometheus.Registerer) *HTTPMetrics {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}
	m := &HTTPMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Histogram of HTTP request durations in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status"},
		),
		registerer: registerer,
	}

	registerer.MustRegister(m.RequestsTotal, m.RequestDuration)

	return m
}

func (m *HTTPMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		status := c.Writer.Status()
		labels := prometheus.Labels{
			"method": c.Request.Method,
			"path":   c.FullPath(),
			"status": formatStatus(status),
		}

		m.RequestsTotal.With(labels).Inc()
		m.RequestDuration.With(labels).Observe(time.Since(start).Seconds())
	}
}

func formatStatus(status int) string {
	return strconv.Itoa(status)
}
