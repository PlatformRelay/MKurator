# Kurator delta audit — 2026-06-03

**Audit time (UTC):** 2026-06-03 ~02:34  
**Scope:** `origin/main` through `41a768e` (coordination gate: `hack/.agent-coordination.json` had empty `notes`, `push_allowed: false`)  
**Method:** Read-only — `git log`, `gh run list`, ROADMAP/AGENTS/docs, terminal log `/tmp/ci-e2e-run2.log`, no cluster mutations.

---

## Executive summary

Phase 5 **core auth** (ChannelAuthRule, AuthorityRecord, mqrest, e2e specs) is **implemented on `main`**, but **release and CI truth lag documentation**:

- Git tag **`v0.5.2`** exists (`c0ea77e`) with **no GitHub Release** (API 404).
- **E2E has not completed green** on recent `main` pushes (failures + cancellations; Helm path fails at deploy).
- **Local `task ci:e2e`** ran ~33 min and **failed** (metrics / controller readiness), not recorded in coordination JSON.
- **Renovate merges** landed; **Integration** on latest SHAs fails on **`go tool kustomize` / missing `go.sum`** entry.
- Open backlog: **USERMAP API**, **PCF implementation**, **scheduled nightly e2e** (not in repo).

---

## Delta matrix (claimed vs observed)

| Area | Documented / tagged state | Observed (2026-06-03) | Delta |
|------|---------------------------|------------------------|-------|
| Phase 5 core auth | Shipped; exit criteria met (`v0.5.2`) | CRDs, reconcilers, mqrest, kind e2e specs present on `main` | **Aligned** for code |
| ROADMAP: CI green before `v0.5.2` | `[x]` Confirm CI + Integration + E2E green | E2E: **0 successes** in last 30 `e2e.yaml` runs on `main`; Integration **failure** on `e71fb05`, `41a768e` | **Doc overstated** |
| ROADMAP: `task ci:e2e` locally | `[ ]` remaining | Local run **FAILED** — `e2e_test.go:286` (metrics log substring); pod readiness **HTTP 500** | **Still open**; evidence in `/tmp/ci-e2e-run2.log` |
| `v0.5.2` release | ROADMAP: "published on GitHub" | Tag exists; **`gh release view v0.5.2` → 404**; `gh release list` shows latest **v0.5.0** | **Tag without release assets** |
| Helm RBAC (auth CRDs) | `[x]` `helm-verify-rbac.sh` in `helm:lint` | `charts/kurator/templates/clusterrole.yaml` lists `authorityrecords`, `channelauthrules`; script on `main` | **Aligned** |
| `internal/` coverage ≥90% | ROADMAP Phase 2/4 bar; `Taskfile.test.yml` `min=90` | Enforced locally/CI via `task test:run`; Codecov **`target: auto` only** (no CI gate) | **Local gate yes; remote soft** |
| Renovate | Weekly workflow + grouped PRs | **6 renovate branches merged** to `main` (`988623c`…); remote `origin/renovate/*` still exist; Renovate workflow **success** at 01:28Z after earlier failures | **Merged; watch go.sum fallout** |
| AGENTS.md on `main` | Phase 5 CRDs in overview | `origin/main` AGENTS lists Queue, Topic, Channel, **ChannelAuthRule**, **AuthorityRecord** | **Aligned** (`bc4688d` docs sync) |
| USERMAP / SSLPEERMAP | Schema-valid; MQSC deferred | OpenAPI enum includes types; `PHASE5_AUTH_SKETCH.md` defers fields; no `buildSetChannelAuthMQSC` for USERMAP | **Open** |
| PCF adapter | ADR-0017 proposed; CHANGELOG scaffold | `internal/adapter/mqpcf` — all methods `errNotImplemented` | **Scaffold only** |
| Nightly e2e | Parent audit item | **No** `cron` / workflow for scheduled e2e; only `push`/`workflow_dispatch` on `e2e.yaml` | **Not implemented** |
| Helm e2e namespace | — | `e71fb05` fix: Helm e2e owns `kurator-system`; CI Helm job still **failed** deploy (`deploy_helpers.go:97`) on `c0ea77e` run | **Partial fix; CI still red** |

---

## Doc truth table

| Document | Claim | Truth on `main` | Action |
|----------|-------|-----------------|--------|
| [ROADMAP.md](../ROADMAP.md) Phase 5 | CI/E2E green before `v0.5.2` checked | **False** for current `main` E2E history | Uncheck or qualify with SHA/date |
| [ROADMAP.md](../ROADMAP.md) Phase 5 | `v0.5.2` published on GitHub | Tag only; **no release page** | Publish release or fix wording |
| [ROADMAP.md](../ROADMAP.md) Phase 5 | `task ci:e2e` green locally | **Failed** this session | Keep `[ ]` until green log |
| [RELEASE.md](../RELEASE.md) | Tag drives release workflow | Tag `v0.5.2` present; release workflow not evidenced | Run release for tag or delete/re-tag per policy |
| [README.md](../README.md) | USERMAP etc. schema + admission | Accurate; reconciler/MQSC not implemented | OK |
| [PHASE5_AUTH_SKETCH.md](../PHASE5_AUTH_SKETCH.md) | USERMAP deferred | Matches code | OK |
| [AGENTS.md](../../AGENTS.md) | Phase 5 resources, mqrest-only | Matches `main` | OK |
| [CHANGELOG.md](../../CHANGELOG.md) | Unreleased / v0.5.2 section | No `0.5.2` heading found at audit time | Align with tag/release |

---

## CI status (`origin/main`)

### Recent commits (newest first)

| SHA | Subject |
|-----|---------|
| `41a768e` | chore(hack): add parallel agent coordination |
| `e71fb05` | fix(test): let Helm e2e own kurator-system namespace |
| `988623c` | Merge renovate/major-github-actions (+ 5 more renovate merges) |
| `bc4688d` | docs: track AGENTS.md and sync roadmap for v0.5.2 |
| `c0ea77e` | chore(release): prepare v0.5.2 (**tag `v0.5.2`**) |

### Workflow conclusions (representative)

| Workflow | Latest `41a768e` | Prior `e71fb05` | `988623c` (renovate merge) | `c0ea77e` (`v0.5.2` prep) |
|----------|------------------|-----------------|---------------------------|---------------------------|
| **CI** | queued / in progress | queued | failure (CI job) | success |
| **Integration** | **failure** | **failure** | failure | success |
| **E2E** | in progress | failure | cancelled | **failure** (kustomize + helm jobs) |

**Integration failure root cause (`41a768e`, `e71fb05`, `988623c`):**

```text
missing go.sum entry for module providing package sigs.k8s.io/kustomize/kustomize/v5
task: Command "go tool -n kustomize" failed
```

**E2E failure patterns:**

- **Kustomize job:** metrics / Manager specs (same class as local failure).
- **Helm job (`c0ea77e`):** 3 failures — Manager BeforeAll, Queue BeforeAll, ChannelAuthRule BeforeAll — all at `deploy_helpers.go:97` (Helm deploy path / namespace lifecycle).

**E2E success count:** **0** in last 30 `e2e.yaml` runs on `main` (as of audit).

---

## Local e2e proof (coordination + terminal)

| Source | Result |
|--------|--------|
| `hack/.agent-coordination.json` `notes` | **Empty** — no sibling-recorded proof |
| Terminal / `KURATOR_E2E_MQ=1 task ci:e2e` → `/tmp/ci-e2e-run2.log` | **FAIL** — `[BeforeSuite] PASSED`; **Manager** spec failed: expected log `"Serving metrics server"` (`e2e_test.go:286`); controller pod had **readiness probe HTTP 500** |
| MQ integration specs in same run | Not reached / skipped after Manager failure |

**Conclusion:** Local full `ci:e2e` is **not** green; ROADMAP checkbox for local e2e remains accurate.

---

## Renovate state

| Item | Status |
|------|--------|
| Merged to `main` | `go-dependencies`, `kubernetes-packages`, `terraform`, `platform-tools`, `major-github-actions`, `ghcr.io-devcontainers-features-docker-in-docker-3.x` |
| Open remote branches | `origin/renovate/*` still listed (docker-images, github-actions, golang-1.x, ibm-mq, …) |
| `renovate.yaml` schedule | Weekly Mon 04:00 UTC + `workflow_dispatch` |
| Last Renovate workflow | **success** @ 2026-06-03T01:28:30Z (after 3 failures 01:11–01:13Z) |
| Risk | Post-merge **go.sum** / `go tool kustomize` breakage in Integration |

---

## AGENTS.md on `main`

- Entry point lists **ChannelAuthRule** and **AuthorityRecord** in scope and architecture diagram.
- Testing table includes e2e auth scenarios and `task ci:e2e`.
- Consistent with Phase 5 code; does **not** claim CI/E2E currently green.

---

## Helm RBAC

- `hack/helm-verify-rbac.sh` compares Helm `ClusterRole` to `config/rbac/role.yaml` for `messaging.kurator.dev`.
- Chart template includes **authorityrecords** and **channelauthrules** (verbs: get/list/watch/create/update/patch/delete + status/finalizers).
- Wired via `task helm:lint` per ROADMAP — **structure aligned**; independent of failing e2e deploy.

---

## Coverage

| Layer | Gate |
|-------|------|
| `task test:run` | **≥90%** on `./internal/...` (`Taskfile.test.yml`, `min=90`) |
| CI `ci.yaml` test job | Produces `coverage.out`, Codecov upload |
| `codecov.yml` | `target: auto` — **no hard fail** on regression |
| ROADMAP | States ≥90% met for Phase 2/4 — assume holds if CI `test` job succeeds |

---

## `v0.5.2` tag

| Check | Result |
|-------|--------|
| Annotated/lightweight tag | **`v0.5.2`** → `c0ea77e` (`chore(release): prepare v0.5.2`) |
| Ancestor of current `main` | **Yes** |
| GitHub Release | **Missing** (404) |
| Release workflow artifacts | Not verified green for this tag |

---

## Open items (USERMAP, PCF, nightly e2e)

### USERMAP / extended CHLAUTH

- CRD enum: `USERMAP`, `SSLPEERMAP`, `QMGRMAP`, `BLOCKADDR` in OpenAPI.
- **No** `clientUser` / `mcaUser` / `sslPeer` on CRD; **no** mqrest MQSC builder tests per [PHASE5_AUTH_SKETCH.md](../PHASE5_AUTH_SKETCH.md).
- ROADMAP: optional follow-up.

### PCF

- [ADR-0017](../adr/0017-pcf-adapter-behind-mqadmin.md) **Proposed**.
- `internal/adapter/mqpcf`: compile-time `mqadmin.Admin` impl; **all methods stubbed** (`errNotImplemented`).
- Production path remains **mqrest** only.

### Nightly e2e

- **Not present** in `.github/workflows/` (no scheduled e2e).
- E2E triggers: `push` to `main` (with paths-ignore), `pull_request`, `workflow_dispatch`.
- Candidate: add `schedule` cron on `e2e.yaml` or dedicated workflow (document in ROADMAP / CICD if desired).

---

## Test pyramid gaps

| Tier | Coverage | Gap |
|------|----------|-----|
| Unit + httptest | Strong on mqrest, controllers | USERMAP MQSC paths absent (intentional) |
| envtest | Webhooks, reconcilers mocked | — |
| Integration (Docker MQ) | Auth delete/update, BLOCKUSER | **CI broken** on kustomize/go.sum; extended CHLAUTH types untested |
| e2e kind | Queue/Topic/Channel/Auth specs exist | **CI never green** recently; **local ci:e2e red**; Helm deploy BeforeAll failures |
| Helm e2e | CI job on main push | **Red** at deploy; namespace fix (`e71fb05`) not sufficient alone |
| PCF | Scaffold only | No integration/e2e tier |

---

## Top 10 remaining (priority order)

1. Fix **Integration** `go.sum` / `go tool kustomize` after Renovate merges.
2. Stabilize **CI E2E** (kustomize + Helm): deploy readiness, metrics spec, `deploy_helpers.go` Helm path.
3. Get **`task ci:e2e` green locally** and record SHA in ROADMAP.
4. Align **ROADMAP** checkboxes with CI reality (CI green before tag; `v0.5.2` "published").
5. **Publish GitHub Release** for `v0.5.2` (or retag per [RELEASE.md](../RELEASE.md) after green CI).
6. Re-run full pipeline on **`e71fb05`+** after Integration fix; confirm Helm e2e.
7. **USERMAP** API + mqrest + integration when product needs it.
8. **PCF** adapter implementation (ADR-0017) — optional environments.
9. Add **scheduled nightly e2e** (workflow + docs) for flake detection.
10. Phase 4 optional: TLS channel drift, mqweb DISPLAY-limited queue attrs.

---

## P0 / P1 / P2 todos

### P0 (release / CI integrity)

- [ ] Repair `go.mod`/`go.sum` so `go tool kustomize` works in Integration (and anywhere else).
- [ ] Green **Integration** + **E2E** on `main` for current HEAD.
- [ ] Correct ROADMAP: uncheck false "CI green before v0.5.2" until evidenced.
- [ ] Decide **v0.5.2** release: publish via `release.yaml` or document tag-only state.

### P1 (Phase 5 closure / operator quality)

- [ ] Local **`task ci:e2e`** green; attach log path/SHA to ROADMAP.
- [ ] Fix e2e **metrics** assertion / controller readiness (500) — blocks Manager spec.
- [ ] Helm e2e: verify `deploy:helm` + CRD registration after namespace ownership fix.
- [ ] CHANGELOG section for **0.5.2** matching tag.

### P2 (backlog)

- [ ] USERMAP / SSLPEERMAP CR fields + mqrest + tests.
- [ ] PCF command wiring behind `MQAdmin`.
- [ ] Scheduled **nightly e2e** workflow.
- [ ] Codecov strict floor (optional, matches ROADMAP narrative).
- [ ] Close stale `origin/renovate/*` branches after merges.

---

## Coordination / publish

| Field | Value |
|-------|-------|
| `push_allowed` | **false** — audit file **not pushed** |
| `pipeline_check_in_progress` | true at audit start |
| Deliverable | `docs/plans/DELTA_AUDIT_2026-06-03.md` (local only unless parent sets `push_allowed`) |

---

## References (commands)

```bash
git log origin/main -15 --oneline
gh run list --branch main --limit 20
gh release list --limit 5
git tag -l 'v0.5.*'
# Local e2e log: /tmp/ci-e2e-run2.log
```
