# ADR-0003: Decouple connection details with QueueManagerConnection

- **Status**: Accepted
- **Date**: 2026-06-02

## Context

A `Queue` (and future object CRs) needs to know *which* Queue Manager to act on
and *how* to reach it: endpoint, TLS trust, and credentials. We could inline
those details on every object CR, or model the connection separately. Many
objects typically share one Queue Manager, and credentials must never live in
specs.

## Decision

We will model connection details as a dedicated **`QueueManagerConnection`** CR
that holds the endpoint, TLS settings (`caSecretRef`, `insecureSkipVerify`), and
a `credentialsSecretRef`. Object CRs such as `Queue` reference it via a
`connectionRef`. Credentials and CA material always come from referenced
`Secret`s, never inline.

## Consequences

- Many `Queue` objects share a single connection definition; endpoint/credential
  changes happen in one place.
- Credentials stay in `Secret`s, satisfying the "no inline secrets" requirement
  (NFR SEC-1) and keeping object specs portable across environments.
- The operator caches one pooled mqweb client per connection and rebuilds it
  when the connection spec or referenced Secret changes (NFR PERF-2).
- `QueueManagerConnection` gets its own reconciler and a `Ready` condition
  (connection reachable via `Ping`), giving clear, separable status.
- Slight indirection: applying a `Queue` requires a `QueueManagerConnection` to
  exist first; the `Queue` reports a clear condition until its connection is
  Ready.

## Alternatives considered

- **Inline endpoint + secretRef on every object CR**: fewer objects, but
  duplicated connection data, scattered updates, and a larger blast radius for
  changes. Rejected.
- **Operator-wide static config (flags/env) for a single Queue Manager**:
  simplest, but precludes managing multiple Queue Managers and bakes
  environment into the deployment. Rejected as too rigid.
