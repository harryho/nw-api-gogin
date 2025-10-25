package auth

import "testing"

func TestErrorWrapping(t *testing.T) {
	wrapped := NewError(ErrorInvalidScope, "invalid scope", nil)
	if wrapped.Error() != "invalid scope" {
		t.Fatalf("expected message 'invalid scope', got %q", wrapped.Error())
	}

	err := NewError(ErrorInvalidCredentials, "invalid username or password", wrapped)
	if err.Unwrap() != wrapped {
		t.Fatalf("expected error to unwrap to wrapped error")
	}

	if appErr, ok := AsError(err); !ok || appErr.Code != ErrorInvalidCredentials {
		t.Fatalf("expected auth error with invalid credentials code")
	}
}
