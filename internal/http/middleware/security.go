package middleware

import "github.com/gin-gonic/gin"

// SecurityConfig defines default security headers.
type SecurityConfig struct {
	ContentTypeOptions      bool
	FrameOptions            string
	ReferrerPolicy          string
	ContentSecurityPolicy   string
	StrictTransportSecurity string
	XSSProtection           string
	PermissionsPolicy       string
}

// DefaultSecurityConfig returns opinionated defaults for secure headers.
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		ContentTypeOptions:      true,
		FrameOptions:            "DENY",
		ReferrerPolicy:          "no-referrer",
		ContentSecurityPolicy:   "default-src 'self'",
		StrictTransportSecurity: "max-age=63072000; includeSubDomains; preload",
		XSSProtection:           "0",
		PermissionsPolicy:       "geolocation=(), microphone=(), camera=()",
	}
}

// SecurityHeaders applies common security headers. When running in non-TLS environments,
// consider overriding StrictTransportSecurity to avoid signalling HTTPS requirements.
func SecurityHeaders(cfg SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.ContentTypeOptions {
			c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		}
		if cfg.FrameOptions != "" {
			c.Writer.Header().Set("X-Frame-Options", cfg.FrameOptions)
		}
		if cfg.ReferrerPolicy != "" {
			c.Writer.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
		}
		if cfg.ContentSecurityPolicy != "" {
			c.Writer.Header().Set("Content-Security-Policy", cfg.ContentSecurityPolicy)
		}
		if cfg.StrictTransportSecurity != "" {
			c.Writer.Header().Set("Strict-Transport-Security", cfg.StrictTransportSecurity)
		}
		if cfg.XSSProtection != "" {
			c.Writer.Header().Set("X-XSS-Protection", cfg.XSSProtection)
		}
		if cfg.PermissionsPolicy != "" {
			c.Writer.Header().Set("Permissions-Policy", cfg.PermissionsPolicy)
		}
		c.Next()
	}
}
