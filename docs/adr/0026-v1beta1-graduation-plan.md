# ADR-0026: `v1beta1` graduation plan — conversion scope and deprecation policy

- **Status**: Accepted
- **Date**: 2026-06-18

## Context

Phase 8a–8c are complete on `v1alpha1`:

- **8a** — typed drift-checked fields alongside `spec.attributes` per
  [ADR-0021](0021-attribute-api-shape.md); mutual exclusivity enforced by CEL and
  validating webhooks.
- **8b** — [API_STABILITY.md](../API_STABILITY.md) documents alpha guarantees and
  graduation criteria.
- **8c** — optional DISPLAY capability probing per [ADR-0024](0024-mqsc-command-construction-hygiene.md)
  §4; no API shape change.

[ADR-0009](0009-validating-admission-webhooks.md) explicitly excludes conversion
webhooks today. Graduation to **`messaging.mkurator.dev/v1beta1`** for all six
kinds (`QueueManagerConnection`, `Queue`, `Topic`, `Channel`, `ChannelAuthRule`,
`AuthorityRecord`) is blocked until a **conversion webhook** and hub-spoke
multi-version CRDs are designed, implemented, and proven in CI.

**Bake-time criterion** (API_STABILITY graduation #1): hybrid typed fields shipped
on `v1alpha1` in `v0.9.x`–`v0.11.x` without schema churn on promoted fields.
`v0.11.0` + patch `v0.11.1` satisfy the one-minor-release bake before cutting
`v1beta1`.

This ADR records the **plan** for Phase 8d implementation slices (8d-1 onward).
No conversion webhook code ships in 8d-0.

## Decision

We will graduate all six kinds to **`v1beta1`** using a **hub-spoke conversion
model** with the following fixed choices.

### Hub-spoke and storage

| Version | Role | Initial posture |
| --- | --- | --- |
| `v1beta1` | **Hub** — conversion functions live here | `served: true`; target **storage** version after migration window |
| `v1alpha1` | **Spoke** — existing stored objects | `served: true`; **storage** until hub migration is proven |

**First cut:** spec/status shapes on `v1beta1` mirror `v1alpha1` (apiVersion bump
only). No reconcile behaviour change vs `v0.11.x` for equivalent specs.

**Storage migration:** start with dual-version CRDs (`v1alpha1` storage + `v1beta1`
served). Move etcd storage to `v1beta1` only after conversion round-trip envtest
and e2e prove safe migration. Do not run dual storage versions in etcd.

### Conversion scope

Deploy a **conversion webhook** alongside the existing validating webhook
(cert-manager TLS, Kustomize + Helm wiring mirroring `config/webhook`).

**Required conversion paths** (round-trip for each kind):

- `v1alpha1` ↔ `v1beta1`

**Per-kind rules:**

1. **Queue / Topic / Channel** — for each typed field promoted in 8a, if
   `spec.attributes["<key>"]` is set and the typed field is empty, populate the
   typed field (case-insensitive MQSC key per
   [ATTRIBUTE_RECONCILIATION.md](../ATTRIBUTE_RECONCILIATION.md)). If both are
   set, prefer the typed field (admission already rejects new writes with both).
   Copy remaining `attributes` keys that have no typed equivalent verbatim.
2. **ChannelAuthRule / AuthorityRecord** — typed auth fields are primary; copy
   remaining `attributes` map verbatim.
3. **QueueManagerConnection** — no attribute map; straight metadata + connection
   field copy.
4. **Status** — copy `conditions`, `observedGeneration`, `desiredMQSC` unchanged.

Shared conversion helper (not used by reconcile): e.g.
`api/v1beta1/conversion_attributes.go` with `FoldAttributesToTyped` for
Queue/Topic/Channel spoke→hub folding only. Reconcilers keep existing controller
fold logic.

**Out of scope for conversion:**

- No mutating admission webhook ([ADR-0021](0021-attribute-api-shape.md) rejects
  typed→map mirroring on create).
- No PCF adapter ([ADR-0017](0017-pcf-adapter-behind-mqadmin.md)).
- No new MQ object types or reconciler behaviour changes.

### Deprecation policy (`v1beta1` and beyond)

1. **`spec.attributes` remains** on the first `v1beta1` cut — not removed in 8d.
   Map keys with typed equivalents are **deprecated**, not deleted.
2. **New manifests** should prefer typed fields (`kubectl explain` documents them).
3. **Deprecated map keys** (keys that have a typed equivalent):
   - **Conversion** copies map values into typed fields where unambiguous so
     GitOps repos pinned to `v1alpha1` survive one upgrade cycle.
   - **Admission** emits **warnings** when a deprecated map key is used on
     `v1beta1` CRs (exact timeline documented in [UPGRADE.md](../UPGRADE.md) at
     8d-5).
   - **Rejection** of deprecated map keys on `v1beta1` is a **later** release
     (post-8d minor or major bump) with notice in CHANGELOG/UPGRADE.
4. **`v1alpha1` stays served** for at least **one minor release** after `v1beta1`
   ships so clusters and GitOps can migrate `apiVersion` gradually; conversion on
   read handles stored `v1alpha1` objects.

### Implementation slices (ordered)

| Slice | Deliverable |
| --- | --- |
| **8d-0** | This ADR — conversion scope, storage strategy, deprecation timeline |
| **8d-1** | Scaffold `api/v1beta1` types; CRD multi-version |
| **8d-2** | Conversion webhook + cert-manager wiring |
| **8d-3** | Per-kind conversion unit tests (table-driven round-trip) |
| **8d-4** | Helm/Kustomize CRD bundle serves both versions; samples default `v1beta1` |
| **8d-5** | UPGRADE.md migration guide + API_STABILITY graduation checklist |
| **8d-6** | e2e: apply `v1alpha1` → upgrade CRDs → object converts → reconcile green |

**Release:** tag a **minor** release (e.g. `v0.12.0`) when 8d exit criteria in
[ROADMAP.md](../ROADMAP.md#8d--v1beta1-graduation) are met.

### Exit criteria (8d complete)

- All six kinds expose `v1beta1` in CRD (`served: true`; hub storage or proven
  migration from `v1alpha1` storage per above).
- Conversion webhook deployed with validating webhook; envtest round-trip per kind.
- No reconcile behaviour change vs `v0.11.x` on equivalent specs.
- [UPGRADE.md](../UPGRADE.md) documents `apiVersion` bump and attributes
  deprecation warnings.
- [ROADMAP.md](../ROADMAP.md) §8d checkbox complete.

## Consequences

- GitOps repos and etcd objects on `v1alpha1` can migrate without manual YAML
  rewrites once conversion is implemented.
- CEL validation rules must be duplicated or generated for both API versions
  until `v1alpha1` is unserved; golden parity tests mitigate drift.
- Operators must run conversion webhook TLS (same cert-manager posture as
  validating webhooks).
- Map-only manifests remain valid through conversion during deprecation; removal
  of `spec.attributes` is explicitly deferred past the first `v1beta1` cut.
- 8d-0 is docs-only; no cluster or CRD change until 8d-1.

## Alternatives considered

- **Cut `v1beta1` without conversion** — breaks stored objects and GitOps repos;
  violates API_STABILITY graduation criteria. Rejected.
- **Mutating webhook to default typed fields from map on create** — rejected per
  ADR-0009 and ADR-0021 (no mutating webhooks).
- **Remove `spec.attributes` in the first `v1beta1` release** — too disruptive
  for escape-hatch users and long-tail MQSC keys. Rejected; deprecate first.
- **Single-version CRD with only `v1beta1`** — forces big-bang migration; dual
  served versions with conversion is the Kubernetes convention. Rejected.
- **Immediate etcd storage on `v1beta1` before e2e proof** — higher migration
  risk. Rejected; prove round-trip first.

## References

- [API_STABILITY.md](../API_STABILITY.md) — graduation criteria and deprecation policy
- [PROGRAM-PHASE8D.md](../../agent-context/PROGRAM-PHASE8D.md) — coordinator work slices (local)
- [ROADMAP.md](../ROADMAP.md#8d--v1beta1-graduation) — Phase 8d tracking
- [ADR-0009](0009-validating-admission-webhooks.md) — validating-only posture (superseded for conversion by this ADR addendum)
- [ADR-0021](0021-attribute-api-shape.md) — typed fields + escape hatch
- Kubebuilder [Multi-Version API](https://book.kubebuilder.io/multiversion-tutorial.html)
