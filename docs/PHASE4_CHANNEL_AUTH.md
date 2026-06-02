# Phase 4 — channel and authority API sketch

Planning document for Kurator **Phase 4** ([ROADMAP.md](ROADMAP.md)). It maps
reference MQSC from
ibm-messaging/mq-gitops-samples `qmdemo-mqsc-config-map.yaml` (mirrored in [`test/e2e/fixtures/channel-auth-prereq.mqsc`](../test/e2e/fixtures/channel-auth-prereq.mqsc))
and patterns in [IBM_MQ_OBJECTS.md](IBM_MQ_OBJECTS.md) to future CRD fields.

Kurator will reconcile these via the existing **mqweb `/mqsc`** path ([ADR-0002](adr/0002-manage-mq-via-mqweb-rest.md)),
not via IBM’s ConfigMap-at-`QueueManager` bootstrap model.

## Reference MQSC (gitops basic deployment)

Source: IBM `mq-gitops-samples` (Apache-2.0 header in upstream file). Kurator e2e
fixture: [`test/e2e/fixtures/channel-auth-prereq.mqsc`](../test/e2e/fixtures/channel-auth-prereq.mqsc).

```mqsc
DEFINE CHANNEL('DEV.APP.SVRCONN.0TLS') CHLTYPE(SVRCONN) TRPTYPE(TCP) +
  MCAUSER('app') SSLCIPH('') SSLCAUTH(OPTIONAL) REPLACE

SET CHLAUTH('DEV.APP.SVRCONN.0TLS') TYPE(ADDRESSMAP) ADDRESS('*') +
  USERSRC(CHANNEL) CHCKCLNT(REQUIRED) +
  DESCR('Allows connection via APP channel') ACTION(REPLACE)
```

## Proposed resources (names tentative)

### `ServerConnectionChannel` (or `Channel` with `type: svrconn`)

| CRD field | MQSC / behavior | Notes |
|-----------|-----------------|-------|
| `spec.connectionRef` | — | Same as `Queue` — namespaced `QueueManagerConnection` |
| `spec.channelName` | `DEFINE CHANNEL('…')` | IBM MQ object name |
| `spec.transportType` | `TRPTYPE` | Default `TCP` |
| `spec.mcaUser` | `MCAUSER` | Default OS/user for OAM when not mapped |
| `spec.sslCipher` | `SSLCIPH` | Empty string = no TLS on channel (dev only) |
| `spec.sslCaAuth` | `SSLCAUTH` | `OPTIONAL`, `REQUIRED`, etc. |
| `spec.maxMsgLen` | `MAXMSGL` | Optional |
| `spec.shareConv` | `SHARECNV` | Optional |
| `spec.description` | `DESCR` | Optional |

**Reconcile:** `DEFINE CHANNEL(...) CHLTYPE(SVRCONN) ... REPLACE` via `runCommandJSON`
or `runCommand`; **delete:** `DELETE CHANNEL('…')`.

**Observe:** `DISPLAY CHANNEL('…') CHLTYPE(SVRCONN)` (extend `MQAdmin` with
`GetChannel`).

### `ChannelAuthRule` (CHLAUTH)

| CRD field | MQSC | Maps from gitops example |
|-----------|------|---------------------------|
| `spec.connectionRef` | — | |
| `spec.channelName` | channel name in `SET CHLAUTH('…')` | `DEV.APP.SVRCONN.0TLS` |
| `spec.ruleType` | `TYPE` | `ADDRESSMAP` |
| `spec.address` | `ADDRESS` | `*` |
| `spec.userSource` | `USERSRC` | `CHANNEL` |
| `spec.checkClient` | `CHCKCLNT` | `REQUIRED` |
| `spec.description` | `DESCR` | |
| `spec.action` | `ACTION` | `REPLACE` (reconcile), `REMOVE` on delete |

**Reconcile:** `SET CHLAUTH(...) ACTION(REPLACE)`; **delete:** `SET CHLAUTH(...) ACTION(REMOVE)` or `REMOVEALL` per rule scope (design choice: one CR per rule).

Additional rule types from [IBM_MQ_OBJECTS.md §6.3](IBM_MQ_OBJECTS.md#63-channel-authentication-set-chlauth):

| `ruleType` | Typical use |
|------------|-------------|
| `BLOCKUSER` | `USERLIST` — deny privileged IDs |
| `USERMAP` | Map `CLNTUSER` to `MCAUSER` |
| `SSLPEERMAP` | Map TLS DN |
| `QMGRMAP` | Map remote QM name |
| `BLOCKADDR` | Block IPs at listener |

### `AuthorityRecord` (OAM — `SET AUTHREC`)

Not in the gitops basic sample but **P0** in [IBM_MQ_OBJECTS.md §5](IBM_MQ_OBJECTS.md).
Likely shape:

| CRD field | MQSC |
|-----------|------|
| `spec.connectionRef` | — |
| `spec.profile` | `PROFILE('…')` queue or `channel` name |
| `spec.objectType` | `OBJTYPE` — `QUEUE`, `CHANNEL`, … |
| `spec.principal` / `spec.group` | `PRINCIPAL` / `GROUP` |
| `spec.authority` | `AUTHADD` / `AUTHRMV` list — `GET`, `PUT`, `CONNECT`, … |

**Reconcile:** `SET AUTHREC ... AUTHADD(...) ACTION(REPLACE)`; use `setmqaut`-equivalent
semantics documented in IBM_MQ_OBJECTS §5.

## `MQAdmin` port extensions (Phase 4)

```go
// Illustrative — signatures to be finalized in Phase 4.
DefineServerConnectionChannel(ctx context.Context, spec ChannelSpec) error
GetServerConnectionChannel(ctx context.Context, name string) (*ChannelState, error)
DeleteServerConnectionChannel(ctx context.Context, name string) error

SetChannelAuth(ctx context.Context, spec ChannelAuthSpec) error
DeleteChannelAuth(ctx context.Context, spec ChannelAuthSpec) error

SetAuthority(ctx context.Context, spec AuthoritySpec) error
DeleteAuthority(ctx context.Context, spec AuthoritySpec) error
```

Adapter: prefer `runCommandJSON` where qualifier mapping exists; fall back to
`RunMQSC` / `runCommand` for `SET CHLAUTH` and `SET AUTHREC` until JSON coverage
is confirmed per command.

## What we are not copying from IBM samples

| IBM pattern | Kurator approach |
|-------------|------------------|
| `spec.queueManager.mqsc` ConfigMap on `QueueManager` | Per-object CRs + continuous reconcile |
| Dynamic MQSC volume reload (gitops `queue-manager-deployment`) | Operator observes CR spec generation |
| IBM MQ Operator webhook / OLM install | Out of scope — Kurator targets existing mqweb |

## E2e and fixtures

Channel/auth MQSC used to validate mqweb and future client tests lives under
[`test/e2e/fixtures/`](../test/e2e/fixtures/). Queue reconcile e2e does not require
those objects; they document the Phase 4 target state and test `RunMQSC` plumbing.
