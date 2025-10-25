package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimitConfig controls the IP based rate limiter.
type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
	TTL               time.Duration
}

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimit enforces a token bucket rate limit per client IP address.
func RateLimit(log *zap.Logger, cfg RateLimitConfig) gin.HandlerFunc {
	if log == nil {
		log = zap.NewNop()
	}
	if cfg.RequestsPerSecond <= 0 {
		cfg.RequestsPerSecond = 5
	}
	if cfg.Burst <= 0 {
		cfg.Burst = 10
	}
	if cfg.TTL <= 0 {
		cfg.TTL = 10 * time.Minute
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*clientLimiter)
	)

	cleanup := func(now time.Time) {
		for key, entry := range clients {
			if now.Sub(entry.lastSeen) > cfg.TTL {
				delete(clients, key)
			}
		}
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		entry, ok := clients[ip]
		if !ok {
			entry = &clientLimiter{limiter: rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.Burst), lastSeen: now}
			clients[ip] = entry
		} else {
			entry.lastSeen = now
		}
		cleanup(now)
		mu.Unlock()

		if !entry.limiter.Allow() {
			log.Warn("rate limit exceeded", zap.String("client_ip", ip), zap.Float64("rps", cfg.RequestsPerSecond), zap.Int("burst", cfg.Burst))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}

		c.Next()
	}
}
