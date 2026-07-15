# ADR-0027: mqweb admin authentication modes — token-first, mTLS-next, Basic-compatible

- **Status**: Accepted
- **Date**: 2026-07-15
- **Maintainer LGTM**: Konrad Heimel — 2026-07-15 (via `/open-questions`; honest LTPA-vs-mTLS delta reviewed, LTPA-first accepted as a strategic ordering; hard constraints acknowledged: union lands only after v1beta1 storage, LTPA re-auth in-client not TTL-eviction)
- **Deciders**: Konrad Heimel (maintainer) — architecture + security posture change; **maintainer LGTM required** per [GOVERNANCE.md](../../GOVERNANCE.md) (CRD shape + security posture)
- **Relates to**: [ADR-0002](0002-manage-mq-via-mqweb-rest.md) (manage MQ via mqweb REST), [ADR-0003](0003-connection-model.md) (connection model), [ADR-0023](0023-connection-client-cache-lifecycle.md) (client cache lifecycle), [ADR-0025](0025-cel-first-admission-validation.md) (CEL-first validation), [ADR-0026](0026-v1beta1-graduation-plan.md) (v1beta1 graduation)
- **External relation (different surface, does NOT supersede)**: SEVEN `mq-on-k8s` ADR-0009 — Prefer token-based (OIDC/JWT) authentication for IBM MQ (`/Users/A242168/Projects/SEVEN/IBM_MQ/mq-on-k8s/docs/adr/0009-prefer-token-based-authentication.md`)
- **Supersedes**: nothing (additive to the QMC spec)
- **Decision record**: D-2026-07-14-mqweb-auth-modes (operator-answered 2026-07-15; refined via `/design-architecture` 2026-07-15)

<!--
AgDR metadata — decision refined mid-session.
model: claude-opus-4-8 (1M)
date: 2026-07-15
trigger: operator D-2026-07-14-mqweb-auth-modes chose Option A (retain Basic, add auth union, add a NEXT mode),
         re-prioritised toward token-based over mTLS-next. This draft refines that ordering after separating
         the two auth surfaces (channel vs mqweb admin) and pinning the concrete mqweb token mechanism.
operator-confirmed 2026-07-15: (1) keep LTPA-first / mTLS-next despite the honest finding that LTPA is
         login-derived and mTLS out-scores it technically — accepted as a strategic ordering; (2) served-version
         sequencing = land the union only after v1beta1 becomes etcd storage (option 2).
-->

## Context

### Two distinct authentication surfaces — do not conflate them

IBM MQ security involves two independent authentication surfaces. MKurator sits on only one of them as a *client*:

1. **MQ client-channel** (application → queue manager, over `MQCONNX`/MQI/JMS): IBM MQ 9.4 supports JWT/JWKS (OIDC bearer) token authentication of the *application identity*. This is the surface governed by **SEVEN `mq-on-k8s` ADR-0009** ("prefer token-based OIDC/JWT, self-signed transport CA, no OCSP in the connection path"). MKurator *manages policy* for this surface — channel TLS (`Channel.spec.sslClientAuth` / `SSLCAUTH`, see `api/v1beta1/channel_types.go`), `CHLAUTH` rules (`ChannelAuthRule`), and `OAM`/`AUTHREC` (`AuthorityRecord`) — but MKurator does **not** authenticate on the channel as a client. Nothing in this ADR changes that surface.

2. **mqweb admin REST** (MKurator → mqweb, over HTTPS to `/ibmmq/rest/v3/...`): this is the surface **ADR-0027 governs**. It is how the operator issues MQSC (`internal/adapter/mqrest`, per ADR-0002). mqweb runs on WebSphere Liberty; its documented inbound authentication modes for the admin REST API are: **HTTP Basic**, **client certificate (mTLS)**, and **LTPA token** (login with credentials → LTPA cookie + CSRF state, with cookie expiry/re-login). Liberty *can* additionally be fronted by an OIDC/SSO session in some configurations, but whether the target mqweb build serves an OIDC bearer/session for the **admin REST API** is not something we assert here (see Open Questions). **mqweb does not consume MQ channel-JWTs** — so "token-based for mqweb" concretely means **LTPA** (and possibly an OIDC-session path), *not* the channel-JWT/JWKS mechanism from SEVEN ADR-0009.

**Reconciliation with SEVEN ADR-0009**: that ADR's case against client certificates — per-app rotation toil across a whole *application fleet*, and OCSP unreachability in air-gapped warehouses because the PKI CA depends on OCSP — is an argument about *application identity on the channel*. It does **not** transfer to MKurator→mqweb, which is **one operator identity** authenticating over a transport that ADR-0009 itself recommends securing with a **local/self-signed, fast-rotating CA** (no OCSP in path). Under those conditions, mTLS for mqweb is air-gap-safe and rotatable by cert-manager. We therefore evaluate the mqweb surface on its own merits and do not import ADR-0009's channel conclusion wholesale.

### Current state

- The mqweb admin identity is **HTTP Basic only**. `QueueManagerConnectionSpec.CredentialsSecretRef` (`api/v1beta1/queuemanagerconnection_types.go`, and the mirror in `api/v1alpha1/`) is `+kubebuilder:validation:Required`; the referenced Secret carries `username`/`password` (or `mqAdminUser`/`mqAdminPassword`). `internal/adapter/mqrest/client.go` calls `req.SetBasicAuth(...)` on every request (Ping and MQSC round-trips). TLS config (`caSecretRef`, `insecureSkipVerify`) is server-auth only today; there is no client-certificate path.
- The mqrest `ClientFactory` (`internal/adapter/mqrest/factory.go`) caches one client per QMC keyed by `namespace/name`, and invalidates on a fingerprint of `{generation, credentials Secret resourceVersion, CA Secret resourceVersion}` (ADR-0023). ADR-0023 **deliberately rejected TTL/LRU caching** in favour of identity-keyed, resourceVersion-driven replacement.

### Why revisit now

- v1beta1 graduation (ADR-0026, Phase 8d) is the natural — and only clean — window to add fields to the QMC public API. The first v1beta1 cut is an apiVersion bump only (spec mirrors v1alpha1); introducing an auth union means deciding *now* how it interacts with hub-spoke conversion while `v1alpha1` remains storage.
- Basic-only sends shared credentials on every request and offers no path to org-preferred token-based identity. Operators increasingly require a non-password admin identity.

## Options considered (mqweb admin identity)

| Option | Pros | Cons |
| --- | --- | --- |
| **Do nothing (Basic only)** | Zero work; already shipped and proven | No non-password path; sends shared creds on every request; blocks org token direction |
| **A — union + token-first + Basic-compatible** (chosen) | Backward-compatible; explicit auth union; adds a servable token mode (LTPA) first, mTLS next | New CRD surface + CEL rule + conversion work; LTPA cookie/CSRF/expiry state to manage in the client |
| B — union + mTLS-first | Strongest identity (no shared secret); air-gap-safe with local CA | Operator cert provisioning + mqweb-side DN→user registry config; larger first-cut lift; against operator's stated token direction |
| C — Basic + optional mTLS only (no token modes) | Simpler union (two modes) | Skips the org-preferred token direction entirely; re-litigates later |

## Decision

**We choose Option A**: retain HTTP Basic as the default/compatible mode, add an explicit `spec.authentication` **union** to the `QueueManagerConnection` CRD, and prioritise a **token mode first, then client-certificate (mTLS) next** for the *mqweb admin identity* — accepting the added CRD/conversion and in-client session-state complexity, over Basic-only (no token path) and over mTLS-first (larger first lift, against operator direction).

The concrete token mechanism MKurator targets **first is LTPA**, because it is *the* documented, universally-servable non-Basic token path on mqweb/Liberty's REST API and requires no queue-manager-side identity infrastructure. An **OIDC-session** path is explicitly deferred behind capability validation (Open Questions), not committed here.

> **Y-statement**: In the context of MKurator authenticating to the **mqweb admin REST API** (not the MQ channel), facing an operator preference for token-based identity and a need to preserve existing Basic deployments, we chose to **add an explicit `authentication` union to the QMC CRD with Basic (default), LTPA-token (first non-Basic target), and client-certificate/mTLS (next) modes**, and to implement re-authentication *inside the cached mqrest client* rather than via cache eviction, accepting **new public-API surface, a CEL "exactly one of" rule, lossy-conversion risk against `v1alpha1` storage, and LTPA cookie/CSRF/expiry state**, in order to **offer a non-password admin identity aligned with the org token direction while keeping every existing Basic connection working unchanged** — over **Basic-only** (no token path) and **mTLS-first** (heavier first cut, against the stated direction). This governs a *different surface* than SEVEN ADR-0009 (MQ channel), which it relates to but does not supersede.

### Honest security delta (why the ordering rests on strategy, not the technical dimensions)

**LTPA is not passwordless.** The LTPA cookie is obtained by **logging in with the same Basic credentials**. LTPA therefore does **not** remove the credential-rotation burden; its only concrete gain over Basic is not transmitting credentials on every request (the credential is exchanged once per session). The large security delta is to **mTLS** (no shared secret at all) or a true **OIDC** flow. We are honest that token-first-then-mTLS is chosen as a **strategic ordering** (alignment with the org token direction; avoiding, in the first cut, operator client-cert provisioning plus mqweb-side certificate-to-user registry mapping), **not** because LTPA out-scores mTLS on the technical dimensions below. **The operator confirmed this ordering on 2026-07-15 with the honest delta in full view** — LTPA-first is a deliberate strategic choice, not a claim that LTPA is the technical winner. That is where the subjectivity in this decision lives.

### CRD shape (union) and served-version / conversion implication

Add to `QueueManagerConnectionSpec`:

```go
// Authentication selects how MKurator authenticates to the mqweb admin REST API.
// Exactly one mode. When omitted, defaults to basic for backward compatibility.
Authentication *MQWebAuthentication `json:"authentication,omitempty"`

type MQWebAuthentication struct {
    // +kubebuilder:validation:Enum=Basic;LTPA;ClientCert
    Mode       string          `json:"mode"`
    Basic      *BasicAuth      `json:"basic,omitempty"`      // secretRef: username/password
    LTPA       *LTPAAuth       `json:"ltpa,omitempty"`       // secretRef: login username/password (LTPA is login-derived)
    ClientCert *ClientCertAuth `json:"clientCert,omitempty"` // secretRef: tls.crt/tls.key (+ optional CA)
}
```

- **`CredentialsSecretRef` becomes optional** and is retained as the implicit Basic path (when `authentication` is absent, admission treats it as `mode: Basic` reading `CredentialsSecretRef`). This preserves every existing manifest verbatim.
- **Mutual exclusivity** is enforced **CEL-first** per ADR-0025 (`x-kubernetes-validations` "exactly one of `basic`/`ltpa`/`clientCert` set, matching `mode`"; and "not both `authentication` and a Basic-implying `credentialsSecretRef` in conflicting ways"). The validating webhook covers only stateful checks (Secret existence/keys), never structural exclusivity.

#### Served versions & conversion — DECISION: sequence after v1beta1 storage (ADR-0026 dependency)

The union is naturally expressible on the **v1beta1 hub**. Per ADR-0026 the first v1beta1 cut mirrors v1alpha1 and `v1alpha1` stays **storage** until migration is proven. A v1beta1-only union makes **v1beta1 → v1alpha1 down-conversion lossy** and breaks the ADR-0026 round-trip guarantee.

**Decision (operator-confirmed 2026-07-15): option (2) — sequence the union to land only after v1beta1 becomes etcd storage.** This gives the simplest, non-lossy conversion story and preserves the ADR-0026 round-trip guarantee exactly; the cost is that the feature waits on the v1beta1 storage migration. This sequencing is a **hard prerequisite** the implementing lane must honour — no auth-union type change lands while `v1alpha1` is storage.

Alternatives, recorded and rejected for now:
1. **Mirror the union onto `v1alpha1`** as well — keeps round-trip exact and ships before storage migration, at the cost of duplicated CEL per version (per ADR-0026 §Consequences). **Documented fallback** if the feature is needed before storage migration.
3. **Preserve via conversion annotations** on the v1alpha1 spoke — fragile; **rejected**.

## Trade-off matrix (mqweb admin auth modes)

Weights sum to 100. Scores 1 (poor) – 5 (excellent), oriented so higher = better for MKurator→mqweb. **Weights are subjective**; the largest subjective lever is "org token-direction alignment", which is *strategic*, not technical (see honest-delta note above).

| Criterion (weight) | Basic | Client-cert (mTLS) | LTPA | OIDC-session |
| --- | :--: | :--: | :--: | :--: |
| **mqweb support confidence** (20) | 5 (shipped) | 4 (high) | 4 (high — documented REST path) | 2 (LOW — unverified for admin REST) |
| **Air-gap / offline viability, no OCSP in path** (15) | 5 | 4 (local CA, no OCSP) | 5 | 3 (depends on IdP reachability) |
| **Security posture** (18) | 2 (shared creds every request) | 5 (no shared secret) | 3 (login-derived; creds still exist) | 4 (true token flow) |
| **Rotation / secret-lifecycle burden** (12) | 2 | 3 (cert-manager rotates one cert) | 2 (still password-rooted) | 4 (issuer-managed) |
| **Implementation complexity — cookie/CSRF/expiry + ADR-0023 cache interaction** (20) | 5 (trivial) | 3 (TLS client cert wiring) | 2 (cookie+CSRF+expiry re-login inside client) | 1 (session + IdP integration) |
| **Blast radius on public CRD + conversion** (10) | 5 (none) | 3 (new mode + secret keys) | 3 (new mode) | 2 (new mode + uncertain shape) |
| **Org token-direction alignment** (5, strategic) | 1 | 2 | 4 | 5 |
| **Weighted total** | **3.75** | **3.86** | **3.15** | **2.60** |

**What falls out of the numbers**: on the technical dimensions alone, **client-cert (mTLS) edges out Basic**, and **LTPA does not lead**. The decision to order **LTPA first** is therefore *not* claimed to be the matrix winner — it is a deliberate strategic override (token direction + first-cut avoidance of operator cert provisioning and mqweb DN→user registry config), confirmed by the operator with the numbers visible. We record this honestly rather than re-weighting until LTPA "wins". mTLS remains the **next** mode precisely because it scores highest on the technical merits.

## Consequences

### CRD / API (both served versions + conversion + cache)

- `QueueManagerConnectionSpec` gains `authentication` (union); `CredentialsSecretRef` relaxes from Required to optional with a CEL default-to-Basic rule (ADR-0025). New CEL rules must exist for **every served version** per ADR-0026; golden parity tests must pin v1alpha1↔v1beta1.
- **Conversion**: honour the sequencing decision above (land after v1beta1 storage); add a round-trip envtest for the union (ADR-0026 slice 8d-3 style) before `v1alpha1` is unserved.
- **Cache (ADR-0023) — sharpest constraint**:
  - The `cacheFingerprint` in `factory.go` (`{generation, credRV, caRV}`) **must gain the resourceVersion of any new auth Secret** (the client-cert keypair Secret; the LTPA login Secret if distinct) — otherwise rotation of the new material will not invalidate the cached client.
  - **LTPA session state (cookie + CSRF token + expiry) expires on a timer independent of generation/resourceVersion.** ADR-0023 rejected TTL caching, so re-login must be handled **transparently inside the cached client** (e.g. re-authenticate on a 401 and retry once, refreshing cookie+CSRF), **not** by evicting/replacing the cache entry. Implementing LTPA via TTL eviction would directly contradict ADR-0023 and must be rejected in review.
  - `internal/adapter/mqrest/client.go` gains an auth strategy seam so `SetBasicAuth` (two call sites) is replaced by a per-mode request decorator (Basic header / LTPA cookie+CSRF / mTLS handled at transport). CSRF handling (`ibm-mq-rest-csrf-token`) already exists for mutations; LTPA adds a *login* CSRF/state dimension.

### Security review & governance

- Touches the **public CRD API** and **security posture** → **maintainer LGTM required** before merge (GOVERNANCE.md). Update `docs/ASSURANCE-CASE.md` (auth claims), `SECURITY.md`, and `docs/development/guidelines.md` §"TLS and credentials" (NFR SEC-1/SEC-2/SEC-5): all new secret material stays in referenced Secrets (never inline), never logged.

### Implementing story / lane must do

1. Honour the served-version sequencing: **the union lands only after v1beta1 becomes etcd storage** — no auth-union type change before then.
2. Add the union + CEL exclusivity (ADR-0025) with per-version golden tests.
3. Refactor `mqrest` request auth into a per-mode strategy; implement **LTPA login → cookie+CSRF → transparent re-login on 401** inside the cached client (no TTL cache).
4. Extend `cacheFingerprint` with the new auth-Secret resourceVersion(s); extend `secret_watch.go` to watch the new Secret ref(s).
5. Keep Basic the default; prove an existing Basic manifest reconciles unchanged (regression).
6. mTLS as the **next** slice after LTPA lands.

### Open questions / validate against a real mqweb

- **OIDC-session for the admin REST API**: does the *target mqweb build/version* actually serve an OIDC bearer/session for `/ibmmq/rest/v3/admin/...`, or is **LTPA the only non-Basic token path**? Confidence is **LOW**; this ADR does **not** assert IBM support and **gates** the OIDC mode on validating against the real mqweb + current IBM/Liberty docs. LTPA-first stands regardless.
- **LTPA specifics to validate**: cookie name (`LtpaToken2`) and lifetime on the target build; whether login requires a distinct CSRF/login endpoint vs the mutation CSRF header already handled; behaviour on cookie expiry mid-batch (drives the transparent-re-login retry design).
- **mTLS specifics to validate**: mqweb/Liberty client-certificate authentication config and the **certificate DN → MQ user** registry mapping required on the mqweb side (a deployment prerequisite, not something MKurator configures) — mirrors ADR-0002's "mqweb enabled is a prerequisite, not a goal" stance.
