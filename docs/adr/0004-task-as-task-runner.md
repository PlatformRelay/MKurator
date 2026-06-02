# ADR-0004: Use Task as the task runner

- **Status**: Accepted
- **Date**: 2026-06-02

## Context

We need a single, discoverable entry point for build/test/lint/codegen/deploy
workflows that works identically for humans, pre-commit hooks, and CI. The two
common choices are **Make** (used by cert-manager) and **Task** (used by rko).
cert-manager's Makefile is powerful but heavily modularised (klone-synced
`make/_shared` modules, self-upgrade bots) — more machinery than this project
warrants.

## Decision

We will use **[Task](https://taskfile.dev)** with two files:
`Taskfile.yml` (build/deploy) and `Taskfile.test.yml` (tests). Go-based tools
are pinned via `go.mod` `tool` directives and invoked with `go tool`. Local
cluster provisioning lives under `hack/kind-cluster` and is wrapped by
`cluster:*` tasks.

## Consequences

- One vocabulary (`task <name>`) for everyone; `task --list` is self-documenting.
- CI parity is trivial: each CI job calls the same `task` target a developer
  runs locally (see [../CICD.md](../CICD.md)).
- YAML task definitions with `deps`, `sources`/`generates` caching, and includes
  are readable and avoid Make's tab/escaping sharp edges.
- We forgo Make's ubiquity (Task must be installed), accepted as a documented
  prerequisite.

## Alternatives considered

- **Make**: ubiquitous and what cert-manager uses, but its quoting/tab rules and
  the temptation toward heavy module systems add friction for a small project.
- **Shell scripts only**: no dependency graph, no caching, poor discoverability.
  We still use small scripts under `hack/`, wrapped by Task.
