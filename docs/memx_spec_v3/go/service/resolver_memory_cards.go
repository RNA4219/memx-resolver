package service

import (
	"fmt"
	"sort"
	"strings"
)

const maxMemoryStatementRunes = 500

// BuildResolverMemoryCards converts chunks into prompt-ready memory cards.
func BuildResolverMemoryCards(chunks []ResolverChunk, limit int) []ResolverMemoryCard {
	return BuildRankedResolverMemoryCards(chunks, "", limit, 0)
}

// BuildRankedResolverMemoryCards converts chunks into ranked prompt-ready memory cards.
func BuildRankedResolverMemoryCards(chunks []ResolverChunk, query string, limit int, tokenBudget int) []ResolverMemoryCard {
	return BuildRankedResolverMemoryCardsWithWeights(chunks, query, limit, tokenBudget, DefaultMemoryCardRankingWeights(), nil)
}

// BuildRankedResolverMemoryCardsWithWeights converts chunks into ranked prompt-ready memory cards using configurable weights.
func BuildRankedResolverMemoryCardsWithWeights(chunks []ResolverChunk, query string, limit int, tokenBudget int, weights MemoryCardRankingWeights, feedback map[string]int) []ResolverMemoryCard {
	weights = normalizeMemoryCardRankingWeights(weights)
	cards := make([]ResolverMemoryCard, 0, len(chunks))
	for _, chunk := range chunks {
		hydrateResolverChunkMemoryFields(&chunk)
		statements := parseListItems(chunk.Body)
		if len(statements) == 0 {
			statements = []string{chunk.Body}
		}
		for idx, statement := range statements {
			statement = trimMemoryStatement(statement)
			if statement == "" {
				continue
			}
			cards = append(cards, ResolverMemoryCard{
				CardID:        fmt.Sprintf("card:%s:%03d", chunk.ChunkID, idx+1),
				DocID:         chunk.DocID,
				ChunkID:       chunk.ChunkID,
				MemoryType:    chunk.MemoryType,
				Cue:           chunk.Cue,
				Statement:     statement,
				HeadingPath:   append([]string(nil), chunk.HeadingPath...),
				Importance:    chunk.Importance,
				TokenEstimate: estimateTokens(statement),
			})
		}
	}
	for idx := range cards {
		cards[idx].Score = scoreMemoryCard(cards[idx], query, weights)
		if feedback != nil {
			cards[idx].Score += feedback[cards[idx].CardID] * weights.FeedbackBoost
			cards[idx].Score += feedback["type:"+cards[idx].MemoryType] * weights.FeedbackBoost
		}
	}
	sort.SliceStable(cards, func(i, j int) bool {
		if cards[i].Score == cards[j].Score {
			if cards[i].TokenEstimate == cards[j].TokenEstimate {
				return cards[i].CardID < cards[j].CardID
			}
			return cards[i].TokenEstimate < cards[j].TokenEstimate
		}
		return cards[i].Score > cards[j].Score
	})
	return applyMemoryCardLimits(cards, limit, tokenBudget)
}

// DefaultMemoryCardRankingWeights returns conservative defaults tuned for LLM prompts.
func DefaultMemoryCardRankingWeights() MemoryCardRankingWeights {
	return MemoryCardRankingWeights{
		ImportanceRequired:    100,
		ImportanceRecommended: 80,
		ImportanceReference:   60,
		MemoryTypeBase:        10,
		QueryExact:            120,
		QueryTerms:            80,
		CueMatch:              50,
		HeadingMatch:          40,
		ShortCardBonus:        10,
		FeedbackBoost:         8,
	}
}

func normalizeMemoryCardRankingWeights(weights MemoryCardRankingWeights) MemoryCardRankingWeights {
	defaults := DefaultMemoryCardRankingWeights()
	if weights.ImportanceRequired == 0 {
		weights.ImportanceRequired = defaults.ImportanceRequired
	}
	if weights.ImportanceRecommended == 0 {
		weights.ImportanceRecommended = defaults.ImportanceRecommended
	}
	if weights.ImportanceReference == 0 {
		weights.ImportanceReference = defaults.ImportanceReference
	}
	if weights.MemoryTypeBase == 0 {
		weights.MemoryTypeBase = defaults.MemoryTypeBase
	}
	if weights.QueryExact == 0 {
		weights.QueryExact = defaults.QueryExact
	}
	if weights.QueryTerms == 0 {
		weights.QueryTerms = defaults.QueryTerms
	}
	if weights.CueMatch == 0 {
		weights.CueMatch = defaults.CueMatch
	}
	if weights.HeadingMatch == 0 {
		weights.HeadingMatch = defaults.HeadingMatch
	}
	if weights.ShortCardBonus == 0 {
		weights.ShortCardBonus = defaults.ShortCardBonus
	}
	if weights.FeedbackBoost == 0 {
		weights.FeedbackBoost = defaults.FeedbackBoost
	}
	return weights
}

func trimMemoryStatement(statement string) string {
	statement = strings.TrimSpace(statement)
	if statement == "" {
		return ""
	}
	runes := []rune(statement)
	if len(runes) <= maxMemoryStatementRunes {
		return statement
	}
	return strings.TrimSpace(string(runes[:maxMemoryStatementRunes])) + "..."
}

func scoreMemoryCard(card ResolverMemoryCard, query string, weights MemoryCardRankingWeights) int {
	score := 0
	switch card.Importance {
	case "required":
		score += weights.ImportanceRequired
	case "recommended":
		score += weights.ImportanceRecommended
	default:
		score += weights.ImportanceReference
	}
	score += weights.MemoryTypeBase * (9 - memoryTypeRank(card.MemoryType))
	query = strings.TrimSpace(query)
	if query != "" {
		switch {
		case textContainsFold(card.Statement, query):
			score += weights.QueryExact
		case allQueryTermsMatch(card.Statement, query):
			score += weights.QueryTerms
		}
		if textContainsFold(card.Cue, query) {
			score += weights.CueMatch
		}
		if textContainsFold(strings.Join(card.HeadingPath, " "), query) {
			score += weights.HeadingMatch
		}
	}
	if card.TokenEstimate > 0 && card.TokenEstimate <= 80 {
		score += weights.ShortCardBonus
	}
	return score
}

func memoryTypeRank(memoryType string) int {
	switch memoryType {
	case "acceptance":
		return 0
	case "constraint":
		return 1
	case "done":
		return 2
	case "procedure":
		return 3
	case "dependency":
		return 4
	case "decision":
		return 5
	case "risk":
		return 6
	case "concept":
		return 7
	default:
		return 8
	}
}

func allQueryTermsMatch(text string, query string) bool {
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return false
	}
	for _, term := range terms {
		if !textContainsFold(text, term) {
			return false
		}
	}
	return true
}

func applyMemoryCardLimits(cards []ResolverMemoryCard, limit int, tokenBudget int) []ResolverMemoryCard {
	out := make([]ResolverMemoryCard, 0, len(cards))
	usedTokens := 0
	for _, card := range cards {
		if limit > 0 && len(out) >= limit {
			break
		}
		if tokenBudget > 0 {
			if card.TokenEstimate > tokenBudget {
				continue
			}
			if usedTokens+card.TokenEstimate > tokenBudget {
				continue
			}
			usedTokens += card.TokenEstimate
		}
		out = append(out, card)
	}
	return out
}
