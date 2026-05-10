// Package broker provides intent-aware MCP tool selection.
//
// A Broker selects a relevant subset of MCP tools from a registry based on a
// detected intent, optional caller hints, and a priority-ordered set of rules.
// This replaces hardcoded exclude-pattern filtering with a flexible,
// rule-driven approach so consumers (chat clients, agent runtimes, etc.) can
// keep per-turn tool payloads small and on-topic.
//
// Two implementations are envisioned:
//
//   - LocalBroker: in-process, config-driven (this package).
//   - Remote broker service: HTTP-based, cross-app (future work).
//
// In addition to selection, the package ships keyword-based intent detection
// (DetectIntent), token-budget estimation and pruning (EstimateToolTokens,
// PruneToTokenBudget), progressive-discovery scoring helpers
// (ScoreByKeywords, ScoreByIntent), and an optional enrichment pipeline
// (Enricher, ComposeOverrideBlock) that lets callers attach per-tool Hints
// (preconditions, anti-patterns, output shape) which are surfaced as a
// markdown override block in SelectResult.
//
// The package has no dependencies on other libraries in this repo and only
// one external dependency (gopkg.in/yaml.v3) for parsing rule files.
package broker
