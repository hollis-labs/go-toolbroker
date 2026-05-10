// Package main demonstrates loading rules from a YAML file at runtime.
//
// Many real applications keep their tool-broker rules under version control
// alongside service config. broker.LoadRulesFromFile auto-detects the format
// from the file extension (.yaml, .yml, .json).
//
// Run from the repo root with:
//
//	go run ./examples/yaml-rules
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hollis-labs/go-toolbroker/broker"
)

const rulesYAML = `rules:
  # Globally hide tools from a noisy/internal server.
  - name: hide-admin-server
    intent: "*"
    priority: 10
    match:
      servers:
        - admin
    action:
      type: exclude

  # On a "search" intent (or any "search_*" intent), include only tools
  # tagged "search". The Intent field supports glob patterns.
  - name: search-tools
    intent: search*
    priority: 20
    match:
      tags:
        - search
    action:
      type: include

  # On a "manage_tasks" intent, include only tools from the tasks server.
  - name: manage-tasks
    intent: manage_tasks
    priority: 20
    match:
      servers:
        - tasks
    action:
      type: include
`

func main() {
	// Write the rule file to a temp location for the example. In a real app
	// this would be a config file shipped alongside the binary.
	dir, err := os.MkdirTemp("", "toolbroker-yaml-example-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	rulesPath := filepath.Join(dir, "rules.yaml")
	if err := os.WriteFile(rulesPath, []byte(rulesYAML), 0o644); err != nil {
		panic(err)
	}

	// Load the rules.
	rules, err := broker.LoadRulesFromFile(rulesPath)
	if err != nil {
		panic(fmt.Errorf("load rules: %w", err))
	}
	fmt.Printf("loaded %d rules from %s\n", len(rules), rulesPath)

	// Build a broker and register tools.
	b := broker.NewLocalBroker(nil, rules)
	b.RegisterTools([]broker.ToolDefinition{
		{Name: "tasks_list", Description: "List tasks", Server: "tasks"},
		{Name: "tasks_create", Description: "Create a task", Server: "tasks"},
		{Name: "search_docs", Description: "Search documentation", Server: "docs", Tags: []string{"search"}},
		{Name: "admin_reset_db", Description: "Reset DB (dangerous)", Server: "admin"},
	})

	// Try a few intents and print the result.
	for _, intent := range []string{"general", "manage_tasks", "search_docs"} {
		result, err := b.SelectTools(context.Background(), intent, nil)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\nintent=%s -> %d/%d tools (rationale: %s)\n",
			intent, result.Count, result.Total, result.Rationale)
		for _, t := range result.Tools {
			fmt.Printf("  - %s\n", t.Name)
		}
	}
}
