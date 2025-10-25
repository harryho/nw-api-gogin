package auth

import "testing"

func TestStaticAuthenticator_Success(t *testing.T) {
	authenticator, err := NewStaticAuthenticator(map[string]struct {
		Password  string
		Principal Principal
	}{
		"admin": {
			Password:  "secret",
			Principal: Principal{Subject: "admin", Scopes: []string{"admin", "viewer"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	principal, err := authenticator.Authenticate(nil, "admin", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if principal.Subject != "admin" {
		t.Fatalf("expected subject admin, got %q", principal.Subject)
	}
	if len(principal.Scopes) != 2 {
		t.Fatalf("expected 2 scopes, got %d", len(principal.Scopes))
	}
}

func TestStaticAuthenticator_InvalidCredentials(t *testing.T) {
	authenticator, err := NewStaticAuthenticator(map[string]struct {
		Password  string
		Principal Principal
	}{
		"admin": {
			Password:  "secret",
			Principal: Principal{Subject: "admin", Scopes: []string{"admin", "viewer"}},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = authenticator.Authenticate(nil, "admin", "wrong")
	if err == nil {
		t.Fatalf("expected error for wrong password")
	}
	if appErr, ok := AsError(err); !ok || appErr.Code != ErrorInvalidCredentials {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}

	_, err = authenticator.Authenticate(nil, "unknown", "secret")
	if err == nil {
		t.Fatalf("expected error for unknown user")
	}
	if appErr, ok := AsError(err); !ok || appErr.Code != ErrorInvalidCredentials {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestStaticAuthenticator_EmptyUsers(t *testing.T) {
	if _, err := NewStaticAuthenticator(map[string]struct {
		Password  string
		Principal Principal
	}{}); err == nil {
		t.Fatalf("expected error for empty users")
	}
}
