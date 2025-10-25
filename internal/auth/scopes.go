package auth

import (
	"sort"
	"strings"
)

func sanitizeScopes(scopes []string) []string {
	if len(scopes) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(scopes))
	result := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		s := normalizeScope(scope)
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		result = append(result, s)
	}
	if len(result) == 0 {
		return nil
	}
	sort.Strings(result)
	return result
}

func normalizeScope(scope string) string {
	return strings.TrimSpace(scope)
}

func parseScopeString(scope string) []string {
	return sanitizeScopes(strings.Fields(scope))
}

func hasAllScopes(requested, allowed []string) bool {
	if len(requested) == 0 {
		return true
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, scope := range allowed {
		allowedSet[scope] = struct{}{}
	}
	for _, scope := range requested {
		if _, ok := allowedSet[scope]; !ok {
			return false
		}
	}
	return true
}

func ParseScopeString(scope string) []string {
	return parseScopeString(scope)
}

func SanitizeScopes(scopes []string) []string {
	return sanitizeScopes(scopes)
}

func HasAllScopes(requested, allowed []string) bool {
	return hasAllScopes(requested, allowed)
}
