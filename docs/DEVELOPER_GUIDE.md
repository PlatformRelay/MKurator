# Developer guide

Day-to-day **what to regenerate** and **what to test** when you change APIs,
reconcilers, or the MQ boundary. For install, Task targets, and local clusters
see [DEVELOPMENT.md](DEVELOPMENT.md); for system design see
[ARCHITECTURE.md](ARCHITECTURE.md).

| | Deep dive |
|---|-----------|
| 🧩 | [GO_MODULE.md](GO_MODULE.md) — module layout, import layers, `go-arch-lint` |
| ⚙️ | [OPERATOR_RUNTIME.md](OPERATOR_RUNTIME.md) — manager wiring, probes, concurrency |

Doc index: [README.md](index.md) · Agent entry: [../AGENTS.md](https://github.com/conduit-ops/MKurator/blob/main/AGENTS.md)

## On this page

| | Section |
|---|---------|
| 📐 | [I changed a CRD field](#i-changed-a-crd-field--what-do-i-regenerate) |
| 🔄 | [I changed reconcile logic](#i-changed-reconcile-logic--what-tests) |
| 🔌 | [I extended the MQAdmin port](#i-extended-the-mqadmin-port--what-tests) |
| 🎭 | [Mocks and the MQAdmin port](#mocks-and-the-mqadmin-port) |
| ✅ | [Before you commit](#before-you-commit) |

---

## I changed a CRD field — what do I regenerate?

Edit types and kubebuilder markers in `api/v1alpha1/`, then run codegen in this
order (or rely on `task verify`, which runs the same steps and fails on drift):

| Step | Command | Produces / updates |
|------|---------|-------------------|
| 1 | `task generate` | `api/*/zz_generated.deepcopy.go`; **mockery** mocks if `.mockery.yaml` lists interfaces |
| 2 | `task manifests` | `config/crd/bases/*.yaml`, `config/rbac/role.yaml`, webhook manifests under `config/` |
| 3 | *(automatic in `task verify`)* | `hack/helm-sync-crds.sh` → `charts/mkurator/crds/` |
| 4 | If CR **spec** OpenAPI changed | `task test:schema:update` → `test/schema/golden/*.spec.openapi.yaml` |
| 5 | If `config/samples/` CR YAML changed | `task samples:sync` → `charts/mkurator/samples/resources/` |

One-liner after API edits:

```sh
task generate && task manifests && task test:schema:update
```

Then `task verify` before commit (also runs schema contract tests).

### Also update by hand (not generated)

| Change | Where |
|--------|--------|
| Admission rules for new/renamed fields | `internal/validation/` + `*_test.go` (table-driven) |
| Webhook handler (thin) | `internal/webhook/v1alpha1/` — delegates to validation |
| Reconciler mapping CR spec → port types | `internal/controller/*_controller.go` |
| Drift / DEFINE vs DISPLAY policy | [ATTRIBUTE_RECONCILIATION.md](ATTRIBUTE_RECONCILIATION.md) and reconciler helpers |
| User-facing samples | `config/samples/`, [config/samples/README.md](https://github.com/conduit-ops/MKurator/blob/main/config/samples/README.md) |
| Helm-only sample files | `charts/mkurator/samples/resources/` (README, Secret template) |

OpenAPI **contract** tests live in [`test/schema/`](https://github.com/conduit-ops/MKurator/blob/main/test/schema/README.md): they
diff committed CRD YAML against golden spec fragments — no cluster. Adding a new
CR kind requires a row in `test/schema/extract.go` plus a golden file.

---

## I changed reconcile logic — what tests?

Pick the **lowest tier** that exercises the behaviour ([ADR-0011](adr/0011-layered-testing-strategy.md)).

| You changed | Add or extend |
|-------------|----------------|
| Pure helpers (`needsUpdate`, status, events, shared wait/enqueue) | `internal/controller/*_test.go` — stdlib `testing`, `t.Parallel()` where safe |
| A reconciler path (define, drift, delete, requeue, finalizer) | `internal/controller/*_unit_test.go` — `fake` client + **mock `Admin`**; or Ginkgo `*_reconciler_test.go` with **envtest** + mocks |
| Controller registration / wiring | `wiring_envtest_test.go`, `events_envtest_test.go` |
| Validation only | `internal/validation/*_test.go` — no cluster, no MQ |
| Admission webhook wiring | `internal/webhook/v1alpha1/` envtest suite |

**Default PR loop** (no Queue Manager):

```sh
task test:run    # unit + envtest, -race, internal/ coverage floor
```

Envtest loads CRDs from `config/crd/bases` (`internal/controller/suite_test.go`).
It uses a real Kubernetes API server from `setup-envtest`; **MQ is always mocked**.

### File naming convention

| Pattern | Tier | Typical use |
|---------|------|-------------|
| `*_unit_test.go` | Unit | Single reconciler with `client/fake` + `mqadmintest.NewMockAdmin` |
| `*_reconciler_test.go` | envtest | Ginkgo specs creating CRs via `k8sClient`, `MQFactory` mocked |
| `*_envtest_test.go` | envtest | Cross-cutting (events, wiring) |
| `reconcile_shared_test.go` | Unit | Shared reconcile helpers |

Do **not** call real mqweb from `internal/controller` tests.

---

## I extended the MQAdmin port — what tests?

The port is `internal/mqadmin.Admin` and `Factory` ([ARCHITECTURE.md](ARCHITECTURE.md#the-mqadmin-port)).
Reconcilers must use typed methods only — not raw `RunMQSC` ([ADR-0014](adr/0014-mq-error-taxonomy-and-requeue.md)).

| Step | Action |
|------|--------|
| 1 | Add method(s) and domain types on `internal/mqadmin/admin.go` |
| 2 | `task generate` (runs `task test:generate` → mockery → `test/mocks/mqadmin/`) |
| 3 | Implement in `internal/adapter/mqrest/` |
| 4 | **Unit**: `internal/adapter/mqrest/*_test.go` with `httptest` server where HTTP/MQSC parsing matters |
| 5 | **Integration**: `test/integration/mq/*_integration_test.go` — `//go:build integration`, `task test:integration` |
| 6 | Reconciler unit/envtest: expect new mock calls in controller tests |
| 7 | **e2e** (optional, slow): kind + live MQ — only when end-to-end install path must be proven; see [DEVELOPMENT.md](DEVELOPMENT.md#test-tiers) |

Integration is gated by `KURATOR_INTEGRATION_MQ=1` (set by `task test:integration`).
It does not run in `task test:run` or default `go test ./...`.

---

## Mocks and the MQAdmin port

### Where mocks live

| Path | Contents |
|------|----------|
| [`.mockery.yaml`](https://github.com/conduit-ops/MKurator/blob/main/.mockery.yaml) | mockery v3 config (`template: testify`) |
| [`test/mocks/mqadmin/`](https://github.com/conduit-ops/MKurator/tree/main/test/mocks/mqadmin) | Generated `MockAdmin`, `MockFactory` — **do not edit by hand** |
| [`internal/mqadmin/`](https://github.com/conduit-ops/MKurator/tree/main/internal/mqadmin) | Real `Admin` / `Factory` interfaces and domain errors |

Regenerate mocks:

```sh
task test:generate    # or: task generate (includes mockery)
```

`task verify` diffs `test/mocks/` after regeneration; drift fails CI.

### How reconcilers use the port

Production code obtains an `Admin` per connection:

```text
Reconciler.MQFactory.ForConnection(ctx, qmc) → (Admin, error)
```

Tests inject mocks at the **factory** boundary (same field name on reconcilers):

1. `mockFactory := mqadmintest.NewMockFactory(t)` (or `GinkgoT()`).
2. `mockFactory.EXPECT().ForConnection(...).Return(mockAdmin, nil)`.
3. `mockAdmin := mqadmintest.NewMockAdmin(t)`.
4. `mockAdmin.EXPECT().DefineQueue(...).Return(nil)` (etc.).

Example (envtest Ginkgo): [`internal/controller/queue_reconciler_test.go`](https://github.com/conduit-ops/MKurator/blob/main/internal/controller/queue_reconciler_test.go).

Example (stdlib unit): [`internal/controller/queue_controller_unit_test.go`](https://github.com/conduit-ops/MKurator/blob/main/internal/controller/queue_controller_unit_test.go).

`NewMockFactory` / `NewMockAdmin` register `t.Cleanup` to assert expectations.
Use `github.com/stretchr/testify/mock` matchers (`mock.Anything`, typed `QueueSpec`, …).

### REST adapter tests without mockery

`internal/adapter/mqrest` unit tests often use an **`httptest.Server`** to fake
mqweb responses — no Kubernetes, no mockery. Use integration tier when behaviour
depends on a real IBM MQ container ([`hack/mq-docker/README.md`](https://github.com/conduit-ops/MKurator/blob/main/hack/mq-docker/README.md)).

---

## Before you commit

```sh
task verify      # deepcopy, CRDs, RBAC, mocks, Helm CRDs, samples, schema goldens
task lint
task test:run
```

See [DEVELOPMENT.md — Before you push](DEVELOPMENT.md#before-you-push) and
[CONTRIBUTING.md](CONTRIBUTING.md) for commit format.

---

## Related docs

| Document | Use when |
|----------|----------|
| [DEVELOPMENT.md](DEVELOPMENT.md) | Local stack, Task reference, integration/e2e env vars |
| [CICD.md](CICD.md) | What CI runs on PRs vs scheduled jobs |
| [GO_MODULE.md](GO_MODULE.md) | Package boundaries and allowed imports |
| [OPERATOR_RUNTIME.md](OPERATOR_RUNTIME.md) | Manager, leader election, health, metrics |
| [test/schema/README.md](https://github.com/conduit-ops/MKurator/blob/main/test/schema/README.md) | CRD OpenAPI golden workflow |
