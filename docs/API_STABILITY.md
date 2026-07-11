# API stability

This document states what the **`messaging.mkurator.dev`** API guarantees today,
how **`v1alpha1`** and **`v1beta1`** relate, and what remains before full
graduation. It satisfies Phase 8b in [ROADMAP.md](ROADMAP.md) and NFR **API-1** in
[NON_FUNCTIONAL_REQUIREMENTS.md](NON_FUNCTIONAL_REQUIREMENTS.md).

## Current version

| Item | Value |
| --- | --- |
| API group | `messaging.mkurator.dev` |
| Served versions | **`v1alpha1`** + **`v1beta1`** (all six kinds) |
| Storage version (etcd) | **`v1alpha1`** until hub migration is proven in CI ([ADR-0026](adr/0026-v1beta1-graduation-plan.md)) |
| Stability (Kubernetes meaning) | **`v1alpha1`** — alpha; **`v1beta1`** — beta (field-level stability improving) |
| MQ parameter surface | Typed spec fields (preferred on `v1beta1`) + `spec.attributes` escape hatch |
| Admission | CRD CEL (`x-kubernetes-validations`) + validating webhooks ([ADR-0025](adr/0025-cel-first-admission-validation.md)) |
| Webhooks | **Validating** (referential checks, unknown-attribute warnings) + **conversion** (`v1alpha1` ↔ `v1beta1`); no mutating webhooks ([ADR-0009](adr/0009-validating-admission-webhooks.md), [ADR-0026](adr/0026-v1beta1-graduation-plan.md)) |

Kinds: `QueueManagerConnection`, `Queue`, `Topic`, `Channel`, `ChannelAuthRule`,
`AuthorityRecord`.

New manifests should use **`apiVersion: messaging.mkurator.dev/v1beta1`**; existing
`v1alpha1` YAML and stored objects remain valid through conversion.

## What `v1alpha1` guarantees

Between tagged releases on `main`, the project aims for **deliberate, documented**
changes only:

1. **Reconcile semantics** for fields documented in
   [ATTRIBUTE_RECONCILIATION.md](ATTRIBUTE_RECONCILIATION.md) and kind-specific
   guides — drift-checked keys are corrected on the queue manager; define-only
   keys are applied on create/update but not compared on DISPLAY.
2. **OpenAPI validation** on the CRD schema (enums, patterns, CEL rules) rejects
   structurally invalid specs at admission time when the API server or webhook is
   available.
3. **Breaking changes** are called out in commit messages (`!` or
   `BREAKING CHANGE:`), [CHANGELOG.md](https://github.com/platformrelay/MKurator/blob/main/CHANGELOG.md), and [UPGRADE.md](UPGRADE.md)
   before a release tag ([CONTRIBUTING.md](CONTRIBUTING.md#breaking-changes),
   [GOVERNANCE.md](https://github.com/platformrelay/MKurator/blob/main/GOVERNANCE.md)).
4. **Status shape** (`conditions`, `observedGeneration`, `desiredMQSC` where
   present) remains the observability contract; new condition reasons may appear
   but existing `Synced` / `Ready` semantics are not removed without a breaking
   release.

## What `v1alpha1` does *not* guarantee

- **Field-level stability** — names, types, and requiredness of spec fields may
  change until consumers migrate to **`v1beta1`**.
- **Map-only MQ parameters forever** — [ADR-0021](adr/0021-attribute-api-shape.md)
  adds typed spec fields alongside `spec.attributes`; on **`v1beta1`**, map keys
  with typed equivalents are deprecated (warnings now, rejection later).
- **Silent compatibility** — typos in `spec.attributes` keys are not caught by
  OpenAPI; unknown keys may receive admission **warnings** but still apply if MQ
  accepts them.
- **Webhook availability as a hard dependency for basic validation** — stateless
  rules live in CEL; referential checks (`connectionRef`, cross-CR references)
  require the validating webhook ([ADR-0025](adr/0025-cel-first-admission-validation.md)).

## Planned maturation (Phase 8)

Phase 8 tracks on [ROADMAP.md](ROADMAP.md#phase-8--api-maturation-v1beta1-readiness):

| Track | Deliverable | Status | ADR |
| --- | --- | --- | --- |
| **8a** | Typed fields for drift-checked MQ attributes + `spec.attributes` escape hatch; mutual exclusivity (CEL); internal fold into the attribute map before `mqadmin` | **Done** | [ADR-0021](adr/0021-attribute-api-shape.md) |
| **8b** | This stability statement (published) | **Done** | — |
| **8c** | Optional DISPLAY capability probing | **Done** | [ADR-0024](adr/0024-mqsc-command-construction-hygiene.md) §4 |
| **8d** | `v1beta1` for all six kinds + conversion webhook + migration docs + e2e proof | **In progress** (8d-0–8d-4 done; 8d-5/8d-6 open) | [ADR-0026](adr/0026-v1beta1-graduation-plan.md) |

During **8a**, existing manifests that use only `spec.attributes` remain valid.
New typed fields are optional; setting both a typed field and the same key in
`attributes` is rejected at admission (no silent merge). The first promoted field
is `Queue.spec.maxDepth` (alternative to `attributes.maxdepth`).

## Graduation to `v1beta1`

The graduation plan is recorded in **[ADR-0026](adr/0026-v1beta1-graduation-plan.md)**
(hub-spoke conversion, storage migration, deprecation timeline, implementation
slices 8d-0–8d-6).

**Completed (on `main` before `v0.12.0`):**

1. **Hybrid attribute surface (8a)** shipped on `v1alpha1` and baked for at least
   **one minor release** without schema churn on promoted fields — **met** by
   `v0.11.0` + `v0.11.1`.
2. **Conversion webhook** converts stored/read objects between `v1alpha1` and
   `v1beta1` for all six kinds — **implemented** (8d-2); envtest round-trip per
   kind (8d-3); dual-version CRD bundle and samples defaulting to `v1beta1` (8d-4).

**Completed at the `v0.12.0` tag:**

3. **Deprecation policy** documented in [UPGRADE.md](UPGRADE.md) — migration
   guide and `spec.attributes` timeline (**8d-5**; this doc sync).
4. **v1beta1 validating admission** — warnings for deprecated map keys and
   referential `connectionRef` checks on `v1beta1` creates/updates (**8d-5b**).
5. CI **e2e migration proof** — apply `v1alpha1` CR, upgrade CRDs, assert
   conversion + reconcile green (**8d-6**).

**Deferred (optional, post-`v0.12.0`):**

6. **etcd storage flip** to `v1beta1` hub — separate step now that 8d-6 is green;
   do not run dual storage versions.

The **8d exit criteria** are met as of `v0.12.0`; pin the operator and CRD bundle
to a **release tag** and read CHANGELOG/UPGRADE before upgrading.

## Deprecation policy (`v1beta1`)

When a drift-checked attribute gains a typed spec field:

1. **Prefer the typed field** in new manifests (`kubectl explain` documents it).
2. **`spec.attributes["<key>"]` is deprecated** for that parameter on **`v1beta1`**
   (admission warning in `v0.12.x`; rejection in a later release — timeline in
   [UPGRADE.md](UPGRADE.md#specattributes-deprecation-timeline)).
3. **Conversion** copies map values into typed fields where unambiguous so
   existing GitOps repos keep working through one upgrade cycle.

On **`v1alpha1`**, map-only manifests remain valid with no deprecation warnings
until you bump `apiVersion`.

## Environment prerequisites

| Dependency | Supported / required |
| --- | --- |
| Kubernetes | **1.29+** for CRD CEL validation ([INSTALL_AND_USE.md](INSTALL_AND_USE.md)) |
| IBM MQ / mqweb | Administrative REST **v3**; adapter behaviour documented in [IBM_MQ_REST_API.md](IBM_MQ_REST_API.md) |

## Related documents

| Document | Role |
| --- | --- |
| [ATTRIBUTE_RECONCILIATION.md](ATTRIBUTE_RECONCILIATION.md) | Drift-checked vs define-only MQ keys (today's contract) |
| [adr/0021-attribute-api-shape.md](adr/0021-attribute-api-shape.md) | Typed fields + escape hatch decision |
| [adr/0025-cel-first-admission-validation.md](adr/0025-cel-first-admission-validation.md) | CEL vs webhook split |
| [adr/0026-v1beta1-graduation-plan.md](adr/0026-v1beta1-graduation-plan.md) | `v1beta1` hub-spoke conversion and deprecation policy |
| [UPGRADE.md](UPGRADE.md) | Release-to-release migration steps |
| [FAQ.md](FAQ.md) | Short pointers for operators |
