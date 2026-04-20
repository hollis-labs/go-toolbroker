package broker

import (
	"context"
	"strings"
	"testing"
)

// mapEnricher is a test Enricher backed by an in-memory map.
type mapEnricher map[string]Hints

func (m mapEnricher) LookupByToolName(ctx context.Context, name string) (Hints, bool, error) {
	h, ok := m[name]
	return h, ok, nil
}

func TestComposeOverrideBlock_EmptyWhenNoEnrichment(t *testing.T) {
	tools := []string{"tool_a", "tool_b"}
	enr := mapEnricher{}
	got, err := ComposeOverrideBlock(context.Background(), tools, enr)
	if err != nil {
		t.Fatalf("ComposeOverrideBlock: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty block, got:\n%s", got)
	}
}

func TestComposeOverrideBlock_SingleTool(t *testing.T) {
	tools := []string{"tool_a"}
	enr := mapEnricher{
		"tool_a": {OutputShape: "list of records"},
	}
	got, err := ComposeOverrideBlock(context.Background(), tools, enr)
	if err != nil {
		t.Fatalf("ComposeOverrideBlock: %v", err)
	}
	if !strings.Contains(got, "## Tool Overrides") {
		t.Errorf("missing header:\n%s", got)
	}
	if !strings.Contains(got, "tool_a") {
		t.Errorf("missing tool name:\n%s", got)
	}
	if !strings.Contains(got, "list of records") {
		t.Errorf("missing output shape:\n%s", got)
	}
}

func TestComposeOverrideBlock_PartialEnrichment(t *testing.T) {
	tools := []string{"tool_a", "tool_b", "tool_c"}
	enr := mapEnricher{
		"tool_b": {
			AntiPatterns: []string{"IDs are ULIDs"},
			OutputShape:  "array of x",
		},
	}
	got, err := ComposeOverrideBlock(context.Background(), tools, enr)
	if err != nil {
		t.Fatalf("ComposeOverrideBlock: %v", err)
	}
	if !strings.Contains(got, "tool_b") {
		t.Errorf("missing enriched tool: %s", got)
	}
	if strings.Contains(got, "tool_a") || strings.Contains(got, "tool_c") {
		t.Errorf("unexpectedly included non-enriched tools: %s", got)
	}
	if !strings.Contains(got, "ULID") {
		t.Errorf("missing anti-pattern content: %s", got)
	}
}

func TestComposeOverrideBlock_EmptyToolList(t *testing.T) {
	got, err := ComposeOverrideBlock(context.Background(), nil, mapEnricher{})
	if err != nil {
		t.Fatalf("ComposeOverrideBlock: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty, got: %s", got)
	}
}

func TestComposeOverrideBlock_NilEnricher(t *testing.T) {
	tools := []string{"tool_a"}
	got, err := ComposeOverrideBlock(context.Background(), tools, nil)
	if err != nil {
		t.Fatalf("ComposeOverrideBlock: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty with nil enricher, got: %s", got)
	}
}

func TestComposeOverrideBlock_NilContext(t *testing.T) {
	// A nil context must not panic and must not propagate to the Enricher:
	// the public helper substitutes context.Background() defensively.
	tools := []string{"tool_a"}
	enr := mapEnricher{"tool_a": {OutputShape: "list of records"}}
	//nolint:staticcheck // intentionally passing nil to exercise the guard
	got, err := ComposeOverrideBlock(nil, tools, enr)
	if err != nil {
		t.Fatalf("ComposeOverrideBlock: %v", err)
	}
	if !strings.Contains(got, "list of records") {
		t.Errorf("expected populated block, got: %q", got)
	}
}

func TestComposeOverrideBlock_FiltersBlankEntries(t *testing.T) {
	// Whitespace-only slice entries shouldn't produce dangling sections like
	// "preconditions: " with no content; whitespace-only OutputShape should
	// be dropped; a tool whose Hints contain only blanks is skipped entirely.
	tools := []string{"tool_a", "tool_b"}
	enr := mapEnricher{
		"tool_a": {
			OutputShape:   "  ",
			Preconditions: []string{"", "  "},
			AntiPatterns:  []string{"real warning", ""},
		},
		"tool_b": {
			Preconditions: []string{"", "\t"},
		},
	}
	got, err := ComposeOverrideBlock(context.Background(), tools, enr)
	if err != nil {
		t.Fatalf("ComposeOverrideBlock: %v", err)
	}
	if strings.Contains(got, "preconditions: ") && !strings.Contains(got, "preconditions: real") {
		t.Errorf("dangling preconditions section in output:\n%s", got)
	}
	if !strings.Contains(got, "real warning") {
		t.Errorf("missing real content:\n%s", got)
	}
	if strings.Contains(got, "tool_b") {
		t.Errorf("tool_b had only blank entries and should have been skipped:\n%s", got)
	}
}
