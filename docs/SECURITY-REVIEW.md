# Security self-review

Structured self-review of MKurator security posture for OpenSSF Best Practices documentation.
Maintainer-conducted review — not an independent third-party audit.

## Review metadata

| Field | Value |
| --- | --- |
| **Date** | 2026-06-06 |
| **Scope** | `main` — operator, Helm chart, CRDs/webhooks, CI/release pipelines |
| **Reviewers** | Konrad Heimel (maintainer, self-review) |
| **Method** | ADR walkthrough, CI control inventory, secret/TLS/RBAC path review |
| **Related docs** | [ASSURANCE-CASE.md](ASSURANCE-CASE.md), [SECURITY.md](https://github.com/conduit-ops/MKurator/blob/main/SECURITY.md) |

## Findings summary

| ID | Severity | Finding | Status |
| --- | --- | --- | --- |
| SR-01 | Info | NFR SEC-* documented and traced to CI | **Closed** |
| SR-02 | Info | gitleaks, govulncheck, CodeQL in CI | **Closed** — 2026-06 OSS maturity work |
| SR-03 | Info | Release cosign + SBOM + asset sign-blob + attestations | **Closed** — ADR-0016 update |
| SR-04 | Info | RBAC audit (Polaris/kubeaudit) in CI | **Closed** — `audit-rbac` job |
| SR-05 | Low | Solo maintainer — no second reviewer | **Accepted** — GOVERNANCE |
| SR-06 | Low | Validating webhook `failurePolicy: Fail` SPOF | **Accepted** — CEL migration planned |

No critical or high-severity defects identified in reviewed paths during this self-review.

## Residual risks

| Risk | Likelihood | Impact | Notes |
| --- | --- | --- | --- |
| Maintainer unavailability | Low | High | Succession in [GOVERNANCE.md](https://github.com/conduit-ops/MKurator/blob/main/GOVERNANCE.md) |
| Misconfigured cluster RBAC | Medium | High | Document least-privilege; `task audit:rbac` |
| Adopter uses insecure TLS in production | Medium | High | Dev-only gate on annotation |
| Zero-day in Go dependency | Low | Medium | govulncheck + SCA policy |

## Recommendations

1. Repeat review after next tagged release or security-relevant ADR.
2. Migrate stateless validation to CEL CRD rules to reduce webhook blast radius.
3. Enable SonarCloud after **conduit-ops** org migration.

## Sign-off

| Role | Name | Date |
| --- | --- | --- |
| Maintainer | Konrad Heimel | 2026-06-06 |
