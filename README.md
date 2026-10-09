# go-toolbroker

## Moved to substrate

This standalone repository is deprecated. New development lives in the
[`github.com/hollis-labs/substrate/agent`](https://github.com/hollis-labs/substrate/tree/agent/v0.2.0/agent)
module, released as **`agent/v0.2.0`**.

```sh
go get github.com/hollis-labs/substrate/agent@v0.2.0
```

Follow the [package and API migration guide](https://github.com/hollis-labs/substrate/blob/agent/v0.2.0/agent/toolbroker/MIGRATION.md) when updating imports;
the consolidation can include API changes. Existing standalone tags and history
are preserved. The documentation below describes the standalone releases and
is retained for historical reference. Applications migrate separately; this
redirect does not deploy or update any consumer.

[![Go Reference](https://pkg.go.dev/badge/github.com/hollis-labs/go-toolbroker.svg)](https://pkg.go.dev/github.com/hollis-labs/go-toolbroker)

`go-toolbroker` is an in-process Go library for **intent-aware MCP tool selection**. Given a registry of MCP tool definitions and a priority-ordered set of rules, it selects the subset of tools relevant to a user's detected intent — replacing hardcoded exclude lists with a flexible rule engine. It also ships keyword-based intent detection, token-budget estimation, progressive-discovery scoring helpers, and an optional per-tool enrichment pipeline (`Hints` → markdown override block).

Consumers (chat clients, agent runtimes, anything that talks to many MCP servers in one process) use it to keep per-turn tool payloads small and on-topic.

## Status

Pre-1.0. The public API is stable enough to be embedded in production code; expect minor additive changes between v0.x releases. A future remote/HTTP broker implementation is on the roadmap; only `LocalBroker` ships today.

## Install

```bash
go get github.com/hollis-labs/go-toolbroker
```

```go
import "github.com/hollis-labs/go-toolbroker/broker"
```

## Quickstart

```go
package main

import (
    "context"
    "fmt"

    "github.com/hollis-labs/go-toolbroker/broker"
)

func main() {
    // 1. Define a tiny rule set (or load from a file with broker.LoadRulesFromFile).
    rules := []broker.Rule{
        {
            Name:     "hide-noisy-build-tools",
            Intent:   "*",
            Priority: 10,
            Match:    broker.Match{Patterns: []string{"build_*"}},
            Action:   broker.Action{Type: "exclude"},
        },
        {
            Name:     "show-build-tools-on-build-intent",
            Intent:   "run_build",
            Priority: 20,
            Match:    broker.Match{Patterns: []string{"build_*"}},
            Action:   broker.Action{Type: "include"},
        },
    }

    // 2. Create a broker and register your discovered MCP tools.
    b := broker.NewLocalBroker(nil, rules)
    b.RegisterTools([]broker.ToolDefinition{
        {Name: "tasks_list", Description: "List tasks", Server: "tasks"},
        {Name: "tasks_create", Description: "Create a task", Server: "tasks"},
        {Name: "build_go_project", Description: "Build a Go project", Server: "ci"},
    })

    // 3. Detect intent from the user's message and ask for relevant tools.
    intents := broker.DetectIntent("create a new task for the backend")
    intent := "*"
    if len(intents) > 0 {
        intent = intents[0].Name
    }
    result, err := b.SelectTools(context.Background(), intent, nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("selected %d/%d tools: %s\n", result.Count, result.Total, result.Rationale)

    // 4. (Optional) Prune further to fit a token budget.
    pruned := broker.PruneToTokenBudget(result.Tools, 40000)
    _ = pruned
}
```

For runnable end-to-end programs, see the [`examples/`](examples/) directory:

- `examples/basic` — minimal selection with hand-written rules.
- `examples/yaml-rules` — load rules from a YAML file at runtime.
- `examples/enrichment` — wire an `Enricher` so `SelectResult.OverrideBlock` is populated with per-tool hints.

## API Overview

All exports live in the single `broker` package.

### Core broker (`broker.go`, `local.go`)

- `Broker` interface — `SelectTools(ctx, intent, hints)` and `AllTools()`.
- `LocalBroker` — rule-based, in-process implementation of `Broker`.
- `NewLocalBroker(tools, rules, opts...) *LocalBroker`
- `(*LocalBroker).RegisterTools`, `LoadRules`, `AllTools`, `SelectTools`
- Data types: `ToolDefinition`, `ToolSummary`, `SelectResult`.

### Rules and configuration (`rule.go`, `config.go`)

- `Rule`, `Match`, `Action` — describe priority-ordered selection rules.
- `Config` — top-level YAML/JSON shape (`{rules: [...]}`).
- `LoadConfig(path)` — parse a JSON config file.
- `LoadRulesFromFile(path)` — parse YAML or JSON rules; format auto-detected by extension.
- `DefaultRules()` — returns example rules embedded from `broker/default-rules.yaml`. The bundled defaults reference a specific MCP toolset and are intended as a worked example of the rule format, not as production defaults — most consumers will write their own rules.

### Intent detection (`intent.go`)

- `Intent{Name, Confidence, Keywords}`
- `DetectIntent(message) []Intent` — keyword-based (no ML), ranked by confidence.

### Progressive discovery / scoring (`discovery.go`)

- `ScoreByKeywords(tools, keywords, maxResults)`
- `ScoreByIntent(tools, intent, maxResults)`
- `FindByNames(tools, names)`
- `TokenizeIntent(intent)`
- `MinKeywordScore` constant.

### Token budget (`budget.go`)

- `EstimateToolTokens(tools) int`
- `PruneToTokenBudget(tools, budgetTokens) []ToolDefinition`
- `DefaultTokenBudgetPct`, `DefaultContextWindowTokens` constants.

### Enrichment (`hints.go`, `enricher.go`, `override.go`)

- `Hints{Preconditions, AntiPatterns, ChainsWith, OutputShape}` — per-tool metadata that shapes the agent's use of a tool without changing the tool's schema.
- `MarshalHints` / `UnmarshalHints` — JSON codec for storage in a consumer-owned table.
- `Enricher` interface — `LookupByToolName(ctx, name) (Hints, bool, error)`. Consumers back this with their own storage (SQLite, in-memory, remote service, ...). A `NopEnricher` is provided as a safe default.
- `ComposeOverrideBlock(ctx, toolNames, enr) (string, error)` — produces a compact markdown `## Tool Overrides` section summarizing hints for the named tools. Tools without hints are skipped.
- `WithEnricher(Enricher) Option` — functional option on `NewLocalBroker`. When set, `SelectTools` populates `SelectResult.OverrideBlock` (markdown ready to append to the per-turn system prompt) from the selected tools' hints.

Storage for enrichment records is **consumer-owned** — this package ships only the interface, composition, and the no-op default. A consumer that already persists per-tool metadata can satisfy `Enricher` with a single method.

## Logging

`go-toolbroker` uses the standard library `log/slog` package on the default handler. Only the enrichment compose-error path logs today (a `slog.WarnContext` at `"toolbroker: compose override block failed"`); selection never logs on the happy path. Enrichment failures are logged and then swallowed so selection remains successful — logging is observation, not control flow.

Callers who want structured logging should configure their own handler via `slog.SetDefault(slog.New(...))` at process start; the broker will honor it automatically.

## Selection algorithm

`LocalBroker.SelectTools` (see `broker/local.go`):

1. Sort rules by `Priority` descending.
2. Filter to rules whose `Intent` glob (via `path.Match`, with `""`/`"*"` = all) matches the incoming intent.
3. Start with the full tool set. For each applicable rule, apply `exclude` (blacklist), `include` (whitelist, cumulative), or `summarize` (currently behaves like `include` — reserved for future progressive disclosure).
4. If any include rule fired, only explicitly included tools are kept. Exclusions always win over includes.
5. Tool-name patterns support both `path.Match` glob semantics and a `HasPrefix` fallback for patterns ending in `*` (so `build_*` matches `build_go_project`).

`LocalBroker` uses an internal `sync.RWMutex`, so registration, rule loading, and tool selection can be called concurrently from multiple goroutines.

For deeper integration patterns — including intent detection feeding tool routing, token budgeting, and a `request_tools` progressive-discovery meta-tool — see [`docs/integration-guide.md`](docs/integration-guide.md).

## Dependencies

Only one external dependency:

- `gopkg.in/yaml.v3` — YAML parsing for rule files and the embedded default-rules.

Everything else is the Go standard library (`context`, `encoding/json`, `embed`, `os`, `path`, `sort`, `strings`, `log/slog`, `sync`).

## Testing

```bash
go test ./...
go test -race ./...
```

No external services, fixtures, or environment variables are required. Tests use `t.TempDir()` for any file I/O and cover broker selection, rule loading (JSON/YAML/YML), default rules, intent detection, scoring, token budgeting, and enrichment composition.

## License

MIT — see [LICENSE](LICENSE).
