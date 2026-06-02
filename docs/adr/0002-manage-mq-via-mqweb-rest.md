# ADR-0002: Manage MQ via the mqweb REST API behind an MQAdmin port

- **Status**: Accepted
- **Date**: 2026-06-02

## Context

The operator must talk to an existing IBM MQ Queue Manager to define/inspect/
delete MQSC objects (queues today; authorities, channels, topics later). IBM
offers two integration paths:

- **mqweb Administrative REST API** over HTTPS — including a `/mqsc` endpoint
  that executes arbitrary MQSC (`runCommand` / `runCommandJSON`). See
  [../IBM_MQ_REST_API.md](../IBM_MQ_REST_API.md).
- **PCF** via `ibm-messaging/mq-golang`, which requires the native MQ C client
  and CGO.

We value a pure-Go, easily testable, slim-image build, and we want to keep the
door open to PCF if a future environment lacks mqweb.

## Decision

We will manage MQ through the **mqweb REST API**, primarily the
`/v3/admin/action/qmgr/{qmgr}/mqsc` endpoint, and place all MQ interaction
behind a single Go interface — the **`MQAdmin` port** (`internal/mqadmin`). The
REST client (`internal/adapter/mqrest`) is the only implementation today.

## Consequences

- The build stays **CGO-free and static** (`CGO_ENABLED=0`), yielding a slim
  distroless image and trivial cross-compilation.
- Tests are easy: reconcilers run against a mockery mock of `MQAdmin`; the
  adapter is tested against an `httptest` server. No native client, no MQ broker
  in unit tests.
- Transport is firewall-friendly HTTPS; we depend on mqweb being enabled on the
  target Queue Manager (an explicit prerequisite and non-goal to deploy).
- The port seam means a **PCF adapter could be added later** implementing the
  same interface, with zero controller changes.
- We must handle REST specifics: CSRF header on mutations, MQSC response parsing
  (`overallCompletionCode`/`commandResponse`), and version targeting (`v3`).

## Alternatives considered

- **PCF via `mq-golang`**: native, feature-complete, but requires CGO + bundling
  the MQ C client, complicating builds, images, and tests. Rejected as the
  default; retained as a possible future adapter behind `MQAdmin`.
- **REST native resource endpoints only** (`/queue`, `/channel`): limited object
  coverage (no topics/auth/CHLAUTH). We standardise on `/mqsc` for a single
  reconcile path and use native GETs only where their JSON attribute model
  helps.
