# ADR-0014: MQ error taxonomy and requeue strategy

- **Status**: Accepted (amended 2026-08-06, REQ-REL-2026-08)
- **Date**: 2026-06-02

## Context

Reconcilers must react differently to “bad MQSC syntax”, “QM down”, and “object
not found”. String-matching HTTP bodies in controllers does not scale and breaks
when mqweb wording changes. controller-runtime already provides rate-limited
requeues; we must not hot-loop on permanent failures.

MQ interaction is behind the `MQAdmin` port ([ADR-0002](0002-manage-mq-via-mqweb-rest.md)).

## Decision

We will classify errors at the **`internal/mqadmin` port boundary** (implemented
in `mqrest`, consumed by controllers):

| Class | Types / signals | Controller behaviour |
|-------|-----------------|----------------------|
| **Terminal** | `*TerminalError` (`ErrTerminal`), invalid MQSC, 4xx auth/validation | Failing status condition; Warning Event ([ADR-0015](0015-kubernetes-events-on-transitions.md)); return **without** unbounded requeue — **except** the QMC terminal-error carve-out below |
| **Transient** | 5xx, timeouts, QM unavailable (503) | Return error to trigger controller-runtime **backoff requeue** |
| **NotFound** | `*NotFoundError` (`ErrNotFound`) | Ensure: treat as create needed; Delete: treat as already gone |

### QMC terminal-error recovery carve-out (AUTH-14)

Auth failures (e.g. 401 Unauthorized) motivated this carve-out: they are classified terminal by
the mqweb client, but the underlying credentials are **mutable** — an operator rotates the Secret
and the controller must self-heal without manual intervention.

The Secret watch (AUTH-14, `internal/controller/secret_watch.go`) is the fast path: a Secret
update enqueues all referencing QMCs within seconds. However, a transient hub read failure in
`secretEnqueueMapper` can silently drop the enqueue, leaving the QMC permanently stuck with no
future trigger.

**Carve-out rule**: the `QueueManagerConnection` reconciler returns
`ctrl.Result{RequeueAfter: TerminalRetryInterval()}` (default 2 minutes, configurable via
`--terminal-retry-interval` flag) after setting a terminal-error status. This is a bounded
backstop — the watch is the preferred trigger — and is intentionally slower than
`TransientRequeueInterval` to avoid hot-looping on genuinely misconfigured (non-rotatable)
credentials.

**Scope**: the backstop applies to **every** non-transient QMC error, not only auth. `fail()`
takes the `TerminalRetryInterval` path whenever the error is not `ErrTransient`, so a QMC that
went terminal on, say, an unreachable-endpoint misconfiguration is also retried every 2 minutes.
This is deliberate — the QMC's inputs (Secret, endpoint, TLS material) are all mutable, so no
terminal QMC state is known to be permanent — but it is broader than the auth-rotation case that
prompted it.

This carve-out does **not** apply to workload reconcilers (Queue, Channel, Topic, etc.), where
terminal errors from MQSC misconfiguration must remain one-shot with no requeue.

Principles:

- Wrap with context: `fmt.Errorf("define queue: %w", err)`.
- Inspect with `errors.Is` / `errors.As` only — no substring checks in reconcilers.
- Never panic in reconcile.
- **Workload reconcilers** register a **watch** on `QueueManagerConnection` status
  so queues/topics/channels requeue when a connection becomes `Ready` instead of
  relying solely on periodic requeue.

`RunMQSC` on the REST client is for fixtures/e2e and future work — not part of
`Admin`; reconcilers use typed port methods only.

### Amendment — never `(Result{}, nil)` on error (2026-08-06, REQ-REL-2026-08)

A live incident showed a third class the original table missed: an error that is
*neither* `TerminalError` nor `TransientError` (e.g. a bare
`context.DeadlineExceeded` escaping the adapter) was treated like terminal —
returned as `(ctrl.Result{}, nil)` — and, because the workload predicates only
enqueue on generation / `reconcile-requested-at` / lifecycle change, the CR
wedged permanently with no log and no event surviving the 1h TTL. The amended
policy for workload reconcilers (`setSyncedError`):

- **Transient** → `RequeueAfter: TransientRequeueInterval()` (30s default), as before.
  Context deadline expiry/cancellation is classified transient at the adapter
  boundary (`roundTrip`, `sleepWithContext`, LTPA login).
- **Terminal** → the *only* no-requeue path; the condition reason defaults to
  `TerminalError` (a specific `TerminalError.Reason` still wins) so no-retry
  states are distinguishable from retryable `Error`.
- **Anything else** (including missing connection/Secret) → the error is returned
  to controller-runtime for rate-limited backoff, which also logs it at ERROR.
- Every `Synced=False` transition that does not return an error emits its own
  ERROR log line with the error and scheduled retry interval, so a not-synced CR
  is always greppable.

## Consequences

- New mqweb failure modes map to port errors in one place (`mqrest`).
- Terminal misconfiguration surfaces once with a stable `Reason` on status/Events.
- Transient outages self-heal without manual CR edits when MQ returns.
- Tests assert classification via mock `Admin` errors and adapter unit tests.

## Alternatives considered

- **Classify in each reconciler**: duplicated logic. Rejected.
- **Always requeue forever**: hides terminal MQSC mistakes. Rejected.
- **Custom workqueue rate limiter per CR**: unnecessary; controller-runtime defaults
  suffice with terminal vs transient returns.

## References

- `internal/mqadmin/admin.go` — `TerminalError`, `NotFoundError`
- [OPERATOR_RUNTIME.md](../OPERATOR_RUNTIME.md#error-handling-and-requeue-adr-0014)
