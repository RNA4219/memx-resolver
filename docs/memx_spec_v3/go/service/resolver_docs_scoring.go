package service

import (
	"sort"
	"strings"
)

// scoreResolverDocuments scores documents based on feature, taskID, and topic matching.
func scoreResolverDocuments(docs []ResolverDocument, feature string, taskID string, topic string) []scoredResolverDoc {
	scored := make([]scoredResolverDoc, 0)
	for _, doc := range docs {
		score := 0
		reason := ""

		if taskID != "" && containsFold(doc.TaskIDs, taskID) {
			score += 1000
			reason = "task dependency matched"
		}
		if feature != "" && containsFold(doc.FeatureKeys, feature) {
			score += 900
			if reason == "" {
				reason = "feature key matched"
			}
		}

		query := firstNonEmpty(topic, feature)
		if query != "" {
			if containsFold(doc.Tags, query) {
				score += 300
				if reason == "" {
					reason = "tag matched"
				}
			}
			if textContainsFold(doc.Title, query) ||
				textContainsFold(doc.Summary, query) ||
				textContainsFold(doc.Body, query) ||
				textContainsFold(doc.SourcePath, query) {
				score += 200
				if reason == "" {
					reason = "topic matched"
				}
			}
		}

		if score == 0 {
			continue
		}
		if reason == "" {
			reason = "relevance matched"
		}
		scored = append(scored, scoredResolverDoc{Doc: doc, Score: score, Why: reason})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].Score == scored[j].Score {
			return scored[i].Doc.UpdatedAt > scored[j].Doc.UpdatedAt
		}
		return scored[i].Score > scored[j].Score
	})
	return scored
}

func scoreResolverDocumentsWithBoost(docs []ResolverDocument, feature string, taskID string, topic string, boosts map[string]int) []scoredResolverDoc {
	scored := scoreResolverDocuments(docs, feature, taskID, topic)
	byDocID := make(map[string]int, len(scored))
	for idx := range scored {
		if boost := boosts[scored[idx].Doc.DocID]; boost > 0 {
			scored[idx].Score += boost
			if scored[idx].Why == "" || scored[idx].Why == "topic matched" {
				scored[idx].Why = "fts matched"
			}
		}
		byDocID[scored[idx].Doc.DocID] = idx
	}
	for _, doc := range docs {
		boost := boosts[doc.DocID]
		if boost <= 0 {
			continue
		}
		if _, ok := byDocID[doc.DocID]; ok {
			continue
		}
		scored = append(scored, scoredResolverDoc{Doc: doc, Score: boost, Why: "fts matched"})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		if scored[i].Score == scored[j].Score {
			return scored[i].Doc.UpdatedAt > scored[j].Doc.UpdatedAt
		}
		return scored[i].Score > scored[j].Score
	})
	return scored
}

// filterResolverChunks filters chunks by heading and query.
func filterResolverChunks(chunks []ResolverChunk, heading string, query string, limit int) []ResolverChunk {
	filtered := make([]ResolverChunk, 0, len(chunks))
	for _, chunk := range chunks {
		// When heading is specified, match only the chunk's heading name, not the full path
		if heading != "" && !textContainsFold(chunk.Heading, heading) {
			continue
		}
		if query != "" &&
			!textContainsFold(chunk.Body, query) &&
			!textContainsFold(chunk.Heading, query) &&
			!textContainsFold(strings.Join(chunk.HeadingPath, " > "), query) {
			continue
		}
		filtered = append(filtered, chunk)
	}

	if limit <= 0 || limit > len(filtered) {
		limit = len(filtered)
	}
	return filtered[:limit]
}

func filterResolverChunksByFTS(chunks []ResolverChunk, scores map[string]int) []ResolverChunk {
	if len(scores) == 0 {
		return nil
	}
	filtered := make([]ResolverChunk, 0, len(chunks))
	for _, chunk := range chunks {
		if scores[chunk.ChunkID] > 0 {
			filtered = append(filtered, chunk)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if scores[filtered[i].ChunkID] == scores[filtered[j].ChunkID] {
			return filtered[i].Ordinal < filtered[j].Ordinal
		}
		return scores[filtered[i].ChunkID] > scores[filtered[j].ChunkID]
	})
	return filtered
}

// filterResolverDocumentsForSearch applies structured search filters.
func filterResolverDocumentsForSearch(docs []ResolverDocument, req DocsSearchRequest) []ResolverDocument {
	if len(req.DocTypes) == 0 && len(req.Tags) == 0 && len(req.FeatureKeys) == 0 {
		return docs
	}
	filtered := make([]ResolverDocument, 0, len(docs))
	for _, doc := range docs {
		if len(req.DocTypes) > 0 && !containsFold(req.DocTypes, doc.DocType) {
			continue
		}
		if len(req.Tags) > 0 && !intersectsFold(doc.Tags, req.Tags) {
			continue
		}
		if len(req.FeatureKeys) > 0 && !intersectsFold(doc.FeatureKeys, req.FeatureKeys) {
			continue
		}
		filtered = append(filtered, doc)
	}
	return filtered
}

// pickTopChunkIDs picks top chunk IDs based on importance and query.
func pickTopChunkIDs(chunks []ResolverChunk, query string, limit int) []string {
	selected := filterResolverChunks(chunks, "", query, 0)
	if len(selected) == 0 {
		selected = chunks
	}

	sort.SliceStable(selected, func(i, j int) bool {
		if selected[i].Importance == selected[j].Importance {
			return selected[i].Ordinal < selected[j].Ordinal
		}
		return chunkImportanceRank(selected[i].Importance) < chunkImportanceRank(selected[j].Importance)
	})

	if limit <= 0 || limit > len(selected) {
		limit = len(selected)
	}
	ids := make([]string, 0, limit)
	for _, chunk := range selected[:limit] {
		ids = append(ids, chunk.ChunkID)
	}
	return ids
}

// chunkImportanceRank returns numeric rank for importance.
func chunkImportanceRank(v string) int {
	switch v {
	case "required":
		return 0
	case "recommended":
		return 1
	default:
		return 2
	}
}

func hydrateResolverChunkMemoryFields(chunk *ResolverChunk) {
	if chunk.MemoryType == "" {
		chunk.MemoryType = inferMemoryType(chunk.Heading, chunk.HeadingPath)
	}
	if chunk.Cue == "" {
		chunk.Cue = buildChunkCue(chunk.HeadingPath)
	}
}

func filterResolverMemoryCards(cards []ResolverMemoryCard, memoryTypes []string) []ResolverMemoryCard {
	if len(memoryTypes) == 0 {
		return cards
	}
	filtered := make([]ResolverMemoryCard, 0, len(cards))
	for _, card := range cards {
		if containsFold(memoryTypes, card.MemoryType) {
			filtered = append(filtered, card)
		}
	}
	return filtered
}
