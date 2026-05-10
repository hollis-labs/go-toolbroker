package broker

import (
	"strings"
)

// Intent represents a detected intent with its confidence score.
type Intent struct {
	Name       string   `json:"name"`
	Confidence float64  `json:"confidence"`
	Keywords   []string `json:"keywords"`
}

// intentKeywords maps intent names to their trigger keywords.
// Each keyword is lowercased for case-insensitive matching.
var intentKeywords = map[string][]string{
	"create-task": {
		"create task", "new task", "add task", "make task",
		"create ticket", "new ticket", "file ticket",
	},
	"list-tasks": {
		"list tasks", "show tasks", "what tasks", "my tasks",
		"task list", "pending tasks", "open tasks",
	},
	"search-context": {
		"search context", "find context", "look up", "search for",
		"context search", "query context", "find in context",
	},
	"write-context": {
		"write context", "save context", "store context",
		"update context", "record context", "remember this",
	},
	"run-blueprint": {
		"run blueprint", "execute blueprint", "start blueprint",
		"trigger blueprint", "launch blueprint", "run build",
		"build project", "deploy",
	},
	"check-health": {
		"check health", "health check", "status check",
		"is it up", "service status", "ping",
	},
	"explore-project": {
		"explore project", "project overview", "what is this project",
		"show project", "project status", "sprint status",
	},
	"search-memory": {
		"search memory", "recall", "what do you know about",
		"remember", "memory search", "find memory",
	},
	"manage-tasks": {
		"manage task", "update task", "edit task", "move task",
		"transition task", "close task", "delete task", "assign task",
	},
	"plan-sprint": {
		"plan sprint", "new sprint", "create sprint", "sprint planning",
		"backlog grooming", "prioritize backlog", "next sprint",
	},
	"schedule-automation": {
		"schedule", "automate", "cron", "recurring",
		"pipeline", "scheduled run", "nightly",
	},
	// Build / test / lint / release intents — match rules in default-rules.yaml.
	"run-tests": {
		"run tests", "run test", "go test", "npm test", "test suite",
		"unit tests", "integration tests", "test project", "check tests",
	},
	"run-build": {
		"run build", "go build", "npm build", "build project",
		"compile", "make build", "build all",
	},
	"audit": {
		"audit", "compliance", "check compliance", "run audit",
		"security audit", "code audit", "otel compliance",
		"api contract", "dependency audit", "portfolio audit",
	},
	"backup": {
		"backup", "back up", "snapshot", "save backup",
		"database backup", "config backup", "time machine",
		"dotfile backup", "agentrc backup",
	},
	"release": {
		"release", "tag release", "cut release", "release notes",
		"publish release", "ship release", "version bump",
	},
	"docker": {
		"docker", "container", "docker build", "docker compose",
		"containerize", "dockerfile", "build image",
	},
	"cleanup": {
		"cleanup", "clean up", "nightly cleanup", "prune",
		"remove stale", "garbage collect", "housekeeping",
	},
	"generate-report": {
		"generate report", "standup", "standup report", "status report",
		"codebase statistics", "archaeology report", "migration diff",
		"dependency graph", "service graph", "blueprint index",
	},
	"search-code": {
		"search code", "grep", "find in code", "cross project grep",
		"search across projects", "code search", "find usage",
	},
	"check-drift": {
		"check drift", "dependency drift", "drift detection",
		"version drift", "outdated dependencies", "stale dependencies",
	},
}

// DetectIntent analyzes a message and returns matching intents ranked by confidence.
// Uses keyword extraction (not ML) to infer intents.
// Returns an empty slice for empty input.
func DetectIntent(message string) []Intent {
	if strings.TrimSpace(message) == "" {
		return nil
	}

	lower := strings.ToLower(message)
	var results []Intent

	for intentName, keywords := range intentKeywords {
		var matched []string
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				matched = append(matched, kw)
			}
		}
		if len(matched) > 0 {
			// Confidence is based on fraction of keywords matched,
			// scaled to a 0.0–1.0 range. More keyword hits = higher confidence.
			confidence := float64(len(matched)) / float64(len(keywords))
			// Boost confidence for longer keyword matches (more specific).
			for _, kw := range matched {
				wordCount := len(strings.Fields(kw))
				if wordCount >= 2 {
					confidence += 0.1
				}
			}
			// Cap at 1.0.
			if confidence > 1.0 {
				confidence = 1.0
			}
			results = append(results, Intent{
				Name:       intentName,
				Confidence: confidence,
				Keywords:   matched,
			})
		}
	}

	// Sort by confidence descending (stable sort to keep deterministic ordering).
	sortIntents(results)
	return results
}

// sortIntents sorts intents by confidence descending, then by name for stability.
func sortIntents(intents []Intent) {
	for i := 1; i < len(intents); i++ {
		for j := i; j > 0; j-- {
			if intents[j].Confidence > intents[j-1].Confidence ||
				(intents[j].Confidence == intents[j-1].Confidence && intents[j].Name < intents[j-1].Name) {
				intents[j], intents[j-1] = intents[j-1], intents[j]
			} else {
				break
			}
		}
	}
}
