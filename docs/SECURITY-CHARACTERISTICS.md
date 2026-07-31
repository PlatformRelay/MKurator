# Security characteristics

A reference map of MKurator's security-relevant characteristics across **design &
architecture**, **runtime**, and **CI/CD & supply chain**. Unlike the point-in-time
[Security review](SECURITY-REVIEW.md) (a dated maintainer self-review) and the
[Assurance case](ASSURANCE-CASE.md) (the claim/argument/evidence structure), this page is
the *living* catalogue of controls with pointers to the code, ADRs, and workflows that
implement them, so it can be kept current as the system evolves.

| Field | Value |
| --- | --- |
| **Last updated** | 2026-07-31 |
| **Applies to** | `main` (post v0.15.0) |
| **Scope** | The operator (controller-manager), its admission webhooks, the mqweb REST adapter, and the CI/CD supply chain. Excludes the customer's IBM MQ deployment and cluster-level controls (RBAC grants, network policy, secret backends) which are the deploying operator's responsibility. |

---

## 1. Trust boundaries

```
   ┌── cluster admin / CRD authors ──┐        ┌─────────── IBM MQ estate ───────────┐
   │  QueueManagerConnection, Queue, │        │  mqweb REST endpoint (per QMC)      │
   │  Channel, Topic, AuthorityRecord│        │  presents a server TLS certificate  │
   └───────────────┬─────────────────┘        └──────────────────┬──────────────────┘
                   │ kube-apiserver (authn/z, admission)          │ HTTPS (TLS)
        ┌──────────▼───────────┐  watch/patch   ┌─────────────────▼─────────────────┐
        │ MKurator controller  │◄──────────────►│ mqweb REST adapter (internal/adapter│
        │  (least-priv ClusterRole)              │ /mqrest) — Basic / LTPA / mTLS     │
        └──────────┬───────────┘                └────────────────────────────────────┘
                   │ reads referenced Secrets (credentials, CA, client certs)
        ┌──────────▼───────────┐
        │ Kubernetes Secrets   │  (namespaced, RBAC-gated)
        └──────────────────────┘
```

Untrusted / attacker-influenced inputs the design must tolerate: **CRD spec fields**
(validated by admission webhooks), **mqweb REST responses** (parsed defensively; see
fuzzing below), and **referenced Secret contents**. The controller holds no long-lived
inbound listener other than the admission webhook and the metrics endpoint.

---

## 2. Design & architecture

### 2.1 Transport & TLS to IBM MQ
All queue-manager administration flows over the **mqweb REST API over HTTPS**
([ADR-0002](adr/0002-manage-mq-via-mqweb-rest.md), [ADR-0017](adr/0017-pcf-adapter-behind-mqadmin.md)).
The adapter validates the mqweb **server certificate** against a configurable CA bundle
(from the referenced Secret); hostname/SAN mismatches surface as TLS errors rather than
being silently trusted. Transport is firewall-friendly HTTPS — no bespoke MQ ports.

### 2.2 Authentication to IBM MQ
Three modes, enum-validated on the CRD
([ADR-0027](adr/0027-mqweb-authentication-modes.md); `internal/adapter/mqrest/auth.go`):

| Mode | Material | Notes |
| --- | --- | --- |
| **Basic** | user + password Secret | mqweb basic auth over TLS |
| **LTPA** | mqweb login → LTPA cookie | cookie cached; re-login on `MQWB0104E`/`MQWB0112E`, single-retry (no loops) |
| **ClientCert** | client key/cert Secret | **mutual TLS** to mqweb |

### 2.3 Admission webhooks
Validating admission webhooks ([ADR-0009](adr/0009-validating-admission-webhooks.md);
`config/webhook/`) reject malformed/unsafe CRs at admission time, with CEL and Go
validators. Webhook serving TLS is provisioned via **cert-manager** (`webhook-server-cert`);
the controller waits for the cert to be Ready before serving.

### 2.4 Least-privilege RBAC
The operator's `ClusterRole` (`config/rbac/role.yaml`) enumerates explicit
resources/verbs — **no `"*"` wildcard resources or verbs**. Leader election is a namespaced
`Role`; metrics auth is a separate scoped `ClusterRole`. The RBAC surface is audited in CI
(`audit-rbac` job → Polaris + kubeaudit).

### 2.5 Runtime hardening
The shipped pod runs hardened (`charts/mkurator/values.yaml`):
`runAsNonRoot: true`, `readOnlyRootFilesystem: true`, `allowPrivilegeEscalation: false`,
`seccompProfile: RuntimeDefault`, and **all Linux capabilities dropped**. The manager binary
is built **CGO-free / static**.

### 2.6 Secret handling

- Credentials are read from **namespaced Kubernetes Secrets**, RBAC-gated; never inlined in
  CRDs.
- Structured logs pass through a **redaction handler** (`internal/logging/redact.go`,
  `redactHandler`) that replaces sensitive values with `[REDACTED]`.
- Secrets are **never written to CR `.status` or Events** — reconcile-error classification
  (`internal/controller/events.go`) emits categorised reasons, not credential material.
- Per [ADR-0023](adr/0023-connection-client-cache-lifecycle.md), the connection *release*
  path reads **no Secrets**, bounding where secret material is touched.

### 2.7 Connection / credential cache lifecycle
Cached mqweb transports are keyed and **rotated on credential fingerprint change**
([ADR-0023](adr/0023-connection-client-cache-lifecycle.md)); rotation closes stale transports
rather than leaking them. Requeue jitter uses non-cryptographic `math/rand` **by design**
(scheduling only, no security dependence — Sonar `S2245` accepted-safe; `drift_resync.go`).

---

## 3. CI/CD & supply chain

Every control below runs in GitHub Actions (`.github/workflows/`). All third-party actions
are **pinned by commit SHA**; workflow tokens follow least privilege (deny-all or read-only
baselines with per-job opt-in — OpenSSF *Token-Permissions* = 10).

| Class | Control | Where |
| --- | --- | --- |
| **SAST** | CodeQL (Go) on every push/PR | `codeql.yaml`; `golangci-lint` incl. `gosec` |
| **SCA (deps)** | `govulncheck` (call-graph) **and blocks PRs** via the required `test` job; standalone `Vulncheck` workflow + weekly cron | `ci.yaml` (`task vuln:check`), `vulncheck.yaml` |
| **SCA policy** | `osv-scanner` with a curated `osv-scanner.toml` (justified `IgnoredVulns` for unreachable/unfixable advisories) | `osv-scanner.toml` |
| **Dep updates** | Dependabot **and** Renovate | `.github/dependabot.yml`, `renovate.json` |
| **Secret scanning** | gitleaks on every PR (full history) | `ci.yaml` (`gitleaks` job) |
| **Static quality** | SonarCloud analysis + Go coverage; accepted-safe/test-only findings suppressed as config-as-code | `sonar-project.properties` |
| **Fuzzing** | Native Go fuzzing of the untrusted mqweb DISPLAY parser (`FuzzParseMQSCDisplayAttributes`); a real crash corpus fails the job with no retry | `ci.yaml` (`fuzz` job), `internal/adapter/mqrest/mqsc_display_parse_fuzz_test.go` |
| **RBAC audit** | Polaris + kubeaudit | `ci.yaml` (`audit-rbac`) |
| **Signed releases** | **cosign keyless (OIDC) signatures**, **SLSA provenance** attestation, and **SBOM** (SPDX) attestation on the container image; signed Helm chart artifacts | `release.yaml` |
| **Merge gates** | Protected `main` ruleset — required checks, 1 approval, dismiss-stale, rebase-only; admin PR bypass reserved for the maintainer | [ADR-0020](adr/0020-merge-gate-matrix.md) |

### OpenSSF Scorecard posture
Maturity posture is tracked ([ADR-0019](adr/0019-oss-maturity-posture.md)); the Scorecard runs
weekly + on push (`scorecard.yaml`). Actionable checks are kept green (Token-Permissions,
Signed-Releases, SAST, Security-Policy, Pinned-Dependencies, Dependency-Update-Tool, Fuzzing).
Some low scores are **structural, not defects**, and are accepted:

- **Maintained / Contributors** — repository age (< 90 days) and a single maintaining org.
- **Code-Review** — an artifact of the solo maintainer's `--admin` rebase-merge flow (the
  independent-review gate is enforced by process, see ADR-0020), not absence of review.
- **Branch-Protection** — the Scorecard token cannot read *rulesets* (this repo uses rulesets,
  not classic branch protection), yielding a false internal error.

---

## 4. Residual risks & deployment-owned controls

- **Network egress** is deployment-owned: the controller egresses to each mqweb endpoint
  (HTTPS) and the kube-apiserver. A cluster `NetworkPolicy` restricting egress is **not**
  shipped by the chart and should be applied by the deploying operator per their environment.
- **changelog-sync** self-heal requires a GitHub App as a `protect-main` **bypass actor**;
  until activated it is inert (the `sync` job skips cleanly). The bypass actor is a
  privileged trust anchor and should be a dedicated, least-privilege App.
- The [Security review](SECURITY-REVIEW.md) self-review is point-in-time; refresh it
  alongside this page when the security surface changes.

---

## 5. Verify it yourself

```bash
task vuln:check                     # govulncheck (call-graph) — 0 findings
osv-scanner --lockfile=go.mod       # SCA vs curated osv-scanner.toml
task lint                           # golangci-lint incl. gosec
go test -run='^$' -fuzz='^FuzzParseMQSCDisplayAttributes$' -fuzztime=30s ./internal/adapter/mqrest/
task audit-rbac 2>/dev/null || bash hack/audit-rbac.sh   # Polaris + kubeaudit
gh api repos/platformrelay/MKurator/actions/workflows --jq '.workflows[] | .name'  # SAST/scans present
```

*Maintenance: keep the “Last updated” date and the control table in sync with
`.github/workflows/` and the referenced ADRs whenever the security surface changes.*
