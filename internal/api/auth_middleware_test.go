package api

import (
	"context"
	"encoding/json"
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

func TestAuthMiddleware_InvalidAuthorizationHeader(t *testing.T) {
	tokenSvc := &tokenServiceStub{
		validateTokenFn: func(ctx context.Context, token string) (*auth.Claims, error) {
			t.Fatalf("validate should not be called")
			return nil, nil
		},
	}

	recorder := httptest.NewRecorder()
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(BearerAuthScopes, []string{"viewer"}) })
	r.Use(gin.HandlerFunc(AuthMiddleware(tokenSvc)))
	r.GET("/protected", func(c *gin.Context) {
		t.Fatalf("handler should not run for invalid authorization header")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token invalid")

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != string(auth.ErrorInvalidToken) {
		t.Fatalf("expected error code invalid_token, got %q", resp.Code)
	}
}

func TestAuthMiddleware_InvalidScopeFromService(t *testing.T) {
	tokenSvc := &tokenServiceStub{
		validateTokenFn: func(ctx context.Context, token string) (*auth.Claims, error) {
			return nil, auth.NewError(auth.ErrorInvalidScope, "requires admin", nil)
		},
	}

	recorder := httptest.NewRecorder()
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(BearerAuthScopes, []string{"viewer"}) })
	r.Use(gin.HandlerFunc(AuthMiddleware(tokenSvc)))
	r.GET("/protected", func(c *gin.Context) {
		t.Fatalf("handler should not run when scope is invalid")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer token")

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != string(auth.ErrorInvalidScope) {
		t.Fatalf("expected error code invalid_scope, got %q", resp.Code)
	}
	if resp.Message != "requires admin" {
		t.Fatalf("expected message 'requires admin', got %q", resp.Message)
	}
}

func TestAuthMiddleware_ServiceUnavailable(t *testing.T) {
	recorder := httptest.NewRecorder()
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set(BearerAuthScopes, []string{"viewer"}) })
	r.Use(gin.HandlerFunc(AuthMiddleware(nil)))
	r.GET("/protected", func(c *gin.Context) {
		t.Fatalf("handler should not run when service unavailable")
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", recorder.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Message != "authentication service unavailable" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}
