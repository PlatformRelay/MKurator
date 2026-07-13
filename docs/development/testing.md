# Testing strategy

Test pyramid, coverage floors, and CI merge gates. Detailed tier commands:
[DEVELOPMENT.md](../DEVELOPMENT.md#test-tiers). Authoritative ADR:
[ADR-0011](../adr/0011-layered-testing-strategy.md); merge-gate matrix in [ADR-0020](../adr/0020-merge-gate-matrix.md).

## L0–L5 pyramid

| Tier | Scope | Blocks merge? | Command |
| --- | --- | --- | --- |
| **L0** | Unit — reconcilers, mqrest adapter, validation (mocks / httptest) | Yes | `task test:run` (unit packages) |
| **L1** | envtest — controllers + API server + admission | Yes | `task test:run` (Ginkgo suites) |
| **L2** | Schema / contract — golden OpenAPI fragments | Yes | `task test:schema` (inside `task verify`) |
| **L3** | Integration — live mqweb vs Docker IBM MQ | Yes (path-filtered CI) | `task test:integration` |
| **L4** | e2e — operator on kind + real QM | Optional on PR; recommended on `main` | `task test:e2e` |
| **L5** | Bench / soak / fuzz | Opt-in | Manual / future CI |

Mkurator maps the historical **four tiers** (unit → envtest → integration → e2e) to L0–L4; L2 is the
schema golden layer; L5 is reserved for performance and fuzz work.

## Coverage

- **Floors** (both enforced in CI via `Taskfile.test.yml`, each with its own coverprofile):
  - `./internal/...` — ~**91%** statement coverage (`coverage.out`).
  - `./api/...` — **75%** statement coverage (`coverage-api.out`), covering the hub CEL
    validation, v1alpha1↔v1beta1 conversion, attribute folding, and generated deepcopy.
- Coverage regressions are investigated — not silently ignored.
- Race detector (`-race`) on L0/L1 in CI.

## Machine lock

`task test:integration`, `task test:e2e`, and `task ci:e2e` share
`hack/kind-cluster/.state/locks/exclusive-test.lock` — one MQ suite at a time per host.

## What to run when

| Change touches | Minimum tier |
| --- | --- |
| Reconcile logic, conditions | L1 envtest |
| mqrest adapter | L0 unit + L3 integration |
| CRD OpenAPI / markers | L2 schema + `task verify` |
| Install / Helm path | L4 e2e when feasible |
| Docs only | preflight + CI (no integration/e2e) |

## CI workflow mapping

| Workflow | Tiers |
| --- | --- |
| `ci.yaml` `test` | L0, L1, govulncheck |
| `ci.yaml` `verify` | L2 |
| `integration.yaml` | L3 |
| `e2e.yaml` | L4 |
| `nightly.yaml` | L3 + L4 (extended) |

See [CICD.md](../CICD.md) for path filters and required checks.

## Related documents

| Document | Owns |
| --- | --- |
| [DEVELOPER_GUIDE.md](../DEVELOPER_GUIDE.md) | Change matrix — what to regenerate |
| [guidelines.md](guidelines.md) | Definition of done |
| [test/schema/README.md](https://github.com/platformrelay/MKurator/blob/main/test/schema/README.md) | Golden OpenAPI contracts |
