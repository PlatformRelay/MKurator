# Integration test tier audit

**Status:** Audit complete (2026-06-03).  
**Audience:** Maintainers and agents extending the MQ adapter or test tiers.  
**Related:** [DEVELOPMENT.md](../DEVELOPMENT.md) · [CICD.md](../CICD.md) · [ADR-0011](../adr/0011-layered-testing-strategy.md)

---

## Summary

Integration tests **run in CI** via a dedicated workflow (not `ci.yaml`). They exercise the **`mqadmin.Admin` / `mqrest` adapter** against live mqweb in Docker IBM MQ — **25 tests** in `test/integration/mq/`, adapter-only scope. Main blindspots: auth **delete/update** paths are not explicitly asserted, **alias/remote** queues are integration-only (not e2e), and **drift/reconcilers** are intentionally out of scope.

---

## 1. CI wiring

| Item | Detail |
|------|--------|
| **Runs in CI?** | **Yes** |
| **Workflow** | [`.github/workflows/integration.yaml`](../../.github/workflows/integration.yaml) — **separate from** [`ci.yaml`](../../.github/workflows/ci.yaml) |
| **Triggers** | `pull_request`, `push` to `main`, `workflow_dispatch` |
| **Path filters** | **Skipped** when changes are only: `**.md`, `docs/**`, `charts/**/README.md` |
| **Branch filter** | `push` limited to `main`; PRs have no branch restriction |
| **Steps** | `task mq:integration:up` → `task mq:integration:wait` → `task test:integration` → `task mq:integration:down` (always) |
| **Timeout** | 30 minutes |
| **Local parity** | `task ci:integration` or `task test:integration:local` |

`ci.yaml` runs unit + envtest only (`task test:run`). Integration is **not** part of that workflow.

---

## 2. Package layout and scope

Only one package under `test/integration/`: **`test/integration/mq/`** (4 files, all `//go:build integration`).

| File | What it tests |
|------|----------------|
| `config.go` | Env config (`KURATOR_INTEGRATION_MQ_*`), `mqrest.NewClient`, name helpers |
| `client_integration_test.go` | Ping; queue local/alias/remote CRUD + replace; not-found; idempotent delete; unsupported type; `RunMQSC`; `ClientFactory.ForConnection` + Ping |
| `topic_channel_integration_test.go` | Topic CRUD + replace; channel (svrconn) CRUD + replace; not-found; idempotent delete; unsupported channel type |
| `auth_integration_test.go` | `SetChannelAuth` + `GetChannelAuth` (ADDRESSMAP); `SetAuthority` + `GetAuthority` (QUEUE); not-found for both |

**25** `TestIntegration_*` functions. Framework: stdlib `testing` (not Ginkgo). Gate: `KURATOR_INTEGRATION_MQ=1` (set by `task test:integration`). Target: `./test/integration/mq/...` with `-tags=integration -race`.

Infrastructure: Docker IBM MQ (`hack/mq-docker/docker-compose.yml`, mqweb on `localhost:9443`).

Design intent matches [ADR-0011](../adr/0011-layered-testing-strategy.md): integration = **adapter contract against live mqweb**, no cluster.

---

## 3. Tier comparison

| Concern | Unit | envtest | **Integration** | e2e |
|---------|------|---------|-----------------|-----|
| **Runs in default CI job** | Yes (`ci.yaml` test) | Yes (same) | **Yes** (`integration.yaml`) | Yes (`e2e.yaml`) |
| **Real mqweb** | No (`httptest`) | No (mock Admin) | **Yes** (Docker MQ) | Yes (kind MQ) |
| **Kubernetes / operator** | No | Yes (API server) | **No** | Yes (full deploy) |
| **CR types** | Mocked | All CRs, mocked MQ | **None** (port only) | Queue, Topic, Channel, CAR, AuthorityRecord, QMC |
| **Reconcilers / drift / finalizers** | Unit tests | envtest + mocks | **No** | Partial (reconcile + delete; no observe-only drift e2e) |
| **Auth delete paths** | httptest | Mock expectations | **Cleanup only, not asserted** | CAR + AuthorityRecord delete verified |
| **Alias / remote queues** | Unit | Mock | **Yes** (adapter) | No (e2e uses local only) |

---

## 4. Blindspots

### CR types — not covered (by design)

- `QueueManagerConnection`, `Queue`, `Topic`, `Channel`, `ChannelAuthRule`, `AuthorityRecord`
- Admission webhooks, status conditions, Events, finalizers

These belong in envtest/e2e, not integration.

### MQ operations — gaps at adapter level

| Area | Integration | Covered elsewhere |
|------|-------------|-------------------|
| Queue local CRUD + replace | Yes | e2e (local) |
| Queue alias / remote | Yes | Not in e2e |
| Topic / channel CRUD + replace | Yes | e2e |
| `SetChannelAuth` / `GetChannelAuth` | Set + get only | Unit httptest; e2e full CAR lifecycle |
| **`DeleteChannelAuth`** | Cleanup only (`_ = c.Delete...`) | Unit + e2e |
| **`SetAuthority` update / replace** | Single set | Unit; e2e |
| **`DeleteAuthority`** | Cleanup only | Unit + e2e |
| CHLAUTH rule types beyond **ADDRESSMAP** | No | Product supports 6 types; only ADDRESSMAP tested live |
| AUTHREC object types beyond **QUEUE** | No | 7 types in API; only QUEUE tested live |
| `ClientFactory.ReleaseConnection` | No | Unit (`factory_release_test.go`) |
| TLS with verified certs / CA bundle | No (always `InsecureSkipVerify`) | Same in e2e |
| Bad credentials / REST error taxonomy vs real MQ | No | e2e (QMC secret rotation) |
| Transient vs terminal errors from real MQ | No | Unit only |

### Drift — not covered

- No observe-only annotation behavior
- No attribute drift detection / replace-on-diff at reconciler level
- Queue/topic/channel **replace-on-update** is tested at adapter level; auth **update** is not

### Delete paths — partial

| Object | Explicit delete test? |
|--------|----------------------|
| Queue / topic / channel | Yes (create → delete → not-found) + idempotent delete |
| CHLAUTH / AUTHREC | **No** — delete only in `t.Cleanup`, never asserted |

---

## 5. Docs accuracy (DEVELOPMENT.md, CICD.md)

| Claim | Accurate? |
|-------|-----------|
| Integration in dedicated workflow, Docker MQ, `task test:integration` | **Yes** |
| Covers queue, topic, channel, CHLAUTH, AUTHREC via mqweb | **Yes** at adapter level |
| Path filters skip docs-only changes | **Yes** |
| `task ci:integration` = CI parity | **Yes** |
| Build tag + `KURATOR_INTEGRATION_MQ=1` gate | **Yes** |
| e2e covers full CR reconcile + delete when `KURATOR_E2E_MQ=1` | **Yes** (`test/e2e/mq_e2e_test.go`) |
| Branch protection requires integration | **Hedged correctly** — depends on GitHub settings |

**Minor doc gaps (not wrong, just incomplete):**

1. Neither doc states that integration auth **delete/update** are not explicitly tested (only set + get + cleanup).
2. Neither notes that **alias/remote queues** are integration-only (e2e uses local queues only).
3. CICD could mention integration runs on **`push` to `main` only** (PRs run on all branches).

No material inaccuracies found; this plan doc adds blindspot detail rather than fixing errors.

---

## 6. Recommendations

### CI — no enablement needed

Integration is already on PRs and `main` pushes (non-docs). No action unless you want it inside `ci.yaml` (not recommended — keeps slow/MQ jobs isolated).

### Add integration tests (highest value)

1. **`TestIntegration_DeleteChannelAuth`** — set → delete → `GetChannelAuth` → `ErrNotFound`
2. **`TestIntegration_DeleteAuthority`** — same for AUTHREC
3. **`TestIntegration_ChannelAuth_UpdateViaReplace`** / **`TestIntegration_Authority_UpdateViaReplace`** — mirror queue/topic patterns
4. **Idempotent auth delete** — parallel to queue/topic/channel tests

Optional: second CHLAUTH rule type or AUTHREC object type if product supports them in Phase 5.

### Document intentional exclusions

- Integration = **adapter contract only**; no K8s, reconcilers, drift, webhooks
- Auth delete/update gaps and where they’re covered (unit httptest + e2e)
- Alias/remote queues: integration yes, e2e no

### Do not duplicate upward

Avoid pushing reconciler/drift/finalizer tests into integration; envtest + e2e already own those.

---

## Quick reference

| Question | Answer |
|----------|--------|
| **CI yes/no** | **Yes** — `.github/workflows/integration.yaml` |
| **Package count** | 1 (`test/integration/mq/`) |
| **Test count** | 25 |
| **Scope** | Adapter-only (`mqadmin.Admin` / `mqrest`) |
