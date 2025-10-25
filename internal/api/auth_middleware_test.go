package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/glb/nw-api-gogin/internal/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddleware_Success(t *testing.T) {
	tokenSvc := &tokenServiceStub{
		validateTokenFn: func(ctx context.Context, token string) (*auth.Claims, error) {
			return &auth.Claims{
				Scopes:           []string{"viewer"},
				RegisteredClaims: jwt.RegisteredClaims{Subject: "user"},
			}, nil
		},
	}

	recorder := httptest.NewRecorder()
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(BearerAuthScopes, []string{"viewer"})
	})
	r.Use(gin.HandlerFunc(AuthMiddleware(tokenSvc)))
	called := false
	r.GET("/protected", func(c *gin.Context) {
		if _, ok := PrincipalFromContext(c); !ok {
			t.Fatalf("expected principal in context")
		}
		called = true
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer good-token")

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", recorder.Code)
	}
	if !called {
		t.Fatalf("expected handler to be invoked")
	}
}

func TestAuthMiddleware_Unauthorized(t *testing.T) {
	tokenSvc := &tokenServiceStub{}

	recorder := httptest.NewRecorder()
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(BearerAuthScopes, []string{"viewer"})
	})
	r.Use(gin.HandlerFunc(AuthMiddleware(tokenSvc)))
	r.GET("/protected", func(c *gin.Context) {
		t.Fatalf("handler should not run without token")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}
}

func TestAuthMiddleware_ForbiddenScope(t *testing.T) {
	tokenSvc := &tokenServiceStub{
		validateTokenFn: func(ctx context.Context, token string) (*auth.Claims, error) {
			return &auth.Claims{
				Scopes:           []string{"viewer"},
				RegisteredClaims: jwt.RegisteredClaims{Subject: "user"},
			}, nil
		},
	}

	recorder := httptest.NewRecorder()
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(BearerAuthScopes, []string{"admin"})
	})
	r.Use(gin.HandlerFunc(AuthMiddleware(tokenSvc)))
	r.GET("/admin", func(c *gin.Context) {
		t.Fatalf("handler should not run without admin scope")
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.Header.Set("Authorization", "Bearer token")

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}
}
