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
- Credentials live in Kubernetes `Secret`s referenced by `QueueManagerConnection` only ([NFR SEC-1](../NON_FUNCTIONAL_REQUIREMENTS.md)).

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
