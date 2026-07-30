# Glossary

Terms used throughout MKurator documentation and the IBM MQ ecosystem.

| Term | Meaning |
| --- | --- |
| **AUTHREC** | Object authority record — MQ OAM permissions for a principal or group on a profile (queue, channel, etc.). Reconciled via `AuthorityRecord` CRs using `SET AUTHREC` MQSC. |
| **CHLAUTH** | Channel authentication record — controls who may connect over a channel. Reconciled via `ChannelAuthRule` CRs using `SET CHLAUTH` MQSC. |
| **Channel (SVRCONN)** | Server-connection channel — client attachment point. MKurator manages `CHLTYPE(SVRCONN)` channels via the `Channel` CR. |
| **DEFINE** | MQSC command family that creates or replaces an object definition (`DEFINE QLOCAL`, `DEFINE CHANNEL`, etc.). |
| **DISPLAY** | MQSC command that reads current object attributes; used for drift detection against CR spec. |
| **Drift** | Difference between desired CR spec and live MQ object attributes after out-of-band changes. |
| **envtest** | controller-runtime test environment with a real apiserver/etcd — used for reconciler and admission tests without a full cluster. |
| **Finalizer** | Kubernetes metadata hook that delays CR deletion until the operator completes MQ cleanup. |
| **mqadmin** | Go port (interface) abstracting MQ administration operations; reconcilers depend on this, not on HTTP details. |
| **MQSC** | MQ script command language (`runmqsc`, REST `/mqsc` endpoint). MKurator translates CR spec into MQSC executed through mqweb. |
| **mqweb** | IBM MQ Administrative REST API served over HTTPS — the sole transport adapter (`mqrest`) uses this API today. |
| **OAM** | Object Authority Manager — MQ component enforcing AUTHREC permissions. |
| **PCF** | Programmable Command Format — binary MQ admin protocol; optional `mqpcf` adapter exists but mqweb REST is primary. |
| **Profile** | Named object or pattern in CHLAUTH/AUTHREC (for example queue name or channel name pattern). |
| **QMC** | Shorthand for **`QueueManagerConnection`** — CR describing endpoint, queue manager name, and credentials Secret. |
| **Queue Manager (QM)** | IBM MQ instance holding queues, channels, and topics. MKurator does not install QMs — it administers an existing one. |
| **QLOCAL / QALIAS / QREMOTE** | Queue types reconciled by the `Queue` CR (`spec.type`: `local`, `alias`, `remote`). |
| **Reconcile loop** | controller-runtime pattern: watch CR → compare desired vs actual → call MQ admin → update `.status`. |
| **runmqsc** | CLI to run MQSC against a queue manager — useful for manual verification on kind or production. |
| **TOPIC** | Publish/subscribe topic object reconciled by the `Topic` CR. |
| **Validating webhook** | Admission hook that rejects invalid CR specs before they are stored in etcd. |
| **v1beta1** | Current MKurator API version (`messaging.mkurator.dev/v1beta1`) — the only served and stored version. The earlier **v1alpha1** was removed in v0.15.0 ([ADR-0029](adr/0029-drop-v1alpha1-hard-cut.md)). |

See also [IBM_MQ_OBJECTS.md](IBM_MQ_OBJECTS.md) (research inventory) and
[ARCHITECTURE.md](ARCHITECTURE.md) (operator design).
