# Repository architectural fitness audit — 2026-06-03

**Audit time (local):** 2026-06-03  
**Scope:** Local toolchain gates (`task verify`, `lint`, `test:run`, `secrets:scan`, `vuln:check`), documented layer model ([ARCHITECTURE.md](../ARCHITECTURE.md), [AGENTS.md](../../AGENTS.md)), import graph under `internal/`. **No e2e / `ci:e2e` executed.**

---

## Tool results

| Tool | Result | Notes |
|------|--------|-------|
| `task verify` | **PASS** | CRDs, mocks, Helm sync, `test/schema` contract |
| `task lint` | **PASS** (after fixes) | Was failing on `test/schema/*` gosec/govet/lll and one `events_envtest_test.go` lll line |
| `task arch:lint` | **N/A** | Task not defined in `Taskfile.yml`; `.go-arch-lint.yml` absent; `go-arch-lint` not in `go.mod` |
| `task arch:graph` | **N/A** | Task not defined; `docs/diagrams/internal-deps.svg` missing |
| `task test:run` | **PASS** | `internal/` coverage **91.2%** (floor 90%) |
| `task secrets:scan` | **PASS** | gitleaks, 228 commits, no leaks |
| `task vuln:check` | **PASS** | govulncheck, no vulnerabilities |

---

## Intended dependency layers

From [ARCHITECTURE.md](../ARCHITECTURE.md) and [AGENTS.md](../../AGENTS.md):

```text
api/v1alpha1
    ↑
internal/mqadmin (port + domain types)
    ↑
internal/adapter/mqrest | mqpcf (implement Admin; wired from cmd/)
    ↑
internal/controller (thin; MQ via Admin only — no HTTP/MQ details)
internal/validation, internal/webhook → api; validation must not call mqweb
cmd/ → wires manager, factory, webhooks
```

Cross-cutting: `internal/logging`, `internal/metrics`, `internal/health` — no imports of `adapter/*` or `controller`.

---

## Observed import graph (2026-06-03)

| Package | Imports (project) |
|---------|-------------------|
| `internal/controller` | `api`, **`adapter/mqrest`**, `metrics`, `mqadmin` |
| `internal/adapter/mqrest` | `api`, `metrics`, `mqadmin` |
| `internal/adapter/mqpcf` | `mqadmin` (stub; not wired in `cmd/main.go`) |
| `internal/validation` | `api`, `mqadmin` |
| `internal/webhook/v1alpha1` | `api`, `validation` |
| `internal/health` | `api` |
| `cmd` | `api`, `adapter/mqrest`, `controller`, `health`, `logging`, `webhook` |

Manual review: **no cycles**; **one documented layer breach** (controller → mqrest).

---

## Fixes applied this session

| File | Change |
|------|--------|
| `test/schema/contract_test.go` | gosec file modes (0750/0600), gosec nolint for fixture paths, govet shadow, lll line breaks |
| `test/schema/extract.go` | Six CRD `DefaultCases`, gosec nolint, govet shadow on unmarshal |
| `test/schema/golden/*.yaml` | OpenAPI spec fragments for all v1alpha1 CRDs |
| `internal/controller/events_envtest_test.go` | lll: split long GVK guard line |
| `Taskfile.test.yml` | (on `main` before audit commit) Split `test:run` coverage: non-internal pkgs without `-coverprofile`; `./internal/...` with `-p 1` |

---

## P0 — merge / correctness / security

| ID | Issue | Evidence | Recommended fix |
|----|-------|----------|-----------------|
| — | *None open from this audit* | All requested local gates green after lint fixes | — |

*(E2E/integration CI health is tracked in [DELTA_AUDIT_2026-06-03.md](./DELTA_AUDIT_2026-06-03.md); out of scope here per instruction.)*

---

## P1 — architectural fitness & doc/tooling drift

| ID | Issue | Evidence | Recommended fix |
|----|-------|----------|-----------------|
| P1-1 | **Controllers import `mqrest` directly** | `queue_controller.go`, `topic_controller.go`, `channel_controller.go`, `channelauthrule_controller.go`, `authorityrecord_controller.go` call `mqrest.Format*MQSC` and drift key helpers | Move MQSC formatting + drift key lists to `internal/mqadmin` (or `internal/mqsc`) so controllers depend only on the port; keep `mqrest` as adapter + factory in `cmd/` |
| P1-2 | **`task arch:lint` / `task arch:graph` documented but missing** | [AGENTS.md](../../AGENTS.md) task table; no `.go-arch-lint.yml`, no `docs/diagrams/` | Add `.go-arch-lint.yml` matching layers above, wire tasks in `Taskfile.yml`, optionally chain `arch:lint` into `task lint` |
| P1-3 | **`task lint` does not run depguard / arch** | [AGENTS.md](../../AGENTS.md) says “depguard + go-arch-lint”; `.golangci.yaml` has neither | Enable `depguard` rules for layer edges **or** rely on go-arch-lint only; align AGENTS.md with reality |
| P1-4 | ~~**CRD OpenAPI contract covers Queue only**~~ | **Addressed this session:** six `DefaultCases` rows + goldens committed | Keep goldens updated when CRD markers change (`task test:schema:update`) |
| P1-5 | **`mqpcf` scaffold unused at runtime** | `internal/adapter/mqpcf` implements stub `Admin`; [ADR-0017](../adr/0017-pcf-adapter-behind-mqadmin.md) | Acceptable placeholder; document “not linked in main” in ADR or arch graph when P1-2 lands |
| P1-6 | **ARCHITECTURE.md component table incomplete** | Diagram/table omit `ChannelAuthRule`, `AuthorityRecord` reconcilers | Update components table + mermaid to match `cmd/main.go` registrations |

---

## P2 — hygiene (optional)

| ID | Issue | Recommended fix |
|----|-------|-----------------|
| P2-1 | `hack/.agent-coordination.json` often dirty locally | Gitignore or document as session-local only |
| P2-2 | `validation` → `mqadmin` for attribute keys | Fine for admission; if mqadmin grows, split shared constants package |

---

## Suggested next actions (ordered)

1. Land P1-4 (schema goldens for all CRDs) — low risk, high signal for API drift.
2. Implement P1-2 + P1-3 so layer rules are enforced in CI, not manual audit.
3. Refactor P1-1 in a focused PR (controller decoupling from mqrest).
4. Refresh ARCHITECTURE.md (P1-6) when auth CRDs are stable on main.

---

## References

- [ARCHITECTURE.md](../ARCHITECTURE.md) — MQAdmin port, thin reconcilers
- [AGENTS.md](../../AGENTS.md) — task matrix, testing tiers
- [DELTA_AUDIT_2026-06-03.md](./DELTA_AUDIT_2026-06-03.md) — CI/E2E delta (same day)
- [ADR-0017](../adr/0017-pcf-adapter-behind-mqadmin.md) — mqpcf scaffold
