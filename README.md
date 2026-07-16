<p align="center">
  <img src="docs/images/mkurator-logo.png" alt="MKurator logo" width="200">
</p>

# MKurator

[![CI](https://github.com/platformrelay/MKurator/actions/workflows/ci.yaml/badge.svg)](https://github.com/platformrelay/MKurator/actions/workflows/ci.yaml)
[![Preflight](https://github.com/platformrelay/MKurator/actions/workflows/preflight.yaml/badge.svg)](https://github.com/platformrelay/MKurator/actions/workflows/preflight.yaml)
[![E2E](https://github.com/platformrelay/MKurator/actions/workflows/e2e.yaml/badge.svg)](https://github.com/platformrelay/MKurator/actions/workflows/e2e.yaml)
[![Documentation](https://img.shields.io/badge/documentation-GitHub%20Pages-2ea44f?logo=readthedocs&logoColor=white)](https://platformrelay.github.io/MKurator/)
[![Docs CI](https://github.com/platformrelay/MKurator/actions/workflows/docs.yaml/badge.svg)](https://github.com/platformrelay/MKurator/actions/workflows/docs.yaml)
[![CodeQL](https://github.com/platformrelay/MKurator/actions/workflows/codeql.yaml/badge.svg)](https://github.com/platformrelay/MKurator/actions/workflows/codeql.yaml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/platformrelay/MKurator/badge)](https://securityscorecards.dev/viewer/?uri=github.com/platformrelay/MKurator)
[![License: MIT](https://img.shields.io/github/license/platformrelay/MKurator)](https://github.com/platformrelay/MKurator/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/platformrelay/MKurator)](https://github.com/platformrelay/MKurator/releases)
[![codecov](https://codecov.io/gh/platformrelay/MKurator/graph/badge.svg)](https://codecov.io/gh/platformrelay/MKurator)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=PlatformRelay_MKurator&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=PlatformRelay_MKurator)
[![Go](https://img.shields.io/github/go-mod/go-version/platformrelay/mkurator)](https://pkg.go.dev/github.com/platformrelay/mkurator)
[![Go Reference](https://pkg.go.dev/badge/github.com/platformrelay/MKurator.svg)](https://pkg.go.dev/github.com/platformrelay/mkurator)
[![Go Report Card](https://goreportcard.com/badge/github.com/platformrelay/mkurator)](https://goreportcard.com/report/github.com/platformrelay/mkurator)
[![Container](https://img.shields.io/badge/ghcr.io-platformrelay%2Fmkurator-2496ED?logo=docker&logoColor=white)](https://github.com/orgs/platformrelay/packages?repo_name=MKurator)

A Kubernetes operator for declaratively managing **resources on an existing
IBM MQ Queue Manager** — queues, topics, SVRCONN channels; users/authorities and
more later.

**Full documentation:** [platformrelay.github.io/MKurator](https://platformrelay.github.io/MKurator/) — install
guides, examples, ADRs, and engineering standards.

> Status: **`v1beta1` API graduated** (v0.12.0) — all six kinds serve both
> `v1alpha1` and `v1beta1` behind a conversion webhook. **Phase 9 breadth on
> `main`** — SDR/RCVR channels, AUTHREC channel/namelist profiles, DISPLAY
> capability probing for local queues, with Docker integration and kind e2e
> coverage. Latest release: **`v0.13.0`**. Extended CHLAUTH rule types remain in
> the [roadmap](docs/ROADMAP.md#phase-5--user--authority-management).

## What ships today

| Custom resource | MQ objects | Notes |
|-----------------|------------|-------|
| `QueueManagerConnection` | (connectivity) | Ping + credentials from a referenced `Secret` |
| `Queue` | `QLOCAL`, `QALIAS`, `QREMOTE` | `spec.type`: `local` (default), `alias`, `remote` |
| `Topic` | `TOPIC` | Drift-checked attributes per [ATTRIBUTE_RECONCILIATION.md](docs/ATTRIBUTE_RECONCILIATION.md) |
| `Channel` | `CHANNEL` … `CHLTYPE(SVRCONN)` | Other channel types planned later |
| `ChannelAuthRule` | `CHLAUTH` | `ADDRESSMAP` exercised in kind e2e; `BLOCKUSER` in Docker integration; `USERMAP`, `SSLPEERMAP`, `QMGRMAP`, `BLOCKADDR` accepted by schema and admission, MQ-validated at apply time |
| `AuthorityRecord` | `SET AUTHREC` (OAM) | Queue profile + principal/group authorities |

**API versions:** all six kinds serve **`v1alpha1`** and **`v1beta1`** behind a
conversion webhook; new manifests should use `v1beta1`. **Access control** covers
`SET CHLAUTH` (one rule per CR) and `SET AUTHREC` for queue/channel-style
profiles; see [PHASE5_AUTH_SKETCH.md](docs/PHASE5_AUTH_SKETCH.md) for the
rule-type roadmap. **API stability:**
[docs/API_STABILITY.md](docs/API_STABILITY.md) (per-version guarantees and the
`v1beta1` graduation path).

**Repository:** [github.com/platformrelay/MKurator](https://github.com/platformrelay/MKurator) — Go module
[`github.com/platformrelay/MKurator`](https://pkg.go.dev/github.com/platformrelay/mkurator), images
`ghcr.io/platformrelay/mkurator` ([ADR-0006](docs/adr/0006-project-name-kurator.md), superseded by [ADR-0018](docs/adr/0018-project-rename-mkurator.md)). Your
local clone directory may differ from the module/repo name (for example
`IBM-Message-Queue-Operator`).

### What CI proves

| Tier | Scope |
|------|-------|
| Unit + envtest | Reconcilers and adapter (mocked MQ); validating admission; Queue, Topic, Channel, auth CRs, QMC |
| Docker integration | Queue, Topic, Channel, AUTHREC against live mqweb; CHLAUTH **`ADDRESSMAP`** (GET, replace, delete) and **`BLOCKUSER`** (GET) |
| kind e2e (`KURATOR_E2E_MQ=1`) | Queue, Topic, Channel, AuthorityRecord reconcile + delete; CHLAUTH **`ADDRESSMAP`** and **`BLOCKUSER`** `ChannelAuthRule` reconcile + delete on live `QM1` |

Details and commands: [DEVELOPMENT.md#test-tiers](docs/DEVELOPMENT.md#test-tiers).

Latest tagged release: [GitHub Releases](https://github.com/platformrelay/MKurator/releases)
(current badge above). `main` may include fixes not yet in a tag. See
[CHANGELOG.md](CHANGELOG.md) for version history (generated from Conventional Commits).

## What it does

- Reconciles custom resources (`Queue`, `Topic`, `Channel`, `ChannelAuthRule`,
  `AuthorityRecord`) into MQSC objects on a running Queue Manager.
- Talks to the Queue Manager through the **IBM MQ Administrative REST API**
  (`mqweb`) over HTTPS — pure Go, no CGO.
- Reports status via conditions and cleans up via finalizers.

It does **not** deploy or operate Queue Manager installations; the Queue
Manager is assumed to already exist and expose `mqweb`.

## How it works

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/assets/reconcile-flow-dark.svg">
  <img src="docs/assets/reconcile-flow-light.svg" alt="CR reconciler to mqweb to queue manager" width="920">
</picture>

You declare desired MQ objects as Kubernetes custom resources. The operator watches those CRs,
validates specs at admission, and reconcilers call the **mqrest** adapter to execute MQSC through
**mqweb** on your queue manager. Status conditions report success, drift correction, and errors.
Full design: [ARCHITECTURE.md](docs/ARCHITECTURE.md) · published docs:
[platformrelay.github.io/MKurator](https://platformrelay.github.io/MKurator/).

## Repository structure

[Kubebuilder v4](https://book.kubebuilder.io/) layout — thin reconcilers, an
[`MQAdmin`](internal/mqadmin) port, and an [`mqweb`](internal/adapter/mqrest)
adapter. Full design: [ARCHITECTURE.md](docs/ARCHITECTURE.md) · extended map:
[AGENTS.md](AGENTS.md#repository-layout).

```text
mkurator/
├── 📦 api/{v1alpha1,v1beta1}/       CRD types, deepcopy + conversion (QMC, Queue, Topic, Channel, auth)
├── 🚀 cmd/                          Manager entrypoint (controller-runtime)
├── 🧠 internal/
│   ├── controller/                  Reconcilers (thin) + unit/envtest suites
│   ├── validation/                  Admission validation rules (pure functions)
│   ├── webhook/{v1alpha1,v1beta1}/  Validating webhook handlers (per version)
│   ├── webhook/conversion/          v1alpha1 ↔ v1beta1 conversion webhook
│   ├── mqadmin/                     MQAdmin port — interface + domain errors
│   ├── adapter/mqrest/              mqweb REST client (sole adapter today)
│   ├── logging/                     Structured logging helpers
│   └── metrics/                     Prometheus metrics
├── ⚙️  config/                       Kustomize — CRDs, RBAC, manager, webhook, samples
├── ⎈  charts/mkurator/                Publishable Helm chart + sample CRs
├── 🧪 test/
│   ├── integration/                 Docker MQ tests (build tag `integration`)
│   ├── e2e/                         kind + live QM1 (build tag `e2e`)
│   └── mocks/                       mockery-generated MQAdmin mocks
├── 🔧 hack/
│   ├── kind-cluster/                Local platform: kind + Terraform + IBM MQ Helm
│   ├── mq-docker/                   Standalone IBM MQ container for integration CI
│   └── *.sh                         verify, release assets, tool install helpers
├── 📚 docs/                         Guides, ADRs, MQ research (see docs/README.md)
├── Taskfile.yml                     Primary task runner (`task local:up`, …)
└── AGENTS.md                        Go conventions + agent entry point
```

## Install and use

**Start here:** [docs/INSTALL_AND_USE.md](docs/INSTALL_AND_USE.md) — install the
operator (Release manifests, Helm), connect to your queue manager, manage queues,
[kubectl diagnostics](docs/INSTALL_AND_USE.md#diagnostics-and-troubleshooting),
and uninstall.

Sample YAML with annotations:
[config/samples/README.md](config/samples/README.md).

```sh
# After task deploy:helm or task local:up — preferred one-shot sample apply:
task deploy:samples
kubectl get qmc,mq,tp,chl,car,auth -n mkurator-system
```

**`task deploy:samples`** is the supported path on kind: it ensures the
`mkurator-system` namespace exists and server-side-applies
`charts/mkurator/samples/resources/` (Secret + all sample CRs). Annotated reference
YAML lives under `config/samples/` — edit there, then `task samples:sync`.

## Local development (contributors)

**Tool install:** [docs/LOCAL_SETUP.md](docs/LOCAL_SETUP.md) — Go, Task, Docker, kind,
Terraform, and verification by tier.

**Canonical reference:** [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) — inner loop,
local platform, task reference, test tiers, URLs, and credentials.

```sh
task local:up      # cluster + IBM MQ + operator (Helm) + sample CRs
task local:info    # URLs, credentials, CR status
task local:down    # tear everything down
```

Verify reconciliation with [docs/IBM_MQ_101.md](docs/IBM_MQ_101.md) (`runmqsc`, MQ console).

## Documentation

**Published site:** [platformrelay.github.io/MKurator](https://platformrelay.github.io/MKurator/) — install guides,
examples, ADRs, and engineering standards.

| | Doc |
|---|-----|
| 🎯 **Use MKurator** | [Quick start](docs/QUICKSTART.md) · [Install and use](docs/INSTALL_AND_USE.md) · [FAQ](docs/FAQ.md) · [Glossary](docs/GLOSSARY.md) |
| 🛠️ **Develop** | [Development guide](docs/DEVELOPMENT.md) · [Contributing](CONTRIBUTING.md) · [CI/CD](docs/CICD.md) |
| 🏗️ **Design** | [Architecture](docs/ARCHITECTURE.md) · [ADRs](docs/adr/) · [Changelog](CHANGELOG.md) |

Full index: **[docs/README.md](docs/README.md)**.

## Community

| | Link |
|---|------|
| Contributing | [CONTRIBUTING.md](CONTRIBUTING.md) |
| Code of Conduct | [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) |
| Governance | [GOVERNANCE.md](GOVERNANCE.md) |
| Security | [SECURITY.md](SECURITY.md) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |

## License

MIT — see [LICENSE](LICENSE).
