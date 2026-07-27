# Engineering guidelines

Operator *how well* — error handling, robustness, security behaviour, and definition of done.
Product NFR IDs live in [NON_FUNCTIONAL_REQUIREMENTS.md](../NON_FUNCTIONAL_REQUIREMENTS.md).

## Error taxonomy

MQ and network failures are classified per [ADR-0014](../adr/0014-mq-error-taxonomy-and-requeue.md):

| Class | Behaviour | Example |
| --- | --- | --- |
| **Transient** | Requeue with backoff; no terminal condition | mqweb 5xx, connection reset |
| **Terminal** | Set `Ready=False` with clear message; no hot loop | Auth failure, invalid MQSC |
| **Configuration** | Surface in status; user must fix spec/Secret | Missing Secret, bad endpoint |

Never log credentials or full mqweb bodies at default log levels ([NFR SEC-5](../NON_FUNCTIONAL_REQUIREMENTS.md)).

## TLS and credentials

- All mqweb traffic is **HTTPS** with certificate verification on by default ([NFR SEC-2](../NON_FUNCTIONAL_REQUIREMENTS.md)).
- Custom CA material comes from a referenced `Secret` (`caSecretRef`).
- `insecureSkipVerify` is **opt-in**, annotation-guarded, dev-only — never default in samples for production paths.
- Credentials live in Kubernetes `Secret`s referenced by `QueueManagerConnection` only ([NFR SEC-1](../NON_FUNCTIONAL_REQUIREMENTS.md)). This holds for the `spec.authentication` union too: every mode references its material by `secretRef`, never inline. Never fold union secret material into any other field or log it.
- The mqweb admin auth **mode** is selected by the optional `spec.authentication` union ([ADR-0027](../adr/0027-mqweb-authentication-modes.md)); when omitted it defaults to Basic reading `credentialsSecretRef`. Exactly one credential source must be present (CEL-enforced). Structural exclusivity of the union is **CEL-first** ([ADR-0025](../adr/0025-cel-first-admission-validation.md)); the validating webhook only performs stateful checks (Secret existence/keys), never structural exclusivity.
- **ClientCert / mTLS (`mode: ClientCert`, AUTH-16):** authentication is at the TLS transport — the `tls.crt`/`tls.key` keypair from a `kubernetes.io/tls` Secret is loaded onto `tls.Config.Certificates` and **no `Authorization` header** is sent (the strategy is a no-op decorator, not nil — nil would fall back to Basic). Server-auth trust (`caSecretRef`) is independent and unchanged. Error taxonomy for this mode: a **malformed/mismatched keypair** is a **configuration-class** error surfaced at client build (`Reason=InvalidClientCertificate`, non-transient, no hot loop); a **server-side rejection of the certificate** (untrusted/expired/no DN mapping) surfaces as a **terminal Unauthorized** whose message points at the mqweb DN→user-registry-mapping prerequisite MKurator does **not** configure (ADR-0002 / ADR-0009 "documented, not implemented"; no OCSP). A purely local network failure stays **transient**. The keypair Secret's `tls.crt`/`tls.key` shape check lives in the **stateful webhook** (`internal/validation/`), never in CEL.
- The mqrest `ClientFactory` re-reads the v1beta1 hub to resolve the union's effective credentials Secret, and the client cache fingerprint keys off **that** Secret's resourceVersion (the secret actually used) ([ADR-0023](../adr/0023-connection-client-cache-lifecycle.md)). Because the fingerprint keys off the union's effective Secret, rotating a `spec.authentication.*.secretRef` Secret changes `credRV` and triggers replace-on-mismatch in `ForConnection` (the old transport's idle connections are closed).
- **Union-secret rotation is watch-driven (AUTH-14, closes the AUTH-12 gap):** the Secret watch re-reads the v1beta1 hub for each candidate `QueueManagerConnection` and matches **every** authentication-union member ref (`basic`/`ltpa`/`clientCert` `.secretRef.name`) in addition to the spoke's `credentialsSecretRef`/`caSecretRef` — including the stripped-data informer path (`secretContentChanged` fires on resourceVersion). So rotating a union secret now enqueues the owning QMC and invalidates the cached client on the next reconcile, without needing an unrelated spec change or operator restart. If the v1beta1 hub is unreadable (not registered / NotFound), the watch degrades to spoke-only matching. `ReleaseConnection` still reads no Secrets, so a deleted union Secret never blocks the finalizer (ADR-0023 rule 3). The matching is generic over union members, so LTPA (AUTH-13) and ClientCert (AUTH-16) inherit the watch and fingerprint for free — a rotated `clientCert.secretRef` keypair Secret enqueues the owning QMC and swaps the cached client (closing the old transport) with no extra wiring. There is still no periodic-resync `RequeueAfter` backstop on the QMC Ready path — recovery is event-driven via the Secret watch, which is the parity-with-Basic guarantee AUTH-14 provides.

## Reconciliation robustness

- Reconcilers are **idempotent** ([NFR REL-1](../NON_FUNCTIONAL_REQUIREMENTS.md)): repeated passes converge; no side effects from duplicate work.
- **Finalizers** delete MQ objects before CR removal ([ADR-0013](../adr/0013-finalizers-and-deletion.md)).
- **Drift policy** for queues/topics/channels uses DISPLAY vs DEFINE matrices ([ATTRIBUTE_RECONCILIATION.md](../ATTRIBUTE_RECONCILIATION.md)).
- Periodic requeue is a **backstop** for mqweb freshness; watch-driven triggers (CR, QMC, referenced Secrets) are primary.

## Webhook availability

Validating admission uses `failurePolicy: Fail` ([ADR-0009](../adr/0009-validating-admission-webhooks.md)). Stateless rules should migrate to CEL CRD validations over time to shrink the blast radius.

## Definition of done

A change is done when:

1. The right **test tier** is updated (see [testing.md](testing.md)).
2. Generated artifacts are fresh (`task verify`).
3. Lint and format are clean (`task lint`, `task format:check`).
4. Non-obvious decisions have an **ADR** or update an existing one.
5. User-facing behaviour is reflected in docs/samples when applicable.

## Related documents

| Document | Owns |
| --- | --- |
| [coding-standards.md](coding-standards.md) | Go style, lint, CI gates |
| [testing.md](testing.md) | Test pyramid and coverage |
| [../CONTRIBUTING.md](https://github.com/platformrelay/MKurator/blob/main/CONTRIBUTING.md) | PR process and DCO |
| [../AGENTS.md](https://github.com/platformrelay/MKurator/blob/main/AGENTS.md) | AI agent workflow |
