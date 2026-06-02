# Local kind dev cluster

A one-command local environment for developing and testing the Kurator
operator. It provisions a [kind](https://kind.sigs.k8s.io/) cluster and,
via Terraform + Helm, installs:

- **HAProxy Ingress** (NodePort 30080/30443, mapped to the host).
- **cert-manager** (for future operator webhook certificates).
- **Argo CD** (GitOps UI; initial admin password written to `.state/argocd.env`).
- **kube-prometheus-stack** (Prometheus + **Grafana**).
- **IBM MQ** queue manager (`QM1`) from the [upstream IBM MQ Helm chart](https://ibm-messaging.github.io/mq-helm),
  with mqweb exposed through a Terraform-managed HAProxy Ingress (the upstream
  chart's Ingress targets nginx and is not used).

TLS uses a [mkcert](https://github.com/FiloSottile/mkcert) wildcard certificate
for `*.localhost`.

## Prerequisites

- A container runtime: Docker (recommended), nerdctl, or Podman
- [`kind`](https://kind.sigs.k8s.io/), [`kubectl`](https://kubernetes.io/docs/tasks/tools/),
  [`helm`](https://helm.sh/), [`terraform`](https://developer.hashicorp.com/terraform),
  [`mkcert`](https://github.com/FiloSottile/mkcert), and [`task`](https://taskfile.dev)

## Usage

From the repository root:

```sh
task cluster:up       # kind + TLS + Terraform; prints URLs
task cluster:info     # re-print access URLs
task cluster:cleanup  # terraform destroy (keeps kind cluster)
task cluster:down     # destroy Terraform + delete kind + wipe .state
```

`task cluster:up` is idempotent: an existing `kurator` cluster is reused; other
kind clusters blocking NodePorts 30080/30443 are removed automatically.

## Access (after `task cluster:up`)

| What | URL | Credentials |
|------|-----|---------------|
| Argo CD | https://argocd.localhost:30443/ | `admin` — password in `.state/argocd.env` |
| IBM MQ web console | https://mq.localhost:30443/ibmmq/console/ | `admin` / `passw0rd` |
| IBM MQ admin REST | https://mq.localhost:30443/ibmmq/rest/v2/admin/qmgr | `admin` / `passw0rd` |
| Grafana | https://grafana.localhost:30443/ | `admin` / `admin` |

In-cluster: `https://ibm-mq.ibm-mq.svc:9443` (`QueueManagerConnection.endpoint`).

## Notes

- Cluster name defaults to `kurator` (`CLUSTER_NAME` env var overrides).
- State lives under `hack/kind-cluster/.state/` (git-ignored).
- IBM MQ uses the Advanced for Developers license (`license: accept`) for local dev only.
