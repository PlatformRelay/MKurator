# ADR-0007: Structured logging with logr and slog

- **Status**: Accepted
- **Date**: 2026-06-02

## Context

Observability NFRs require **structured, JSON-capable logs** with per-object
context and configurable verbosity ([OBS-4](../NON_FUNCTIONAL_REQUIREMENTS.md)),
while security NFRs forbid credentials and credentialed HTTP bodies in logs at
default levels ([SEC-5](../NON_FUNCTIONAL_REQUIREMENTS.md)). Operability expects
runtime tuning without rebuilds ([OPS-4](../NON_FUNCTIONAL_REQUIREMENTS.md)).

The operator is built on **controller-runtime** ([ADR-0001](0001-use-kubebuilder-controller-runtime.md)).
That stack standardises on the **`logr` interface**: reconcilers, the manager,
the client, and webhooks all emit through `logr.Logger` values carried on
`context.Context`. The Kubebuilder scaffold wires **`zap`** as the default
implementation today (`ctrl.SetLogger(zap.New(...))` in `cmd/main.go`), with
`Development: true` until production flags are set.

Go 1.21 added **`log/slog`** to the standard library. The project already pins
**`github.com/go-logr/logr` v1.4+**, which provides `logr.FromSlogHandler` to
bridge a `slog.Handler` into `logr`. The operator author is familiar with `slog`
and wants a deliberate choice rather than accepting the scaffold default by
inertia.

We need one logging story that:

- Integrates with controller-runtime’s context injection and verbosity (`V()`).
- Produces logs aggregators can parse (JSON in cluster; readable text locally).
- Keeps reconcilers testable without a real log sink.
- Stays lean ([ADR-0005](0005-keep-tooling-lean.md)): no second logging stack in
  business logic, no custom log facades unless they earn their keep.

## Decision

### 1. Application code uses `logr` only — never `slog` or `zap` directly

All operator code under `internal/` and reconcilers obtains a logger from context:

```go
log := log.FromContext(ctx)
log.Info("defined queue", "queue", spec.Name, "connection", connRef)
```

Use `logr`’s structured key/value pairs (even-length arguments). Do **not** import
`log/slog`, `go.uber.org/zap`, or `k8s.io/klog` in reconcilers, the `MQAdmin`
port, or the `mqrest` adapter.

**Rationale:** controller-runtime attaches reconcile-scoped values to the context
when controllers use its patterns; tests can substitute `logr` discards or
`logr/logtest`; and the Kubernetes ecosystem (client, admission, leader election)
already speaks `logr`.

### 2. Bootstrap uses `slog` as the sole sink implementation

Only `cmd/main.go` (and, if needed, a tiny `internal/logging` package for handler
construction) touches `log/slog`. Wire the manager like this:

```go
handler := /* JSON or text, level from flags/env */
ctrl.SetLogger(logr.FromSlogHandler(handler))
```

Replace the Kubebuilder-default `zap.New(zap.UseFlagOptions(&opts))` setup.

**Rationale:** `slog` is stdlib, matches team familiarity, and satisfies OBS-4
without learning zap’s flag surface. `logr` remains the stable API at the
controller boundary; `slog` is an implementation detail confined to process
startup.

### 3. Output format and level are configurable (flags / env)

| Mode | Handler | Typical use |
|------|---------|-------------|
| **Production (cluster)** | `slog.NewJSONHandler(os.Stdout, …)` | Loki/Elastic/Cloud Logging |
| **Development (local)** | `slog.NewTextHandler(os.Stderr, …)` or JSON to stdout | `task run` / `make run` |

Expose at least:

- **Log level** — maps to `slog.HandlerOptions.Level` (default `Info`; `debug`
  enables verbose reconcile detail via `log.V(1)`).
- **Log format** — `json` vs `text` (default `json` when `KUBERNETES_SERVICE_HOST`
  is set, else `text` for local ergonomics, overridable by flag).

Drop zap-specific flags (`-zap-devel`, `-zap-log-level`, etc.) in favour of
`--log-level`, `--log-format`, and `--log-config`, plus `KURATOR_LOG_*`
environment variables. Full precedence and examples are in
[LOGGING.md](../LOGGING.md); see also [DEVELOPMENT.md](../DEVELOPMENT.md) and
the manager Deployment manifest.

### 4. Conventions for fields and verbosity

**Per-reconcile context** — at the top of each `Reconcile`, enrich the logger and
put it back on the context:

```go
log := log.FromContext(ctx).
    WithValues(
        "controller", "queue",
        "namespace", req.Namespace,
        "name", req.Name,
    )
// Add resource-specific keys, e.g. queueManagerConnection, queueName
ctx = log.IntoContext(ctx, log)
```

Prefer **lowerCamelCase** keys aligned with Kubernetes logging guidance
(`namespace`, `name`, `queue`, `connection`, `reconcileID` when not already
present from controller-runtime).

| Level | Use for |
|-------|---------|
| **Error** | Reconcile failed after retries exhausted, or terminal setup errors (`log.Error(err, "msg", keys…)`). |
| **Info** | Lifecycle: started/finished reconcile, object created/deleted on MQ, connection established. |
| **V(1)** | High-volume detail: drift detected, MQSC command class (not full body), requeue reason. |
| **V(2)+** | Deep debugging only; may include sanitised HTTP status/latency, never bodies with secrets. |

Do **not** log: Secret data, basic-auth passwords, CSRF tokens, full `Authorization`
headers, or raw mqweb request/response bodies. Log **references** (`secret`,
`namespace`/`name`) and **outcomes** (HTTP status, MQ reason code, object name).

### 5. Security: defensive handler (follow-up in Phase 2)

Add a wrapping `slog.Handler` (in `internal/logging` if non-trivial) that replaces
known sensitive attribute keys and strips credential substrings from string values
before write. Unit-test it ([SEC-5](../NON_FUNCTIONAL_REQUIREMENTS.md)). This is
belt-and-suspenders on top of discipline in call sites.

### 6. Testing

- **Unit / envtest:** rely on default `logr` discards or `logr/logtest` when a test
  must assert log output; do not configure `slog` in every test file.
- **E2e:** scrape JSON logs from the manager pod and spot-check required keys
  (`controller`, `namespace`, `name`) on a happy-path reconcile.

## Consequences

**Positive**

- One idiomatic API (`logr`) across controllers and adapters; one modern sink
  (`slog`) at the edge.
- JSON logs in production without a project-specific logging framework.
- Aligns with ARCHITECTURE.md / NFR observability and security bars.
- Easier for contributors who know `slog` — they implement handlers in `cmd/`,
  not scattered `slog.Info` calls.

**Negative / neutral**

- Deviates from the stock Kubebuilder zap flags; local runbooks and upstream
  docs must reference our `-log-*` flags instead.
- `zap` may remain an **indirect** module dependency via controller-runtime until
  upstream drops it; we simply stop importing it in our code.
- `logr`’s `V(n)` verbosity mapping to `slog` levels must be configured once in
  bootstrap (document the mapping in `internal/logging`).

**Follow-up work**

- ~~Phase 1: replace zap bootstrap, configurable slog in `internal/logging`, flags/env/file,
  Deployment defaults, and logging unit tests~~ (done; see [LOGGING.md](../LOGGING.md)).
- Phase 2: reconciler logging conventions per LOG-4 in all controllers.
- Phase 2: extend redacting handler tests as mqrest logging lands.
- Phase 2: optional `log/slog` **test** usage only inside `mqrest` adapter tests if
  asserting debug output — still not in production adapter code (use `logr` there).

## Alternatives considered

### Use `slog` directly everywhere (no `logr`)

**Rejected.** controller-runtime, `client-go`, and envtest integrate with `logr`.
Bypassing it means fighting the framework, losing context-scoped reconcile loggers,
and duplicating wiring that `log.FromContext(ctx)` already provides.

### Keep Kubebuilder’s zap backend; use `logr` in application code only

**Viable but not chosen.** Lowest migration cost and matches most operator
tutorials. Rejected because the maintainer prefers `slog`, zap’s CLI surface is
heavier than needed for this project, and `logr.FromSlogHandler` is stable on
our pinned `logr` version. Staying on zap would be acceptable if slog bridging
proved problematic — that would be a small ADR amendment, not a reconciler rewrite.

### `zap` inside `internal/` as well as bootstrap

**Rejected.** Two APIs in the codebase violates ADR-0005 and confuses tests.

### `k8s.io/klog`

**Rejected.** Unstructured heritage API; not a fit for JSON observability goals.

### Heavy logging framework (zerolog, logrus, etc.)

**Rejected.** Extra dependency and style divergence from Go stdlib + Kubernetes
norms without benefit for this operator’s size.
