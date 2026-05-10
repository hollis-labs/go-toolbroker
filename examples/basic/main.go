// Package main demonstrates the minimal go-toolbroker usage pattern:
// hand-written rules, in-memory tool registration, and intent-driven
// selection.
//
// Run from the repo root with:
//
//	go run ./examples/basic
package main

import (
	"context"
	"fmt"

	"github.com/hollis-labs/go-toolbroker/broker"
)

func main() {
	// Define a small rule set.
	//
	// Rules behave as follows:
	//   - exclude rules unconditionally drop matching tools (always win).
	//   - include rules act as a per-intent whitelist: when any include
	//     rule fires for the requested intent, only tools touched by an
	//     include rule survive (exclusions still apply on top).
	//   - Priority orders rationale and rule application; it does NOT let
	//     an include override an exclude on the same tool.
	rules := []broker.Rule{
		// Globally hide an internal/admin server's tools.
		{
			Name:     "hide-admin-server",
			Intent:   "*",
			Priority: 10,
			Match:    broker.Match{Servers: []string{"admin"}},
			Action:   broker.Action{Type: "exclude"},
		},
		// On a "manage_tasks" intent, surface only task-server tools.
		{
			Name:     "manage-tasks-include",
			Intent:   "manage_tasks",
			Priority: 20,
			Match:    broker.Match{Servers: []string{"tasks"}},
			Action:   broker.Action{Type: "include"},
		},
		// On a "search" intent, surface anything tagged "search".
		{
			Name:     "search-include",
			Intent:   "search",
			Priority: 20,
			Match:    broker.Match{Tags: []string{"search"}},
			Action:   broker.Action{Type: "include"},
		},
	}

	// Create a broker and register a representative tool set. In a real
	// embedding these would come from MCP server discovery.
	b := broker.NewLocalBroker(nil, rules)
	b.RegisterTools([]broker.ToolDefinition{
		{Name: "tasks_list", Description: "List open tasks", Server: "tasks"},
		{Name: "tasks_create", Description: "Create a task", Server: "tasks"},
		{Name: "search_docs", Description: "Search documentation", Server: "docs", Tags: []string{"search"}},
		{Name: "search_code", Description: "Search source code", Server: "code", Tags: []string{"search"}},
		{Name: "admin_reset_db", Description: "Reset the database", Server: "admin"},
	})

	for _, intent := range []string{"general", "manage_tasks", "search"} {
		result, err := b.SelectTools(context.Background(), intent, nil)
		if err != nil {
			panic(err)
		}
		fmt.Printf("intent=%-13s -> %d/%d tools (rationale: %s)\n",
			intent, result.Count, result.Total, result.Rationale)
		for _, t := range result.Tools {
			fmt.Printf("  - %s (%s)\n", t.Name, t.Server)
		}
	}

	// Token-budget pruning is independent of selection. Pass any
	// []ToolDefinition and a token budget; tools are dropped tail-first
	// until the estimate fits.
	result, _ := b.SelectTools(context.Background(), "general", nil)
	pruned := broker.PruneToTokenBudget(result.Tools, 200)
	fmt.Printf("\npruned 'general' selection to ~200 tokens -> %d tools\n", len(pruned))
}
