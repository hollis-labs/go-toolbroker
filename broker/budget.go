package broker

import "encoding/json"

// DefaultTokenBudgetPct is the default fraction of context window reserved for tool definitions.
const DefaultTokenBudgetPct = 0.20

// DefaultContextWindowTokens is a sensible default context window size.
const DefaultContextWindowTokens = 200000

// EstimateToolTokens estimates total tokens for a set of tool definitions
// by serializing each to JSON and dividing by 4 (approximation consistent
// with most tokenizers for structured text).
func EstimateToolTokens(tools []ToolDefinition) int {
	total := 0
	for _, t := range tools {
		data, err := json.Marshal(t)
		if err != nil {
			n := len(t.Name) + len(t.Description)
			if n == 0 {
				n = 4
			}
			total += n / 4
			continue
		}
		n := len(data) / 4
		if n == 0 {
			n = 1
		}
		total += n
	}
	return total
}

// PruneToTokenBudget removes tools from the end of the slice (lowest priority)
// until the total estimated tokens fits within the given budget.
// At least one tool is always retained. Returns a new slice.
func PruneToTokenBudget(tools []ToolDefinition, budgetTokens int) []ToolDefinition {
	if len(tools) == 0 {
		return tools
	}

	result := append([]ToolDefinition(nil), tools...)
	total := EstimateToolTokens(result)
	if total <= budgetTokens {
		return result
	}

	for len(result) > 1 && total > budgetTokens {
		last := result[len(result)-1]
		data, _ := json.Marshal(last)
		tokens := len(data) / 4
		if tokens == 0 {
			tokens = 1
		}
		total -= tokens
		result = result[:len(result)-1]
	}

	return result
}
