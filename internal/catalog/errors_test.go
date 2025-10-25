package catalog

import (
	"errors"
	"net/http"
	"testing"
)

func TestErrorMethods(t *testing.T) {
	inner := errors.New("boom")
	wrapped := Wrap(inner, ErrorConflict, "conflict occurred", http.StatusConflict)

	if wrapped.Error() != "conflict occurred" {
		t.Fatalf("expected message 'conflict occurred', got %q", wrapped.Error())
	}
	if !errors.Is(wrapped, inner) {
		t.Fatalf("expected wrapped error to contain inner error")
	}
	if wrapped.Unwrap() != inner {
		t.Fatalf("expected unwrap to return inner error")
	}
	if wrapped.Status != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, wrapped.Status)
	}
}

func TestNewConflictAndInternalError(t *testing.T) {
	conflictErr := NewConflictError("duplicate", errors.New("constraint"))
	if conflictErr.Code != ErrorConflict {
		t.Fatalf("expected conflict code, got %s", conflictErr.Code)
	}
	if conflictErr.Status != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, conflictErr.Status)
	}

	internalErr := NewInternalError("failure", nil)
	if internalErr.Code != ErrorInternal {
		t.Fatalf("expected internal code, got %s", internalErr.Code)
	}
	if internalErr.Status != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, internalErr.Status)
	}
}

func TestAsErrorRecognizesWrapped(t *testing.T) {
	wrapped := Wrap(errors.New("root"), ErrorConflict, "oops", http.StatusBadRequest)
	got, ok := AsError(wrapped)
	if !ok {
		t.Fatalf("expected AsError to recognize wrapped error")
	}
	if got.Code != ErrorConflict {
		t.Fatalf("expected conflict code, got %s", got.Code)
	}
}
