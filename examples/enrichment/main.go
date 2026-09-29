// Package main demonstrates the enrichment pipeline: per-tool Hints
// (preconditions, anti-patterns, output shape, chaining) composed into
// a markdown override block that callers can append to the system prompt.
//
// Storage for the Hints is consumer-owned. This example uses a tiny
// in-memory map; a real application would back the Enricher with SQLite,
// a remote service, or whatever store already holds tool metadata.
//
// Run from the repo root with:
//
//	go run ./examples/enrichment
package main

import (
	"context"
	"fmt"

	"github.com/hollis-labs/go-toolbroker/broker"
)

// inMemoryEnricher is a trivial Enricher backed by a map. It satisfies the
// broker.Enricher interface with a single method.
type inMemoryEnricher struct {
	data map[string]broker.Hints
}

func (e *inMemoryEnricher) LookupByToolName(_ context.Context, name string) (broker.Hints, bool, error) {
	h, ok := e.data[name]
	return h, ok, nil
}

func main() {
	// Register hints for two tools. Tools without hints are silently
	// skipped during composition.
	enr := &inMemoryEnricher{
		data: map[string]broker.Hints{
			"tasks_create": {
				Preconditions: []string{
					"Confirm the task title is non-empty.",
					"Resolve the project_id before calling.",
				},
				AntiPatterns: []string{
					"Do not call this for status updates — use tasks_transition.",
				},
				ChainsWith:  []string{"tasks_get", "tasks_transition"},
				OutputShape: "Returns {id, title, status, created_at}.",
			},
			"build_go_project": {
				Preconditions: []string{
					"Repo must contain go.mod at project_path.",
				},
				OutputShape: "Returns {ok, stdout, stderr, exit_code}.",
			},
		},
	}

	// A minimal include-only rule keeps the example focused on enrichment.
	rules := []broker.Rule{
		{
			Name:     "include-everything",
			Intent:   "*",
			Priority: 10,
			Match:    broker.Match{Patterns: []string{"*"}},
			Action:   broker.Action{Type: "include"},
		},
	}

	// Wire the enricher in via a functional option.
	b := broker.NewLocalBroker(nil, rules, broker.WithEnricher(enr))
	b.RegisterTools([]broker.ToolDefinition{
		{Name: "tasks_create", Description: "Create a task", Server: "tasks"},
		{Name: "tasks_get", Description: "Get a task by id", Server: "tasks"},
		{Name: "build_go_project", Description: "Run go build", Server: "ci"},
		{Name: "tools_without_hints", Description: "An unenriched tool", Server: "misc"},
	})

	result, err := b.SelectTools(context.Background(), "general", nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("selected %d/%d tools\n\n", result.Count, result.Total)
	if result.OverrideBlock == "" {
		fmt.Println("(no override block — no selected tool had hints)")
		return
	}

	fmt.Println("--- override block (append to system prompt) ---")
	fmt.Println(result.OverrideBlock)
	fmt.Println("--- end override block ---")
}
