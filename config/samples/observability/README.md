# Observability samples

User-facing artifacts for Prometheus scrape, alerting, and Grafana dashboards.
Full guide: [docs/OBSERVABILITY.md](../../../docs/OBSERVABILITY.md).

| File | Purpose |
|------|---------|
| [`metrics-helm-values.yaml`](metrics-helm-values.yaml) | Annotated Helm values — enable `ServiceMonitor` + `PrometheusRule` |
| [`grafana-dashboard.json`](grafana-dashboard.json) | Starter Grafana dashboard (import via UI or provisioning) |

**ServiceMonitor:** prefer the Helm chart (`metrics.serviceMonitor.enabled`) or
[`charts/mkurator/samples/values-kind.yaml`](../../../charts/mkurator/samples/values-kind.yaml)
for local kind. Kustomize installs can use [`config/prometheus/monitor.yaml`](../../prometheus/monitor.yaml).

**Metric names** (from `internal/metrics/metrics.go`; cardinality tested in `metrics_test.go`):

- `mkurator_reconcile_total` — labels `controller`, `result` (`success` / `error`)
- `mkurator_reconcile_errors_total` — label `controller`
- `mkurator_mq_operations_total` — labels `operation`, `result`
- `mkurator_drift_detected_total` — label `controller` (workload CRs only)
- `mkurator_mq_circuit_breaker_transitions_total` — labels `from`, `to`

See the Helm values file comments and [docs/OBSERVABILITY.md](../../../docs/OBSERVABILITY.md) for full label values.
