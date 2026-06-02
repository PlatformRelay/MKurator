# Architecture Decision Records

This directory holds **Architecture Decision Records (ADRs)**: short documents
that capture a significant decision, its context, and its consequences. They are
the durable record of *why* the project looks the way it does.

## When to write one

Write an ADR when a decision is non-obvious, hard to reverse, or likely to be
questioned later — e.g. choice of framework, an external protocol, an API
shape, or a tooling trade-off. Don't write one for routine, easily reversible
choices.

## How to add one

1. Copy [`0000-template.md`](0000-template.md) to the next number, e.g.
   `0005-short-title.md`.
2. Fill in Context, Decision, Consequences, and Alternatives.
3. Set the status (`Proposed` → `Accepted` → optionally `Superseded by ADR-NNNN`).
4. Add it to the index below and link it from the relevant doc/code.

ADRs are immutable once Accepted: to change a decision, write a new ADR that
supersedes the old one rather than editing history.

## Index

| ADR | Title | Status |
|-----|-------|--------|
| [0001](0001-use-kubebuilder-controller-runtime.md) | Use Kubebuilder + controller-runtime | Accepted |
| [0002](0002-manage-mq-via-mqweb-rest.md) | Manage MQ via the mqweb REST API behind an MQAdmin port | Accepted |
| [0003](0003-connection-model.md) | Decouple connection details with QueueManagerConnection | Accepted |
| [0004](0004-task-as-task-runner.md) | Use Task as the task runner | Accepted |
| [0005](0005-keep-tooling-lean.md) | Keep tooling lean; borrow discipline, not org overhead | Accepted |

> The module path / API group decision (`github.com/konradheimel/ibm-mq-operator`,
> `messaging.heimel.dev/v1alpha1`) is still a placeholder; record it as an ADR
> once fixed at scaffold time.
