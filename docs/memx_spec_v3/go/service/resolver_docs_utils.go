package service

import (
	"encoding/json"
	"strings"
)

// containsFold checks if values contain target (case-insensitive).
func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), strings.TrimSpace(target)) {
			return true
		}
	}
	return false
}

// textContainsFold checks if text contains query (case-insensitive).
func textContainsFold(text string, query string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(strings.TrimSpace(query)))
}

// uniqueTrimmedStrings returns unique trimmed strings.
func uniqueTrimmedStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

// mustJSON marshals value to JSON string.
func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// decodeStringSlice decodes JSON string slice.
func decodeStringSlice(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil
	}
	return out
}

// compareVersions compares two version strings.
func compareVersions(left string, right string) int {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)
	switch {
	case left == right:
		return 0
	case right == "":
		return 1
	case left == "":
		return -1
	case left > right:
		return 1
	default:
		return -1
	}
}

// firstNonEmpty returns first non-empty string.
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}