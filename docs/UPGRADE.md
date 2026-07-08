# Upgrading MKurator

How to upgrade the MKurator operator between releases without surprising CRD,
webhook, or workload breakage. For first-time install see
[INSTALL_AND_USE.md](INSTALL_AND_USE.md).

Doc index: [README.md](README.md)

## Safe upgrade order

Apply changes in this order on every upgrade:

1. **CRDs** — new fields, new kinds, and schema tightening land here first.
2. **Operator** — controller image, RBAC, webhooks, cert-manager objects, metrics Service.
3. **Your CRs** — only after the new controller is running and webhooks are serving.

Skipping step 1 can leave the API server on an old schema while the controller expects
new fields. Upgrading CRs before the operator can cause admission failures or stale
reconcile behaviour.

```sh
VERSION=0.11.1   # target release

# 1. CRDs (release tarball or chart crds/)
kubectl apply --server-side -f install-crds.yaml
# Helm-only clusters: kubectl apply --server-side -f charts/mkurator/crds/

# 2. Operator
kubectl apply -f install.yaml
# or: helm upgrade --install mkurator … --version "${VERSION}"

kubectl -n mkurator-system rollout status deployment/mkurator-controller-manager
kubectl -n mkurator-system wait --for=condition=Ready certificate/webhook-server-cert --timeout=120s

# 3. Workload CRs (when release notes require spec changes)
kubectl apply -k config/samples/   # or your GitOps manifests
```

## Version-to-version notes

Check [CHANGELOG.md](../CHANGELOG.md) and the GitHub release for breaking changes
before upgrading.

| From | To | Highlights |
|------|-----|------------|
| **0.11.x** | **0.12.x** | **`v1beta1` API** for all six kinds with **conversion webhook** (dual served versions). See [Migrating to v1beta1 (0.11.x → 0.12.x)](#migrating-to-v1beta1-011x--012x) below. |
| **&lt; 0.5.0** | **0.5.0+** | New CRDs: `ChannelAuthRule`, `AuthorityRecord`. Validating webhooks on by default (cert-manager TLS). Review [INSTALL_AND_USE.md](INSTALL_AND_USE.md) auth sections. |
| **0.3.x** | **0.4.0+** | Validating webhooks and QMC delete protection. Ensure cert-manager is installed if using Helm/Kustomize webhook bundles. |
| **0.2.x** | **0.3.0+** | Module and image registry moved to `conduit-ops/MKurator` ([ADR-0006](adr/0006-project-name-kurator.md)). Update `image.repository` / install manifest URLs. |

Semantic versioning: **patch** — bug fixes, safe rolling image bump; **minor** —
new CR fields or kinds, may need CRD apply; **major** (or `feat!` / `BREAKING CHANGE`)
— read release notes and ADRs before upgrading production.

## Migrating to v1beta1 (0.11.x → 0.12.x)

Release **0.12.x** introduces **`messaging.mkurator.dev/v1beta1`** for all six
kinds (`QueueManagerConnection`, `Queue`, `Topic`, `Channel`, `ChannelAuthRule`,
`AuthorityRecord`) per [ADR-0026](adr/0026-v1beta1-graduation-plan.md). Existing
`v1alpha1` manifests and etcd objects continue to work — no big-bang rewrite required.

### Upgrade order

Follow the [safe upgrade order](#safe-upgrade-order) above. For this release the
critical steps are:

1. **CRDs first** — multi-version CRDs add `v1beta1` (`served: true`) alongside
   `v1alpha1` (still **storage** until a later release proves hub migration).
   Apply `install-crds.yaml` or `charts/mkurator/crds/` with server-side apply.
2. **Operator second** — the controller image registers a **conversion webhook**
   in addition to validating webhooks. Wait for rollout **and** webhook TLS before
   changing workload CRs (see below).
3. **Workload CRs last** — optional gradual `apiVersion` bump; stored `v1alpha1`
   objects convert on read.

```sh
VERSION=0.12.0   # target release (when tagged)

# 1. CRDs (includes conversion webhook clientConfig)
kubectl apply --server-side -f install-crds.yaml

# 2. Operator (conversion + validating webhook Deployment)
kubectl apply -f install.yaml
kubectl -n mkurator-system rollout status deployment/mkurator-controller-manager
kubectl -n mkurator-system wait --for=condition=Ready certificate/webhook-server-cert --timeout=120s

# 3. Verify dual versions are served
kubectl explain queue.spec --api-version=messaging.mkurator.dev/v1beta1
kubectl explain queue.spec --api-version=messaging.mkurator.dev/v1alpha1
```

### Conversion webhook TLS

The conversion webhook shares the same cert-manager posture as validating
webhooks ([Validating webhooks and cert-manager](#validating-webhooks-and-cert-manager)):

- cert-manager must be healthy before upgrading.
- The `webhook-server-cert` `Certificate` must reach **Ready** — the API server
  calls conversion over HTTPS using the same serving Secret as validation.
- If conversion fails with TLS errors after cert rotation, restart the controller
  Deployment once.

Conversion is registered on the CRD `spec.conversion` strategy; no separate
cert-manager `Certificate` is required beyond the existing webhook bundle.

### Gradual `apiVersion` bump

Both **`v1alpha1`** and **`v1beta1`** are **served** for at least **one minor
release** after `v0.12.0` ships ([ADR-0026](adr/0026-v1beta1-graduation-plan.md)):

| Posture | Version | Meaning |
|---------|---------|---------|
| Storage (etcd) | `v1alpha1` | Existing objects stay stored as-is until hub migration is proven in CI |
| Served (read/write) | `v1alpha1` + `v1beta1` | `kubectl get` may show either version; conversion handles round-trip |
| Preferred for new YAML | `v1beta1` | Samples in this repo default to `v1beta1` from 8d-4 onward |

**You do not need to rewrite all manifests immediately.** GitOps repos pinned to
`apiVersion: messaging.mkurator.dev/v1alpha1` keep reconciling. When ready,
change the `apiVersion` line (and prefer typed fields — see below); conversion
folds `spec.attributes` map keys into typed fields where unambiguous.

Example:

```yaml
# Before (still valid through the deprecation window)
apiVersion: messaging.mkurator.dev/v1alpha1
kind: Queue
spec:
  attributes:
    maxdepth: "500000"

# After (preferred for new manifests)
apiVersion: messaging.mkurator.dev/v1beta1
kind: Queue
spec:
  maxDepth: 500000
```

### `spec.attributes` deprecation timeline

On **`v1beta1`**, map keys that have a typed equivalent (for example `maxdepth` →
`spec.maxDepth`) are **deprecated**, not removed ([ADR-0021](adr/0021-attribute-api-shape.md),
[API_STABILITY.md](API_STABILITY.md)):

| Phase | Release | Behaviour |
|-------|---------|-----------|
| **Now** | `v0.12.x` | `spec.attributes` remains valid; conversion copies map values into typed fields on read; admission **warnings** when a deprecated map key is used on `v1beta1` creates/updates |
| **Later** | post-`v0.12` minor or major | Deprecated map keys on **`v1beta1`** may be **rejected** at admission; notice in CHANGELOG and this doc before enforcement |
| **Escape hatch** | indefinite on `v1beta1` | Keys with **no** typed equivalent stay in `spec.attributes` |

Setting **both** a typed field and the same key in `attributes` is rejected at
admission (no silent merge). Prefer typed fields in new manifests; use
`kubectl explain queue.spec.maxDepth --api-version=messaging.mkurator.dev/v1beta1`.

Map-only **`v1alpha1`** manifests are unaffected until you bump `apiVersion`.

## CRD schema changes and server-side apply

MKurator CRDs are generated from kubebuilder markers and shipped in release assets
(`install-crds.yaml`) and [`charts/mkurator/crds/`](../charts/mkurator/crds/).

- Prefer **`kubectl apply --server-side`** (or `kubectl apply --server-side --force-conflicts`
  on the first upgrade after a large schema change) so field management stays consistent
  with Helm and GitOps tools.
- **Helm** installs CRDs on first install only; upgrading the chart does not always
  refresh CRDs. Re-apply `install-crds.yaml` or `charts/mkurator/crds/` explicitly when
  the release notes mention API changes.
- Existing CR instances are generally **preserved** across CRD upgrades; new required
  fields may need you to patch resources or rely on webhook defaults.
- If you use **server-side apply** for workload CRs, keep a single field manager
  (your GitOps controller or `kubectl`) to avoid ownership fights on `spec`.

After CRD apply, verify:

```sh
kubectl get crd | grep messaging.mkurator.dev
kubectl explain queue.spec --api-version=messaging.mkurator.dev/v1beta1
# or v1alpha1 while both versions are served
```

## Validating webhooks and cert-manager

With `webhooks.enabled=true` (Helm default), the API server calls MKurator’s
validating webhooks over HTTPS. TLS is provisioned by **cert-manager**:

- Helm creates an `Issuer` + `Certificate` (`webhooks.certManager.create=true`).
- The signed Secret is mounted at `/tmp/k8s-webhook-server/serving-certs`; controller-runtime
  reloads when cert-manager rotates the Secret.

### cert-manager version expectations

MKurator does **not** bundle cert-manager. You must install it in the cluster
(or use a platform that already provides it).

| Environment | Reference version |
|-------------|-------------------|
| Local kind platform | **v1.18.2** (pinned in [`hack/kind-cluster/terraform/cert-manager.tf`](../hack/kind-cluster/terraform/cert-manager.tf)) |
| Production | cert-manager **v1.13+** (use a supported release from [cert-manager.io](https://cert-manager.io/docs/installation/supported-releases/); match your platform’s supported chart) |

Upgrade cert-manager on its own lifecycle **before** or **in parallel with** MKurator
only when release notes require a newer API; otherwise keep cert-manager stable and
upgrade MKurator independently.

### Webhook cert rotation

Rotation is automatic when cert-manager renews the `Certificate`:

1. Confirm cert-manager is healthy: `kubectl -n cert-manager get pods`.
2. Check webhook cert: `kubectl -n mkurator-system describe certificate webhook-server-cert`.
3. After renewal, the controller pod should continue running; if webhooks fail with
   TLS errors, restart the deployment once:  
   `kubectl -n mkurator-system rollout restart deployment/mkurator-controller-manager`.

E2e tests wait for the webhook `Certificate` to be Ready before exercising admission
— replicate that check after upgrades.

### Disabling webhooks (not recommended)

For break-glass only, Helm allows `webhooks.enabled=false`. You lose admission validation;
invalid specs will fail later at reconcile. Do not disable webhooks in production without
a documented reason.

## Operator image upgrade

**Kustomize / manifest install:** apply the new `install.yaml`; the Deployment rolls
out with the pinned `ghcr.io/conduit-ops/mkurator:<version>` image.

**Helm:**

```sh
helm upgrade --install mkurator oci://ghcr.io/conduit-ops/mkurator \
  --version "${VERSION}" \
  --namespace mkurator-system \
  --reuse-values \
  --set image.tag="${VERSION}"
```

Use `--reuse-values` to keep your metrics, webhook, and logging settings; merge in new
defaults from [charts/mkurator/README.md](../charts/mkurator/README.md) when release notes
call them out.

Wait for rollout and webhook availability before changing workload CRs.

## Workload CRs and samples

After the operator is healthy:

- Re-apply GitOps manifests or `kubectl apply` changed CRs.
- New kinds (e.g. auth CRs in 0.5.0) are optional until you need them.
- Sample YAML in this repo: canonical Kubebuilder tree [`config/samples/`](../config/samples/);
  Helm copies are synced via `task samples:sync` (see [config/samples/README.md](../config/samples/README.md)).

## Rollback

1. Re-install the **previous operator manifest or Helm chart version** (same namespace).
2. Only roll back **CRDs** if the release notes say the new schema is backward-compatible
   with the old controller — otherwise keep new CRDs and downgrade the image (may limit
   new fields).
3. Restore workload CRs from Git if needed.

## Uninstall and reinstall

For a clean reinstall, remove workload CRs first (`Queue`, `Topic`, `Channel`,
`ChannelAuthRule`, `AuthorityRecord`, then `QueueManagerConnection`), then the
operator, then CRDs — see [INSTALL_AND_USE.md#uninstall](INSTALL_AND_USE.md#uninstall).

## See also

- [INSTALL_AND_USE.md](INSTALL_AND_USE.md) — install paths and day-2 operations  
- [OBSERVABILITY.md](OBSERVABILITY.md) — metrics and Prometheus  
- [RELEASE.md](RELEASE.md) — maintainer release process  
- [charts/mkurator/README.md](../charts/mkurator/README.md) — Helm values reference  
