package auth

import (
	"context"
	"testing"
	"time"
)

func TestService_IssueAndValidateToken(t *testing.T) {
	svc := newTestService(t)
	token, err := svc.IssueToken(context.Background(), TokenIssueRequest{
		Username: "alice",
		Password: "password",
		Scopes:   []string{"viewer"},
	})
	if err != nil {
		t.Fatalf("issue token failed: %v", err)
	}
	if token.Subject != "user-alice" {
		t.Fatalf("expected subject user-alice, got %q", token.Subject)
	}

	claims, err := svc.ValidateToken(context.Background(), token.Value)
	if err != nil {
		t.Fatalf("validate token failed: %v", err)
	}
	if claims.RegisteredClaims.Subject != "user-alice" {
		t.Fatalf("expected subject user-alice, got %q", claims.RegisteredClaims.Subject)
	}
	if len(claims.Scopes) != 1 || claims.Scopes[0] != "viewer" {
		t.Fatalf("unexpected scopes: %v", claims.Scopes)
	}

	expectedExpiry := svc.now().Add(svc.cfg.AccessTokenTTL)
	if !token.ExpiresAt.Equal(expectedExpiry) {
		t.Fatalf("expected expiry %v, got %v", expectedExpiry, token.ExpiresAt)
	}
}

func TestService_IssueToken_InvalidScope(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.IssueToken(context.Background(), TokenIssueRequest{
		Username: "alice",
		Password: "password",
		Scopes:   []string{"admin"},
	})
	if err == nil {
		t.Fatalf("expected error for invalid scope")
	}
	if appErr, ok := AsError(err); !ok || appErr.Code != ErrorInvalidScope {
		t.Fatalf("expected invalid scope error, got %v", err)
	}
}

func TestService_IssueToken_InvalidCredentials(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.IssueToken(context.Background(), TokenIssueRequest{
		Username: "alice",
		Password: "wrong",
	})
	if err == nil {
		t.Fatalf("expected error for invalid credentials")
	}
	if appErr, ok := AsError(err); !ok || appErr.Code != ErrorInvalidCredentials {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestService_IssueToken_DefaultScopes(t *testing.T) {
	svc := newTestService(t)
	token, err := svc.IssueToken(context.Background(), TokenIssueRequest{
		Username: "alice",
		Password: "password",
		Scopes:   nil,
	})
	if err != nil {
		t.Fatalf("issue token failed: %v", err)
	}
	if len(token.Scopes) != 1 || token.Scopes[0] != "viewer" {
		t.Fatalf("expected default viewer scope, got %v", token.Scopes)
	}
}

func TestService_ValidateToken_Expired(t *testing.T) {
	svc := newTestService(t)
	base := time.Unix(1_000, 0)
	current := base
	svc.now = func() time.Time {
		return current
	}

	token, err := svc.IssueToken(context.Background(), TokenIssueRequest{
		Username: "alice",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("issue token failed: %v", err)
	}

	current = current.Add(svc.cfg.AccessTokenTTL + time.Second)
	_, err = svc.ValidateToken(context.Background(), token.Value)
	if err == nil {
		t.Fatalf("expected error for expired token")
	}
	appErr, ok := AsError(err)
	if !ok || appErr.Code != ErrorInvalidToken {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func TestService_ValidateToken_Invalid(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.ValidateToken(context.Background(), "invalid-token")
	if err == nil {
		t.Fatalf("expected error for invalid token")
	}
	if appErr, ok := AsError(err); !ok || appErr.Code != ErrorInvalidToken {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func TestService_ValidateToken_InvalidAudience(t *testing.T) {
	svc := newTestService(t)
	token, err := svc.IssueToken(context.Background(), TokenIssueRequest{
		Username: "alice",
		Password: "password",
	})
	if err != nil {
		t.Fatalf("issue token failed: %v", err)
	}

	svc.cfg.Audience = []string{"different"}
	_, err = svc.ValidateToken(context.Background(), token.Value)
	if err == nil {
		t.Fatalf("expected error for invalid audience")
	}
	if appErr, ok := AsError(err); !ok || appErr.Code != ErrorInvalidToken {
		t.Fatalf("expected invalid token error, got %v", err)
	}
}

func newTestService(t *testing.T) *service {
	t.Helper()

	authenticator, err := NewStaticAuthenticator(map[string]struct {
		Password  string
		Principal Principal
	}{
		"alice": {
			Password:  "password",
			Principal: Principal{Subject: "user-alice", Scopes: []string{"viewer"}},
		},
	})
	if err != nil {
		t.Fatalf("failed to create authenticator: %v", err)
	}

	keyManager, err := NewHMACKeyManager([]byte("super-secret"), "test-key")
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	svcIface, err := NewService(Config{
		Issuer:         "test-issuer",
		Audience:       []string{"test-audience"},
		AccessTokenTTL: time.Minute,
	}, authenticator, keyManager)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	svc := svcIface.(*service)
	svc.now = func() time.Time { return time.Unix(1_000, 0) }

	return svc
}
