package auth

import (
	"context"
	"testing"
)

func TestHMACKeyManager(t *testing.T) {
	manager, err := NewHMACKeyManager([]byte("secret"), "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	key, err := manager.Current(context.TODO())
	if err != nil {
		t.Fatalf("failed to get current key: %v", err)
	}
	if key.ID != "key1" {
		t.Fatalf("expected key1, got %q", key.ID)
	}

	retrieved, err := manager.Get(context.TODO(), "key1")
	if err != nil {
		t.Fatalf("failed to get key by id: %v", err)
	}
	if retrieved.ID != key.ID {
		t.Fatalf("expected retrieved key to match current key")
	}

	if _, err := manager.Get(context.TODO(), "unknown"); err == nil {
		t.Fatalf("expected error for unknown key id")
	}
}

func TestNewHMACKeyManager_EmptySecret(t *testing.T) {
	if _, err := NewHMACKeyManager(nil, ""); err == nil {
		t.Fatalf("expected error for empty secret")
	}
}
