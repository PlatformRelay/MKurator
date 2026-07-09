# Observability (metrics)

Practical guide to **Prometheus metrics** for MKurator in production. Logging is
documented separately in [LOGGING.md](LOGGING.md).

Doc index: [README.md](index.md) · Install: [INSTALL_AND_USE.md](INSTALL_AND_USE.md)

## What you get out of the box

| Capability | Default | You configure |
|------------|---------|---------------|
| Controller-runtime metrics + MKurator counters | **On** (`metrics.enabled=true`) | Scraping / dashboards |
| HTTPS metrics on port **8443** | **On** (`metrics.secure=true`) | Network policies, TLS trust |
| Kubernetes API auth on `/metrics` | **On** (secure mode) | Prometheus `ServiceMonitor` + RBAC |
| `ServiceMonitor` CR | **Off** | Enable when Prometheus Operator is installed |
| `PrometheusRule` alerts | **Off** | Enable with kube-prometheus-stack labels |
| Structured logs | **On** (JSON, `info`) | [LOGGING.md](LOGGING.md), Helm `logging.*` |

The operator does **not** install Prometheus, Grafana, or the Prometheus Operator.
You need a monitoring stack (or a vendor equivalent) to scrape and alert.

## Quick start: metrics + dashboard

User-facing samples live under [`config/samples/observability/`](https://github.com/conduit-ops/MKurator/tree/main/config/samples/observability):

| Artifact | Purpose |
|----------|---------|
| [`metrics-helm-values.yaml`](https://github.com/conduit-ops/MKurator/blob/main/config/samples/observability/metrics-helm-values.yaml) | Annotated Helm values — `ServiceMonitor` + `PrometheusRule` |
| [`grafana-dashboard.json`](https://github.com/conduit-ops/MKurator/blob/main/config/samples/observability/grafana-dashboard.json) | Starter Grafana dashboard (panels aligned to MKurator metric names) |

### 1. Enable scrape and starter alerts (Helm)

Requires [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
CRDs (e.g. **kube-prometheus-stack**). Adjust `metrics.serviceMonitor.labels` to match
your Prometheus `serviceMonitorSelector`.

```sh
helm upgrade --install mkurator ./charts/mkurator \
  --namespace mkurator-system \
  --create-namespace \
  -f config/samples/observability/metrics-helm-values.yaml
```

Local kind dev (image + logging overrides): use
[`charts/mkurator/samples/values-kind.yaml`](https://github.com/conduit-ops/MKurator/blob/main/charts/mkurator/samples/values-kind.yaml)
instead — it enables the same metrics toggles.

Ensure Prometheus can scrape secure metrics: bind **`{release}-metrics-reader`**
to the Prometheus ServiceAccount (see [RBAC](#rbac-metrics-reader-pattern) below).

### 2. Import the Grafana dashboard

**UI:** Grafana → Dashboards → New → Import → Upload
`config/samples/observability/grafana-dashboard.json`. Select your Prometheus data
source and set the **Namespace** variable (default `mkurator-system`).

**Provisioning (optional):** mount the JSON under your Grafana sidecar or
`dashboardProviders` path; set `uid: mkurator-operator` to avoid duplicates on re-import.

Panels use the custom metrics below plus controller-runtime workqueue /
`controller_runtime_reconcile_total` from the same `/metrics` endpoint.

### 3. Verify

```sh
kubectl -n mkurator-system get servicemonitor,prometheusrule
# In Prometheus: up{namespace="mkurator-system"} and mkurator_reconcile_total
```

See [Verify scraping](#verify-scraping) for port-forward and 403 troubleshooting.

## Metrics endpoint

- **Path:** `/metrics`
- **Port:** `8443` (named port `metrics` on the manager pod)
- **Service:** `{release}-metrics` in the operator namespace (Helm fullname prefix)

When `metrics.secure=true` (default), scrapes must use **HTTPS** and present a
valid Kubernetes service account token (or a subject allowed by RBAC).

### Metrics inventory

All MKurator custom metrics are **counters** registered in
[`internal/metrics/metrics.go`](https://github.com/conduit-ops/MKurator/blob/main/internal/metrics/metrics.go). Label cardinality
is covered by unit tests in [`internal/metrics/metrics_test.go`](https://github.com/conduit-ops/MKurator/blob/main/internal/metrics/metrics_test.go).

| Type | Custom MKurator metrics | Histograms |
|------|-------------------------|------------|
| Counters | 5 (see table below) | — |
| Histograms | — | none today (mqweb latency not instrumented) |

The same `/metrics` endpoint also exposes **controller-runtime** workqueue metrics
(e.g. `workqueue_depth`, `workqueue_adds_total`, `controller_runtime_reconcile_total`,
`controller_runtime_reconcile_time_seconds` histogram) and standard Go/process
collectors. The starter Grafana dashboard uses MKurator counters plus
`controller_runtime_reconcile_total`.

#### MKurator counters

| Metric | Labels | Incremented when |
|--------|--------|------------------|
| `mkurator_reconcile_total` | `controller`, `result` | Every reconcile pass completes (`result`: `success` or `error`) |
| `mkurator_reconcile_errors_total` | `controller` | Reconcile pass returns a non-nil error to the manager |
| `mkurator_mq_operations_total` | `operation`, `result` | mqweb adapter call completes (`result`: `success` or `error`) |
| `mkurator_drift_detected_total` | `controller` | Workload reconcile detects attribute drift on IBM MQ |
| `mkurator_mq_circuit_breaker_transitions_total` | `from`, `to` | mqweb per-connection circuit breaker changes state |

**`controller` label values** (reconcile + drift; drift excludes QMC):

| Value | Reconciler |
|-------|------------|
| `queue` | Queue |
| `topic` | Topic |
| `channel` | Channel |
| `channelauthrule` | ChannelAuthRule |
| `authorityrecord` | AuthorityRecord |
| `queuemanagerconnection` | QueueManagerConnection (reconcile only; no drift counter) |

**`operation` label values** (`mkurator_mq_operations_total`):

| Value | mqweb call |
|-------|------------|
| `ping` | QueueManagerConnection health check |
| `get_queue` / `define_queue` / `delete_queue` | Queue CRUD |
| `get_topic` / `define_topic` / `delete_topic` | Topic CRUD |
| `get_channel` / `define_channel` / `delete_channel` | Channel CRUD |
| `set_channel_auth` / `get_channel_auth` / `delete_channel_auth` | CHLAUTH |
| `set_authority` / `get_authority` / `delete_authority` | OAM AUTHREC |
| `run_mqsc` | Raw MQSC (integration/e2e fixtures only) |

**`from` / `to` label values** (`mkurator_mq_circuit_breaker_transitions_total`):
`closed`, `open`, `half_open` — transitions observed include `closed→open`,
`open→half_open`, `half_open→closed`, and `half_open→open`.

**`result` label values** (reconcile + MQ operation counters): `success`, `error`.
Any non-nil error (including wrapped errors) counts as `error`.

### Drift vs reconcile errors

Attribute **drift** (spec no longer matches MQ) sets workload CR status
`Synced=False` with `Reason=DriftDetected`, increments
`mkurator_drift_detected_total`, and does **not** increment
`mkurator_reconcile_errors_total` — the reconcile pass succeeds after patching status.

Alert on drift via the drift counter, CR conditions, or GitOps checks — not via
reconcile-error metrics. Transient MQ failures and terminal reconcile errors **do**
increment `mkurator_reconcile_errors_total` and often set `Reason=Error`.

## Readiness (`/readyz`, REL-7)

- **Liveness:** `/healthz` — process alive; independent of MQ connectivity.
- **Readiness:** `/readyz` — aggregated `QueueManagerConnection` status:
  - **Ready** when there are no non-deleting QMCs, or at least one has `Ready=True`
    (successful mqweb ping).
  - **Not ready** (HTTP 500) when one or more non-deleting QMCs exist and **none**
    report `Ready=True`, so the Deployment stops routing traffic while every
    configured connection is down.

Implemented in `internal/health` ([NON_FUNCTIONAL_REQUIREMENTS.md](NON_FUNCTIONAL_REQUIREMENTS.md)
REL-7). The Helm alert **MKuratorOperatorNotReady** uses `kube_pod_status_ready` on
the controller-manager pod (requires kube-state-metrics, e.g. kube-prometheus-stack).

## ServiceMonitor scrape configuration

When `metrics.serviceMonitor.enabled=true`, Helm renders
[`charts/mkurator/templates/servicemonitor.yaml`](https://github.com/conduit-ops/MKurator/blob/main/charts/mkurator/templates/servicemonitor.yaml)
in the release namespace. Kustomize installs can use
[`config/prometheus/monitor.yaml`](https://github.com/conduit-ops/MKurator/blob/main/config/prometheus/monitor.yaml) (disabled in
the default overlay).

| Field | Helm default | Notes |
|-------|--------------|-------|
| `spec.endpoints[].path` | `/metrics` | Same path as controller-runtime |
| `spec.endpoints[].port` | `https` | Named port on the metrics Service (8443) |
| `spec.endpoints[].scheme` | `https` | Required when `metrics.secure=true` |
| `spec.endpoints[].interval` | `30s` | Helm value `metrics.serviceMonitor.interval` |
| `spec.endpoints[].scrapeTimeout` | `10s` | Helm value `metrics.serviceMonitor.scrapeTimeout` |
| `spec.endpoints[].bearerTokenFile` | SA token path | Prometheus pod SA must bind `metrics-reader` |
| `spec.endpoints[].tlsConfig.insecureSkipVerify` | `true` | Skip metrics-server cert verify; tighten in prod if needed |
| `spec.selector` | release pod labels | Targets `{release}-metrics` Service |
| `metadata.labels` | chart labels + `metrics.serviceMonitor.labels` | Must match Prometheus `serviceMonitorSelector` |

### Observability checklist

Use this when standing up or reviewing production scrape:

- [ ] `metrics.enabled=true` and `metrics.secure=true` (defaults)
- [ ] Prometheus Operator CRDs installed (e.g. kube-prometheus-stack)
- [ ] `metrics.serviceMonitor.enabled=true` with labels matching `serviceMonitorSelector`
- [ ] ClusterRoleBinding: Prometheus SA → `{release}-metrics-reader`
- [ ] Target **up** in Prometheus: `up{namespace="<release-ns>",service="<release>-metrics"}`
- [ ] Custom counters present after traffic: `mkurator_reconcile_total`, `mkurator_mq_operations_total`
- [ ] Optional: `metrics.prometheusRule.enabled=true` for starter alerts
- [ ] Optional: import [`grafana-dashboard.json`](https://github.com/conduit-ops/MKurator/blob/main/config/samples/observability/grafana-dashboard.json)
- [ ] Drift monitoring: `rate(mkurator_drift_detected_total[15m])` or CR `Synced` conditions

## Enabling Prometheus scrape (Helm)

Requires [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
CRDs (e.g. **kube-prometheus-stack**).

```sh
helm upgrade --install mkurator ./charts/mkurator \
  --namespace mkurator-system \
  --set metrics.serviceMonitor.enabled=true \
  --set metrics.serviceMonitor.labels.release=kube-prometheus-stack
```

Adjust `metrics.serviceMonitor.labels` to match your Prometheus `serviceMonitorSelector`.
The local kind stack uses `release: kube-prometheus-stack` — see
[`charts/mkurator/samples/values-kind.yaml`](https://github.com/conduit-ops/MKurator/blob/main/charts/mkurator/samples/values-kind.yaml).

Optional starter alerts:

```sh
--set metrics.prometheusRule.enabled=true \
--set metrics.prometheusRule.labels.release=kube-prometheus-stack
```

With `metrics.prometheusRule.enabled=true`, the chart installs **PrometheusRule**
`mkurator.rules` including:

| Alert | Severity | Signal |
|-------|----------|--------|
| MKuratorMetricsTargetDown | critical | Metrics Service not scraped |
| MKuratorReconcileErrors | warning | Any controller reconcile errors |
| MKuratorMQOperationErrors | warning | Any mqweb operation errors |
| MKuratorOperatorNotReady | critical | Controller pod not ready (`kube_pod_status_ready`) |
| MKuratorQMCPingFailures | warning | `mkurator_mq_operations_total{operation="ping",result="error"}` |
| MKuratorQMCReconcileErrors | warning | `mkurator_reconcile_errors_total{controller="queuemanagerconnection"}` |
| MKuratorAuthMQOperationErrors | warning | Auth mqweb ops (channel auth + authority) |

**MKuratorOperatorNotReady** depends on cluster-level kube-state-metrics; the MKurator
metrics alerts only need a scrape of `{release}-metrics`.

### Kustomize installs

The repo includes a sample `ServiceMonitor` under `config/prometheus/` (disabled in
default kustomization). For production, either enable that overlay or use the Helm
chart’s `ServiceMonitor` template with equivalent labels.

## RBAC: metrics-reader pattern

Secure metrics use two cluster roles (created by Helm when `metrics.enabled=true`):

| ClusterRole | Purpose |
|-------------|---------|
| `{release}-metrics-reader` | `GET` on non-resource URL `/metrics` |
| `{release}-metrics-auth` | Delegates authentication/authorization to the Kubernetes API |

Prometheus needs permission to scrape. Typical pattern:

1. Create a **dedicated ServiceAccount** for Prometheus in your monitoring namespace.
2. Bind **`{release}-metrics-reader`** to that ServiceAccount (ClusterRoleBinding).
3. Point the `ServiceMonitor` at the MKurator metrics Service; the chart sets
   `bearerTokenFile` on the scrape endpoint so the Prometheus pod’s SA token is used.

E2e validates this flow with `mkurator-metrics-reader` — see
[`test/e2e/e2e_test.go`](https://github.com/conduit-ops/MKurator/blob/main/test/e2e/e2e_test.go).

If scrapes return **403 Forbidden**, check RoleBindings and that Prometheus runs with
the bound ServiceAccount.

### Insecure metrics (not recommended)

`metrics.secure=false` exposes metrics without Kubernetes API auth. Only use behind
strict network policies in isolated environments.

## Verify scraping

```sh
# Metrics Service exists
kubectl -n mkurator-system get svc -l app.kubernetes.io/name=mkurator

# ServiceMonitor (when enabled)
kubectl -n mkurator-system get servicemonitor

# From a debug pod with a bound metrics-reader SA (simplified)
kubectl -n mkurator-system port-forward svc/mkurator-metrics 8443:8443
# curl -k -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" https://127.0.0.1:8443/metrics
```

In Prometheus UI, query `up{namespace="mkurator-system"}` for the MKurator target.

## Dashboards and SLOs

Import the starter dashboard from
[`config/samples/observability/grafana-dashboard.json`](https://github.com/conduit-ops/MKurator/blob/main/config/samples/observability/grafana-dashboard.json)
(see [Quick start](#quick-start-metrics--dashboard)). Ad-hoc PromQL:

- Reconcile error rate: `rate(mkurator_reconcile_errors_total[5m])` by `controller`
- MQ ping failures: `rate(mkurator_mq_operations_total{operation="ping",result="error"}[5m])`
- Auth MQ failures: `rate(mkurator_mq_operations_total{operation=~"set_channel_auth|get_channel_auth|delete_channel_auth|set_authority|get_authority|delete_authority",result="error"}[5m])`
- MQ operation failures (all): `rate(mkurator_mq_operations_total{result="error"}[5m])`
- Drift detections: `rate(mkurator_drift_detected_total[15m])` by `controller`
- Circuit breaker opens: `increase(mkurator_mq_circuit_breaker_transitions_total{from="closed",to="open"}[15m])`
- Target up: `up` for the metrics Service
- Drift (CR conditions): `kubectl get … -o jsonpath='…conditions[?(@.type=="Synced")].reason'` for `DriftDetected`

Align alerting with [NON_FUNCTIONAL_REQUIREMENTS.md](NON_FUNCTIONAL_REQUIREMENTS.md) (OBS-*).

## See also

- [`config/samples/observability/`](https://github.com/conduit-ops/MKurator/tree/main/config/samples/observability) — Helm values + Grafana JSON  
- [LOGGING.md](LOGGING.md) — log levels and formats  
- [charts/mkurator/README.md](https://github.com/conduit-ops/MKurator/blob/main/charts/mkurator/README.md) — Helm values table  
- [ARCHITECTURE.md](ARCHITECTURE.md) — metrics component overview  
- [UPGRADE.md](UPGRADE.md) — upgrade order (operator before changing scrape config)  
