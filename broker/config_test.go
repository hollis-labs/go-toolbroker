package broker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Write a temp config file.
	dir := t.TempDir()
	configPath := filepath.Join(dir, "tool-broker.json")
	configJSON := `{
		"rules": [
			{
				"name": "test-exclude",
				"intent": "*",
				"priority": 10,
				"match": {
					"patterns": ["hadron_bp_*"]
				},
				"action": {
					"type": "exclude"
				}
			},
			{
				"name": "test-include",
				"intent": "search",
				"priority": 20,
				"match": {
					"patterns": ["cortex_*"]
				},
				"action": {
					"type": "include"
				}
			}
		]
	}`
	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if len(cfg.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(cfg.Rules))
	}

	r0 := cfg.Rules[0]
	if r0.Name != "test-exclude" {
		t.Errorf("rule 0 name: expected 'test-exclude', got %q", r0.Name)
	}
	if r0.Intent != "*" {
		t.Errorf("rule 0 intent: expected '*', got %q", r0.Intent)
	}
	if r0.Priority != 10 {
		t.Errorf("rule 0 priority: expected 10, got %d", r0.Priority)
	}
	if r0.Action.Type != "exclude" {
		t.Errorf("rule 0 action: expected 'exclude', got %q", r0.Action.Type)
	}
	if len(r0.Match.Patterns) != 1 || r0.Match.Patterns[0] != "hadron_bp_*" {
		t.Errorf("rule 0 patterns: expected [hadron_bp_*], got %v", r0.Match.Patterns)
	}

	r1 := cfg.Rules[1]
	if r1.Name != "test-include" {
		t.Errorf("rule 1 name: expected 'test-include', got %q", r1.Name)
	}
	if r1.Action.Type != "include" {
		t.Errorf("rule 1 action: expected 'include', got %q", r1.Action.Type)
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(configPath, []byte("not json"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDefaultRules(t *testing.T) {
	rules := DefaultRules()
	if len(rules) == 0 {
		t.Fatal("expected non-empty default rules")
	}

	// Verify the first rule is the blueprint exclusion.
	found := false
	for _, r := range rules {
		if r.Name == "hide-blueprint-runners" {
			found = true
			if r.Action.Type != "exclude" {
				t.Errorf("hide-blueprint-runners should be exclude, got %q", r.Action.Type)
			}
			if len(r.Match.Patterns) != 1 || r.Match.Patterns[0] != "hadron_bp_*" {
				t.Errorf("unexpected patterns: %v", r.Match.Patterns)
			}
			break
		}
	}
	if !found {
		t.Error("missing 'hide-blueprint-runners' rule in defaults")
	}

	// Verify intent-specific rules exist.
	intents := map[string]bool{}
	for _, r := range rules {
		if r.Intent != "*" && r.Intent != "" {
			intents[r.Intent] = true
		}
	}
	for _, want := range []string{"explore_project", "search_memory", "run_build", "manage_tasks", "plan_sprint"} {
		if !intents[want] {
			t.Errorf("missing default rule for intent %q", want)
		}
	}
}

func TestLoadRulesFromFileYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "rules.yaml")
	yamlContent := `rules:
  - name: test-yaml-exclude
    intent: "*"
    priority: 10
    match:
      patterns:
        - "hadron_bp_*"
    action:
      type: exclude
  - name: test-yaml-include
    intent: search
    priority: 20
    match:
      patterns:
        - "cortex_*"
    action:
      type: include
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}

	rules, err := LoadRulesFromFile(yamlPath)
	if err != nil {
		t.Fatalf("LoadRulesFromFile YAML: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Name != "test-yaml-exclude" {
		t.Errorf("rule 0 name: expected 'test-yaml-exclude', got %q", rules[0].Name)
	}
	if rules[0].Action.Type != "exclude" {
		t.Errorf("rule 0 action: expected 'exclude', got %q", rules[0].Action.Type)
	}
	if rules[1].Name != "test-yaml-include" {
		t.Errorf("rule 1 name: expected 'test-yaml-include', got %q", rules[1].Name)
	}
	if rules[1].Priority != 20 {
		t.Errorf("rule 1 priority: expected 20, got %d", rules[1].Priority)
	}
}

func TestLoadRulesFromFileYML(t *testing.T) {
	dir := t.TempDir()
	ymlPath := filepath.Join(dir, "rules.yml")
	ymlContent := `rules:
  - name: yml-rule
    intent: "*"
    priority: 5
    match:
      patterns: ["test_*"]
    action:
      type: exclude
`
	if err := os.WriteFile(ymlPath, []byte(ymlContent), 0644); err != nil {
		t.Fatalf("write yml: %v", err)
	}

	rules, err := LoadRulesFromFile(ymlPath)
	if err != nil {
		t.Fatalf("LoadRulesFromFile .yml: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Name != "yml-rule" {
		t.Errorf("expected 'yml-rule', got %q", rules[0].Name)
	}
}

func TestLoadRulesFromFileJSON(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "rules.json")
	jsonContent := `{"rules": [{"name": "json-rule", "intent": "*", "priority": 10, "match": {"patterns": ["test_*"]}, "action": {"type": "exclude"}}]}`
	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("write json: %v", err)
	}

	rules, err := LoadRulesFromFile(jsonPath)
	if err != nil {
		t.Fatalf("LoadRulesFromFile JSON: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Name != "json-rule" {
		t.Errorf("expected 'json-rule', got %q", rules[0].Name)
	}
}

func TestLoadRulesFromFileUnknownExtension(t *testing.T) {
	dir := t.TempDir()
	tomlPath := filepath.Join(dir, "rules.toml")
	if err := os.WriteFile(tomlPath, []byte("test"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := LoadRulesFromFile(tomlPath)
	if err == nil {
		t.Fatal("expected error for unknown extension")
	}
	if !strings.Contains(err.Error(), "unsupported config format") {
		t.Errorf("expected 'unsupported config format' error, got: %v", err)
	}
}

func TestLoadRulesFromFileInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "bad.yaml")
	if err := os.WriteFile(yamlPath, []byte(":\n  :\n    - [invalid"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	_, err := LoadRulesFromFile(yamlPath)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestLoadRulesFromFileNotFound(t *testing.T) {
	_, err := LoadRulesFromFile("/nonexistent/rules.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestDefaultRulesLoadFromYAML(t *testing.T) {
	rules := DefaultRules()
	if len(rules) < 10 {
		t.Errorf("expected 10+ default rules, got %d", len(rules))
	}

	// Verify key rules exist.
	names := make(map[string]bool)
	for _, r := range rules {
		names[r.Name] = true
	}
	for _, want := range []string{
		"hide-blueprint-runners", "create-task", "list-tasks",
		"manage-tasks", "plan-sprint", "search-context",
		"write-context", "explore-project", "search-memory",
		"run-blueprint", "run-build", "check-health", "schedule-automation",
	} {
		if !names[want] {
			t.Errorf("missing default rule %q", want)
		}
	}
}

func TestDefaultRulesWithBroker(t *testing.T) {
	// Verify default rules work correctly with the broker.
	b := NewLocalBroker(testTools(), DefaultRules())

	// Default intent should exclude hadron_bp_* tools.
	result, err := b.SelectTools(nil, "general", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, tool := range result.Tools {
		if len(tool.Name) >= 10 && tool.Name[:10] == "hadron_bp_" {
			t.Errorf("default rules should exclude %q", tool.Name)
		}
	}
	expectedCount := len(testTools()) - 5 // 5 hadron_bp_* tools
	if result.Count != expectedCount {
		t.Errorf("expected %d tools with default rules, got %d", expectedCount, result.Count)
	}
}
