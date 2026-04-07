# Tool Broker — agent boot

## Agent Auto-Boot Override

This repo boots as `worker`. **Skip profile selection** — go directly to:

4. Read `.agentrc/agent-boot.md`
5. Read `.agentrc/boot/worker.md`
6. Read `.agentrc/bootstrap.md`
7. Follow the worker profile instructions — emit boot confirmation and begin work

## Project Overview

Tool Broker is a Go service that routes and manages tool calls across the Fragments Engine ecosystem. It provides configurable routing rules, local tool execution, and broker-level configuration management.

## Build & Test

```bash
go build ./...
go test ./...
```

## Architecture

- `broker/` — Core broker logic: routing, configuration, local execution, rules
