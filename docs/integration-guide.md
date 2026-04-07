# Tool Broker Integration Guide

How to integrate the tool broker into any Go application
for intent-aware MCP tool selection, token budget management,
and progressive discovery.

## 1. Initialize the Broker with YAML Rules

The broker ships with embedded default rules. You can also load custom rules
from a YAML or JSON file.

```go
import "github.com/hollis-labs/tool-broker/broker"

// Option A: Use built-in defaults (recommended starting point).
b := broker.NewLocalBroker(nil, broker.DefaultRules())

// Option B: Load custom rules from a YAML file.
rules, err := broker.LoadRulesFromFile("~/.config/fragments/tool-broker-rules.yaml")
if err != nil {
    log.Fatalf("load rules: %v", err)
}
b := broker.NewLocalBroker(nil, rules)

// Option C: Merge defaults with custom overrides.
rules := broker.DefaultRules()
custom, _ := broker.LoadRulesFromFile("my-rules.yaml")
rules = append(rules, custom...)
b := broker.NewLocalBroker(nil, rules)
```

After creating the broker, register your MCP tools:

```go
// Convert discovered MCP tools into broker ToolDefinitions.
var brokerTools []broker.ToolDefinition
for _, entry := range discoveredTools {
    brokerTools = append(brokerTools, broker.ToolDefinition{
        Name:        entry.tool.Name,
        Description: entry.tool.Description,
        InputSchema: entry.tool.InputSchema,
        Server:      entry.serverName,
    })
}
b.RegisterTools(brokerTools)
```

## 2. Intent Detection Feeding Tool Routing

The broker includes a keyword-based intent detector. Use it to infer the user's
intent from their message, then pass the top intent to `SelectTools`:

```go
// Detect intent from user message.
intents := broker.DetectIntent(userMessage)

intent := "*" // fallback: apply global rules only
if len(intents) > 0 {
    intent = intents[0].Name
}

// Select tools relevant to the detected intent.
result, err := b.SelectTools(ctx, intent, nil)
if err != nil {
    log.Printf("broker error: %v", err)
    // Fall back to all tools on error.
}

// result.Tools contains only the tools relevant to the intent.
// result.Rationale explains which rules were applied.
log.Printf("selected %d/%d tools: %s", result.Count, result.Total, result.Rationale)
```

### End-to-end flow

```
User message
    |
    v
DetectIntent(message)
    |
    v
[]{Intent{Name, Confidence, Keywords}}
    |
    v (top intent)
broker.SelectTools(ctx, intent, hints)
    |
    v
SelectResult{Tools, Count, Total, Rationale}
    |
    v
Inject selected tools into LLM system prompt
```

### How mentat-chat uses this today

In `internal/mcp/manager.go`, the `Manager` holds a `*broker.LocalBroker`:

```go
type Manager struct {
    servers map[string]Transport
    tools   []toolEntry
    Broker  *broker.LocalBroker // intent-aware tool broker
    mu      sync.RWMutex
}
```

Tool selection is done via `GetToolsForIntent`:

```go
func (m *Manager) GetToolsForIntent(intent string, hints []string) []provider.ToolDefinition {
    result, err := m.Broker.SelectTools(context.Background(), intent, hints)
    if err != nil {
        return m.getAllToolsLocked() // safe fallback
    }

    defs := make([]provider.ToolDefinition, 0, len(result.Tools))
    for _, t := range result.Tools {
        defs = append(defs, provider.ToolDefinition{
            Name:        fmt.Sprintf("mcp__%s__%s", t.Server, t.Name),
            Description: t.Description,
            InputSchema: t.InputSchema,
        })
    }
    return defs
}
```

## 3. Code Examples

### Basic usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/hollis-labs/tool-broker/broker"
)

func main() {
    // Create broker with default Fragments Engine rules.
    b := broker.NewLocalBroker(nil, broker.DefaultRules())

    // Register tools (normally from MCP server discovery).
    b.RegisterTools([]broker.ToolDefinition{
        {Name: "volon_tasks_list", Description: "List tasks", Server: "volon"},
        {Name: "volon_task_create", Description: "Create a task", Server: "volon"},
        {Name: "cortex_context_view", Description: "View context", Server: "cortex"},
        {Name: "hadron_bp_build_volon", Description: "Build volon", Server: "hadron"},
        {Name: "hadron_run_enqueue", Description: "Enqueue a run", Server: "hadron"},
    })

    // Detect intent and select tools.
    intents := broker.DetectIntent("create a new task for the backend")
    if len(intents) > 0 {
        fmt.Printf("Detected intent: %s (confidence: %.2f)\n", intents[0].Name, intents[0].Confidence)
    }

    intent := "*"
    if len(intents) > 0 {
        intent = intents[0].Name
    }

    result, _ := b.SelectTools(context.Background(), intent, nil)
    fmt.Printf("Selected %d/%d tools\n", result.Count, result.Total)
    for _, t := range result.Tools {
        fmt.Printf("  - %s (%s)\n", t.Name, t.Server)
    }
}
```

### Loading rules at startup

```go
func initBroker(configPath string) (*broker.LocalBroker, error) {
    var rules []Rule

    if configPath != "" {
        var err error
        rules, err = broker.LoadRulesFromFile(configPath)
        if err != nil {
            return nil, fmt.Errorf("load broker rules: %w", err)
        }
    } else {
        rules = broker.DefaultRules()
    }

    return broker.NewLocalBroker(nil, rules), nil
}
```

### Using hints for refinement

```go
// Hints provide additional context beyond the intent.
// They can be project names, capability tags, or keywords.
result, _ := b.SelectTools(ctx, "manage_tasks", []string{"volon", "sprint-12"})
```

## 4. Config Reference

### Rule structure (YAML)

```yaml
rules:
  - name: rule-name          # unique identifier (required)
    intent: intent_name       # which intent triggers this rule (required)
                              # use "*" for all intents
    priority: 20              # higher = evaluated first (default: 0)
    match:                    # criteria for selecting tools
      patterns:               # glob patterns on tool names (OR logic)
        - "volon_task_*"
        - "volon_tasks_list"
      tags:                   # match tools with any of these tags (OR logic)
        - "tasks"
        - "write"
      servers:                # match tools from these MCP servers (OR logic)
        - "volon"
    action:                   # what to do with matched tools
      type: include           # "include", "exclude", or "summarize"
```

### Match logic

- **Within a field** (patterns, tags, servers): OR logic. Any match is sufficient.
- **Across fields**: AND logic. All non-empty fields must match.
- **Empty fields**: not considered (match everything).

### Action types

| Type | Behavior |
|------|----------|
| `include` | Only matched tools are kept (whitelist). Multiple include rules are cumulative. |
| `exclude` | Matched tools are removed. Excludes always win over includes. |
| `summarize` | Tools are kept but marked for progressive disclosure (future). |

### Priority

Rules are evaluated in descending priority order. Higher priority rules are
applied first. When multiple rules apply:

1. Exclude rules remove tools from the set.
2. If any include rules fired, only explicitly included tools are kept.
3. Excluded tools are always removed, even if an include rule matched them.

### Supported file formats

| Extension | Format |
|-----------|--------|
| `.yaml`, `.yml` | YAML |
| `.json` | JSON |

Use `broker.LoadRulesFromFile(path)` — format is auto-detected by extension.

### Default rules

The broker ships with 13 default rules covering common Fragments Engine intents.
See `broker/default-rules.yaml` for the full list. Key defaults:

- **hide-blueprint-runners**: Excludes all `hadron_bp_*` tools (global).
- **create-task**: Routes to `volon_task_create`, `volon_tasks_list`, `volon_backlog_capture`.
- **search-context**: Routes to `cortex_*` tools.
- **run-blueprint**: Routes to `hadron_blueprints_list`, `hadron_run_*` tools.
- **check-health**: Routes to `hadron_health`, `volon_health`.

### Supported intents

| Intent | Description |
|--------|-------------|
| `create_task` | Creating new tasks or tickets |
| `list_tasks` | Viewing existing tasks |
| `manage_tasks` | Updating, transitioning, or deleting tasks |
| `plan_sprint` | Sprint planning and backlog grooming |
| `search_context` | Searching Cortex context records |
| `write_context` | Writing or updating context |
| `explore_project` | Getting project overview and status |
| `search_memory` | Recalling stored knowledge |
| `run_blueprint` | Executing Hadron blueprints |
| `run_build` | Running builds via Hadron |
| `check_health` | Checking service health |
| `schedule_automation` | Setting up scheduled or recurring runs |

## 5. Token Budget Management

The broker includes utilities for estimating tool token costs and pruning
tool sets to fit within a context window budget.

```go
tools := result.Tools

// Estimate total tokens consumed by tool definitions.
tokens := broker.EstimateToolTokens(tools)

// Prune to fit within 20% of a 200k token context window.
budget := int(0.20 * 200000) // 40,000 tokens
pruned := broker.PruneToTokenBudget(tools, budget)
// pruned always retains at least 1 tool.
```

## 6. Progressive Discovery (Keyword Scoring)

When a consumer has too many tools to send in one prompt, use progressive
discovery: send tool summaries first, then resolve full schemas on demand.

```go
// Find tools relevant to a natural-language intent.
matched := broker.ScoreByIntent(allTools, "I need to create a task", 10)

// Or use pre-tokenized keywords directly.
matched := broker.ScoreByKeywords(allTools, []string{"task", "create"}, 10)

// Find tools by exact name (for LLM tool requests).
found := broker.FindByNames(allTools, []string{"volon_task_create", "volon_tasks_list"})
```

### Progressive Discovery Pattern

```
1. Initial prompt:
   - SelectTools(intent) → small set of relevant tools
   - If > threshold: send only summaries + request_tools meta-tool

2. LLM requests specific tools:
   - FindByNames(allTools, requestedNames) → full schemas
   - ScoreByIntent(allTools, intentDescription, max) → ranked matches

3. Append newly-resolved tools to the active set for next turn.
```

### Building a request_tools Meta-Tool

Each consumer can expose a `request_tools` tool to the LLM. The broker
provides the scoring/lookup primitives; consumers define the tool schema:

```go
requestToolsDef := ToolDefinition{
    Name:        "request_tools",
    Description: "Request tools by name or intent description.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "tool_names": map[string]any{
                "type": "array", "items": map[string]any{"type": "string"},
                "description": "Specific tool names to load",
            },
            "intent": map[string]any{
                "type": "string",
                "description": "Describe what you want to do",
            },
        },
    },
}

// In the tool handler:
func handleRequestTools(allTools []broker.ToolDefinition, input map[string]any) []broker.ToolDefinition {
    var result []broker.ToolDefinition
    if names, ok := input["tool_names"].([]any); ok {
        var nameStrs []string
        for _, n := range names { nameStrs = append(nameStrs, n.(string)) }
        result = append(result, broker.FindByNames(allTools, nameStrs)...)
    }
    if intent, ok := input["intent"].(string); ok && intent != "" {
        result = append(result, broker.ScoreByIntent(allTools, intent, 10)...)
    }
    return result
}
```

## 7. Minimal Consumer Setup

A complete consumer needs just three things:

```go
// 1. Create broker with rules.
b := broker.NewLocalBroker(nil, broker.DefaultRules())

// 2. Register tools from MCP discovery.
b.RegisterTools(discoveredTools)

// 3. Select tools per turn.
intents := broker.DetectIntent(userMessage)
intent := "*"
if len(intents) > 0 {
    intent = intents[0].Name
}
result, _ := b.SelectTools(ctx, intent, nil)
tools := broker.PruneToTokenBudget(result.Tools, tokenBudget)
```

This is the universal pattern used by all Fragments Engine consumers.
Mentat's ToolClient and Volon's ToolClient both follow this pattern,
adding consumer-specific concerns (permissions, builtins, MCP routing)
on top.
