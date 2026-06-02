# Kurator Helm chart

Installs the Kurator operator (controller Deployment, RBAC, and CRDs).

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
task cluster:up
task deploy:helm
kubectl apply -k charts/kurator/samples/resources/
```

`deploy:helm` builds the dev image, loads it into the `kurator` kind cluster, and
installs this chart with [`samples/values-kind.yaml`](samples/values-kind.yaml).

Kustomize install (`task deploy`) remains available for controller-runtime workflows.

## Configuration

| Value | Description | Default |
|-------|-------------|---------|
| `image.repository` | Controller image repository | `kurator-controller-manager` |
| `image.tag` | Image tag | `dev` |
| `leaderElection.enabled` | Pass `--leader-elect` | `true` |
| `metrics.enabled` | Expose HTTPS metrics on :8443 | `true` |
| `logging.level` | `KURATOR_LOG_LEVEL` | `info` |
| `logging.format` | `KURATOR_LOG_FORMAT` | `json` |

## CRDs

CRD manifests live in [`crds/`](crds/). Regenerate from kubebuilder output:

```sh
task manifests
task helm:sync-crds
```

Helm installs CRDs on first install; upgrading CRDs may require a manual `kubectl apply`
when the API changes.

## Publishing

Package the chart for an OCI registry or chart museum:

```sh
task helm:package
# artifact: dist/kurator-0.1.0.tgz
```

Bump `version` in `Chart.yaml` for each release and align `appVersion` with the
controller image tag you publish.
