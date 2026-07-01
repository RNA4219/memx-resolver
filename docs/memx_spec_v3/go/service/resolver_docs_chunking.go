package service

import (
	"fmt"
	"strings"
)

// buildResolverChunks creates chunks from sections.
func buildResolverChunks(docID string, sections []resolverSection, opts ChunkingOptions) []ResolverChunk {
	chunks := make([]ResolverChunk, 0, len(sections))
	ordinal := 1
	for _, section := range sections {
		for _, part := range splitChunkBody(section.Body, opts.MaxChars) {
			body := strings.TrimSpace(part)
			if body == "" {
				continue
			}
			chunks = append(chunks, ResolverChunk{
				ChunkID:       fmt.Sprintf("chunk:%s:%03d", docID, ordinal),
				DocID:         docID,
				Heading:       section.Heading,
				HeadingPath:   append([]string(nil), section.HeadingPath...),
				Ordinal:       ordinal,
				Body:          body,
				TokenEstimate: estimateTokens(body),
				Importance:    section.Importance,
				MemoryType:    inferMemoryType(section.Heading, section.HeadingPath),
				Cue:           buildChunkCue(section.HeadingPath),
			})
			ordinal++
		}
	}
	return chunks
}

// parseSectionsForChunking parses document body into sections for chunking.
func parseSectionsForChunking(title string, body string, docType string, mode string) []resolverSection {
	if mode == "fixed" {
		return []resolverSection{
			{
				Heading:     strings.TrimSpace(title),
				HeadingPath: []string{strings.TrimSpace(title)},
				Body:        strings.TrimSpace(body),
				Importance:  defaultDocImportance(docType),
			},
		}
	}
	return parseResolverSections(title, body, docType)
}

// parseResolverSections parses markdown body into sections based on headings.
func parseResolverSections(title string, body string, docType string) []resolverSection {
	defaultImportance := defaultDocImportance(docType)
	lines := strings.Split(body, "\n")
	path := []string{strings.TrimSpace(title)}
	currentHeading := strings.TrimSpace(title)
	var currentBody strings.Builder
	sections := make([]resolverSection, 0)

	flush := func() {
		text := strings.TrimSpace(currentBody.String())
		if text == "" {
			return
		}
		sections = append(sections, resolverSection{
			Heading:     currentHeading,
			HeadingPath: append([]string(nil), path...),
			Body:        text,
			Importance:  inferChunkImportance(defaultImportance, currentHeading, path),
		})
		currentBody.Reset()
	}

	for _, line := range lines {
		if matches := markdownHeadingPattern.FindStringSubmatch(line); matches != nil {
			flush()
			level := len(matches[1])
			heading := strings.TrimSpace(matches[2])
			if level <= 1 {
				path = []string{heading}
			} else {
				base := []string{title}
				if len(path) > 0 {
					base = append([]string(nil), path...)
				}
				if len(base) >= level {
					base = append([]string(nil), base[:level-1]...)
				}
				base = append(base, heading)
				path = base
			}
			currentHeading = heading
			continue
		}
		currentBody.WriteString(line)
		currentBody.WriteByte('\n')
	}
	flush()

	if len(sections) == 0 {
		sections = append(sections, resolverSection{
			Heading:     title,
			HeadingPath: []string{title},
			Body:        strings.TrimSpace(body),
			Importance:  defaultImportance,
		})
	}
	return sections
}

// splitChunkBody splits body into chunks respecting maxChars.
func splitChunkBody(body string, maxChars int) []string {
	text := strings.TrimSpace(body)
	if text == "" {
		return nil
	}
	if maxChars <= 0 || len([]rune(text)) <= maxChars {
		return []string{text}
	}

	paragraphs := strings.Split(text, "\n\n")
	parts := make([]string, 0)
	var current strings.Builder

	flush := func() {
		chunk := strings.TrimSpace(current.String())
		if chunk != "" {
			parts = append(parts, chunk)
		}
		current.Reset()
	}

	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		candidate := paragraph
		if current.Len() > 0 {
			candidate = current.String() + "\n\n" + paragraph
		}

		if len([]rune(candidate)) <= maxChars {
			if current.Len() > 0 {
				current.WriteString("\n\n")
			}
			current.WriteString(paragraph)
			continue
		}

		if current.Len() > 0 {
			flush()
		}
		for _, hard := range splitRunes(paragraph, maxChars) {
			parts = append(parts, hard)
		}
	}
	flush()
	return parts
}

// splitRunes splits text by rune count.
func splitRunes(text string, maxChars int) []string {
	runes := []rune(text)
	parts := make([]string, 0)
	for start := 0; start < len(runes); start += maxChars {
		end := start + maxChars
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, strings.TrimSpace(string(runes[start:end])))
	}
	return parts
}

// estimateTokens estimates token count from text.
func estimateTokens(text string) int {
	return (len([]rune(text)) + 3) / 4
}

// inferChunkImportance determines chunk importance from heading and path.
func inferChunkImportance(defaultImportance string, heading string, path []string) string {
	joined := strings.ToLower(heading + " " + strings.Join(path, " "))
	switch {
	case strings.Contains(joined, "acceptance"), strings.Contains(joined, "forbidden"), strings.Contains(joined, "definition of done"):
		return "required"
	case strings.Contains(joined, "overview"), strings.Contains(joined, "background"):
		if defaultImportance == "required" {
			return "recommended"
		}
		return defaultImportance
	default:
		return defaultImportance
	}
}

// inferMemoryType assigns a prompt-facing role to a chunk.
func inferMemoryType(heading string, path []string) string {
	joined := strings.ToLower(heading + " " + strings.Join(path, " "))
	switch {
	case strings.Contains(joined, "acceptance"):
		return "acceptance"
	case strings.Contains(joined, "forbidden"), strings.Contains(joined, "constraint"), strings.Contains(joined, "guardrail"):
		return "constraint"
	case strings.Contains(joined, "definition of done"), strings.Contains(joined, "done"):
		return "done"
	case strings.Contains(joined, "dependency"), strings.Contains(joined, "dependencies"):
		return "dependency"
	case strings.Contains(joined, "runbook"), strings.Contains(joined, "execute"), strings.Contains(joined, "procedure"), strings.Contains(joined, "workflow"):
		return "procedure"
	case strings.Contains(joined, "decision"), strings.Contains(joined, "adr"):
		return "decision"
	case strings.Contains(joined, "risk"), strings.Contains(joined, "incident"):
		return "risk"
	case strings.Contains(joined, "overview"), strings.Contains(joined, "background"), strings.Contains(joined, "purpose"):
		return "concept"
	default:
		return "reference"
	}
}

// buildChunkCue creates a compact retrieval cue for LLM prompts.
func buildChunkCue(path []string) string {
	parts := uniqueTrimmedStrings(path)
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " > ")
}

// defaultDocImportance returns default importance for document type.
func defaultDocImportance(docType string) string {
	switch strings.ToLower(docType) {
	case "spec", "requirement", "requirements", "interfaces", "design":
		return "required"
	case "cookbook", "runbook", "birdseye", "task":
		return "recommended"
	default:
		return "reference"
	}
}
