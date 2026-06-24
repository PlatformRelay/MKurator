# Attribute reconciliation model

MKurator applies IBM MQ objects through **mqweb `runCommandJSON`** (`DEFINE … REPLACE`).
Reconcilers compare desired `spec.attributes` to **DISPLAY** results before re-applying.

Implementation lives in `internal/adapter/mqrest/mqsc_params.go` (DISPLAY parameter lists)
and `internal/mqadmin/attrmatch.go` (value comparison). Decision record:
[adr/0010-drift-based-mq-reconciliation.md](adr/0010-drift-based-mq-reconciliation.md).
See [IBM_MQ_OBJECTS.md](IBM_MQ_OBJECTS.md) for MQSC semantics.

## How it works

| Layer | Behaviour |
|-------|-----------|
| **DEFINE** | Any lowercase key in `spec.attributes` is forwarded (numeric coercion where configured; topic `topstr` → mqweb `topicStr`; topic `pub`/`sub` uppercased for DEFINE). |
| **DISPLAY** | Only attributes listed per object type are requested from mqweb (some keywords return `MQWB0120E` on IBM MQ 9.4.x and are omitted). |
| **Drift** | For each desired key, observed DISPLAY value must match (`AttributeValueMatches` — case-insensitive for policies, numeric-normalized for counters). |
| **`Synced=True`** | Object exists and every **desired** key that we can observe matches; define-only keys are not verified after apply. |

### Capability probing (ADR-0024 §4)

Runtime DISPLAY probing for define-only local-queue attributes is wired for
**`share`**: `GetQueue` and `ResolveQueueDriftCheckKeys` probe
`SYSTEM.DEFAULT.LOCAL.QUEUE` once per mqrest client and include the attribute
when mqweb allows it. See [DISPLAY_CAPABILITY_PROBE.md](DISPLAY_CAPABILITY_PROBE.md).
Remaining candidates (`defopts`, `bothresh`, …) stay static-deferred until probed.

## Typed spec fields (Phase 8a)

Per [ADR-0021](adr/0021-attribute-api-shape.md), drift-checked MQ parameters are
also exposed as typed `spec` fields on `Queue`, `Topic`, and `Channel`.
Reconcilers fold non-empty typed fields into the same attribute map before
calling `mqadmin` (`toMQQueueSpec`, `toMQTopicSpec`, `toMQChannelSpec` in
`internal/controller/`). CEL admission rejects setting both the typed field and
the same key in `spec.attributes`.

| CRD | Typed `spec` field | MQ attribute key | Drift (DISPLAY) |
|-----|-------------------|------------------|-----------------|
| `Queue` | `maxDepth` | `maxdepth` | yes |
| `Queue` | `description` | `descr` | yes |
| `Queue` | `defPersistence` | `defpsist` | yes |
| `Queue` | `get`, `put` | `get`, `put` | yes |
| `Queue` | `targetQueue` | `targq` | yes (alias only) |
| `Queue` | `xmitQueue`, `remoteQueueManager` | `xmitq`, `rqmname` | yes (remote only) |
| `Topic` | `topicString` | `topstr` | yes |
| `Topic` | `description` | `descr` | yes |
| `Topic` | `publish`, `subscribe` | `pub`, `sub` | yes |
| `Topic` | `defPersistence` | `defpsist` | yes |
| `Topic` | `publishScope`, `subscribeScope` | `pubscope`, `subscope` | yes |
| `Channel` | `description` | `descr` | yes |
| `Channel` | `maxMsgLength` | `maxmsgl` | yes |
| `Channel` | `transportType` | `trptype` | yes |
| `Channel` | `shareConv` | `sharecnv` | yes |
| `Channel` | `mcaUser` | `mcauser` | yes |
| `Channel` | `maxInstances`, `maxInstancesClient` | `maxinst`, `maxinstc` | yes |
| `Channel` | `sslCipherSpec`, `sslClientAuth` | `sslciph`, `sslcauth` | yes |

Typed fields and `spec.attributes` keys in the tables below refer to the same
reconciliation path once folded.

## Reconciled object types (v1alpha1)

| CRD | MQ object | `spec.type` |
|-----|-----------|-------------|
| `Queue` | `QLOCAL`, `QALIAS`, `QREMOTE` | `local` (default), `alias`, `remote` |
| `Topic` | `TOPIC` | n/a |
| `Channel` | `CHANNEL` | `svrconn` only (default) |
| `QueueManagerConnection` | (connectivity, not MQSC) | n/a |

Shipped: `SET AUTHREC` via `AuthorityRecord` and `SET CHLAUTH` via
`ChannelAuthRule` (all enum `ruleType` values: `ADDRESSMAP`, `BLOCKUSER`,
`USERMAP`, `SSLPEERMAP`, `QMGRMAP`, `BLOCKADDR`). Auth reconcilers compare
desired `spec` to mqweb **GET** (`DISPLAY CHLAUTH` / `DISPLAY AUTHREC`) and
apply **replace-on-diff** (not the DISPLAY attribute matrices below). Field
matrices and CI coverage: [PHASE5_AUTH_SKETCH.md](PHASE5_AUTH_SKETCH.md).

## Attribute coverage by object

### Queue — `type: local` (`QLOCAL`)

| Attribute | DEFINE | Drift (DISPLAY) | Notes |
|-----------|--------|-----------------|-------|
| `maxdepth` | yes | yes | Numeric |
| `descr` | yes | yes | |
| `defpsist` | yes | yes | Case-insensitive match |
| `get`, `put` | yes | yes | Case-insensitive |
| `share` | yes | **probed** | Included in DISPLAY/drift when mqweb reports displayable (see [DISPLAY_CAPABILITY_PROBE.md](DISPLAY_CAPABILITY_PROBE.md)) |
| `defopts`, `bothresh`, `boqname`, `usage` | yes | **no** | DEFINE-only on mqweb 9.4 (`MQWB0120E` on DISPLAY); drift deferred |
| `maxmsglen` | yes | **no** | mqweb 9.4 rejects on DISPLAY (`MQWB0120E`) |
| trigger fields | yes | **no** | Passthrough; not in safe DISPLAY list |
| `cluster`, `clusnl` | yes | **no** | Clustering — future work |

### Queue — `type: alias` (`QALIAS`)

| Attribute | DEFINE | Drift (DISPLAY) | Notes |
|-----------|--------|-----------------|-------|
| `targq` | yes | yes | Target queue name |
| `targtype` | yes | yes | `QUEUE` or `TOPIC` |
| `descr` | yes | yes | |

### Queue — `type: remote` (`QREMOTE`)

| Attribute | DEFINE | Drift (DISPLAY) | Notes |
|-----------|--------|-----------------|-------|
| `rname` | yes | yes | Remote queue name (blank for QM alias) |
| `rqmname` | yes | yes | Remote queue manager |
| `xmitq` | yes | yes | Transmission queue |
| `descr` | yes | yes | |

### Topic (`TOPIC`)

| Attribute | DEFINE | Drift (DISPLAY) | Notes |
|-----------|--------|-----------------|-------|
| `topstr` | yes | yes | Stored as `topicStr` in mqweb JSON |
| `descr` | yes | yes | |
| `pub`, `sub` | yes | yes | Uppercased on DEFINE; case-insensitive drift |
| `defpsist` | yes | yes | |
| `pubscope`, `subscope` | yes | yes | Omitted from DISPLAY if mqweb returns `MQWB0120E` on your QM level |
| `toptype`, `cluster` | yes | **no** | Passthrough only |

### Channel (`CHLTYPE(SVRCONN)`)

| Attribute | DEFINE | Drift (DISPLAY) | Notes |
|-----------|--------|-----------------|-------|
| `trptype` | yes | yes | Case-insensitive |
| `descr` | yes | yes | |
| `maxmsgl` | yes | yes | Numeric |
| `sharecnv` | yes | yes | Numeric |
| `mcauser` | yes | yes | |
| `maxinst`, `maxinstc` | yes | yes | Numeric |
| `sslciph`, `sslcauth` | yes | yes | TLS — Phase 4 DISPLAY drift; `sslcauth` case-insensitive |

### Channel (`CHLTYPE(RCVR)`)

| Attribute | DEFINE | Drift (DISPLAY) | Notes |
|-----------|--------|-----------------|-------|
| `trptype` | yes | yes | Case-insensitive |
| `descr` | yes | yes | |
| `maxmsgl` | yes | yes | Numeric |
| `mcauser` | yes | yes | |
| `sslciph` | yes | yes | TLS cipher spec |

RCVR channels do not support `conname` or `xmitq` (unlike SDR); the partner SDR
dials the listener address configured on the receiving queue manager.

### Channel (`CHLTYPE(SDR)`)

| Attribute | DEFINE | Drift (DISPLAY) | Notes |
|-----------|--------|-----------------|-------|
| `conname` | yes | yes | Required — remote connection name (typed `connName`) |
| `xmitq` | yes | yes | Required — transmission queue (typed `xmitQueue`) |
| `trptype` | yes | yes | Case-insensitive |
| `descr` | yes | yes | |
| `maxmsgl` | yes | yes | Numeric |
| `mcauser` | yes | yes | |
| `sslciph` | yes | yes | TLS cipher spec |

## Auth — GET / drift (Phase 5)

`ChannelAuthRule` and `AuthorityRecord` reconcilers do **not** use the DISPLAY
attribute matrices above. They call `GetChannelAuth` / `GetAuthority` (mqweb
`DISPLAY CHLAUTH` / `DISPLAY AUTHREC`), parse a fixed field set, and
**replace-on-diff** via `SET CHLAUTH` / `SET AUTHREC` unless
`messaging.mkurator.dev/drift-policy=observe-only`.

Implementation: `internal/adapter/mqrest/auth.go` (DISPLAY command + attribute
parsing), `internal/mqadmin/authmatch.go` (desired vs observed).

### `ChannelAuthRule` — CHLAUTH GET

| `spec.ruleType` | DISPLAY CHLAUTH (mqweb) | Parsed from GET | Compared to `spec` (drift) |
|-----------------|-------------------------|-----------------|------------------------------|
| `ADDRESSMAP` | `DISPLAY CHLAUTH('<channel>') TYPE(ADDRESSMAP)`; appends `ADDRESS('…')` when `spec.address` is non-empty | `address`, `usersrc`, `chckclnt`, `descr` | `address`, `userSource`, `checkClient`, `description` |
| `BLOCKUSER` | `DISPLAY CHLAUTH('<channel>') TYPE(BLOCKUSER)` | `userlist`, `descr` | `userList`, `description` |
| `BLOCKADDR` | `DISPLAY CHLAUTH('<channel>') TYPE(BLOCKADDR) ADDLIST` | `addrlist` → `address`, `descr` | `address`, `description` |
| `USERMAP` | `DISPLAY CHLAUTH('<channel>') TYPE(USERMAP)`; appends `CLNTUSER('…')` when `spec.clientUser` is non-empty | `clntuser`, `mcauser`, `usersrc`, `chckclnt`, `descr` | `clientUser`, `mcaUser` (when set), `userSource` (when set), `description` |
| `SSLPEERMAP` | `DISPLAY CHLAUTH('<channel>') TYPE(SSLPEERMAP)`; appends `SSLPEER('…')` when `spec.sslPeerName` is non-empty | `sslpeer`, `mcauser`, `usersrc`, `chckclnt`, `descr` | `sslPeerName`, `mcaUser` (when set), `userSource` (when set), `description` |
| `QMGRMAP` | `DISPLAY CHLAUTH('<channel>') TYPE(QMGRMAP)`; appends `QMNAME('…')` when `spec.remoteQueueManager` is non-empty | `qmname`, `mcauser`, `usersrc`, `chckclnt`, `descr` | `remoteQueueManager`, `mcaUser` (when set), `userSource` (when set), `description` |

Notes:

- `channelName` and `ruleType` identify the rule for GET/SET; they are not
  re-read from DISPLAY for drift (the CR is the source of truth for identity).
- `ADDRESSMAP`-only SET fields (`userSource`, `checkClient`) are ignored on GET
  for `BLOCKUSER` / `BLOCKADDR` (empty desired vs empty observed matches).
- For `USERMAP` / `SSLPEERMAP` / `QMGRMAP`, `mcaUser`, `userSource`, and
  `checkClient` are compared only when non-empty in `spec` (DISPLAY may surface
  MQ defaults such as `CHCKCLNT(ASQMGR)` that the operator does not manage).
- `BLOCKADDR` SET uses MQSC `ADDRLIST(...)`; DISPLAY returns `addrlist`, mapped
  to `spec.address` in the adapter.

### `AuthorityRecord` — AUTHREC GET

| Step | Behaviour |
|------|-----------|
| DISPLAY | `DISPLAY AUTHREC PROFILE('…') OBJTYPE(…) PRINCIPAL('…')` or `GROUP('…')` |
| Parsed | `authlist` → comma-separated authority tokens |
| Drift | `spec.authorities` set must match observed set (`AuthorityNeedsUpdate`; case-insensitive, order-independent) |

`define` keys such as `profile`, `objectType`, `principal` / `group` are identity
for GET; only `authorities` is drift-checked after the rule exists.

**Profile parity (Phase 9 AUTH-9):** `OBJTYPE(QUEUE)`, `OBJTYPE(CHANNEL)`, and
`OBJTYPE(NLIST)` (MQSC `NAMELIST`) share the same SET/DISPLAY/DELETE adapter path
and reconciler logic. The CRD enum uses `NLIST`; mqrest maps it to MQSC
`NAMELIST`. Docker integration and envtest cover queue (`GET`/`PUT`), channel
(`CHG`/`DSP`), and namelist (`CHG`/`DSP`) profiles; topic profiles are also
integration-tested (`SUB`/`DSP`).

## Out of scope (not CRDs today)

| MQ surface | MQSC | Phase |
|------------|------|-------|
| Durable subscription | `DEFINE SUB` | Later |
| Model queue | `QMODEL` | Later |
| Message channels | `CHLTYPE(SVR\|RQSTR\|CLUS*)` | Later |
| Connection auth | `AUTHINFO`, `ALTER QMGR CONNAUTH` | Platform |

**Shipped (Phase 5):** OAM via `AuthorityRecord` (`SET AUTHREC`); channel auth via
`ChannelAuthRule` (`SET CHLAUTH`). Drift uses GET/replace (see [Observe-only](#observe-only-drift-policy)), not DISPLAY attribute matrices.

| MQ surface | CRD | MQSC |
|------------|-----|------|
| OAM | `AuthorityRecord` | `SET AUTHREC` |
| Channel auth | `ChannelAuthRule` | `SET CHLAUTH` |
| Alias / remote queue | `Queue` | `QALIAS`, `QREMOTE` (Phase 4) |

Sketch and rule-type roadmap: [PHASE5_AUTH_SKETCH.md](PHASE5_AUTH_SKETCH.md).

## Manual and out-of-band MQ changes

MKurator is **declarative**: the operator drives IBM MQ toward what your CRs specify. Changes made
outside the operator (MQ console, `runmqsc`, another tool) are handled as follows:

- **Drift-checked attributes** (see tables above) — on the next reconcile, DISPLAY shows a
  difference and the operator issues **DEFINE … REPLACE** to match the CR (unless observe-only;
  see below).
- **Define-only attributes** — manual edits are **not** detected; the CR must change (or you
  must alter a drift-checked field) to trigger a new DEFINE.
- **Objects with no CR** — MKurator does not delete queues, topics, or channels it does not
  manage; it only creates/updates/deletes objects backed by a CR with a finalizer.

## Observe-only drift policy

Set annotation `messaging.mkurator.dev/drift-policy=observe-only` on a `Queue`,
`Topic`, `Channel`, `ChannelAuthRule`, or `AuthorityRecord` to **report drift
without applying** DEFINE/ALTER (or auth GET/replace) to IBM MQ:

| Behaviour | Default (annotation absent) | `observe-only` |
|-----------|----------------------------|----------------|
| DISPLAY / GET | yes | yes |
| DEFINE on missing object | yes | **no** — `Synced=False`, `Reason=DriftDetected` |
| DEFINE on attribute drift | yes | **no** — `Synced=False`, `Reason=DriftDetected`, drift message in `status.message` |
| Auth replace (`SET CHLAUTH` / `SET AUTHREC`) | yes — reconcile applies desired rule | **no** — GET may still run; manual MQ changes are reported, not overwritten |
| `Synced=True` | object exists and drift-checked attrs match | same |
| Deletion | finalizer still deletes MQ object | unchanged |

For `ChannelAuthRule` and `AuthorityRecord`, drift detection uses GET against
mqweb (not DISPLAY attribute matrices). Observe-only skips replace-on-reconcile
when the observed rule differs from `spec`.

Drift comparison for queues, topics, and channels uses only DISPLAY-safe keys per
object type (define-only attributes such as `maxmsglen` are ignored for drift
detection). Implementation: `internal/controller/drift_policy.go`,
`internal/mqadmin/attrmatch.go`.

## Known limitations

1. **Manual MQ changes** to define-only attributes are not detected; re-applying the CR does not force a new DEFINE unless a drift-checked key changes.
2. **mqweb version** — DISPLAY safe lists are tuned for 9.4.x; older queue managers may need list adjustments (see Phase 2 roadmap note on `maxmsglen`).
3. **Open attribute map** — typos in keys fail at MQ apply time with MQSC errors, not Kubernetes schema validation.

## Related docs

- User-facing field tables: [INSTALL_AND_USE.md](INSTALL_AND_USE.md#attribute-reconciliation)
- Delivery plan: [ROADMAP.md](ROADMAP.md)
