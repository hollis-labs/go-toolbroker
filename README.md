# Tool Broker

`go-toolbroker` is an in-process Go library for **intent-aware MCP tool selection**. Given a registry of MCP tool definitions and a set of configurable rules, it selects the subset of tools relevant to a user's detected intent — replacing hardcoded exclude lists with a flexible, priority-ordered rule engine. It also ships keyword-based intent detection, token-budget estimation, and progressive-discovery scoring helpers. Consumers (e.g. chat clients embedding many MCP servers) use it to keep per-turn tool payloads small and on-topic.

## Status

Beta. The public API is stable enough to be embedded in consumer apps (the package godoc and integration guide describe mentat-chat consuming it via `*broker.LocalBroker`), the package comment flags a second "remote broker service" implementation as planned/future, and the package has broad unit-test coverage across every source file. Module path `github.com/hollis-labs/tool-broker` is the canonical standalone repo path and still suggests pre-v1.

## Install

```bash
go get github.com/hollis-labs/tool-broker
```

Then import the broker package:

```go
import "github.com/hollis-labs/tool-broker/broker"
```

## Usage

Minimal end-to-end example (adapted from `docs/integration-guide.md` and `broker/broker_test.go`):

```go
package main

import (
    "context"
    "fmt"

    "github.com/hollis-labs/tool-broker/broker"
)

func main() {
    // 1. Create a broker with the built-in default rules.
    b := broker.NewLocalBroker(nil, broker.DefaultRules())

    // 2. Register MCP tools (normally discovered from live servers).
    b.RegisterTools([]broker.ToolDefinition{
        {Name: "volon_tasks_list", Description: "List tasks", Server: "volon"},
        {Name: "volon_task_create", Description: "Create a task", Server: "volon"},
        {Name: "cortex_context_view", Description: "View context", Server: "cortex"},
        {Name: "hadron_bp_build_volon", Description: "Build volon", Server: "hadron"},
    })

    // 3. Detect intent from the user's message.
    intents := broker.DetectIntent("create a new task for the backend")
    intent := "*"
    if len(intents) > 0 {
        intent = intents[0].Name
    }

    // 4. Ask the broker for the relevant tools.
    result, err := b.SelectTools(context.Background(), intent, nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("selected %d/%d tools: %s\n", result.Count, result.Total, result.Rationale)

    // 5. (Optional) Prune further to fit a token budget.
    pruned := broker.PruneToTokenBudget(result.Tools, 40000)
    _ = pruned
}
```

Rules may also be loaded from disk:

```go
rules, err := broker.LoadRulesFromFile("my-rules.yaml") // .yaml, .yml, or .json
if err != nil { /* handle */ }
b := broker.NewLocalBroker(nil, rules)
```

See `docs/integration-guide.md` for extended examples (custom rules, hints, progressive discovery, and building a `request_tools` meta-tool).

## API Overview

All exports live in the single `broker` package:

**Core broker (`broker/broker.go`, `broker/local.go`)**
- `Broker` interface — `SelectTools(ctx, intent, hints)` and `AllTools()`.
- `LocalBroker` — rule-based, in-process implementation of `Broker`.
- `NewLocalBroker(tools, rules) *LocalBroker`
- `(*LocalBroker).RegisterTools`, `LoadRules`, `AllTools`, `SelectTools`
- Data types: `ToolDefinition`, `ToolSummary`, `SelectResult`.

**Rules and configuration (`broker/rule.go`, `broker/config.go`)**
- `Rule`, `Match`, `Action` — describe priority-ordered selection rules.
- `Config` — top-level YAML/JSON shape (`{rules: [...]}`).
- `LoadConfig(path)` — parse a JSON config file.
- `LoadRulesFromFile(path)` — parse YAML or JSON rules, format auto-detected by extension.
- `DefaultRules()` — returns sensible defaults embedded from `broker/default-rules.yaml` (hides `hadron_bp_*` globally and adds per-intent include rules for the Fragments Engine ecosystem).

**Intent detection (`broker/intent.go`)**
- `Intent{Name, Confidence, Keywords}`
- `DetectIntent(message) []Intent` — keyword-based (no ML), ranked by confidence.

**Progressive discovery / scoring (`broker/discovery.go`)**
- `ScoreByKeywords(tools, keywords, maxResults)`
- `ScoreByIntent(tools, intent, maxResults)`
- `FindByNames(tools, names)`
- `TokenizeIntent(intent)`
- `MinKeywordScore` constant.

**Token budget (`broker/budget.go`)**
- `EstimateToolTokens(tools) int`
- `PruneToTokenBudget(tools, budgetTokens) []ToolDefinition`
- `DefaultTokenBudgetPct`, `DefaultContextWindowTokens` constants.

## Architecture Notes

The broker has one concrete implementation today: `LocalBroker`. Its `SelectTools` algorithm (see `broker/local.go`) is:

1. Sort rules by `Priority` descending.
2. Filter to rules whose `Intent` glob (via `path.Match`, with `""`/`"*"` = all) matches the incoming intent.
3. Start with the full tool set. For each applicable rule, apply `exclude` (blacklist), `include` (whitelist, cumulative), or `summarize` (currently behaves like include — reserved for future progressive disclosure).
4. If any include rule fired, only explicitly included tools are kept. Exclusions always win over includes.
5. Tool-name patterns support both `path.Match` glob semantics and a `HasPrefix` fallback for patterns ending in `*` (so `hadron_bp_*` matches `hadron_bp_build_volon`).

Default rules are compiled into the binary via `go:embed` of `broker/default-rules.yaml`, so `DefaultRules()` has zero runtime file dependencies. A remote HTTP-based broker implementation is mentioned as future work in the package godoc but is not yet present.

`LocalBroker` uses an internal `sync.RWMutex`, so registration, rule loading, and tool selection can be called concurrently from multiple goroutines.

For integration patterns — including intent detection feeding tool routing, token budgeting, and a `request_tools` progressive-discovery meta-tool — see `docs/integration-guide.md`.

## Dependencies

**Framework-internal:** none. This lib has no dependencies on other libs in this repo.

**External (from `go.mod`):**
- `gopkg.in/yaml.v3 v3.0.1` — YAML parsing for rule files and the embedded default-rules.

Everything else is Go standard library (`context`, `encoding/json`, `embed`, `os`, `path`, `sort`, `strings`, etc.).

## Testing

```bash
go test ./...
```

No external services, fixtures, or environment variables are required. Tests use `t.TempDir()` for any file I/O and cover broker selection, rule loading (JSON/YAML/YML), default rules, intent detection, scoring, and token budgeting.

## License

MIT License.
