# Spec — SonarCloud SECURITY remediation (2026-07)

Epic: sonar-security-remediation-2026-07 · **Level:** M / U
Sonar project: `PlatformRelay_MKurator`

---

## REQ-SONAR-MK-01: Drift jitter keeps math/rand (SAFE)

**Priority:** must · **Finding:** `AZ9Qhu4_OI8JwMZvsL9t`, `…L9u` (`go:S2245`) · **Disposition:** SAFE
**Given** `internal/controller/drift_resync.go`
**When** requeue jitter is computed
**Then** `math/rand` remains (non-cryptographic), with an existing gosec nolint **and** a short
comment that Sonar `S2245` is accepted SAFE for this control plane scheduling use
**Test:** `hack/test/sonar_mk_01_drift_rand_comment_test.sh`

**Verify:** `bash hack/test/sonar_mk_01_drift_rand_comment_test.sh`

---

## REQ-SONAR-MK-02: docs + release-gate use job-scoped permissions

**Priority:** must · **Finding:** `AZ9Qhu9jOI8JwMZvsL-a`, `…L-b`, `…L-c` (`githubactions:S8264`)
**Given** `.github/workflows/docs.yaml` and `.github/workflows/release-gate.yaml`
**When** permissions are inspected
**Then** no workflow-level `contents: read` / `actions: read` remain; jobs declare what they need
**Test:** `hack/test/sonar_mk_02_job_permissions_test.sh`

**Verify:** `bash hack/test/sonar_mk_02_job_permissions_test.sh`

---

## REQ-SONAR-MK-03: Devcontainer installer curls enforce HTTPS

**Priority:** must · **Finding:** `AZ-YINt54jfVaSob3S7x`…`S7z` (`shell:S6506`)
**Given** `.devcontainer/post-install.sh`
**When** kubectl/helm (or other) binaries are downloaded
**Then** every `curl` uses `--proto '=https' --tlsv1.2 -fsSL`; piping `get-helm-3` is replaced
with a pinned release asset download (or the script URL is pinned + checksummed)
**Test:** `hack/test/sonar_mk_03_devcontainer_curl_test.sh`

**Verify:** `bash hack/test/sonar_mk_03_devcontainer_curl_test.sh`

---

## REQ-SONAR-MK-04: CI gitleaks curl enforces HTTPS

**Priority:** must · **Finding:** `AZ-YINsw4jfVaSob3S7w` (`githubactions:S6506`)
**Given** `.github/workflows/ci.yaml` Install gitleaks step
**When** curl downloads the release tarball
**Then** the invocation includes `--proto '=https' --tlsv1.2`
**Test:** `hack/test/sonar_mk_04_ci_curl_test.sh`

**Verify:** `bash hack/test/sonar_mk_04_ci_curl_test.sh`

---

## REQ-SONAR-MK-05: Docs pip install refuses source builds

**Priority:** must · **Finding:** `AZ-YINsb4jfVaSob3S7v` (`githubactions:S8541`)
**Given** `.github/workflows/docs.yaml` pip install step
**When** Python deps are installed
**Then** the command includes `--only-binary=:all:` (or an explicit documented exception with
allowlisted packages if a dep has no wheel)
**Test:** `hack/test/sonar_mk_05_pip_only_binary_test.sh`

**Verify:** `bash hack/test/sonar_mk_05_pip_only_binary_test.sh`

---

## REQ-SONAR-MK-06: Test helpers pin PATH for exec

**Priority:** should · **Finding:** `go:S4036` ×38 · **Disposition:** FIX (shared helper)
**Given** `test/e2e/*` and `test/utils/utils.go` command runners
**When** external binaries are executed
**Then** a shared helper sets `PATH=/usr/bin:/bin` (plus repo `bin/` if required) on the
`exec.Cmd` env, clearing the PATH class of findings without disabling the rule globally
**Test:** `test/utils/execenv_test.go`

**Verify:** `go test ./test/utils/ -run ExecEnv -count=1`
