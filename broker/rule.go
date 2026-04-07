package broker

// Rule defines a tool selection or exclusion rule.
// Rules are evaluated in priority order (highest priority first).
// When multiple rules match, their actions are applied cumulatively.
type Rule struct {
	Name     string `json:"name" yaml:"name"`
	Intent   string `json:"intent,omitempty" yaml:"intent,omitempty"`     // which intent triggers this rule ("*" = all)
	Match    Match  `json:"match" yaml:"match"`                           // what tools to match
	Action   Action `json:"action" yaml:"action"`                         // what to do with matches
	Priority int    `json:"priority,omitempty" yaml:"priority,omitempty"` // higher = applied first
}

// Match defines how to identify tools for a rule.
// All non-empty fields must match (AND logic).
// Within each field, any element matching is sufficient (OR logic).
type Match struct {
	Patterns []string `json:"patterns,omitempty" yaml:"patterns,omitempty"` // glob patterns on tool names (path.Match syntax)
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`         // match tools having any of these tags
	Servers  []string `json:"servers,omitempty" yaml:"servers,omitempty"`   // match tools from any of these MCP servers
}

// Action defines what to do with matched tools.
type Action struct {
	// Type is one of:
	//   "include"   - only matched tools are kept (whitelist)
	//   "exclude"   - matched tools are removed (blacklist)
	//   "summarize" - matched tools are kept but marked summary-only
	Type string `json:"type" yaml:"type"`
}
