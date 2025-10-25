package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/glb/nw-api-gogin/internal/auth"
	httpmw "github.com/glb/nw-api-gogin/internal/http/middleware"
)

const principalContextKey = "auth.principal"

// AuthMiddleware enforces bearer authentication and scope requirements extracted from the generated handler wrappers.
func AuthMiddleware(tokens TokenService) MiddlewareFunc {
	return func(c *gin.Context) {
		required := requiredScopes(c)
		if len(required) == 0 {
			c.Next()
			return
		}

		if tokens == nil {
			writeAuthError(c, http.StatusUnauthorized, string(auth.ErrorInvalidToken), "authentication service unavailable")
			c.Abort()
			return
		}

		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			writeAuthError(c, http.StatusUnauthorized, string(auth.ErrorInvalidToken), "missing or invalid authorization header")
			c.Abort()
			return
		}

		token := strings.TrimSpace(authHeader[len("Bearer "):])
		if token == "" {
			writeAuthError(c, http.StatusUnauthorized, string(auth.ErrorInvalidToken), "missing bearer token value")
			c.Abort()
			return
		}

		claims, err := tokens.ValidateToken(c.Request.Context(), token)
		if err != nil {
			status := http.StatusUnauthorized
			code := auth.ErrorInvalidToken
			message := "invalid token"
			if authErr, ok := auth.AsError(err); ok {
				code = authErr.Code
				message = authErr.Message
				if code == auth.ErrorInvalidScope {
					status = http.StatusForbidden
				}
			}
			writeAuthError(c, status, string(code), message)
			c.Abort()
			return
		}

		principal := auth.Principal{Subject: claims.Subject, Scopes: auth.SanitizeScopes(claims.Scopes)}
		if !auth.HasAllScopes(required, principal.Scopes) {
			writeAuthError(c, http.StatusForbidden, string(auth.ErrorInvalidScope), "insufficient scope")
			c.Abort()
			return
		}

		c.Set(principalContextKey, principal)
		c.Next()
	}
}

func requiredScopes(c *gin.Context) []string {
	value, ok := c.Get(BearerAuthScopes)
	if !ok {
		return nil
	}
	scopes, _ := value.([]string)
	return auth.SanitizeScopes(scopes)
}

func writeAuthError(c *gin.Context, status int, code, message string) {
	traceID := httpmw.RequestIDFromContext(c.Request.Context())
	resp := ErrorResponse{Code: code, Message: message}
	if traceID != "" {
		resp.TraceId = &traceID
	}
	c.JSON(status, resp)
}

// PrincipalFromContext retrieves the current principal from the Gin context when available.
func PrincipalFromContext(c *gin.Context) (auth.Principal, bool) {
	value, ok := c.Get(principalContextKey)
	if !ok {
		return auth.Principal{}, false
	}
	principal, ok := value.(auth.Principal)
	return principal, ok
}
