# ADR-0030: Operator distribution via Artifact Hub and OperatorHub

- **Status**: Proposed (maintainer LGTM required — GOVERNANCE / distribution surface)
- **Date**: 2026-08-06
- **Extends**: [ADR-0016](0016-release-supply-chain.md), [ADR-0019](0019-oss-maturity-posture.md)
- **Clarifies**: [ADR-0012](0012-operator-scope-existing-queue-manager.md) does **not** block this packaging

<!-- AgDR: architect role · 2026-08-06 · trigger: hub distribution parity plan (Attune pattern) -->

## Context

MKurator already publishes a signed multi-arch image and signed Helm OCI chart to GHCR
([ADR-0016](0016-release-supply-chain.md)). Adopters discover operators primarily through
**Artifact Hub** (Helm) and **OperatorHub / OpenShift OperatorHub** (OLM bundles). Sibling
reference Attune shows a lean pattern: Chart.yaml annotations + Verified Publisher OCI
metadata (`artifacthub-repo.yml` via `oras push`), hand-templated registry+v1 OLM bundles,
and dual community-operators PRs from release — without in-repo FBC/`opm` catalog builds.

[ADR-0012](0012-operator-scope-existing-queue-manager.md) rejects **Queue Manager lifecycle**
management (install/scale/upgrade QM). That is unrelated to packaging **MKurator itself** for
OperatorHub. OLM here ships the MKurator controller + its CRDs (queues, topics, channels, … on
an existing QM), not an IBM MQ Queue Manager operator.

Forces:

- Keep Helm OCI as the primary install path (GitOps-friendly, already signed).
- Avoid CLOMonitor / Docker Hub chart mirror / Krew (still deferred or N/A per ADR-0005/0019).
- Soft-fail optional hub jobs so a broken OperatorHub PAT never blocks image/chart publish.
- Package identity: **`mkurator`**, channel **`stable`**.

## Options considered

| Option | Pros | Cons |
| --- | --- | --- |
| **A. Full Attune parity** (AH annotations + Verified Publisher oras + hand CSV/bundle + dual PRs) | Matches proven sibling; discoverable on both hubs; digest-pinned CSV | Hand-CSV can drift from Helm/kustomize RBAC unless generated from real sources |
| **B. Artifact Hub only** | Smaller surface; Helm already primary | Misses OpenShift / OperatorHub.io adopters |
| **C. operator-sdk generate bundle / in-repo FBC** | More “official” OLM tooling | Heavier toolchain; FBC ops cost; conflicts with lean tooling (ADR-0005) |
| **D. Do nothing** | Zero maintenance | No hub discoverability beyond GHCR URL knowledge |

### Weighted trade-off (subjective scores 1–5)

| Criterion (weight) | A | B | C | D |
| --- | ---: | ---: | ---: | ---: |
| Adopter discoverability (3) | 5 | 3 | 5 | 1 |
| Operability / lean tooling (3) | 4 | 5 | 2 | 5 |
| Supply-chain continuity w/ ADR-0016 (2) | 5 | 5 | 4 | 3 |
| Blast radius on release success (2) | 4 | 5 | 3 | 5 |
| Team familiarity (Attune pattern) (1) | 5 | 3 | 2 | 1 |
| **Weighted total** | **55** | **46** | **37** | **35** |

Scores for “lean tooling” and “blast radius” are judgment calls; soft-fail + secret-gated
OperatorHub jobs are what keep A’s blast radius acceptable.

## Decision

We chose **option A — full Attune-style Artifact Hub + OperatorHub distribution** because it
maximizes discoverability while staying within our existing GHCR/cosign release path, accepting
hand-maintained OLM templates and registration prerequisites, over B/C/D which either leave
OperatorHub empty, add FBC machinery we do not want, or leave hubs dark.

Contract:

1. **Helm OCI remains primary.** OperatorHub is an additional catalog path, not a replacement.
2. **Artifact Hub:** enrich `charts/mkurator/Chart.yaml` with `artifacthub.io/*` annotations;
   root `artifacthub-repo.yml` (placeholder `repositoryID` until operator registers the OCI repo);
   release step `oras push …/mkurator:artifacthub.io` after helm push + cosign.
3. **OperatorHub:** hand-templated bundle under `config/olm/` (`template/` CSV + annotations,
   `ci.yaml`); `make generate-olm-bundle` copies owned CRDs from `config/crd/bases` and
   substitutes version/date/digest/icon; package name **`mkurator`**, channel **`stable`**.
4. **CSV images digest-pinned** (`relatedImages` + deployment image =
   `ghcr.io/<owner>/mkurator@sha256:…`). Cluster permissions derived from real
   `config/rbac/role.yaml` (+ leader-election rules as needed) — not invented.
5. **`hack/operatorhub-pr.sh`** opens/updates PRs to forks of
   `k8s-operatorhub/community-operators` and
   `redhat-openshift-ecosystem/community-operators-prod` (OpenShift versions annotation
   default **`v4.19`** unless operator overrides). Release job **`operatorhub-pr`** runs only
   when `secrets.OPERATORHUB_PAT` is set; step **`continue-on-error: true`** (soft-fail) so
   core release still succeeds.
6. **ADR-0012 clarification:** QM lifecycle stays out of scope; this ADR packages **MKurator**,
   not MQ. Feature requests for `QueueManager` CRDs remain rejected under ADR-0012.
7. **Skipped:** Krew, Docker Hub chart mirror, in-repo FBC/`opm`, CLOMonitor.

## Consequences

- Enables Artifact Hub Verified Publisher once `repositoryID` is real and `oras push` runs on tag.
- Enables OperatorHub.io + OpenShift catalog submission once org forks + `OPERATORHUB_PAT` exist.
- Adds maintenance: keep CSV RBAC/CRDs in sync with controller markers (generate from bases;
   review RBAC on role changes).
- Registration (Artifact Hub OCI repo IDs, forks, PAT) is an **operator** prerequisite; product
   lanes may land wiring with placeholders but must not open upstream community-operators PRs
   until registration is confirmed.
- Docs/README may describe hub install paths; live badges that 404 must wait until listings exist
  (or use conditional wording).

## Alternatives considered

See options table. **Do nothing** rejected for OSS adoption goals in ADR-0019.
**operator-sdk FBC** rejected as disproportionate for a solo-maintained operator.
