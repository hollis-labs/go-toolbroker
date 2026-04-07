package broker

import "testing"

func TestEstimateToolTokens(t *testing.T) {
	tools := testTools()[:3]
	tokens := EstimateToolTokens(tools)
	if tokens <= 0 {
		t.Errorf("expected positive token count, got %d", tokens)
	}
}

func TestEstimateToolTokensEmpty(t *testing.T) {
	tokens := EstimateToolTokens(nil)
	if tokens != 0 {
		t.Errorf("expected 0 tokens for nil, got %d", tokens)
	}
}

func TestPruneToTokenBudget(t *testing.T) {
	tools := testTools()
	total := EstimateToolTokens(tools)

	// Budget larger than total — no pruning.
	result := PruneToTokenBudget(tools, total+1000)
	if len(result) != len(tools) {
		t.Errorf("expected no pruning, got %d/%d", len(result), len(tools))
	}

	// Budget of 0 — should keep at least 1.
	result = PruneToTokenBudget(tools, 0)
	if len(result) != 1 {
		t.Errorf("expected 1 tool with zero budget, got %d", len(result))
	}

	// Budget half of total — should prune some.
	result = PruneToTokenBudget(tools, total/2)
	if len(result) >= len(tools) {
		t.Errorf("expected fewer tools with half budget, got %d/%d", len(result), len(tools))
	}
	if len(result) < 1 {
		t.Error("expected at least 1 tool")
	}
}

func TestPruneToTokenBudgetEmpty(t *testing.T) {
	result := PruneToTokenBudget(nil, 1000)
	if len(result) != 0 {
		t.Errorf("expected 0 for nil input, got %d", len(result))
	}
}

func TestPruneDoesNotMutateOriginal(t *testing.T) {
	tools := testTools()
	original := len(tools)
	_ = PruneToTokenBudget(tools, 0)
	if len(tools) != original {
		t.Errorf("original slice was mutated: %d → %d", original, len(tools))
	}
}
