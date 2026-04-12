# Audit — go-toolbroker

**Audited:** 2026-04-09
**Auditor:** general-purpose subagent (BOOT_STANDARDIZATION audit)
**Path:** libs/go-toolbroker
**Kind:** lib

## Summary

The library is well-structured and test-covered: a single `broker` package provides intent-aware MCP tool selection, rule loading, token budgeting, and progressive-discovery helpers, with a thoughtful package-level godoc comment and a thorough `docs/integration-guide.md`. The apply session added the missing MIT `LICENSE`, a Go example test, and internal concurrency safety for `LocalBroker`; the module path is now treated as the standalone repo's canonical import path per D2. The remaining unresolved item is the repo-root agent/session files, which were intentionally left untouched.

## Checklist

| # | Check | Status | Notes |
|---|---|---|---|
| 1 | `go.mod` present | pass | `github.com/hollis-labs/tool-broker`, Go 1.25.0 |
| 2 | `README.md` present (before this audit) | fail | Missing. Created by this audit. No rename needed. |
| 3 | `LICENSE` present | pass | MIT `LICENSE` added in this apply session. |
| 4 | `doc.go` with `// Package X ...` godoc comment | partial | No `doc.go`, but `broker/broker.go` has a complete `// Package broker ...` comment describing purpose and future plans. |
| 5 | Module path matches intended repo layout | pass | `github.com/hollis-labs/tool-broker` is treated as the standalone repo's canonical import path per D2; README aligns with that decision. |
| 6 | README has standard sections (title, desc, install, usage, API, examples) | pass | Written by this audit per BOOT_STANDARDIZATION template. |
| 7 | Tests exist (`*_test.go`) | pass | 4 test files: `broker_test.go`, `budget_test.go`, `config_test.go`, `discovery_test.go`, `intent_test.go` (all under `broker/`). Broad coverage of selection, rules, intent detection, scoring, budgeting. |
| 8 | Examples (`example_test.go` or `examples/`) | pass | Added `broker/example_test.go` with `ExampleLocalBroker_SelectTools`. |
| 9 | State/session files NOT misclassified as library docs | pass | `.agentrc/`, `agentrc.yaml`, `CLAUDE.md` present but excluded from README per BOOT_STANDARDIZATION rule 4. |
| 10 | Public API sanity: errors typed/sentinel, context.Context first arg | partial | `SelectTools(ctx context.Context, ...)` takes context first — good. However, no exported sentinel errors or typed error values anywhere in the package; all errors are `fmt.Errorf`-wrapped. Some functions that could benefit from context (e.g. `LoadConfig`, `LoadRulesFromFile`) do not accept one. `LocalBroker` now has an internal `sync.RWMutex`, so the concurrency concern is addressed. |
| 11 | `CHANGELOG.md` present (nice to have) | fail | Missing. |
| 12 | No circular/suspicious deps on other framework libs | pass | Zero framework-internal deps. Only external dep is `gopkg.in/yaml.v3`. |

## Findings — Required Fixes

1. **Add a LICENSE file.**
   - **What:** No LICENSE exists at `libs/go-toolbroker/`.
   - **Why:** A library published under `github.com/hollis-labs/...` without a license is legally unusable by downstream consumers. Blocks any public/external release.
   - **Suggested fix:** Add the framework's standard LICENSE. Completed in this apply session with MIT text.

2. **Confirm or correct the module path.**
   - **What:** `go.mod` declares `module github.com/hollis-labs/tool-broker`, which does not match the framework monorepo layout (`libs/go-toolbroker` under the `framework` repo). Other audited libs in this monorepo should be checked for consistency.
   - **Why:** Inconsistent module paths break `go get` and cross-lib imports within the monorepo; if the intended home is the framework repo, this needs to be updated before release.
   - **Suggested fix:** D2 confirms this is a standalone repo path, so the README and audit now treat `github.com/hollis-labs/tool-broker` as canonical. No `go.mod` rewrite was required.

3. **Remove or re-scope state/session files at the lib root.**
   - **What:** `CLAUDE.md`, `agentrc.yaml`, and the `.agentrc/` directory (with `agent-boot.md`, `bootstrap.md`, `boot/worker.md`, `logs/`, `pcc/`, `tasks/`) live at the library root alongside source. These are agent session state, not library artifacts.
   - **Why:** They will ship inside any published Go module (there is no `.gitignore` exclusion visible), confusing library consumers and inflating the module. BOOT_STANDARDIZATION rule 4 explicitly excludes them from library docs.
   - **Suggested fix:** Defer: relocating repo-local agent state would be a repo-structure decision outside this apply pass. Leave untouched per the session boundary.

4. **Add basic concurrency safety to `LocalBroker`.**
   - **What:** `LocalBroker.tools` and `LocalBroker.rules` are mutated by `RegisterTools` and `LoadRules` but read by `SelectTools` and `AllTools` with no mutex. The integration guide shows it embedded in a `Manager` that uses `sync.RWMutex` externally, but the type itself offers no guarantees.
   - **Why:** Libraries that advertise "in-process, embeddable" should document and ideally enforce their goroutine-safety contract.
   - **Suggested fix:** Completed in this apply session by adding an internal `sync.RWMutex` and documenting concurrent use in the README.

5. **Add an `example_test.go` with the canonical usage pattern.**
   - **What:** No Go-native example tests exist; examples are only in `docs/integration-guide.md` as prose.
   - **Why:** `Example*` functions show up in `go doc` output and pkg.go.dev, and they are compiled on every test run, so they can never go stale.
   - **Suggested fix:** Completed in this apply session with `broker/example_test.go` and `ExampleLocalBroker_SelectTools`.

## Findings — Nice-to-Have

1. **Add a top-level `doc.go`.**
   - **What:** The package comment is attached to `broker/broker.go`. Convention is a dedicated `doc.go` file that holds only the package comment.
   - **Why:** Easier to find, less churn when editing `broker.go`.
   - **Suggested fix:** Move the `// Package broker ...` block to `broker/doc.go`.

2. **Add a `CHANGELOG.md`.**
   - **What:** No version history file.
   - **Why:** Helps consumers track API changes as the library moves toward v1 and the planned remote-broker implementation lands.
   - **Suggested fix:** Seed with a "Unreleased" section and follow Keep-a-Changelog format.

3. **Implement or remove the `summarize` action.**
   - **What:** `local.go` notes that `summarize` "currently behaves like include" and is "reserved for future progressive disclosure." Default rules do not use it.
   - **Why:** Dead/aspirational behavior in the public `Action.Type` surface invites consumer confusion.
   - **Suggested fix:** Either implement the progressive-disclosure semantics and wire them through `SelectResult`, or remove the type from the documented action list and reject it in rule parsing.

4. **Define exported error sentinels.**
   - **What:** All errors are `fmt.Errorf` wraps with no sentinel like `ErrUnsupportedFormat`, `ErrRuleParse`, etc.
   - **Why:** Consumers cannot use `errors.Is` to branch on specific failure modes (e.g. "unsupported config format" vs "file not found").
   - **Suggested fix:** Introduce `var Err... = errors.New(...)` in `config.go` and wrap with `%w`.

5. **Add CI lint coverage visibility.**
   - **What:** `.golangci.yml` is present at lib root but there is no README badge or mention of lint/CI status.
   - **Why:** Signals maintained-ness to downstream consumers.
   - **Suggested fix:** Once CI is set up, add status badges to README (out of scope for this audit — noted for the apply session).

## Prior Documentation

- **`README.md`:** did not exist before this audit. No rename needed; the new README was written directly.
- **`docs/integration-guide.md`:** preserved as-is. This is real architecture/usage documentation and is referenced from the new README.
- **Session/state files (excluded from library docs per BOOT_STANDARDIZATION rule 4):**
  - `CLAUDE.md` — agent boot/profile override and a short project overview.
  - `agentrc.yaml` — project metadata for the agent harness.
  - `.agentrc/agent-boot.md`, `.agentrc/bootstrap.md`, `.agentrc/boot/worker.md` — boot/worker profile instructions.
  - `.agentrc/logs/`, `.agentrc/pcc/`, `.agentrc/tasks/` — empty at audit time; session state directories.
  - `.golangci.yml`, `lefthook.yml` — tooling configs (not library docs; noted here only for completeness).

## Public API Snapshot

All in package `broker` (module `github.com/hollis-labs/tool-broker`).

**`broker/broker.go`**
- interface `Broker` — `SelectTools(ctx, intent, hints) (*SelectResult, error)`, `AllTools() []ToolSummary`
- struct `ToolDefinition{Name, Description, InputSchema, Server, Tags, CostTier}`
- struct `ToolSummary{Name, Description, Server, Tags}`
- struct `SelectResult{Tools, Count, Total, Intent, Rationale}`

**`broker/local.go`**
- struct `LocalBroker` (implements `Broker`)
- func `NewLocalBroker(tools []ToolDefinition, rules []Rule) *LocalBroker`
- method `(*LocalBroker).RegisterTools(tools []ToolDefinition)`
- method `(*LocalBroker).LoadRules(rules []Rule)`
- method `(*LocalBroker).AllTools() []ToolSummary`
- method `(*LocalBroker).SelectTools(ctx context.Context, intent string, hints []string) (*SelectResult, error)`

**`broker/rule.go`**
- struct `Rule{Name, Intent, Match, Action, Priority}`
- struct `Match{Patterns, Tags, Servers}`
- struct `Action{Type}` — values: `"include"`, `"exclude"`, `"summarize"`

**`broker/config.go`**
- struct `Config{Rules []Rule}`
- func `LoadConfig(path string) (*Config, error)` — JSON only
- func `LoadRulesFromFile(path string) ([]Rule, error)` — auto-detects `.yaml`/`.yml`/`.json`
- func `DefaultRules() []Rule` — loads embedded `default-rules.yaml`

**`broker/intent.go`**
- struct `Intent{Name, Confidence, Keywords}`
- func `DetectIntent(message string) []Intent`

**`broker/discovery.go`**
- const `MinKeywordScore = 1`
- func `ScoreByKeywords(tools []ToolDefinition, keywords []string, maxResults int) []ToolDefinition`
- func `ScoreByIntent(tools []ToolDefinition, intent string, maxResults int) []ToolDefinition`
- func `FindByNames(tools []ToolDefinition, names []string) []ToolDefinition`
- func `TokenizeIntent(intent string) []string`

**`broker/budget.go`**
- const `DefaultTokenBudgetPct = 0.20`
- const `DefaultContextWindowTokens = 200000`
- func `EstimateToolTokens(tools []ToolDefinition) int`
- func `PruneToTokenBudget(tools []ToolDefinition, budgetTokens int) []ToolDefinition`

**Embedded asset**
- `broker/default-rules.yaml` — embedded via `//go:embed`; source of `DefaultRules()`. Contains ~30 rules covering blueprint-runner exclusion and intent-specific inclusion for the Fragments Engine ecosystem.

## Open Questions

1. **Intended module path?** Is `github.com/hollis-labs/tool-broker` the canonical import path, or should it be re-homed under the framework monorepo's module namespace? This affects both the `go.mod` fix and any cross-lib imports.
2. **Release status?** The package godoc calls out a planned "remote broker service" implementation. Is the current version `0.x` (beta) or already considered stable for internal use? README currently says "beta" based on evidence, but human confirmation would be ideal.
3. **Should `summarize` be removed or implemented?** The action type is public and accepted by rule parsing but behaves identically to `include` today.
4. **License preference?** MIT, Apache-2.0, or other? The rest of the framework monorepo should set precedent.
5. **Concurrency contract?** Is `LocalBroker` expected to be goroutine-safe out of the box, or does it delegate that responsibility to callers (as mentat-chat's `Manager` appears to do via its own `sync.RWMutex`)?
