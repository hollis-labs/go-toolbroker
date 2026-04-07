// Package broker provides intent-aware MCP tool selection.
//
// The broker selects relevant tools from a registry based on intent,
// hints, and configurable rules. This replaces hardcoded exclude-pattern
// filtering with a flexible, rule-driven approach.
//
// Two implementations are planned:
//   - LocalBroker: in-process, config-driven (this package)
//   - Remote broker service: HTTP-based, cross-app (future)
package broker

import "context"

// Broker selects relevant tools based on intent and context.
type Broker interface {
	// SelectTools returns tools relevant to the given intent.
	// hints are optional keywords, project names, or capability tags
	// that refine the selection beyond what intent alone provides.
	SelectTools(ctx context.Context, intent string, hints []string) (*SelectResult, error)

	// AllTools returns lightweight summaries of every registered tool.
	AllTools() []ToolSummary
}

// ToolDefinition is the full tool schema (name, description, inputSchema).
// This mirrors the MCP tool schema and costs ~150 tokens when serialized.
type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema,omitempty"`
	Server      string         `json:"server,omitempty"`   // which MCP server provides this
	Tags        []string       `json:"tags,omitempty"`     // capability tags
	CostTier    string         `json:"cost_tier,omitempty"` // "free", "low", "high"
}

// ToolSummary is a lightweight view of a tool (~30 tokens vs ~150 for full definition).
// Used for progressive disclosure: the LLM sees summaries first and can
// request full schemas on demand.
type ToolSummary struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Server      string   `json:"server,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// SelectResult contains the selected tools and metadata about the selection.
type SelectResult struct {
	Tools     []ToolDefinition `json:"tools"`
	Count     int              `json:"count"`               // number of tools selected
	Total     int              `json:"total"`               // total tools available
	Intent    string           `json:"intent"`
	Rationale string           `json:"rationale,omitempty"` // explains which rules were applied
}
