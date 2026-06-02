# Kurator Helm chart

Installs the Kurator operator (controller Deployment, RBAC, and CRDs).

**Managing MQ queues after install:** see [docs/INSTALL_AND_USE.md](../../docs/INSTALL_AND_USE.md)
and [samples/resources/README.md](samples/resources/README.md) (same content as
[`config/samples/`](../../config/samples/README.md)).

## Prerequisites

- Kubernetes 1.28+
- Helm 3
- An existing IBM MQ queue manager with **mqweb** enabled

## Install

```sh
helm upgrade --install kurator . \
  --namespace kurator-system \
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

`deploy:helm` builds the dev image, loads it into the `kurator` kind cluster, and
installs this chart with [`samples/values-kind.yaml`](samples/values-kind.yaml).
`deploy:samples` applies [`samples/resources/`](samples/resources/) (Secret,
`QueueManagerConnection`, `Queue` for `QM1`).

Kustomize install (`task deploy`) remains available for controller-runtime workflows.

## Configuration

| Value | Description | Default |
|-------|-------------|---------|
| `image.repository` | Controller image repository | `kurator-controller-manager` |
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
[docs/INSTALL_AND_USE.md](../../docs/INSTALL_AND_USE.md#how-kurator-manages-mq-objects).

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

- `kurator_reconcile_total{controller,result}`
- `kurator_reconcile_errors_total{controller}`
- `kurator_mq_operations_total{operation,result}`

On a cluster with **kube-prometheus-stack** (the local kind platform), enable
scraping and alerts:

```sh
helm upgrade --install kurator . \
  --namespace kurator-system \
  -f samples/values-kind.yaml
```

`values-kind.yaml` sets `metrics.serviceMonitor` and `metrics.prometheusRule`
labels to `release: kube-prometheus-stack` for operator discovery.

## Publishing

Package the chart for an OCI registry or chart museum:

```sh
task helm:package
# artifact: dist/kurator-0.5.0.tgz
```

Bump `version` in `Chart.yaml` for each release and align `appVersion` with the
controller image tag you publish.
