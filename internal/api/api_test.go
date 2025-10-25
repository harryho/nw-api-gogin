package api

import "testing"

func TestHandlersRegistered(t *testing.T) {
	if _, err := GetSwagger(); err != nil {
		t.Fatalf("expected swagger to load, got error: %v", err)
	}
}
