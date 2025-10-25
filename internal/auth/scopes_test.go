package auth

import "testing"

func TestSanitizeScopes(t *testing.T) {
	result := SanitizeScopes([]string{" viewer ", "", "viewer", "admin"})
	expected := []string{"admin", "viewer"}

	if len(result) != len(expected) {
		t.Fatalf("expected %d scopes, got %d", len(expected), len(result))
	}
	for i, scope := range expected {
		if result[i] != scope {
			t.Fatalf("expected scope %q at index %d, got %q", scope, i, result[i])
		}
	}
}

func TestParseScopeString(t *testing.T) {
	result := ParseScopeString(" viewer manager viewer ")
	expected := []string{"manager", "viewer"}

	if len(result) != len(expected) {
		t.Fatalf("expected %d scopes, got %d", len(expected), len(result))
	}
	for i, scope := range expected {
		if result[i] != scope {
			t.Fatalf("expected scope %q at index %d, got %q", scope, i, result[i])
		}
	}
}

func TestHasAllScopes(t *testing.T) {
	allowed := []string{"viewer", "manager"}
	if !HasAllScopes([]string{"viewer"}, allowed) {
		t.Fatalf("expected viewer to be allowed")
	}
	if HasAllScopes([]string{"admin"}, allowed) {
		t.Fatalf("expected admin to be denied")
	}
}
