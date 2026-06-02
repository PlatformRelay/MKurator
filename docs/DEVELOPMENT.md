# Development

How to set up, build, test, and run **Kurator** locally. For
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

### Codegen verification (`hack/verify.sh`)

`task verify` runs `hack/verify.sh`, which implements the **generate / verify**
discipline from [CICD.md](CICD.md):

1. Snapshot committed generated artifacts (`config/crd/bases`, `config/rbac`,
   `api/*/zz_generated.deepcopy.go`).
2. Regenerate with `controller-gen`.
3. `diff` snapshot vs working tree and fail on drift.

This catches the common mistake of editing API types or kubebuilder markers
without re-running `task generate && task manifests`. It is unrelated to
`go mod verify` (module checksum verification in `task install`).

### Task vs Makefile

**Task is the canonical entry point** ([ADR-0004](adr/0004-task-as-task-runner.md)):
humans, pre-commit, and CI all run `task <target>`. The root `Makefile` is
**Kubebuilder scaffold** — it ships with `kubebuilder init` and is still used
by the default e2e suite (`make docker-build`, `make deploy`). Prefer `task`
for day-to-day work; ignore `make` unless you are running that scaffold as-is.
A future cleanup can rewire e2e to `task deploy` and trim the Makefile.

Build the manager binary (CGO-free, static):

```sh
task build
```

### Logging

Logging is configured via YAML file, environment variables, or flags (see
[LOGGING.md](LOGGING.md)). Quick local examples:

```sh
# Human-readable logs on stderr (default when not in a pod)
go run ./cmd/main.go --log-format=text --log-level=debug

# JSON to stdout (production-style)
go run ./cmd/main.go --log-format=json --log-level=info

# File-based config
export KURATOR_LOG_CONFIG=config/samples/logging-config.yaml
go run ./cmd/main.go
```

In the cluster, the manager Deployment sets `KURATOR_LOG_FORMAT=json` and
`KURATOR_LOG_LEVEL=info` by default.

## Local platform (kind + IBM MQ)

The `hack/kind-cluster` tree provisions a complete environment: a kind cluster
with **ingress-nginx**, **cert-manager**, an optional **kube-prometheus-stack**,
and a real **IBM MQ** Queue Manager exposing `mqweb` — wired with **Terraform**
and trusted TLS from **mkcert**.

Cluster name: `kurator` (override with `CLUSTER_NAME` if you have an existing
`ibm-mq-operator` cluster from before the rename). State (kubeconfig, TLS) is
written to `hack/kind-cluster/.state/`.

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

## Deploying a queue manager for Kurator

Kurator requires an **existing** queue manager with **mqweb** enabled. It does not
install or upgrade Queue Managers. Choose one of the options below; then point a
`QueueManagerConnection` at the in-cluster mqweb URL (or a reachable equivalent).

See also [REFERENCES.md.example](REFERENCES.md.example) (copy to gitignored
`docs/REFERENCES.md` when using a local `references/` clone).

### Option A — mq-helm on kind (recommended for local dev)

This is what `hack/kind-cluster` provisions: the upstream
[ibm-messaging/mq-helm](https://github.com/ibm-messaging/mq-helm) chart with
`web.enable: true` and a `mq-credentials` Secret (`mqAdminPassword` / `mqAppPassword`).

```sh
task cluster:up
# In-cluster endpoint for QueueManagerConnection:
#   https://ibm-mq.ibm-mq.svc:9443
kubectl apply -k config/samples
```

The operator pod resolves cluster DNS; your laptop can use the ingress URLs from
`task cluster:info` for manual REST/console checks.

### Option B — IBM MQ Operator (OpenShift or EKS preview)

Use when the queue manager is already managed by IBM’s operator. Kurator only needs
mqweb credentials and network reachability from the operator namespace.

**OpenShift:** install the IBM Operator Catalog (`icr.io/cpopen/ibm-operator-catalog`)
and the **IBM MQ** operator from OperatorHub. See
[ibm-messaging/mq-gitops-samples](https://github.com/ibm-messaging/mq-gitops-samples/tree/main/queue-manager-basic-deployment) (local clone under `references/`, gitignored).

**Amazon EKS (preview):** Helm chart and CRD in
[ibm-messaging/mq-operator-eks-preview-2025](https://github.com/ibm-messaging/mq-operator-eks-preview-2025); no
controller source is published.

Minimum `QueueManager` fields for mqweb (adapt namespace/license as required):

| Field | Purpose |
|-------|---------|
| `spec.web.enabled: true` | Enables IBM MQ Console and REST APIs on the QM pod |
| `spec.web.console.authentication.provider` / `authorization.provider` | e.g. `manual` for basic registry (see gitops `qmdemo-qm.yaml`) |
| `spec.pki.keys` / `spec.pki.trust` | TLS material for the QM pod (often from cert-manager Secrets) |
| `spec.queueManager.name` | QM name — must match `QueueManagerConnection.spec.queueManager` |
| `spec.queueManager.mqsc` | Optional bootstrap MQSC via ConfigMap (channels, CHLAUTH at install). Kurator reconciles **additional** objects later via CRs. |

On **EKS**, disable OpenShift-only routes in the `QueueManager` spec (see
[Ingress for IBM MQ Console and REST APIs](https://github.com/ibm-messaging/mq-operator-eks-preview-2025/blob/main/configuring_Ingress_and_LoadBalancers/Ingress_for_IBM_MQ_Console_and_REST_APIs.md)):

- `spec.web.route.enabled: false`
- `spec.queueManager.route.enabled: false`
- `spec.queueManager.metrics.serviceMonitor.enabled: false` (unless you run Prometheus Operator)

Create a Kubernetes Secret with mqweb admin credentials. Kurator accepts `username` +
`password` or `mqAdminPassword` (see `internal/adapter/mqrest/factory.go`).

Example `QueueManagerConnection` (same as [samples](../config/samples/)):

```yaml
spec:
  queueManager: QM1   # or your spec.queueManager.name
  endpoint: https://<qm-service>.<namespace>.svc:9443
  tls:
    caSecretRef:
      name: <ca-secret>   # omit insecureSkipVerify in production
  credentialsSecretRef:
    name: mq-credentials
```

### Option C — Other environments

Any queue manager (VM, container, Cloud Pak) works if mqweb is reachable and
admin credentials are in a referenced Secret. Use [IBM_MQ_REST_API.md](IBM_MQ_REST_API.md)
for CSRF, TLS, and MQSC endpoint details.

### Deploy the operator

**Helm (recommended for kind):**

```sh
task deploy:helm     # build image, load into kind, helm install charts/kurator
task deploy:samples  # Secret + QueueManagerConnection + Queue for QM1
```

**Kustomize (Kubebuilder default):**

```sh
task deploy          # build image, load into kind, apply CRDs + manager
kubectl apply -k config/samples
```

Both paths target namespace `kurator-system` and expect mqweb at
`https://ibm-mq.ibm-mq.svc:9443` after `task cluster:up`. Chart details:
[charts/kurator/README.md](../charts/kurator/README.md).

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

**IBM MQ e2e scenarios** (queue reconcile, channel/auth fixtures) run only when
`KURATOR_E2E_MQ=1` is set and the kind platform with IBM MQ is up. Without that,
the scaffold e2e suite (controller pod, metrics) still runs. MQ-specific tests use
defaults aligned with `hack/kind-cluster` (`QM1`, `admin` / `passw0rd`, endpoint
`https://ibm-mq.ibm-mq.svc:9443`). Override with `KURATOR_E2E_MQ_*` env vars
documented in [`test/e2e/fixtures/README.md`](../test/e2e/fixtures/README.md).

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
