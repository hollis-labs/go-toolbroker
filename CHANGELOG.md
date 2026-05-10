# Changelog

All notable changes to `go-toolbroker` are documented in this file. The
format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and
this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
(pre-1.0: minor bumps may include behavior-shaping doc/example changes that
do not break the public Go API).

## v0.2.0 — 2026-05-10

### Added
- `examples/` directory with three runnable end-to-end programs:
  - `examples/basic` — minimal selection with hand-written rules.
  - `examples/yaml-rules` — load rules from a YAML file at runtime.
  - `examples/enrichment` — wire an `Enricher` so `SelectResult.OverrideBlock`
    is populated with per-tool hints.
- `broker/doc.go` containing the canonical package godoc comment.
- `.gitignore` covering Go build artifacts and agent/session files.
- `README.md` badge for [pkg.go.dev](https://pkg.go.dev/github.com/hollis-labs/go-toolbroker).

### Changed
- README rewritten for a public audience: removed internal-application
  framing, added a self-contained quickstart, linked `examples/`, clarified
  the role of `DefaultRules()` as a worked example rather than a production
  default.
- `docs/integration-guide.md` updated to use the canonical
  `github.com/hollis-labs/go-toolbroker/broker` import path and reframed
  away from a single embedder's perspective.
- Package-level godoc moved from `broker/broker.go` to a dedicated
  `broker/doc.go` and expanded to summarize the full public surface
  (selection, intent detection, scoring, budgeting, enrichment).
- `broker/default-rules.yaml` header reframed: the bundled rules are now
  documented as an example aligned to a specific MCP toolset, not as
  defaults that are useful in arbitrary consumer environments.
- Inline godoc on `LocalBroker`, `DefaultRules`, and intent-keyword
  comments scrubbed of internal-application names.

### Removed
- Repo-internal agent/session artifacts: `.agentrc/`, `agentrc.yaml`,
  `CLAUDE.md`, `AUDIT_RESULTS.md`, `lefthook.yml`. These never belonged in
  a public module's tree; consumers see no API change.

### Public API
- No breaking changes to the Go API. The contents of the rule set returned
  by `DefaultRules()` are unchanged in this release; they are now clearly
  documented as illustrative.

## v0.1.0 — 2026-04-19

### Added
- `Hints` struct + JSON codec for per-tool metadata (`Preconditions`, `AntiPatterns`, `ChainsWith`, `OutputShape`).
- `Enricher` interface + `NopEnricher` no-op default.
- `ComposeOverrideBlock` — produces a compact markdown "## Tool Overrides" section from selected tools' Hints.
- `WithEnricher(Enricher)` functional option on `NewLocalBroker`.
- `SelectResult.OverrideBlock string` — populated when an enricher is wired.
- `log/slog` logging on enrichment compose errors (default handler; non-fatal — selection continues with an empty `OverrideBlock`).

### Changed
- `NewLocalBroker` gains a `...Option` variadic parameter (backward-compatible).
- README import examples updated to `github.com/hollis-labs/go-toolbroker` (matches current module path).

### Consumer notes
- Storage for enrichment records (e.g., SQLite `tool_enrichments` table) stays consumer-owned. Implement `Enricher` against your own store and pass it via `WithEnricher`.
- Callers wanting structured logging should configure their own `slog` handler via `slog.SetDefault` at process start.
