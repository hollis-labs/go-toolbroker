package broker

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"sync"
)

// LocalBroker is a rule-based tool broker that runs in-process.
// It applies configurable rules to select tools based on intent and hints.
// This is the primary implementation for embedding in applications like mentat-chat.
type LocalBroker struct {
	mu    sync.RWMutex
	tools []ToolDefinition
	rules []Rule
}

// Verify LocalBroker implements Broker.
var _ Broker = (*LocalBroker)(nil)

// NewLocalBroker creates a broker with the given tools and rules.
func NewLocalBroker(tools []ToolDefinition, rules []Rule) *LocalBroker {
	return &LocalBroker{
		tools: append([]ToolDefinition(nil), tools...),
		rules: append([]Rule(nil), rules...),
	}
}

// RegisterTools adds tools to the broker's registry.
func (b *LocalBroker) RegisterTools(tools []ToolDefinition) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tools = append(b.tools, tools...)
}

// LoadRules replaces the current rule set.
func (b *LocalBroker) LoadRules(rules []Rule) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.rules = append([]Rule(nil), rules...)
}

// AllTools returns lightweight summaries of every registered tool.
func (b *LocalBroker) AllTools() []ToolSummary {
	b.mu.RLock()
	defer b.mu.RUnlock()
	summaries := make([]ToolSummary, len(b.tools))
	for i, t := range b.tools {
		summaries[i] = ToolSummary{
			Name:        t.Name,
			Description: t.Description,
			Server:      t.Server,
			Tags:        t.Tags,
		}
	}
	return summaries
}

// SelectTools implements Broker. It applies rules in priority order to select
// tools relevant to the given intent.
//
// Algorithm:
//  1. Sort rules by priority (descending — highest first).
//  2. Separate rules into applicable ones (intent matches) and skip the rest.
//  3. If no rules are applicable, return all tools (safe default).
//  4. Start with all tools. For each applicable rule:
//     - "exclude": remove matching tools from the set.
//     - "include": mark matching tools as explicitly included.
//     - "summarize": keep matching tools but could be used for progressive disclosure.
//  5. If any "include" rules fired, only keep tools that were explicitly included
//     (plus any that no include rule's patterns covered — include acts as a whitelist
//     for its matched scope, not a global filter).
//  6. Build result with rationale.
func (b *LocalBroker) SelectTools(_ context.Context, intent string, hints []string) (*SelectResult, error) {
	b.mu.RLock()
	tools := append([]ToolDefinition(nil), b.tools...)
	rules := append([]Rule(nil), b.rules...)
	b.mu.RUnlock()

	total := len(tools)
	if total == 0 {
		return &SelectResult{
			Intent:    intent,
			Rationale: "no tools registered",
		}, nil
	}

	// Sort rules by priority descending.
	sorted := make([]Rule, len(rules))
	copy(sorted, rules)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})

	// Find applicable rules for this intent.
	var applicable []Rule
	for _, r := range sorted {
		if intentMatches(r.Intent, intent) {
			applicable = append(applicable, r)
		}
	}

	// No applicable rules — return everything.
	if len(applicable) == 0 {
		result := &SelectResult{
			Tools:     tools,
			Count:     total,
			Total:     total,
			Intent:    intent,
			Rationale: "no rules matched intent; returning all tools",
		}
		return result, nil
	}

	// Track which tools to keep/remove.
	excluded := make(map[int]bool)   // index -> excluded
	included := make(map[int]bool)   // index -> explicitly included
	hasIncludeRule := false
	var appliedRules []string

	for _, r := range applicable {
		switch r.Action.Type {
		case "exclude":
			for i, t := range tools {
				if !excluded[i] && toolMatchesRule(t, r.Match, hints) {
					excluded[i] = true
				}
			}
			appliedRules = append(appliedRules, fmt.Sprintf("exclude(%s)", r.Name))

		case "include":
			hasIncludeRule = true
			for i, t := range tools {
				if toolMatchesRule(t, r.Match, hints) {
					included[i] = true
				}
			}
			appliedRules = append(appliedRules, fmt.Sprintf("include(%s)", r.Name))

		case "summarize":
			// For now, summarize acts like include — keeps tools in the set.
			// Future: mark these tools for summary-only progressive disclosure.
			for i, t := range tools {
				if toolMatchesRule(t, r.Match, hints) {
					included[i] = true
				}
			}
			appliedRules = append(appliedRules, fmt.Sprintf("summarize(%s)", r.Name))
		}
	}

	// Build result set.
	var selected []ToolDefinition
	for i, t := range tools {
		if excluded[i] {
			continue
		}
		// If include rules were used, only keep explicitly included tools.
		if hasIncludeRule && !included[i] {
			continue
		}
		selected = append(selected, t)
	}

	rationale := fmt.Sprintf("applied rules: %s", strings.Join(appliedRules, ", "))

	return &SelectResult{
		Tools:     selected,
		Count:     len(selected),
		Total:     total,
		Intent:    intent,
		Rationale: rationale,
	}, nil
}

// intentMatches checks if a rule's intent pattern matches the given intent.
// Supports "*" as a wildcard for all intents, and empty intent matches everything.
func intentMatches(pattern, intent string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	// Support glob matching on intent strings.
	matched, err := path.Match(pattern, intent)
	if err != nil {
		// Bad pattern — treat as no match.
		return false
	}
	return matched
}

// toolMatchesRule checks if a tool matches the rule's match criteria.
// Match fields use OR within each field, AND across fields.
// If a field is empty, it is not considered (matches everything).
func toolMatchesRule(tool ToolDefinition, m Match, hints []string) bool {
	// Check patterns (glob on tool name).
	if len(m.Patterns) > 0 {
		nameMatched := false
		for _, pattern := range m.Patterns {
			// Support both glob and prefix matching.
			if matched, err := path.Match(pattern, tool.Name); err == nil && matched {
				nameMatched = true
				break
			}
			// Also support prefix patterns like "hadron_bp_*" matching "hadron_bp_build_volon"
			// path.Match requires exact segment matching, so also try HasPrefix
			// for patterns ending in "*".
			if strings.HasSuffix(pattern, "*") {
				prefix := strings.TrimSuffix(pattern, "*")
				if strings.HasPrefix(tool.Name, prefix) {
					nameMatched = true
					break
				}
			}
		}
		if !nameMatched {
			return false
		}
	}

	// Check tags.
	if len(m.Tags) > 0 {
		tagMatched := false
		for _, mt := range m.Tags {
			for _, tt := range tool.Tags {
				if mt == tt {
					tagMatched = true
					break
				}
			}
			if tagMatched {
				break
			}
		}
		if !tagMatched {
			return false
		}
	}

	// Check servers.
	if len(m.Servers) > 0 {
		serverMatched := false
		for _, s := range m.Servers {
			if s == tool.Server {
				serverMatched = true
				break
			}
		}
		if !serverMatched {
			return false
		}
	}

	return true
}
