# Change: sonar-security-remediation-2026-07 — SonarCloud SECURITY findings

## Story map (parallel sessions)

| Story | REQs |
| --- | --- |
| CI-11…CI-16 | MK-01…MK-06 (see `agent-context/daily-backlog.md`) |

## Why

SonarCloud project `PlatformRelay_MKurator` reports **48 open SECURITY issues**
(10 MAJOR · 38 MINOR). Zero hotspots `TO_REVIEW`. The bulk (38) are `go:S4036`
PATH warnings on e2e/`test/utils` helpers — real class of bug in production
code, low practical risk in CI-controlled runners. This change dispositions
every finding and specs the meaningful FIX work with **Test:** + **Verify:**.

First committed OpenSpec change in mkurator (scaffold + remediation).

## Scope

- Disposition all 48 SECURITY issues.
- FIX CI/devcontainer HTTPS + permissions + pip binary constraint.
- SAFE dispositions for math/rand jitter and e2e PATH (with optional pin helper).

## Non-goals

- Rewriting the entire e2e helper surface in one PR (may land as a shared
  `test/utils/execenv` helper referenced by KO-style PATH pin).
- Production controller crypto RNG for drift jitter (not a security boundary).

## Analysis summary

| ID | Rule | Sev | Where | Disposition | Why |
|----|------|-----|-------|-------------|-----|
| MK-01 | `go:S2245` | MAJOR ×2 | `drift_resync.go` | **SAFE** | `math/rand` jitter for requeue — already `//nolint:gosec G404`; not a crypto/auth boundary |
| MK-02 | `githubactions:S8264` | MAJOR ×3 | `docs.yaml`, `release-gate.yaml` | **FIX** | Move workflow-level read permissions to jobs |
| MK-03 | `shell:S6506` | MAJOR ×3 | `.devcontainer/post-install.sh` | **FIX** | `curl`/`get-helm-3` without HTTPS protocol pin — add `--proto '=https' --tlsv1.2`; pin helm install script hash or vendor |
| MK-04 | `githubactions:S6506` | MAJOR | `ci.yaml` gitleaks curl | **FIX** | Same curl HTTPS enforce |
| MK-05 | `githubactions:S8541` | MAJOR | `docs.yaml` pip | **FIX** | Add `--only-binary=:all:` (or document why source builds are required) |
| MK-06 | `go:S4036` | MINOR ×38 | `test/e2e/*`, `test/utils` | **SAFE→pin** | Introduce shared PATH pin helper for all `exec.Command` in test helpers; clears Sonar without claiming production risk |

## Counterpoints considered

- *"Ignore all test/ PATH findings."* Rejected as a blanket Sonar exclusion —
  a one-line env helper is cheaper and teaches the pattern.
- *"Switch drift jitter to crypto/rand."* Rejected — wasteful and implies a
  threat model that does not exist; SAFE disposition is correct.

## Links

- SonarCloud: https://sonarcloud.io/project/issues?id=PlatformRelay_MKurator&impactSoftwareQualities=SECURITY
