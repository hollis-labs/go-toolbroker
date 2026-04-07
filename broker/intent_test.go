package broker

import (
	"testing"
)

func TestDetectIntentCreateTask(t *testing.T) {
	intents := DetectIntent("create a new task for cortex")
	if len(intents) == 0 {
		t.Fatal("expected at least one intent")
	}
	if intents[0].Name != "create-task" {
		t.Errorf("expected top intent 'create-task', got %q", intents[0].Name)
	}
}

func TestDetectIntentEmptyInput(t *testing.T) {
	intents := DetectIntent("")
	if len(intents) != 0 {
		t.Errorf("expected empty slice for empty input, got %d intents", len(intents))
	}
}

func TestDetectIntentWhitespaceOnly(t *testing.T) {
	intents := DetectIntent("   \t\n  ")
	if len(intents) != 0 {
		t.Errorf("expected empty slice for whitespace input, got %d intents", len(intents))
	}
}

func TestDetectIntentSearchContext(t *testing.T) {
	intents := DetectIntent("search context for deployment notes")
	if len(intents) == 0 {
		t.Fatal("expected at least one intent")
	}
	found := false
	for _, i := range intents {
		if i.Name == "search-context" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'search-context' in results, got %v", intentNames(intents))
	}
}

func TestDetectIntentRunBlueprint(t *testing.T) {
	intents := DetectIntent("run blueprint to build the project")
	if len(intents) == 0 {
		t.Fatal("expected at least one intent")
	}
	found := false
	for _, i := range intents {
		if i.Name == "run-blueprint" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'run-blueprint' in results, got %v", intentNames(intents))
	}
}

func TestDetectIntentCheckHealth(t *testing.T) {
	intents := DetectIntent("check health of the services")
	if len(intents) == 0 {
		t.Fatal("expected at least one intent")
	}
	if intents[0].Name != "check-health" {
		t.Errorf("expected top intent 'check-health', got %q", intents[0].Name)
	}
}

func TestDetectIntentCaseInsensitive(t *testing.T) {
	intents := DetectIntent("CREATE A NEW TASK please")
	if len(intents) == 0 {
		t.Fatal("expected at least one intent")
	}
	if intents[0].Name != "create-task" {
		t.Errorf("expected 'create-task', got %q", intents[0].Name)
	}
}

func TestDetectIntentConfidenceOrdering(t *testing.T) {
	// A message matching multiple keywords for one intent should rank higher.
	intents := DetectIntent("create task and new task for the project")
	if len(intents) == 0 {
		t.Fatal("expected at least one intent")
	}
	// create-task should be top since it matches multiple keywords.
	if intents[0].Name != "create-task" {
		t.Errorf("expected 'create-task' as top intent, got %q", intents[0].Name)
	}

	// Verify descending confidence order.
	for i := 1; i < len(intents); i++ {
		if intents[i].Confidence > intents[i-1].Confidence {
			t.Errorf("intents not sorted by confidence: %s(%.2f) > %s(%.2f)",
				intents[i].Name, intents[i].Confidence,
				intents[i-1].Name, intents[i-1].Confidence)
		}
	}
}

func TestDetectIntentKeywordsPopulated(t *testing.T) {
	intents := DetectIntent("create a new task")
	if len(intents) == 0 {
		t.Fatal("expected at least one intent")
	}
	if len(intents[0].Keywords) == 0 {
		t.Error("expected non-empty keywords for matched intent")
	}
}

func TestDetectIntentNoMatch(t *testing.T) {
	intents := DetectIntent("the quick brown fox jumps over the lazy dog")
	// No keywords should match this unrelated sentence.
	if len(intents) != 0 {
		t.Errorf("expected no intents for unrelated input, got %v", intentNames(intents))
	}
}

func TestDetectIntentManageTasks(t *testing.T) {
	intents := DetectIntent("update task status to done")
	found := false
	for _, i := range intents {
		if i.Name == "manage-tasks" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'manage-tasks' for 'update task', got %v", intentNames(intents))
	}
}

func TestDetectIntentPlanSprint(t *testing.T) {
	intents := DetectIntent("let's plan sprint for next week")
	found := false
	for _, i := range intents {
		if i.Name == "plan-sprint" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'plan-sprint', got %v", intentNames(intents))
	}
}

func TestDetectIntentScheduleAutomation(t *testing.T) {
	intents := DetectIntent("schedule a nightly build")
	found := false
	for _, i := range intents {
		if i.Name == "schedule-automation" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'schedule-automation', got %v", intentNames(intents))
	}
}

func TestDetectIntentConfidenceCappedAtOne(t *testing.T) {
	// Even with many keyword matches, confidence should not exceed 1.0.
	intents := DetectIntent("create task new task add task make task create ticket new ticket file ticket")
	if len(intents) == 0 {
		t.Fatal("expected at least one intent")
	}
	for _, i := range intents {
		if i.Confidence > 1.0 {
			t.Errorf("confidence for %q exceeds 1.0: %.2f", i.Name, i.Confidence)
		}
	}
}

// ── Hadron blueprint intent tests ──────────────────────────────────────────

func TestDetectIntentRunTests(t *testing.T) {
	intents := DetectIntent("run tests for the volon project")
	assertHasIntent(t, intents, "run-tests")
}

func TestDetectIntentRunBuild(t *testing.T) {
	intents := DetectIntent("go build the cortex backend")
	assertHasIntent(t, intents, "run-build")
}

func TestDetectIntentAudit(t *testing.T) {
	intents := DetectIntent("run an audit on API contracts")
	assertHasIntent(t, intents, "audit")
}

func TestDetectIntentBackup(t *testing.T) {
	intents := DetectIntent("backup all databases")
	assertHasIntent(t, intents, "backup")
}

func TestDetectIntentRelease(t *testing.T) {
	intents := DetectIntent("tag release for v2.0")
	assertHasIntent(t, intents, "release")
}

func TestDetectIntentDocker(t *testing.T) {
	intents := DetectIntent("docker build the volon image")
	assertHasIntent(t, intents, "docker")
}

func TestDetectIntentGenerateReport(t *testing.T) {
	intents := DetectIntent("generate a standup report")
	assertHasIntent(t, intents, "generate-report")
}

func TestDetectIntentSearchCode(t *testing.T) {
	intents := DetectIntent("grep for ToolBroker across projects")
	assertHasIntent(t, intents, "search-code")
}

func TestDetectIntentCheckDrift(t *testing.T) {
	intents := DetectIntent("check dependency drift in go modules")
	assertHasIntent(t, intents, "check-drift")
}

func TestDetectIntentCleanup(t *testing.T) {
	intents := DetectIntent("run nightly cleanup and prune stale data")
	assertHasIntent(t, intents, "cleanup")
}

func assertHasIntent(t *testing.T, intents []Intent, name string) {
	t.Helper()
	if len(intents) == 0 {
		t.Fatalf("expected at least one intent for %q", name)
	}
	for _, i := range intents {
		if i.Name == name {
			return
		}
	}
	t.Errorf("expected %q in results, got %v", name, intentNames(intents))
}

// intentNames returns a slice of intent names for test error messages.
func intentNames(intents []Intent) []string {
	names := make([]string, len(intents))
	for i, intent := range intents {
		names[i] = intent.Name
	}
	return names
}
