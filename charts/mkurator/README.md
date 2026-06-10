# MKurator Helm chart

Installs the MKurator operator (controller Deployment, RBAC, and CRDs).

**Managing MQ queues after install:** see [docs/INSTALL_AND_USE.md](../../docs/INSTALL_AND_USE.md)
and [samples/resources/README.md](samples/resources/README.md) (same content as
[`config/samples/`](../../config/samples/README.md)).

## Prerequisites

- Kubernetes 1.28+
- Helm 3
- An existing IBM MQ queue manager with **mqweb** enabled

## Install

```sh
helm upgrade --install mkurator . \
  --namespace mkurator-system \
  --create-namespace
```

## Local kind development

With the platform from [`hack/kind-cluster`](../../hack/kind-cluster/README.md):

```sh
task local:up      # recommended: cluster + this chart + sample CRs
# or step by step:
task cluster:up
task deploy:helm
task deploy:samples
```

`deploy:helm` builds the dev image, loads it into the `mkurator` kind cluster, and
installs this chart with [`samples/values-kind.yaml`](samples/values-kind.yaml).
`deploy:samples` applies [`samples/resources/`](samples/resources/) (Secret,
`QueueManagerConnection`, `Queue` for `QM1`).

Kustomize install (`task deploy`) remains available for controller-runtime workflows.

## Configuration

Helm validates critical values against [`values.schema.json`](values.schema.json)
(image repository/tag, metrics TLS and ServiceMonitor toggles, webhook cert-manager
settings). The schema catches empty image fields and invalid enums at install time;
operational prerequisites (cert-manager running, Prometheus Operator labels, production
image tags) are spelled out in the post-install [`templates/NOTES.txt`](templates/NOTES.txt)
output.

| Value | Description | Default |
|-------|-------------|---------|
| `image.repository` | Controller image repository | `mkurator-controller-manager` |
| `image.tag` | Image tag | `dev` |
| `leaderElection.enabled` | Pass `--leader-elect` | `true` |
| `metrics.enabled` | Expose Prometheus metrics on `:8443` | `true` |
| `metrics.secure` | HTTPS metrics with kube authn/authz | `true` |
| `metrics.serviceMonitor.enabled` | Create a `ServiceMonitor` (needs Prometheus Operator) | `false` |
| `metrics.prometheusRule.enabled` | Create alerting rules | `false` |
| `logging.level` | `KURATOR_LOG_LEVEL` | `info` |
| `logging.format` | `KURATOR_LOG_FORMAT` | `json` |
| `webhooks.enabled` | Install validating admission webhooks and webhook Service | `true` |
| `webhooks.certManager.create` | Create cert-manager `Issuer` + `Certificate` for webhook TLS (requires [cert-manager](https://cert-manager.io/); installed on the [kind platform](../../hack/kind-cluster/README.md)) | `true` |
| `webhooks.certManager.secretName` | Secret mounted at `/tmp/k8s-webhook-server/serving-certs` | `webhook-server-cert` |

When `webhooks.enabled=true`, **cert-manager must be running** in the cluster so the
serving certificate becomes Ready. Admission behaviour (invalid spec rejection,
unknown-attribute warnings) is described in
[docs/INSTALL_AND_USE.md](../../docs/INSTALL_AND_USE.md#how-it-works).

## CRDs

CRD manifests live in [`crds/`](crds/). Regenerate from kubebuilder output:

```sh
task manifests
task helm:sync-crds
```

Helm installs CRDs on first install; upgrading CRDs may require a manual `kubectl apply`
when the API changes.

## Metrics and alerting

The controller exposes Prometheus metrics on `/metrics` (port `8443`, HTTPS with
Kubernetes API authentication by default). Custom metrics include:

- `mkurator_reconcile_total{controller,result}`
- `mkurator_reconcile_errors_total{controller}`
- `mkurator_mq_operations_total{operation,result}`

On a cluster with **kube-prometheus-stack** (the local kind platform), enable
scraping and alerts:

```sh
helm upgrade --install mkurator . \
  --namespace mkurator-system \
  -f samples/values-kind.yaml
```

`values-kind.yaml` sets `metrics.serviceMonitor` and `metrics.prometheusRule`
labels to `release: kube-prometheus-stack` for operator discovery.

When `metrics.prometheusRule.enabled=true`, the chart creates a **PrometheusRule**
with starter alerts: metrics target down, reconcile/MQ operation errors,
controller pod not ready (`kube_pod_status_ready`), QMC ping and reconcile errors,
and auth mqweb operation errors (channel auth + authority). See
[docs/OBSERVABILITY.md](../../docs/OBSERVABILITY.md) for expressions and drift vs
error metrics.

## Publishing

Package the chart for an OCI registry or chart museum:

```sh
task helm:package
# artifact: dist/mkurator-0.5.0.tgz
```

Bump `version` in `Chart.yaml` for each release and align `appVersion` with the
controller image tag you publish.
