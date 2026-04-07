package broker

import (
	"context"
	"testing"
)

// testTools returns a representative set of tools for testing.
func testTools() []ToolDefinition {
	return []ToolDefinition{
		{Name: "volon_tasks_list", Description: "List tasks", Server: "volon", Tags: []string{"tasks"}},
		{Name: "volon_task_get", Description: "Get a task", Server: "volon", Tags: []string{"tasks"}},
		{Name: "volon_task_create", Description: "Create a task", Server: "volon", Tags: []string{"tasks", "write"}},
		{Name: "volon_sprints_list", Description: "List sprints", Server: "volon", Tags: []string{"sprints"}},
		{Name: "volon_backlog_list", Description: "List backlog", Server: "volon", Tags: []string{"backlog"}},
		{Name: "cortex_context_view", Description: "View context", Server: "cortex", Tags: []string{"context"}},
		{Name: "cortex_context_namespaces_list", Description: "List namespaces", Server: "cortex", Tags: []string{"context"}},
		{Name: "cortex_context_write", Description: "Write context", Server: "cortex", Tags: []string{"context", "write"}},
		{Name: "hadron_blueprints_list", Description: "List blueprints", Server: "hadron", Tags: []string{"blueprints"}},
		{Name: "hadron_run_enqueue", Description: "Enqueue a run", Server: "hadron", Tags: []string{"runs"}},
		{Name: "hadron_bp_build_volon", Description: "Build volon", Server: "hadron", Tags: []string{"blueprint-runner"}},
		{Name: "hadron_bp_build_cortex", Description: "Build cortex", Server: "hadron", Tags: []string{"blueprint-runner"}},
		{Name: "hadron_bp_build_hadron", Description: "Build hadron", Server: "hadron", Tags: []string{"blueprint-runner"}},
		{Name: "hadron_bp_lint_go", Description: "Lint Go project", Server: "hadron", Tags: []string{"blueprint-runner"}},
		{Name: "hadron_bp_test_go", Description: "Test Go project", Server: "hadron", Tags: []string{"blueprint-runner"}},
	}
}

func TestSelectToolsNoRules(t *testing.T) {
	b := NewLocalBroker(testTools(), nil)
	result, err := b.SelectTools(context.Background(), "anything", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != len(testTools()) {
		t.Errorf("expected %d tools, got %d", len(testTools()), result.Count)
	}
	if result.Total != len(testTools()) {
		t.Errorf("expected total %d, got %d", len(testTools()), result.Total)
	}
	if result.Intent != "anything" {
		t.Errorf("expected intent 'anything', got %q", result.Intent)
	}
}

func TestSelectToolsExcludeRule(t *testing.T) {
	rules := []Rule{
		{
			Name:     "hide-bp",
			Intent:   "*",
			Priority: 10,
			Match:    Match{Patterns: []string{"hadron_bp_*"}},
			Action:   Action{Type: "exclude"},
		},
	}
	b := NewLocalBroker(testTools(), rules)
	result, err := b.SelectTools(context.Background(), "general", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should exclude the 5 hadron_bp_* tools.
	expectedCount := len(testTools()) - 5
	if result.Count != expectedCount {
		t.Errorf("expected %d tools, got %d", expectedCount, result.Count)
	}

	// Verify none of the excluded tools are present.
	for _, tool := range result.Tools {
		if len(tool.Name) >= 10 && tool.Name[:10] == "hadron_bp_" {
			t.Errorf("tool %q should have been excluded", tool.Name)
		}
	}
}

func TestSelectToolsIntentSpecificInclude(t *testing.T) {
	rules := []Rule{
		{
			Name:     "explore",
			Intent:   "explore_project",
			Priority: 20,
			Match: Match{
				Patterns: []string{
					"volon_tasks_list",
					"volon_sprints_list",
					"cortex_context_view",
				},
			},
			Action: Action{Type: "include"},
		},
	}
	b := NewLocalBroker(testTools(), rules)

	// With matching intent, only included tools should be returned.
	result, err := b.SelectTools(context.Background(), "explore_project", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 3 {
		t.Errorf("expected 3 tools for explore_project, got %d", result.Count)
		for _, tool := range result.Tools {
			t.Logf("  got: %s", tool.Name)
		}
	}

	// With non-matching intent, all tools should be returned (no rules match).
	result2, err := b.SelectTools(context.Background(), "other_intent", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result2.Count != len(testTools()) {
		t.Errorf("expected all %d tools for other_intent, got %d", len(testTools()), result2.Count)
	}
}

func TestSelectToolsPatternMatching(t *testing.T) {
	rules := []Rule{
		{
			Name:     "cortex-only",
			Intent:   "search_memory",
			Priority: 20,
			Match:    Match{Patterns: []string{"cortex_*"}},
			Action:   Action{Type: "include"},
		},
	}
	b := NewLocalBroker(testTools(), rules)
	result, err := b.SelectTools(context.Background(), "search_memory", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should include cortex_context_view, cortex_context_namespaces_list, cortex_context_write.
	if result.Count != 3 {
		t.Errorf("expected 3 cortex tools, got %d", result.Count)
		for _, tool := range result.Tools {
			t.Logf("  got: %s", tool.Name)
		}
	}
	for _, tool := range result.Tools {
		if tool.Server != "cortex" {
			t.Errorf("expected only cortex tools, got %q (server: %s)", tool.Name, tool.Server)
		}
	}
}

func TestSelectToolsServerMatch(t *testing.T) {
	rules := []Rule{
		{
			Name:     "volon-only",
			Intent:   "volon_work",
			Priority: 20,
			Match:    Match{Servers: []string{"volon"}},
			Action:   Action{Type: "include"},
		},
	}
	b := NewLocalBroker(testTools(), rules)
	result, err := b.SelectTools(context.Background(), "volon_work", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Count != 5 {
		t.Errorf("expected 5 volon tools, got %d", result.Count)
	}
	for _, tool := range result.Tools {
		if tool.Server != "volon" {
			t.Errorf("expected only volon tools, got %q (server: %s)", tool.Name, tool.Server)
		}
	}
}

func TestSelectToolsTagMatch(t *testing.T) {
	rules := []Rule{
		{
			Name:     "write-tools",
			Intent:   "write_stuff",
			Priority: 20,
			Match:    Match{Tags: []string{"write"}},
			Action:   Action{Type: "include"},
		},
	}
	b := NewLocalBroker(testTools(), rules)
	result, err := b.SelectTools(context.Background(), "write_stuff", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// volon_task_create and cortex_context_write have "write" tag.
	if result.Count != 2 {
		t.Errorf("expected 2 write tools, got %d", result.Count)
		for _, tool := range result.Tools {
			t.Logf("  got: %s (tags: %v)", tool.Name, tool.Tags)
		}
	}
}

func TestSelectToolsPriorityOrdering(t *testing.T) {
	rules := []Rule{
		{
			Name:     "low-priority-include-all",
			Intent:   "test_priority",
			Priority: 5,
			Match:    Match{Patterns: []string{"*"}},
			Action:   Action{Type: "include"},
		},
		{
			Name:     "high-priority-exclude-bp",
			Intent:   "test_priority",
			Priority: 50,
			Match:    Match{Patterns: []string{"hadron_bp_*"}},
			Action:   Action{Type: "exclude"},
		},
	}
	b := NewLocalBroker(testTools(), rules)
	result, err := b.SelectTools(context.Background(), "test_priority", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The exclude rule should remove hadron_bp_* even though include-all matches everything.
	// Exclude takes precedence because excluded tools are filtered out regardless of include.
	for _, tool := range result.Tools {
		if len(tool.Name) >= 10 && tool.Name[:10] == "hadron_bp_" {
			t.Errorf("tool %q should have been excluded by high-priority rule", tool.Name)
		}
	}
}

func TestSelectToolsCombinedExcludeAndInclude(t *testing.T) {
	rules := []Rule{
		{
			Name:     "exclude-bp",
			Intent:   "*",
			Priority: 10,
			Match:    Match{Patterns: []string{"hadron_bp_*"}},
			Action:   Action{Type: "exclude"},
		},
		{
			Name:     "include-volon",
			Intent:   "manage_tasks",
			Priority: 20,
			Match:    Match{Patterns: []string{"volon_task_*", "volon_tasks_list"}},
			Action:   Action{Type: "include"},
		},
	}
	b := NewLocalBroker(testTools(), rules)
	result, err := b.SelectTools(context.Background(), "manage_tasks", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should include volon_task_get, volon_task_create, volon_tasks_list.
	// Should NOT include hadron_bp_* (excluded) or non-volon-task tools (not included).
	if result.Count != 3 {
		t.Errorf("expected 3 tools, got %d", result.Count)
		for _, tool := range result.Tools {
			t.Logf("  got: %s", tool.Name)
		}
	}
}

func TestAllTools(t *testing.T) {
	b := NewLocalBroker(testTools(), nil)
	summaries := b.AllTools()

	if len(summaries) != len(testTools()) {
		t.Errorf("expected %d summaries, got %d", len(testTools()), len(summaries))
	}

	// Verify summaries have the right fields.
	for _, s := range summaries {
		if s.Name == "" {
			t.Error("summary has empty name")
		}
		if s.Description == "" {
			t.Error("summary has empty description")
		}
	}
}

func TestRegisterToolsAndLoadRules(t *testing.T) {
	b := NewLocalBroker(nil, nil)

	if len(b.AllTools()) != 0 {
		t.Error("expected no tools initially")
	}

	b.RegisterTools(testTools()[:3])
	if len(b.AllTools()) != 3 {
		t.Errorf("expected 3 tools after register, got %d", len(b.AllTools()))
	}

	b.RegisterTools(testTools()[3:5])
	if len(b.AllTools()) != 5 {
		t.Errorf("expected 5 tools after second register, got %d", len(b.AllTools()))
	}

	b.LoadRules([]Rule{
		{Name: "test", Intent: "*", Match: Match{Patterns: []string{"volon_*"}}, Action: Action{Type: "include"}},
	})
	result, err := b.SelectTools(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// All 5 registered tools are volon_*, so all should be included.
	if result.Count != 5 {
		t.Errorf("expected 5 tools, got %d", result.Count)
	}
}

func TestSelectToolsEmptyRegistry(t *testing.T) {
	b := NewLocalBroker(nil, nil)
	result, err := b.SelectTools(context.Background(), "anything", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 0 {
		t.Errorf("expected 0 tools, got %d", result.Count)
	}
	if result.Rationale != "no tools registered" {
		t.Errorf("unexpected rationale: %q", result.Rationale)
	}
}
