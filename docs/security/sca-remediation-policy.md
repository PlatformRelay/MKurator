# SCA remediation policy

Software Composition Analysis (SCA) for **third-party dependencies** — `go.mod`, GitHub Actions,
and container base images. Application SAST and secret scanning are governed separately in
[coding-standards.md](../development/coding-standards.md) and [SECURITY.md](https://github.com/platformrelay/MKurator/blob/main/SECURITY.md).

## OSPS-VM-05.01 compliance

This policy defines **remediation thresholds** for SCA findings on vulnerabilities and licenses.

## Remediation thresholds

**Clock starts** when a finding is first reported by govulncheck, Dependabot/Renovate, Trivy, or
manual SBOM review.

### Vulnerability findings

| Severity band | Remediation threshold | If exceeded |
| --- | --- | --- |
| **Critical** (CVSS ≥ 9.0) | **7 calendar days** | Release blocker |
| **High** (7.0–8.9) | **30 calendar days** | Release blocker |
| **Medium** (4.0–6.9) | **90 calendar days** | Track in issue/Dependabot |
| **Low** (&lt; 4.0) | Next minor release | Best-effort |

**Zero-tolerance gates:**

| Finding | Merge | Tagged release |
| --- | --- | --- |
| Reachable vulnerability (`govulncheck ./...`) | Must pass `task vuln:check` in CI | Must pass on release commit |
| Fixable CRITICAL/HIGH in release image (Trivy) | N/A | Release workflow fails |

### License findings

MKurator is [MIT-licensed](https://github.com/platformrelay/MKurator/blob/main/LICENSE).

| License class | Examples | Action |
| --- | --- | --- |
| **Allow** | MIT, ISC, BSD, Apache-2.0 | Permitted |
| **Review** | MPL-2.0, LGPL (library use) | Review within 90 days; SBOM at release |
| **Deny** | GPL, AGPL, proprietary, UNKNOWN | Remove/replace before merge |

## Detection tools

| Tool | Finds | When | Location |
| --- | --- | --- | --- |
| **govulncheck** | Go CVEs in imported packages | PR + weekly | `ci.yaml` `test`, `vulncheck.yaml` |
| **Dependabot** | Actions + gomod updates | Weekly | `.github/dependabot.yml` |
| **Renovate** | IBM MQ chart/image, Taskfile tools | Weekly | `renovate.json` |
| **Trivy** | Image CRITICAL/HIGH | Release tag | `release.yaml` |
| **Release SBOM** | SPDX inventory | Release | `dist/sbom.spdx.json` |

## Enforcement model

- **Merge:** maintainer requires green CI including govulncheck on PRs.
- **Release:** Trivy + SBOM review before `v*.*.*` tag; cosign signatures on image, chart, and release assets.

Exceptions require maintainer approval documented in release notes or `.trivyignore` with comment.

## Related documents

- [ADR-0016](../adr/0016-release-supply-chain.md)
- [ASSURANCE-CASE.md](../ASSURANCE-CASE.md)
