# Logging

How **MKurator** emits structured logs, how to configure them at runtime, and the
rules every contributor must follow. The architectural choice is recorded in
[ADR-0007](adr/0007-structured-logging-logr-slog.md); observability and security
requirements are in [NON_FUNCTIONAL_REQUIREMENTS.md](NON_FUNCTIONAL_REQUIREMENTS.md)
(OBS-4, SEC-5, OPS-4).

## Stack

| Layer | Technology | Where |
|-------|------------|--------|
| Application code | [`logr`](https://github.com/go-logr/logr) | Reconcilers, `internal/*` — `log.FromContext(ctx)` |
| Sink | [`log/slog`](https://pkg.go.dev/log/slog) | `internal/logging` only |
| Wiring | `ctrl.SetLogger(logr.FromSlogHandler(...))` | `cmd/main.go` via `logging.Setup` |

Do **not** import `zap`, `klog`, or `slog` in reconcilers or adapters.

## Configuration

Settings merge in this order (**later wins**):

1. Built-in defaults
2. YAML config file (if a path is set)
3. Environment variables
4. CLI flags (non-empty values only)

### Defaults

| Setting | In cluster | Local (no `KUBERNETES_SERVICE_HOST`) |
|---------|------------|--------------------------------------|
| `level` | `info` | `info` |
| `format` | `json` | `text` |

### Config file

Path via `--log-config` or `KURATOR_LOG_CONFIG`. YAML example:

```yaml
# See config/samples/logging-config.yaml
level: info    # debug | info | warn | error
format: json   # json | text
```

### Environment variables

| Variable | Purpose |
|----------|---------|
| `KURATOR_LOG_CONFIG` | Path to a YAML file (same schema as above) |
| `KURATOR_LOG_LEVEL` | `debug`, `info`, `warn`, or `error` |
| `KURATOR_LOG_FORMAT` | `json` or `text` |

### CLI flags

| Flag | Purpose |
|------|---------|
| `--log-config` | Path to YAML config file |
| `--log-level` | Override level (empty = not set) |
| `--log-format` | Override format (empty = not set) |

### Deployment example

```yaml
env:
  - name: KURATOR_LOG_LEVEL
    value: info
  - name: KURATOR_LOG_FORMAT
    value: json
  - name: KURATOR_LOG_CONFIG
    value: /etc/mkurator/logging.yaml
volumeMounts:
  - name: logging-config
    mountPath: /etc/mkurator
    readOnly: true
```

Mount a `ConfigMap` at `/etc/mkurator/logging.yaml` to change logging without
rebuilding the image.

## Requirements

These are enforced by review, tests, and (where noted) CI.

| ID | Requirement | Verification |
|----|-------------|--------------|
| LOG-1 | Application code uses `logr` only; bootstrap uses `slog` via `internal/logging`. | Review, import lint |
| LOG-2 | Logs are structured (JSON in production) with stable `lowerCamelCase` keys. | OBS-4; sample output tests |
| LOG-3 | Level and format are configurable at runtime without rebuild (file, env, flags). | OPS-4; config unit tests |
| LOG-4 | Per-reconcile loggers include `controller`, `namespace`, and `name` (plus resource-specific keys). | Review (all reconcilers) |
| LOG-5 | No Secret values, passwords, tokens, CSRF headers, or credentialed HTTP bodies at default levels. | SEC-5; redacting handler + review |
| LOG-6 | `log.V(1)` is used for high-volume detail; disabled when level is `info` or higher. | Level filter tests |

## Guidelines

### Levels and `logr` verbosity

| User `level` | `slog` level | `log.Info` | `log.V(1).Info` | `log.Error` |
|--------------|--------------|------------|-----------------|-------------|
| `error` | Error | hidden | hidden | shown |
| `warn` | Warn | hidden | hidden | shown |
| `info` | Info | shown | hidden | shown |
| `debug` | Debug | shown | shown | shown |

Use **Error** for terminal failures, **Info** for lifecycle events (reconcile
finished, object created on MQ), **V(1)** for drift/requeue/MQSC command class
(without full bodies), **V(2)+** only for deep troubleshooting.

### Field names

Use **lowerCamelCase** keys consistent with Kubernetes logging:

- Always (per reconcile): `controller`, `namespace`, `name`
- Queue resources: `queue`, `connection` (QueueManagerConnection ref)
- Connection resources: `endpoint` (host only, no credentials), `secret` (name ref)

### Safe logging

**Do log:** object names, namespace, HTTP status, MQ reason codes, Secret
**references** (`namespace`/`name`).

**Do not log:** Secret `.Data`, basic-auth passwords, `Authorization` headers,
`ibm-mq-rest-csrf-token`, full mqweb request/response bodies.

The redacting handler masks common sensitive **attribute keys** (`password`,
`token`, `authorization`, etc.) as a safety net; do not rely on it instead of
discipline at call sites.

### Reconcile pattern (all reconcilers)

```go
log := log.FromContext(ctx).WithValues(
    "controller", "queue",
    "namespace", req.Namespace,
    "name", req.Name,
)
ctx = log.IntoContext(ctx, log)
log.Info("reconciling queue")
// ...
log.V(1).Info("mqsc command issued", "command", "DEFINE QLOCAL", "queue", q.Spec.Name)
```

### Tests

- Unit tests: use `logr` discards or `internal/logging` test helpers; avoid
  configuring global `slog` in every file.
- When asserting output, write to an `io.Writer` via `logging.NewLogger` test API
  rather than stdout.

## Sample output

**JSON** (`format: json`, `level: info`):

```json
{"time":"2026-06-02T12:00:00.000Z","level":"INFO","msg":"starting manager","logger":"setup"}
```

**Text** (`format: text`, local dev):

```text
time=2026-06-02T12:00:00.000Z level=INFO msg="starting manager" logger=setup
```

With debug enabled, a verbose reconcile line might look like:

```json
{"time":"...","level":"DEBUG","msg":"requeue requested","controller":"queue","namespace":"default","name":"orders","reason":"transient MQ error"}
```

## Related documents

- [ADR-0007](adr/0007-structured-logging-logr-slog.md) — decision record
- [ARCHITECTURE.md](ARCHITECTURE.md) — runtime logging row
- [DEVELOPMENT.md](DEVELOPMENT.md) — local run and flags
