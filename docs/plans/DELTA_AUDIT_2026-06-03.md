# Kurator delta audit — 2026-06-03 (evening refresh)

**Audit time (UTC):** 2026-06-03 ~03:20  
**Scope:** `origin/main` through **`77e49d4`** (`ci(workflows): reusable GHA caches`); coordination gate `push_allowed: false`  
**Method:** Read-only — `git log`, `gh run list` / failed-job logs, ROADMAP/CICD/docs, `hack/.agent-coordination.json`; **no e2e execution**.

---

## Executive summary

Since the early audit (`41a768e`), **`main` gained Phase C CI/pyramid work, preflight, nightly + release-gate workflows, reusable caches, admission/envtest hardening, and the Docker MQ GROUP AUTHREC fix** — but **merge-blocking E2E is still red** on the last fully observed push at **`6d55292`** (Ginkgo duplicate `SynchronizedBeforeSuite`). **Integration** on that SHA failed on **`TestIntegration_GetAuthority_Group`** (GROUP test removed in **`ea8690c`**, not yet in that push’s ancestry order for CI — fix is on current `main`).

| Signal | Status |
|--------|--------|
| **CI + Preflight** | Green on recent pushes (`6d55292`, `762ec45`, `77e49d4` in flight) |
| **Integration** | Red at `6d55292` (GROUP); **expected green** on `77e49d4` after `ea8690c` |
| **E2E (kustomize + Helm on main)** | **Red** at `6d55292` — suite structure error; **pending re-run** after e2e `BeforeSuite` merge lands on `origin/main` |
| **Nightly / release gate** | **Shipped** (workflows + docs); first scheduled nightly not yet a green signal |
| **Local `task ci:e2e`** | Not re-run this audit; prior session log still red (metrics class) |

---

## Shipped since early audit (June 2026)

### Phase C — CI pyramid and e2e scope

| Commit | What |
|--------|------|
| `e460d48` | E2e pyramid trim; Ginkgo labels (`smoke`, `mq`, `slow`); adapter-heavy cases shifted toward integration |
| `ad0cbeb` | **PR:** `KURATOR_E2E_LABEL_FILTER=(smoke \|\| mq) && !slow`; **main push:** full kustomize + **Helm on same cluster** after kustomize |
| `dac64ed` | [CICD.md](../CICD.md), [E2E_SPEEDUP_PROPOSAL.md](./E2E_SPEEDUP_PROPOSAL.md) — Phase C marked implemented |

### Nightly workflow

| Commit | What |
|--------|------|
| `762ec45` | [`.github/workflows/nightly.yaml`](../../.github/workflows/nightly.yaml) — Mon **03:00 UTC** integration + e2e (kustomize + optional Helm); `cancel-in-progress: false` |

### Release gate

| Commit | What |
|--------|------|
| `9aa001e` | [`.github/workflows/release-gate.yaml`](../../.github/workflows/release-gate.yaml) + `hack/ci/wait-release-gate-checks.sh`; [RELEASE.md](../RELEASE.md) automated gate section |

### Reusable GHA caches

| Commit | What |
|--------|------|
| `77e49d4` | Composite actions: `go-cache`, `tools-bin`, `mq-docker-image`, `helm-cache` wired in `ci`, `preflight`, `integration`, `e2e`, `nightly` ([CICD.md](../CICD.md) cache table) |

### Admission (CAR rule types)

| Commit | What |
|--------|------|
| `6d55292` | Table-driven **`internal/validation`** tests for `ChannelAuthRule` rule types (schema enums; USERMAP still deferred for MQSC) |

### Envtest — `events.k8s.io`

| Commit | What |
|--------|------|
| `f9d22b7` | RBAC: controller may create **`events.k8s.io`** |
| `355f6c9` | `internal/controller/events_envtest_test.go` — assert reconcile events on Ready/Synced transitions |

### GROUP AUTHREC (Docker integration)

| Commit | What |
|--------|------|
| `ea8690c` | Drop **`TestIntegration_GetAuthority_Group`** — IBM MQ dev image has no OS groups (`AMQ8871E`); GROUP MQSC remains in **mqrest unit tests** |
| `d64835c` | CHLAUTH/AUTHREC MQSC goldens in mqrest unit tests |

### Other CI / test infrastructure (same window)

| Commit | What |
|--------|------|
| `5e99461` | **`go.sum` / `go tool kustomize`** — tidy after Renovate (fixes early-audit Integration `missing go.sum` class) |
| `7fb1323` | **`preflight.yaml`** fail-fast before heavy jobs |
| `6e24af2` | Parallel MQ Ginkgo nodes + split deploy `SynchronizedBeforeSuite` (**introduced duplicate BeforeSuite bug** — see red items) |

---

## Still red / pending

### E2E on `6d55292` (observed)

| Workflow | Run | Result | Root cause |
|----------|-----|--------|------------|
| **E2E** | `26861264759` | **failure** (~6m) | Ginkgo: **two** `SynchronizedBeforeSuite` nodes in `test/e2e/e2e_suite_test.go` (lines 41 and 71 at that SHA) — suite aborts before specs |

```text
Ginkgo only allows you to define one suite setup node.
… already have [SynchronizedBeforeSuite] at e2e_suite_test.go:41
… trying to add at e2e_suite_test.go:71
```

**Pending:** Re-run **E2E + Integration** on `origin/main` after the **single merged `SynchronizedBeforeSuite`** fix is pushed (local working tree has the merge; not on `77e49d4` at audit time). Coordinator: **`push_allowed: false`** until green pipeline.

### Integration on `6d55292` (observed)

| Workflow | Run | Result | Root cause |
|----------|-----|--------|------------|
| **Integration** | `26861264747` | **failure** | `TestIntegration_GetAuthority_Group` — fixed on `main` by **`ea8690c`** (after `6d55292` in history) |

### E2E history (context)

- **`6e24af2`:** E2E failed after ~10m (deploy/spec class; superseded by struct error above).
- **`67d2d35` / `5da75e5`:** Both e2e jobs failed (pre-Phase-C).
- **Helm on main:** Phase C runs Helm **after** kustomize on same cluster; blocked while kustomize job fails at compile/suite setup.
- **Nightly:** Not required for merge; no green scheduled run recorded yet.

### Release / doc drift (unchanged from early audit)

- Tag **`v0.5.2`** @ `c0ea77e`; GitHub Release page may still lag (verify with `gh release view v0.5.2`).
- [ROADMAP.md](../ROADMAP.md) still lists **CI/Integration/E2E green before `v0.5.2`** as `[x]` — **overstated** until a green E2E run on current `main`.
- **`task ci:e2e` green locally** — still `[ ]` in ROADMAP.

---

## Delta matrix (claimed vs observed) — evening

| Area | Prior audit (~02:34 UTC) | Observed now (`77e49d4` / CI) | Delta |
|------|--------------------------|-------------------------------|-------|
| `go tool kustomize` / go.sum | Integration red | **`5e99461`** on `main`; CI/Preflight green | **Fixed** |
| Phase C pyramid | Not implemented | PR filter + main Helm pass **coded** | **Shipped** |
| Nightly e2e | Not in repo | **`nightly.yaml`** Mon 03:00 UTC | **Shipped** |
| Release gate | Not in repo | **`release-gate.yaml`** + poll script | **Shipped** |
| GHA caches | None | **Four composite actions** on heavy workflows | **Shipped** |
| CAR admission depth | Partial | **`6d55292`** validation tables | **Shipped** |
| events.k8s.io | RBAC gap | **`f9d22b7` + `355f6c9`** envtest | **Shipped** |
| GROUP integration | Failing / flaky | **`ea8690c`** drop GROUP test | **Shipped** (on HEAD; CI at `6d55292` predates) |
| E2E green on `main` | 0 successes | **Still 0** at `6d55292`; **`77e49d4` in progress** | **Open** |
| Duplicate BeforeSuite | — | **`6e24af2` → `6d55292`**; fix **unpushed** | **Open** |
| USERMAP / PCF | Open | Open | Unchanged |
| `v0.5.2` GitHub Release | Missing / 404 | Not re-verified this pass | **Verify before tag** |

---

## CI status (`origin/main`) — representative SHAs

| SHA | Subject | Preflight | CI | Integration | E2E |
|-----|---------|-----------|-----|-------------|-----|
| `77e49d4` | reusable GHA caches | in progress @ audit | in progress | in progress | in progress |
| `6d55292` | CAR admission tables | success | success | **failure** (GROUP) | **failure** (2× BeforeSuite) |
| `762ec45` | nightly workflow | success | success | failure | cancelled (superseded) |
| `6e24af2` | parallel MQ + split deploy | success | success | failure | failure |
| `5e99461` | go.sum tidy | — | — | — | (enables Integration class fix) |

**E2E success count on `main`:** still **0** completed green kustomize runs in recent history through **`6d55292`**.

---

## P0 / P1 / P2 checklist

### P0 (release / CI integrity)

- [x] Repair `go.mod`/`go.sum` so `go tool kustomize` works (**`5e99461`**).
- [ ] Green **Integration** + **E2E** on current `origin/main` HEAD ( **`77e49d4` + BeforeSuite fix** — re-run pending after push).
- [ ] Correct ROADMAP: uncheck or qualify **“CI green before v0.5.2”** until E2E evidenced on HEAD.
- [ ] Decide **`v0.5.2` release**: publish via `release.yaml` / release-gate or document tag-only state.

### P1 (Phase 5 closure / operator quality)

- [ ] Local **`task ci:e2e`** green; attach log path/SHA to ROADMAP.
- [x] **events.k8s.io** RBAC + envtest event assertions (**`f9d22b7`**, **`355f6c9`**).
- [ ] **E2E suite setup:** merge to **one** `SynchronizedBeforeSuite` (deploy + images); verify kustomize + Helm on main push.
- [x] **GROUP** Docker integration — drop unsupported AUTHREC test (**`ea8690c`**).
- [x] **Phase C** PR label filter + main Helm-on-same-cluster (**`ad0cbeb`**, **`e460d48`**).
- [x] **Preflight** fail-fast job (**`7fb1323`**).
- [x] **Nightly** + **release-gate** workflows + docs.
- [x] **GHA caches** for Go/MQ/tools/helm (**`77e49d4`**).
- [x] **CAR rule-type** admission unit tables (**`6d55292`**).
- [ ] CHANGELOG section for **0.5.2** matching tag/release.

### P2 (backlog)

- [ ] USERMAP / SSLPEERMAP CR fields + mqrest + tests.
- [ ] PCF adapter implementation ([ADR-0017](../adr/0017-pcf-adapter-behind-mqadmin.md)).
- [ ] First **green nightly** run (flake signal; non-blocking).
- [ ] Codecov strict floor (optional).
- [ ] Close stale `origin/renovate/*` after merges.

---

## Coordination / publish

| Field | Value |
|-------|-------|
| `push_allowed` | **false** — audit update may commit locally; **do not push** until coordinator sets true and E2E re-run planned |
| `pipeline_check_in_progress` | false @ `2026-06-03T03:03:47Z` |
| `notes` (coordinator) | Wait **`6e24af2` / HEAD** Integration+E2E green; **`77e49d4`** run in flight at evening audit |
| **Pending after push** | Re-run **E2E** (and confirm **Integration**) on `main` after **single BeforeSuite** lands; update this doc with run IDs |

---

## Top remaining (priority)

1. **Push** merged `e2e_suite_test.go` `SynchronizedBeforeSuite` fix; confirm **E2E kustomize** green.
2. Confirm **Integration** green on HEAD (`ea8690c` already on `main`).
3. **Helm e2e** on main push (Phase C) after kustomize green.
4. Align **ROADMAP** checkboxes with CI reality; run **release-gate** on green SHA before next tag.
5. **`task ci:e2e` green locally** and record SHA.
6. USERMAP / PCF / nightly green signal (P2).

---

## References

```bash
git log origin/main -20 --oneline
gh run list --branch main --limit 20
gh run view <run-id> --log-failed
git show 6d55292:test/e2e/e2e_suite_test.go | grep -n SynchronizedBeforeSuite
# Coordinator: hack/.agent-coordination.json
```
