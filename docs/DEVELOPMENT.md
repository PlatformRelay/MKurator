# Development

How to set up, build, test, and run the IBM Message Queue Operator locally. For
conventions see [../AGENTS.md](../AGENTS.md); for design see
[ARCHITECTURE.md](ARCHITECTURE.md).

> Status: the Go project is scaffolded in [ROADMAP.md](ROADMAP.md) Phase 1. The
> `hack/kind-cluster` local platform already exists and works today. Where a
> `task` target does not exist yet, the equivalent script is given.

## Prerequisites

| Tool | Why |
|------|-----|
| **Go** (the version in `go.mod`) | Build/test the operator |
| **Task** ([taskfile.dev](https://taskfile.dev)) | Single entry point for all workflows |
| **Docker** (or **nerdctl**/**Podman**) | Container runtime for kind and image builds |
| **kind** | Local Kubernetes cluster |
| **kubectl** | Talk to the cluster |
| **Terraform** | Provision the local platform (ingress, cert-manager, monitoring, IBM MQ) |
| **Helm** | Used by Terraform to install charts |
| **mkcert** | Trusted local TLS for `*.localhost` |

Go-based tools (controller-gen, kustomize, mockery, ginkgo, setup-envtest,
golangci-lint) are pinned via `go.mod` `tool` directives and invoked with
`go tool <name>` — no separate install needed.

Optional: **direnv** to auto-export `KUBECONFIG` for the local cluster.

## The inner loop

Fast feedback without a cluster (mocks + envtest):

```sh
task install      # download/verify modules
task generate     # deepcopy + mocks
task manifests    # CRDs + RBAC
task lint         # golangci-lint v2
task test:run     # unit + envtest (Ginkgo, -race, coverage)
```

`task verify` re-runs codegen into a scratch area and fails if anything is stale
— run it before committing (pre-commit does this automatically).

Build the manager binary (CGO-free, static):

```sh
task build
```

## Local platform (kind + IBM MQ)

The `hack/kind-cluster` tree provisions a complete environment: a kind cluster
with **ingress-nginx**, **cert-manager**, an optional **kube-prometheus-stack**,
and a real **IBM MQ** Queue Manager exposing `mqweb` — wired with **Terraform**
and trusted TLS from **mkcert**.

Cluster name: `ibm-mq-operator`. State (kubeconfig, TLS) is written to
`hack/kind-cluster/.state/`.

### Bring it up

Via Task (once the root Taskfile lands):

```sh
task cluster:up      # kind + mkcert TLS + terraform apply (ingress, cert-manager, monitoring, IBM MQ)
```

Or directly with the scripts (works today):

```sh
cd hack/kind-cluster
./scripts/kind-up.sh           # create cluster; writes .state/kubeconfig.yaml
./scripts/mkcert-gen.sh        # wildcard *.localhost cert -> .state/tls.env
./scripts/terraform-apply.sh   # ingress-nginx, cert-manager, monitoring, IBM MQ
./scripts/info.sh              # print URLs and credentials
export KUBECONFIG="$(pwd)/.state/kubeconfig.yaml"
```

### Endpoints

| Target | URL |
|--------|-----|
| IBM MQ web console | `https://mq.localhost:30443/ibmmq/console/` |
| IBM MQ admin REST | `https://mq.localhost:30443/ibmmq/rest/v3/admin/qmgr` |
| Grafana (if monitoring enabled) | `https://grafana.localhost:30443/` |
| In-cluster mqweb (for `QueueManagerConnection.endpoint`) | `https://ibm-mq.ibm-mq.svc:9443` |

Defaults (override via Terraform variables): Queue Manager `QM1`, MQ admin user
`admin`, namespace `ibm-mq`, credentials in the `mq-credentials` Secret. These
are **local-dev defaults only** — never reuse them anywhere real.

### Deploy the operator

```sh
task deploy        # build image, load into kind, apply CRDs + manager
# point a QueueManagerConnection at https://ibm-mq.ibm-mq.svc:9443
kubectl apply -k config/samples
```

### Tear down

```sh
task cluster:down                      # delete the kind cluster
# or, keep the cluster but remove provisioned resources:
cd hack/kind-cluster && ./scripts/cleanup.sh
DELETE_CLUSTER=true ./scripts/cleanup.sh   # also delete the cluster
```

## Test tiers

| Tier | Scope | Needs a cluster? | Command |
|------|-------|------------------|---------|
| **Unit** | Reconciler logic + REST adapter vs mocks / `httptest` | No | `task test:run` |
| **envtest** | Controller + API against a real API server (`setup-envtest`), `MQAdmin` mocked | No (downloads control-plane binaries) | `task test:run` |
| **e2e** | Operator in kind against live IBM MQ; asserts real MQSC | Yes (`task cluster:up`) | `task test:e2e` |

Guidelines:

- Unit + envtest must stay fast and hermetic; mock the `MQAdmin` port, never hit
  a real Queue Manager.
- e2e is gated behind a build tag (`//go:build e2e`) so it does not run in the
  default `go test ./...`.
- Keep coverage high on `internal/`; CI reports it.

## Troubleshooting

- **kind can't start / wrong runtime**: the scripts auto-detect docker →
  nerdctl → podman; override with `KIND_EXPERIMENTAL_PROVIDER`.
- **TLS not trusted in browser**: run `mkcert -install` once in an interactive
  shell, then re-run `./scripts/mkcert-gen.sh`.
- **IBM MQ pod slow to start**: the chart waits up to ~15 min; check
  `kubectl -n ibm-mq get pods` and `kubectl -n ibm-mq logs`.
- **mqweb 401/403**: confirm the `mq-credentials` Secret and the `MQWebAdmin`
  role mapping; see [IBM_MQ_REST_API.md](IBM_MQ_REST_API.md).
- **envtest binaries missing**: `task test:run` provisions them via
  `setup-envtest`; ensure network access on first run.

## Before you push

1. `task verify` — generated artifacts are fresh.
2. `task lint` — clean.
3. `task test:run` — green.
4. Conventional Commit with a gitmoji (see [../AGENTS.md](../AGENTS.md)).
