# Changelog

All notable changes to `go-toolbroker` are documented in this file.

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
