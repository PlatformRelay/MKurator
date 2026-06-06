# ADR-0020: CI merge-gate matrix (L0–L5 ↔ workflows)

- **Status**: Accepted
- **Date**: 2026-06-06
- **Extends**: [ADR-0011](0011-layered-testing-strategy.md), [CICD.md](../CICD.md)

## Context

[ADR-0011](0011-layered-testing-strategy.md) defines four test tiers (unit → envtest → integration → e2e).
The OSS maturity work ([ADR-0019](0019-oss-maturity-posture.md)) maps these to an L0–L5 label set aligned
with sibling project kollect, and adds security CI jobs that must stay explicit in merge policy.

## Decision

### Merge-gate matrix

| Tier | CI / workflow | Required on PR? | Required on `main`? |
| --- | --- | --- | --- |
| **Preflight** | `preflight.yaml` | Yes | Yes |
| **L0–L1 + vuln** | `ci.yaml` → `test` | Yes | Yes |
| **L2 schema** | `ci.yaml` → `verify` | Yes | Yes |
| **Lint / format** | `ci.yaml` → `lint` | Yes | Yes |
| **Build** | `ci.yaml` → `build`, `docker-build`, `helm-lint` | Yes | Yes |
| **Secrets** | `ci.yaml` → `gitleaks` | Yes | Yes |
| **RBAC audit** | `ci.yaml` → `audit-rbac` | Yes | Yes |
| **L3 integration** | `integration.yaml` | When path filter runs | When path filter runs |
| **L4 e2e** | `e2e.yaml` | Optional (latency) | Recommended |
| **L5 soak/bench** | — | No | No |
| **Nightly** | `nightly.yaml` | No | No |
| **SAST** | `codeql.yaml` | Runs on PR; not branch-protection required initially | Yes (signal) |
| **Scorecard** | `scorecard.yaml` | No | Weekly + push |
| **Vulncheck schedule** | `vulncheck.yaml` | No | Weekly |

Docs-only PRs (markdown / `docs/**` / chart README) skip integration and e2e per existing path filters.

### Release gate (unchanged)

Tagged releases still require [release-gate.yaml](../CICD.md) polling green **CI**, **Integration**, and
**E2E (kustomize)** on the release SHA before maintainer tags.

## Consequences

- Branch protection table in [CICD.md](../CICD.md) lists **`audit-rbac`** as a required check.
- [testing.md](../development/testing.md) L0–L5 table references this ADR.
- Optional future: dedicated `e2e-webhook-path.yaml` when admission surface grows (MKR-TST-04).

## Alternatives considered

- **Require e2e on every PR**: rejected — ~90 min latency; optional until team accepts cost.
- **Fold Scorecard into required checks immediately**: deferred — iterate on findings first.
