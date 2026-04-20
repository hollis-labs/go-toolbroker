package broker

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Hints carries tool-specific guidance surfaced to the LLM via the broker's
// override block. Stored externally to the ToolDefinition so Hints can be
// updated via data / probe-agent enrichment without touching tool
// implementations. Each field is optional.
type Hints struct {
	// Preconditions are assertions the agent should verify before invoking
	// the tool ("Call list/search first to get a real ID").
	Preconditions []string `json:"preconditions,omitempty"`

	// AntiPatterns are ways the agent commonly misuses the tool that the
	// broker wants to prevent ("IDs are ULIDs, not file paths").
	AntiPatterns []string `json:"anti_patterns,omitempty"`

	// ChainsWith lists other tools this one is commonly paired with
	// ("dev_glob before dev_read").
	ChainsWith []string `json:"chains_with,omitempty"`

	// OutputShape is a one-line description of what the tool returns,
	// so the agent can plan the next call without re-reading the full result.
	OutputShape string `json:"output_shape,omitempty"`
}

// IsEmpty returns true if none of the hint fields carry content. Slices of
// only whitespace-only entries (e.g. []string{""} or []string{"  "}) are
// treated as empty, as is a whitespace-only OutputShape — those inputs would
// otherwise render as dangling "preconditions: " lines in the override block.
func (h Hints) IsEmpty() bool {
	return !hasNonBlankEntry(h.Preconditions) &&
		!hasNonBlankEntry(h.AntiPatterns) &&
		!hasNonBlankEntry(h.ChainsWith) &&
		strings.TrimSpace(h.OutputShape) == ""
}

// hasNonBlankEntry reports whether ss contains at least one entry that is
// non-empty after trimming whitespace.
func hasNonBlankEntry(ss []string) bool {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return true
		}
	}
	return false
}

// MarshalHints encodes Hints to a JSON string suitable for storage by
// consumers that persist per-tool enrichment records.
func MarshalHints(h Hints) (string, error) {
	b, err := json.Marshal(h)
	if err != nil {
		return "", fmt.Errorf("marshal hints: %w", err)
	}
	return string(b), nil
}

// UnmarshalHints decodes a stored JSON string into a Hints struct. Empty
// input or "{}" yields a zero-value Hints with no error.
func UnmarshalHints(s string) (Hints, error) {
	if s == "" {
		return Hints{}, nil
	}
	var h Hints
	if err := json.Unmarshal([]byte(s), &h); err != nil {
		return Hints{}, fmt.Errorf("unmarshal hints: %w", err)
	}
	return h, nil
}
