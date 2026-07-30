# ADR-0029: Drop `v1alpha1` entirely — hard cut, no soft-migration window

- **Status**: Accepted
- **Date**: 2026-07-30
- **Supersedes**: [ADR-0026](0026-v1beta1-graduation-plan.md) §Deprecation-policy
  item 4 (`v1alpha1` "stays served for at least one minor release")
- **Closes**: [ADR-0027](0027-mqweb-authentication-modes.md) §Served-versions &
  conversion sequencing question
- **Governs**: Phase 8e removal slices (8e-1 onward)

## Context

Phase 8d graduated all six kinds
(`QueueManagerConnection`, `Queue`, `Topic`, `Channel`, `ChannelAuthRule`,
`AuthorityRecord`) to `messaging.mkurator.dev/v1beta1` with a hub-spoke
conversion webhook per [ADR-0026](0026-v1beta1-graduation-plan.md). As of
2026-07-30 the CRDs ship both versions with `v1beta1` as etcd **storage**
(`config/crd/bases/messaging.mkurator.dev_*.yaml`: every kind has `v1beta1`
`served: true` + `storage: true`; `v1alpha1` `served: true` + `storage: false`).
The storage migration ADR-0026 gated on is therefore **already complete**.

[ADR-0026](0026-v1beta1-graduation-plan.md) §Deprecation-policy item 4
(`docs/adr/0026-v1beta1-graduation-plan.md:99-101`) planned to keep `v1alpha1`
**served for at least one minor release** after `v1beta1` shipped, so GitOps
repos could migrate `apiVersion` gradually with conversion-on-read covering
stored `v1alpha1` objects. That soft-migration window presumed a productive user
base pinned to `v1alpha1`.

Two facts collapse that presumption:

1. **No productive users.** MKurator has no known cluster or GitOps repo pinned
   to `v1alpha1`; the operator (2026-07-29) confirmed a hard cut with no users to
   protect. Keeping a served spoke buys migration safety nobody needs while
   carrying real cost (below).
2. **The conversion spoke is a proven data-loss surface.** The union-auth e2e
   went red because a conversion-webhook round trip through the `v1alpha1` spoke
   wiped `spec.authentication` on finalizer-add — a data-loss bug, not a race
   (fixed in PR #168, commit `d744a6f`, "preserve authentication union across
   v1alpha1 spoke round trip"). Every field added to `v1beta1` that `v1alpha1`
   cannot represent reopens this class of bug for as long as the spoke is served.

This ADR is a governance record: it closes the deprecation-window plan and the
conversion-sequencing question so the 8e removal slices do not silently
contradict [ADR-0026](0026-v1beta1-graduation-plan.md), and the sequencing
constraint [ADR-0027](0027-mqweb-authentication-modes.md) left open is formally
resolved. **8e-0 is docs-only; no CRD or code changes here** — those land in
8e-1 onward.

## Decision

We will **remove `messaging.mkurator.dev/v1alpha1` entirely** — a hard cut with
**no soft-migration window**. Removal deletes the `v1alpha1` API types, its CRD
`spec.versions` entry for all six kinds, and the hub-spoke conversion webhook
(v1beta1 becomes the single served and stored version).

### Supersession of ADR-0026 §Deprecation-policy item 4

[ADR-0026](0026-v1beta1-graduation-plan.md) item 4 ("`v1alpha1` stays served for
at least one minor release … conversion on read handles stored `v1alpha1`
objects") is **superseded**. Rationale: **no productive users** exist to migrate
gradually, and the served spoke is a live data-loss surface (PR #168). The rest
of ADR-0026 (the graduation plan, conversion scope, and `spec.attributes`
deprecation policy items 1–3) stands unchanged — only the served-window
commitment in item 4 is withdrawn.

### Closing the ADR-0027 conversion-sequencing question

[ADR-0027](0027-mqweb-authentication-modes.md) §Served-versions & conversion
(`docs/adr/0027-mqweb-authentication-modes.md:88-98`) made the auth-union type
change a **hard prerequisite**: no auth-union change may land while `v1alpha1` is
storage, because a v1beta1-only union makes v1beta1 → v1alpha1 down-conversion
lossy.

- **Constraint satisfied.** `v1beta1` is already etcd storage on all six kinds
  (verified above), and the auth union shipped on it. The prerequisite was
  honoured.
- **Now moot.** Removing `v1alpha1` deletes the down-conversion direction
  altogether, so the losslessness concern that motivated the sequencing no
  longer exists.
- **Rejected mitigation stays rejected.** ADR-0027 option 3 — "preserve via
  conversion annotations on the `v1alpha1` spoke" — **must NOT be reintroduced**
  by any removal slice. It was rejected as fragile; the hard cut removes the
  spoke it would have annotated, so there is nothing left to mitigate.

### Migration instruction for readers

A reader holding an existing `v1alpha1` manifest migrates by rewriting the
`apiVersion` from `messaging.mkurator.dev/v1alpha1` to
`messaging.mkurator.dev/v1beta1`. On the first `v1beta1` cut the spec is
**identical** to `v1alpha1` (ADR-0026: "spec/status shapes on `v1beta1` mirror
`v1alpha1`, apiVersion bump only"), so no field edits are needed for manifests
that predate any `v1beta1`-only field.

**Consequence — no conversion on read after removal.** Once `v1alpha1` is gone
from `spec.versions`, the API server can no longer convert a stored `v1alpha1`
object on read: any latent object still persisted as `v1alpha1` becomes
**unreadable / unservable**, not silently upgraded. This is why the precondition
below is mandatory, not advisory.

### Precondition — stored-version cleanliness (operator decision C, 2026-07-30)

Per the operator's **assert-clean, no stored-version migration** decision, the
hard cut carries a hard precondition:

- Removing `v1alpha1` from a CRD's `spec.versions` requires that the CRD's
  `status.storedVersions` **no longer list `v1alpha1`**. The API server refuses
  to drop a version still named in `storedVersions`.
- The hard cut **assumes no target cluster has any object stored as `v1alpha1`.**
  This assumption is **verified by a guard e2e in 8e-8** (a cluster with a
  `v1alpha1`-stored object must fail the guard, not silently lose data).
- Any cluster that ran **≤ v0.12** (before `v1beta1` became storage) may still
  hold `v1alpha1`-stored objects. Such a cluster **must first complete the
  stored-object rewrite + `storedVersions` prune** documented in
  [`docs/UPGRADE.md`](../UPGRADE.md) — read every object and re-persist it under
  `v1beta1` (e.g. `kubectl get … -o yaml | kubectl apply -f -` or a storage-
  version migrator), then patch `status.storedVersions` to `[v1beta1]` — before
  upgrading to the release that removes `v1alpha1`.

This precondition is documentation only in 8e-0; the UPGRADE.md prose and the
guard e2e are separate slices (UPGRADE prose in 8e-9, guard e2e in 8e-8).

## Consequences

- The conversion-webhook data-loss surface (PR #168 class) is **eliminated** —
  no served spoke means no lossy round trip to guard against; per-version CEL
  duplication (ADR-0026 §Consequences) also goes away.
- Upgrades become **gated, not gradual**: an operator must land on a `v1beta1`-
  stored, `v1alpha1`-pruned cluster before taking the removal release. A cluster
  that skips the precondition and still stores `v1alpha1` will find those objects
  unreadable after upgrade — hence the 8e-8 guard e2e and the UPGRADE.md
  runbook (8e-9).
- GitOps repos still pinned to `apiVersion: …/v1alpha1` break at admission after
  removal (the version no longer exists) — acceptable given no productive users;
  the one-line `apiVersion` rewrite above is the whole migration.
- `docs/API_STABILITY.md` and `docs/UPGRADE.md` need removal-cut prose; that is
  **out of scope for 8e-0** (owned by 8e-9). This ADR only records the decision
  and preconditions.
- Follow-on removal slices (8e-1…8e-10) may now delete `api/v1alpha1`, the
  conversion webhook wiring, and the `v1alpha1` CRD `spec.versions` entries
  without contradicting ADR-0026 or ADR-0027.

## Alternatives considered

- **Soft / gradual deprecation window** (keep `v1alpha1` served one minor release
  per original ADR-0026 item 4) — rejected. It protects a migration path no
  productive user needs and keeps the conversion spoke — the exact surface that
  produced the red union-auth e2e / PR #168 data-loss bug — alive for another
  release. The trigger for this ADR is precisely that this window is all cost, no
  benefit here.
- **Keep both versions served indefinitely** (no removal) — rejected. Permanently
  duplicates CEL and conversion code, and permanently carries the down-conversion
  losslessness burden ADR-0027 sequenced around. Convention favours retiring the
  alpha once the beta is storage and unused.
- **Automated stored-object migration on upgrade** (operator ships a migrator that
  rewrites `v1alpha1`-stored objects transparently) — rejected for 8e per operator
  decision C: **assert-clean** instead. With no productive users, a guard e2e
  (8e-8) that fails loudly on any `v1alpha1`-stored object plus a manual UPGRADE.md
  runbook is simpler and safer than shipping and maintaining migrator code.
- **Reintroduce ADR-0027 option 3 (conversion annotations on the `v1alpha1`
  spoke)** — rejected then and rejected now; the hard cut removes the spoke, so
  the mitigation has nothing to attach to. Recorded here so no removal slice
  revives it.

## References

- [ADR-0026](0026-v1beta1-graduation-plan.md) — `v1beta1` graduation plan;
  §Deprecation-policy item 4 superseded by this ADR
- [ADR-0027](0027-mqweb-authentication-modes.md) — mqweb auth modes;
  §Served-versions & conversion sequencing closed by this ADR
- [ADR-0009](0009-validating-admission-webhooks.md) — validating-admission posture
- PR #168 (`d744a6f`) — auth-union spoke round-trip data-loss fix (the trigger)
- [`docs/UPGRADE.md`](../UPGRADE.md) — stored-object rewrite + `storedVersions`
  prune runbook (populated in 8e-9)
- [`docs/API_STABILITY.md`](../API_STABILITY.md) — API stability & deprecation
  policy (removal-cut prose in 8e-9)
