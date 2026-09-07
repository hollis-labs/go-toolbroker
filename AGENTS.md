# go-toolbroker

An in-process library for intent-aware MCP tool selection: given a registry of
tool definitions and a priority-ordered rule set, it picks the subset relevant
to a detected intent, so a client talking to many MCP servers can keep per-turn
tool payloads small. It also ships keyword intent detection, token estimation,
discovery scoring and an optional enrichment pipeline. It selects tools; it
never calls one, and it speaks no MCP transport.

## Start Here

- `README.md` has the quickstart; `docs/integration-guide.md` is the fuller
  consumer walkthrough.
- `broker/broker.go` declares the `Broker` interface; `broker/local.go` is the
  only shipped implementation.
- `broker/rule.go` and `broker/config.go` own rule shape and YAML/JSON loading.
- `broker/intent.go` owns keyword intent detection and confidence.
- `broker/budget.go` owns token estimation and budget pruning.
- `broker/enricher.go`, `broker/hints.go` and `broker/override.go` own the
  enrichment pipeline and the markdown override block.
- `examples/basic`, `examples/yaml-rules` and `examples/enrichment` are
  runnable.

## Commands

```bash
gofmt -l .
go vet ./...
go test -race -count=1 ./...
golangci-lint run
```

There is no CI workflow in this repo, so these are the only gate.

## Boundaries

Only `LocalBroker` exists. A remote/HTTP broker is on the roadmap but is not
implemented — do not write code that assumes the `Broker` interface is already
satisfied over a network.

Rules are priority-ordered and evaluated as a set: an exclude and an include
can both match, and the combined outcome is what
`TestSelectToolsCombinedExcludeAndInclude` pins. Treating the first match as
final would change selection for every consumer with layered rules.

The enrichment pipeline must degrade to nothing rather than to something
wrong. A nil enricher, a nil context, partial enrichment, an empty tool list
and blank entries each have their own test, and each must produce a clean
empty or partial override block rather than a malformed one — the block is
injected into a prompt, so malformed output is worse than absent output.

Token estimation and budget pruning decide what a model sees. Pruning drops
tools to fit a budget, so a change to the estimate silently changes which tools
are available at the same budget.

Rule files load from YAML, YML and JSON, and an unknown extension is an error
rather than a guess (`TestLoadRulesFromFileUnknownExtension`).
