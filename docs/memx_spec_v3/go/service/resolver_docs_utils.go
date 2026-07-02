package service

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
	"unicode"
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

// intersectsFold checks whether two string slices share at least one value.
func intersectsFold(left []string, right []string) bool {
	for _, value := range left {
		if containsFold(right, value) {
			return true
		}
	}
	return false
}

// textContainsFold checks if text contains query (case-insensitive).
func textContainsFold(text string, query string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(strings.TrimSpace(query)))
}

func normalizeFTSQuery(query string) string {
	fields := strings.FieldsFunc(query, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && r != '_'
	})
	return strings.Join(uniqueTrimmedStrings(fields), " ")
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

func normalizeBirdseyeRefs(values []string) []string {
	refs := uniqueTrimmedStrings(values)
	for idx, ref := range refs {
		if strings.Contains(ref, ":") {
			continue
		}
		refs[idx] = typedRef("birdseye", "view", ref)
	}
	return refs
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

// decodeJSON decodes JSON into a destination and leaves zero value on invalid input.
func decodeJSON(raw string, dst interface{}) {
	if strings.TrimSpace(raw) == "" {
		return
	}
	_ = json.Unmarshal([]byte(raw), dst)
}

// nonEmptySlice returns a single-item slice for a non-empty string.
func nonEmptySlice(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return []string{value}
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

func inferVersionScheme(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "string"
	}
	if _, ok := parseSemVer(version); ok {
		return "semver"
	}
	if _, ok := parseVersionTime(version); ok {
		return "iso_datetime"
	}
	if isHexRevision(version) {
		return "git_revision"
	}
	return "string"
}

func normalizeVersionScheme(scheme string) string {
	switch strings.ToLower(strings.TrimSpace(scheme)) {
	case "semver":
		return "semver"
	case "iso", "iso_date", "iso_datetime", "datetime":
		return "iso_datetime"
	case "git", "git_revision", "revision":
		return "git_revision"
	default:
		return "string"
	}
}

func compareVersionsWithScheme(left string, right string, scheme string) int {
	scheme = normalizeVersionScheme(scheme)
	switch scheme {
	case "semver":
		if l, ok := parseSemVer(left); ok {
			if r, ok := parseSemVer(right); ok {
				for i := range l {
					if l[i] > r[i] {
						return 1
					}
					if l[i] < r[i] {
						return -1
					}
				}
				return 0
			}
		}
	case "iso_datetime":
		if l, ok := parseVersionTime(left); ok {
			if r, ok := parseVersionTime(right); ok {
				switch {
				case l.After(r):
					return 1
				case l.Before(r):
					return -1
				default:
					return 0
				}
			}
		}
	case "git_revision":
		if strings.EqualFold(strings.TrimSpace(left), strings.TrimSpace(right)) {
			return 0
		}
		return 1
	}
	return compareVersions(left, right)
}

func versionsEquivalent(left string, right string, scheme string) bool {
	return compareVersionsWithScheme(left, right, scheme) == 0
}

func parseSemVer(version string) ([3]int, bool) {
	var out [3]int
	version = strings.TrimPrefix(strings.TrimSpace(strings.ToLower(version)), "v")
	core := strings.SplitN(version, "-", 2)[0]
	core = strings.SplitN(core, "+", 2)[0]
	parts := strings.Split(core, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return out, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

func parseVersionTime(version string) (time.Time, bool) {
	version = strings.TrimSpace(version)
	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		t, err := time.Parse(layout, version)
		if err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func isHexRevision(version string) bool {
	version = strings.TrimSpace(version)
	if len(version) < 7 || len(version) > 40 {
		return false
	}
	for _, r := range version {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
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
