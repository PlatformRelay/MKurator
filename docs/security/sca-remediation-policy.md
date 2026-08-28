# SCA remediation policy

Software Composition Analysis (SCA) for **third-party dependencies** — `go.mod`, the Go toolchain,
GitHub Actions, and container base images. Application SAST and secret scanning are governed separately in
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

## Go toolchain

The Go toolchain is a dependency like any other: its standard library is linked into the shipped
manager binary, and `govulncheck` reports stdlib CVEs as **reachable** findings that trip the
zero-tolerance merge gate above. This section governs how we move it.

### Selection rule

When `task vuln:check` reports a reachable finding whose only fix is a newer Go:

1. Take the **latest patch release on the current minor series** — not the exact `Fixed in` version,
   and not a new minor. The extra patches are already-shipped regression fixes on a line we are
   running; a minor is a different change with a different blast radius.
2. **Never bump the minor under security time pressure.** A minor bump raises the language version
   the compiler applies, recompiles every `go run …@version` and `go tool` binary (`go-arch-lint`,
   `golines`, `goimports`, `controller-gen`, `golangci-lint` incl. the `.custom-gcl.yml` plugin
   build) under a new compiler, and raises the floor for anyone importing this module. Minor bumps
   are their own PR, on their own schedule, with the full L0–L5 gate matrix
   ([ADR-0020](../adr/0020-merge-gate-matrix.md)) as the acceptance test.
3. Go supports the two most recent majors, so staying on the previous minor is not an
   end-of-life corner — it keeps receiving security patches until the minor after next ships.

### `go.mod` is the single source of truth — and its fan-out

`go.mod`'s `go` directive is authoritative. Some pins derive from it automatically; the rest are
**copies that must be moved in the same commit**. A missed copy is the characteristic failure of
this change, because most of them fail silently (wrong compiler, still-green CI).

| Pin | Location | Derived? | Action on a patch bump |
| --- | --- | --- | --- |
| `go` directive | `go.mod` | **source of truth** | Edit |
| Local toolchain | `Taskfile.yml` / `Taskfile.test.yml` (`GOTOOLCHAIN: sh: … sed … go.mod`) | Yes | None |
| CI toolchain | `.github/actions/go-cache/action.yml`, `.github/workflows/release.yaml` (`go-version-file: go.mod`) | Yes | None |
| README/docs badge | `README.md`, `docs/index.md` (shields `go-mod/go-version`) | Yes | None |
| **Release builder image** | `Dockerfile` (`FROM golang:<tag>@sha256:…`) | **No** | Edit tag **and** digest. Use the **image-index** digest (`docker buildx imagetools inspect golang:<tag>` → `application/vnd.oci.image.index.v1+json`), never a per-platform digest — the release builds `linux/amd64` **and** `linux/arm64` ([ADR-0016](../adr/0016-release-supply-chain.md)), and `hack/test/sonar_mk_11_dockerfile_digest_test.sh` only checks the digest is 64 hex chars, not that it resolves for both platforms. |
| **Devcontainer image** | `.devcontainer/devcontainer.json` | **No** | Edit the tag **only**. Renovate's `Devcontainer Go base image` custom manager matches `"image": "golang:(\d+\.\d+\.\d+)-bookworm"`; adding an `@sha256:` suffix silently breaks that regex and the pin stops being tracked. It is out of scope for the Scorecard Pinned-Dependencies guard (not a Dockerfile) by consequence, not by exemption. |
| **Local-tooling guard** | `hack/tools-check.sh` (`[[ "${ver}" != 1.26.* ]]`) | **No** | None on a patch bump; **must be edited on a minor bump**. |
| **Setup docs** | `docs/LOCAL_SETUP.md` (prerequisite table, install prose, `go version` sample) | **No** | Edit. These drift unguarded — no gate reads them. |
| e2e subprocess toolchain | `test/utils/utils.go` (`GOTOOLCHAIN=local`) | **Deliberate exception** | None. e2e subprocesses use whatever `go` is on `PATH`, by design (a derived `GOTOOLCHAIN` re-downloads inside the test). Sound in CI, where `setup-go` installs from `go.mod`; **locally, a stale `PATH` Go silently builds the e2e image with unpatched stdlib.** `task tools:check` is the only warning. |

### Bundle nothing with a toolchain bump

A toolchain bump ships **alone**: `go.mod`, the non-derived pins above, the docs that copy them, and
the policy text recording the decision. No dependency bumps, no generator upgrades, no drive-by fixes
— nothing that could fail independently of the toolchain, so the commit stays revertible in one step.

This rule is earned from `adb9b87` ("bump Go 1.26.4 and sync verify artifacts"), which also carried a
Makefile `kustomize` path fix, a webhook `commonLabels` removal, the `test/utils/utils.go`
`GOTOOLCHAIN=local` change, a unit-test edit and a mock regeneration — five concerns *beyond* the
bump itself. It regenerated CRDs and `zz_generated.deepcopy.go`, and has been miscited since as
proof that "a Go bump requires regenerating CRDs". **It does not** — the real cause is this
section's other warning, not Go:

`adb9b87` changed no `controller-gen` version at all. `sigs.k8s.io/controller-tools` was `v0.19.0` in
`go.mod` both before and after it, while `Makefile`'s `CONTROLLER_TOOLS_VERSION ?= v0.20.1` has stood
unchanged since the initial scaffold (`b27dd50`). The annotations moved because the generator that
ran was the **Makefile-pinned** v0.20.1 rather than the **`go.mod`-pinned** v0.19.0. `adb9b87` is
therefore the historical proof case for the `make manifests` skew warned about below — not an
example of Go forcing regeneration.

Note the skew ran the *other way* then: the Makefile pin was **above** go.mod (v0.20.1 > v0.19.0), so
the annotation was rewritten *upward*. Today it is **below** (v0.20.1 < v0.21.0) and the same
mechanism would rewrite it *downward*. Same failure, opposite direction.

**A patch-level toolchain bump must produce zero generated-artifact drift.** If `task verify`
reports drift, that is a signal to investigate (a `gofmt` behaviour change), not an expected cost.
Regenerate with `task manifests` / `task generate`, never `make manifests` — `Makefile`'s
`CONTROLLER_TOOLS_VERSION` is skewed below the `go.mod` `tool` pin and would rewrite the
`controller-gen.kubebuilder.io/version` annotation downward across every CRD.

### Verifying the toolchain actually moved

The Go build cache key is `${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}` and a `go`-directive
bump does not change `go.sum`, so the cache key is unchanged. This is safe: Go's build-cache action
IDs incorporate the compiler's build ID, so entries compiled by the old toolchain are never reused
by the new one — the cache goes cold, it does not go wrong. `setup-go` installs the toolchain
independently of that cache.

Confirm the toolchain from the job log, not from a green check: `go version` in CI, plus a green
`task vuln:check` (which reads the *building* toolchain's stdlib version, so it cannot pass under
the old one). Look for `Setup go version spec <X>` followed by `Successfully set up Go version <X>`.

`setup-go` resolves a `go` directive carrying a patch component **exactly**, with no `toolchain`
directive present — confirmed on both majors in use here: v6.4.0 (via `.github/actions/go-cache`)
logged `Setup go version spec 1.26.7` → `go version go1.26.7 linux/amd64` on the 1.26.7 bump, and
v7.0.0 (`release.yaml`) logged `Successfully set up Go version 1.26.5` against a `go 1.26.5` module
on the v0.15.1 release run. Adding a `toolchain` directive would break this: `Taskfile.yml` derives
`GOTOOLCHAIN` by sed-ing the `go` line only, so local builds would stay behind while CI moved.

## Reachability triage

`govulncheck` reports findings at three confidence levels. Only the first is gated.

| Result class | Meaning | Gate | Action |
| --- | --- | --- | --- |
| **Symbol-level (reachable)** | A call trace reaches vulnerable code | Blocks merge and release (table above) | Fix now. If every trace bottoms out in stdlib, it is a toolchain bump — see above — and implies **zero application-code change**. |
| **Package-level** | Vulnerable package imported, no reaching trace | None | Severity-band clock starts at first report. Own PR. |
| **Module-level** | Vulnerable module in the graph only | None | Severity-band clock. Renovate/Dependabot. |

Non-reachable does **not** mean unbounded. The severity bands above apply from first report
regardless of reachability, so a non-reachable **High** carries a 30-day clock even though nothing
blocks the merge. Record the clock-start date when deferring.

Two standing qualifiers for this repo:

- **Tool-only dependencies** (reached solely via `go.mod` `tool` directives — `goimports`, `golines`,
  `mockery`, `golangci-lint`) are not linked into the manager binary and do not appear in the
  release image or its SBOM. They still carry the severity clock; they do not constrain a release.
- **Transitive Kubernetes dependencies** are bumped on their own, gated by the tests that cover the
  behaviour they actually reach — which is rarely the suite the package name suggests. Establish the
  entry point with `go mod why -m <module>` (one shortest *import* path — confirm uniqueness with
  `go list -deps`) before choosing a gate: our `cel-go`, for example, is reached only from `cmd` via
  the metrics-endpoint authentication/authorization filter, and is **not** exercised by the ADR-0025
  admission suite (see the register below). Never bump one alongside an urgent fix.

### Open non-reachable findings

This table is the **tracking mechanism of record** for deferred findings, because `renovate.json`
sets `:disableDependencyDashboard` — the band table's "Track in issue/Dependabot" has no
dashboard to point at. Maintained by hand; delete rows when closed, and check it when the
weekly `vulncheck.yaml` run reports a non-reachable finding.

| ID | Module | CVSS / band | Clock start | Due | Notes |
| --- | --- | --- | --- | --- | --- |
| [GO-2026-6094](https://pkg.go.dev/vuln/GO-2026-6094) | `github.com/google/cel-go` v0.29.2 → v0.30.0 | CVSS v4 `AV:N/AC:L/AT:P/PR:N/UI:N/VC:L` — **Medium** | 2026-08-28 | 2026-11-26 | Package-level. Reached **only** via `cmd` → `controller-runtime/pkg/metrics/filters` → `k8s.io/apiserver/pkg/authorization/authorizerfactory` → `.../authorization/cel` — the **metrics-endpoint authentication/authorization filter** (the filter pulls in `authentication/cel` too). There is no direct `cel-go` import in `internal/`, `api/` or `cmd/`. **Not gated by the ADR-0025 admission suite:** CRD `x-kubernetes-validations` are evaluated by the kube-apiserver's own compiled-in cel-go (a separate binary under envtest), so our version cannot affect them. Gate on the metrics-endpoint authorization path instead. `k8s.io/apiserver` v0.36.3 requires only v0.26.0 against our v0.29.2, so MVS divergence is not the risk either; the advisory is JSON private-field exposure via `NativeTypes`/`ParseStructTag`. |
| [GO-2026-6180](https://pkg.go.dev/vuln/GO-2026-6180) | `golang.org/x/mod` v0.37.0 → v0.40.0 | CVE-2026-56864, CVSS v3.1 **7.5 High** | 2026-08-28 | 2026-09-27 | **Module-level**, tool-only (`x/tools/cmd/goimports`); `sumdb` verification, not in the manager binary. High band ⇒ 30 days despite no gate. |
| [GO-2026-6179](https://pkg.go.dev/vuln/GO-2026-6179) | `golang.org/x/mod` v0.37.0 → v0.40.0 | CVE-2026-56865, CVSS v3.1 **8.4 High** | 2026-08-28 | 2026-09-27 | **Module-level**, as above (`sumdb/tlog` tile verification bypass). Same one-line fix as GO-2026-6180. |

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

- [ADR-0005](../adr/0005-keep-tooling-lean.md) — pinned tool versions as an adopted practice
- [ADR-0016](../adr/0016-release-supply-chain.md)
- [ADR-0020](../adr/0020-merge-gate-matrix.md) — the L0–L5 gate matrix a minor bump must clear
- [ASSURANCE-CASE.md](../ASSURANCE-CASE.md)
- [LOCAL_SETUP.md](../LOCAL_SETUP.md) — Go prerequisite (an unguarded copy of the `go.mod` pin)
