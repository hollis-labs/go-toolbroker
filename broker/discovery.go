package broker

import (
	"sort"
	"strings"
)

// MinKeywordScore is the minimum keyword-overlap score a tool must reach
// to be included in ScoreByKeywords results.
const MinKeywordScore = 1

// ScoreByKeywords scores every tool in the slice against the given keywords
// using substring matching on tool name (+2) and description (+1).
// Returns the top maxResults tools sorted by relevance score descending.
// If no tool scores above MinKeywordScore, returns nil.
func ScoreByKeywords(tools []ToolDefinition, keywords []string, maxResults int) []ToolDefinition {
	if len(tools) == 0 || len(keywords) == 0 {
		return nil
	}

	words := normalizeKeywords(keywords)
	if len(words) == 0 {
		return nil
	}

	type scored struct {
		tool  ToolDefinition
		score int
	}

	var candidates []scored
	for _, t := range tools {
		s := scoreToolKeywords(t, words)
		if s >= MinKeywordScore {
			candidates = append(candidates, scored{tool: t, score: s})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if maxResults > 0 && len(candidates) > maxResults {
		candidates = candidates[:maxResults]
	}

	result := make([]ToolDefinition, len(candidates))
	for i, c := range candidates {
		result[i] = c.tool
	}
	return result
}

// ScoreByIntent is a convenience wrapper that tokenizes an intent string
// into keywords and calls ScoreByKeywords. This is the primary entry point
// for consumers doing progressive discovery by intent description.
func ScoreByIntent(tools []ToolDefinition, intent string, maxResults int) []ToolDefinition {
	words := TokenizeIntent(intent)
	return ScoreByKeywords(tools, words, maxResults)
}

// FindByNames returns tools whose Name matches any of the given names (exact match).
// Order follows the input names slice. Unmatched names are silently skipped.
func FindByNames(tools []ToolDefinition, names []string) []ToolDefinition {
	if len(tools) == 0 || len(names) == 0 {
		return nil
	}

	index := make(map[string]ToolDefinition, len(tools))
	for _, t := range tools {
		index[t.Name] = t
	}

	var result []ToolDefinition
	for _, name := range names {
		if t, ok := index[name]; ok {
			result = append(result, t)
		}
	}
	return result
}

// scoreToolKeywords counts how many keywords appear as substrings
// in the tool's name or description (case-insensitive).
func scoreToolKeywords(t ToolDefinition, words []string) int {
	nameLower := strings.ToLower(t.Name)
	descLower := strings.ToLower(t.Description)

	score := 0
	for _, w := range words {
		if strings.Contains(nameLower, w) {
			score += 2 // name match weighted higher
		}
		if strings.Contains(descLower, w) {
			score++
		}
	}
	return score
}

// normalizeKeywords lowercases and deduplicates keywords, filtering empties.
func normalizeKeywords(keywords []string) []string {
	seen := make(map[string]bool, len(keywords))
	var result []string
	for _, kw := range keywords {
		w := strings.ToLower(strings.TrimSpace(kw))
		if w == "" || seen[w] {
			continue
		}
		seen[w] = true
		result = append(result, w)
	}
	return result
}

// TokenizeIntent splits an intent string into lowercase words of 3+ characters,
// filtering common stop words. Exported for consumers that want to pre-tokenize.
func TokenizeIntent(intent string) []string {
	raw := strings.ToLower(intent)
	for _, ch := range []string{",", ".", "!", "?", ";", ":", "'", "\"", "(", ")", "[", "]", "{", "}"} {
		raw = strings.ReplaceAll(raw, ch, " ")
	}

	var words []string
	seen := make(map[string]bool)
	for _, w := range strings.Fields(raw) {
		if len(w) < 3 {
			continue
		}
		if intentStopWords[w] {
			continue
		}
		if seen[w] {
			continue
		}
		seen[w] = true
		words = append(words, w)
	}
	return words
}

var intentStopWords = map[string]bool{
	"the": true, "and": true, "for": true, "are": true, "but": true,
	"not": true, "you": true, "all": true, "can": true, "her": true,
	"was": true, "one": true, "our": true, "out": true, "has": true,
	"its": true, "let": true, "get": true, "make": true, "like": true,
	"just": true, "want": true, "need": true, "from": true, "with": true,
	"this": true, "that": true, "have": true, "will": true, "what": true,
	"when": true, "how": true, "who": true, "which": true, "where": true,
	"why": true, "been": true, "being": true, "would": true, "could": true,
	"should": true, "about": true, "into": true, "some": true, "any": true,
}
