# Roadmap

Phased delivery plan for the IBM Message Queue Operator. Each phase is shippable
on its own and keeps the tree green (build + lint + tests pass). See
[ARCHITECTURE.md](ARCHITECTURE.md) for design and [../AGENTS.md](../AGENTS.md)
for conventions.

## Guiding principles

- Small, atomic, fully-tested increments over big drops.
- Every external interaction sits behind the `MQAdmin` port so it can be mocked.
- The build stays pure Go (`CGO_ENABLED=0`); no native MQ client.

## Phase 0 — Foundations (this step)

- [x] `AGENTS.md` with context and conventions.
- [x] `docs/ARCHITECTURE.md` and `docs/ROADMAP.md`.
- [x] `README.md`.

## Phase 1 — Scaffold & toolchain

- Confirm module path and API group, then scaffold with **Kubebuilder**
  (manager entrypoint, `PROJECT`, empty `api/v1alpha1`, `internal/controller`).
- `Taskfile.yml` + `Taskfile.test.yml` (install, format, lint, manifests,
  generate, build, docker:build, kind:up/down, deploy/undeploy, test:run,
  test:e2e).
- `.golangci.yaml` (v2, linter set per AGENTS.md), `.mockery.yaml`,
  `.pre-commit-config.yaml`, `.editorconfig`, `Dockerfile`.
- GitHub Actions skeleton: lint + unit tests + `govulncheck` on PRs.

Exit criteria: `task build`, `task lint`, and an empty `task test:run` pass
locally and in CI.

## Phase 2 — Core API, adapter & tests

- `api/v1alpha1`: `QueueManagerConnection` and `Queue` types + generated
  deepcopy and CRD manifests; basic validation (kubebuilder markers).
- `internal/mqadmin`: the `MQAdmin` port and domain types.
- `internal/adapter/mqrest`: `mqweb` REST client implementing `MQAdmin`
  (define/inspect/delete queue, ping), with `httptest`-based unit tests.
- `internal/controller`: thin reconcilers for both resources — finalizers,
  drift detection, status conditions (`Ready`, `Synced`, `observedGeneration`).
- Tests: mockery mocks of `MQAdmin`, unit tests for reconcilers, envtest for
  API/controller integration. Maintain high coverage on `internal/`.

Exit criteria: applying samples in envtest drives the expected `MQAdmin` calls;
adapter unit tests cover success + error paths.

## Phase 3 — End-to-end & CI hardening

- e2e suite (`test/e2e`) on **kind** against a real IBM MQ container exposing
  `mqweb`; assert real MQSC objects for create/update/delete.
- Wire e2e into CI (kind in GitHub Actions) on a dedicated job.
- Image publishing workflow; SBOM/vuln scanning as desired.

Exit criteria: `task test:e2e` green locally and in CI against a live Queue
Manager container.

## Phase 4 — User & authority management

- Extend the API toward MQ access control: authority records / channel auth /
  user-style resources (exact CRDs decided when reached).
- Corresponding `MQAdmin` operations, adapter support, and tests at all layers.

## Later / candidate work

- Additional object types (`Topic`, `Channel`, alias/remote queues).
- Optional PCF adapter behind the existing `MQAdmin` port for environments
  without `mqweb`.
- Metrics/dashboards and richer status reporting.
- Documentation site and published install manifests.
