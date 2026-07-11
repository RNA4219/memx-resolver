package service

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

// extractSectionItems extracts list items from sections matching keywords.
func extractSectionItems(sections []resolverSection, keywords []string) []string {
	var out []string
	for _, section := range sections {
		// Match only the heading name itself, not the full path
		headingLower := strings.ToLower(section.Heading)
		matched := false
		for _, keyword := range keywords {
			if strings.Contains(headingLower, keyword) {
				matched = true
				break
			}
		}
		if matched {
			out = append(out, parseListItems(section.Body)...)
		}
	}
	return uniqueTrimmedStrings(out)
}

// parseListItems parses markdown list items from body.
func parseListItems(body string) []string {
	var out []string
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "- "):
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		case strings.HasPrefix(trimmed, "* "):
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "* "))
		default:
			trimmed = trimNumericListPrefix(trimmed)
		}
		if trimmed != "" && trimmed != strings.TrimSpace(line) {
			out = append(out, trimmed)
		}
	}
	if len(out) > 0 {
		return uniqueTrimmedStrings(out)
	}
	text := strings.TrimSpace(body)
	if text == "" {
		return nil
	}
	return []string{text}
}

// trimNumericListPrefix removes numeric list prefix like "1. ".
func trimNumericListPrefix(s string) string {
	idx := strings.Index(s, ". ")
	if idx <= 0 {
		return s
	}
	for _, r := range s[:idx] {
		if r < '0' || r > '9' {
			return s
		}
	}
	return strings.TrimSpace(s[idx+2:])
}

// slugify converts string to URL-safe slug.
func slugify(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || unicode.IsSpace(r):
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// buildDocID creates a document ID from type and title.
func buildDocID(docType string, title string) string {
	slug := slugify(title)
	if slug == "" {
		slug = fmt.Sprintf("doc-%d", time.Now().Unix())
	}
	return fmt.Sprintf("doc:%s:%s", docType, slug)
}

// buildDocumentSummary creates a summary from body.
func buildDocumentSummary(body string) string {
	trimmed := strings.TrimSpace(body)
	if trimmed == "" {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) <= 120 {
		return trimmed
	}
	return strings.TrimSpace(string(runes[:120]))
}
