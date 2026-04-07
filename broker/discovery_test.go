package broker

import "testing"

func TestScoreByKeywords(t *testing.T) {
	tools := testTools()
	result := ScoreByKeywords(tools, []string{"task", "list"}, 5)
	if len(result) == 0 {
		t.Fatal("expected some results")
	}
	// volon_tasks_list should be first — matches both "task" and "list" in name.
	if result[0].Name != "volon_tasks_list" {
		t.Errorf("expected volon_tasks_list first, got %s", result[0].Name)
	}
}

func TestScoreByKeywordsEmpty(t *testing.T) {
	result := ScoreByKeywords(nil, []string{"task"}, 5)
	if result != nil {
		t.Error("expected nil for empty tools")
	}

	result = ScoreByKeywords(testTools(), nil, 5)
	if result != nil {
		t.Error("expected nil for empty keywords")
	}

	result = ScoreByKeywords(testTools(), []string{""}, 5)
	if result != nil {
		t.Error("expected nil for blank keywords")
	}
}

func TestScoreByKeywordsMaxResults(t *testing.T) {
	tools := testTools()
	result := ScoreByKeywords(tools, []string{"volon"}, 2)
	if len(result) > 2 {
		t.Errorf("expected at most 2 results, got %d", len(result))
	}
}

func TestScoreByIntent(t *testing.T) {
	tools := testTools()
	result := ScoreByIntent(tools, "I want to list all tasks", 5)
	if len(result) == 0 {
		t.Fatal("expected some results for 'list all tasks'")
	}
	// Should find task-related tools.
	foundTask := false
	for _, r := range result {
		if r.Name == "volon_tasks_list" {
			foundTask = true
		}
	}
	if !foundTask {
		t.Error("expected volon_tasks_list in results")
	}
}

func TestScoreByIntentEmpty(t *testing.T) {
	result := ScoreByIntent(testTools(), "", 5)
	if result != nil {
		t.Error("expected nil for empty intent")
	}
}

func TestFindByNames(t *testing.T) {
	tools := testTools()
	result := FindByNames(tools, []string{"volon_tasks_list", "cortex_context_view", "nonexistent"})
	if len(result) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result))
	}
	if result[0].Name != "volon_tasks_list" {
		t.Errorf("expected volon_tasks_list first (order follows input), got %s", result[0].Name)
	}
	if result[1].Name != "cortex_context_view" {
		t.Errorf("expected cortex_context_view second, got %s", result[1].Name)
	}
}

func TestFindByNamesEmpty(t *testing.T) {
	result := FindByNames(nil, []string{"x"})
	if result != nil {
		t.Error("expected nil for empty tools")
	}
	result = FindByNames(testTools(), nil)
	if result != nil {
		t.Error("expected nil for empty names")
	}
}

func TestTokenizeIntent(t *testing.T) {
	words := TokenizeIntent("I want to create a new task!")
	// "want" and "the" are stop words; "to", "a" are <3 chars.
	// Should get: "create", "new", "task"
	expected := map[string]bool{"create": true, "new": true, "task": true}
	for _, w := range words {
		if !expected[w] {
			t.Errorf("unexpected word %q", w)
		}
	}
	if len(words) != len(expected) {
		t.Errorf("expected %d words, got %d: %v", len(expected), len(words), words)
	}
}

func TestTokenizeIntentEmpty(t *testing.T) {
	words := TokenizeIntent("")
	if len(words) != 0 {
		t.Errorf("expected empty for empty input, got %v", words)
	}
}
