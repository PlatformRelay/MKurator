# IBM MQ objects — research reference

> **Not the operator contract.** Shipped behaviour is defined by the v1alpha1
> CRDs, [ATTRIBUTE_RECONCILIATION.md](ATTRIBUTE_RECONCILIATION.md), and
> [INSTALL_AND_USE.md](INSTALL_AND_USE.md). Conceptual resource names below
> (`MqQueue`, `Applied`, …) are design notes — they do not match MKurator CRD or
> condition names.

This document inventories the IBM MQ objects most relevant to a **resource-only** Kubernetes operator: queues, access control, client connectivity, and publish/subscribe. It assumes a **vanilla queue manager** created with `crtmqm` (or equivalent container bootstrap) where system and default template objects already exist, and describes how operators and administrators configure application-facing resources using **MQSC** via `runmqsc`.

**Out of scope for this operator (by design):** queue manager lifecycle (`crtmqm`, `strmqm`), pod/deployment topology, listeners at the infrastructure layer (though listener *objects* are noted), and cluster topology between queue managers.

**Target IBM MQ version:** 9.2+ / 9.3+ (multiplatform). Syntax is MQSC unless noted.

**Primary references:**

- [MQSC commands reference](https://www.ibm.com/docs/en/ibm-mq/9.3.x?topic=reference-mqsc-commands)
- [System and default objects](https://www.ibm.com/docs/en/ibm-mq/9.4.x?topic=reference-system-default-objects)
- [Giving access to IBM MQ objects](https://www.ibm.com/docs/en/ibm-mq/9.2.x?topic=SSFKSJ_9.2.0/com.ibm.mq.sec.doc/q013490_.html)

---

## 1. Administration model

### 1.1 How configuration is applied

| Mechanism | Command / tool | Typical use |
|-----------|----------------|-------------|
| Interactive / batch MQSC | `runmqsc <qmgr>` | Human ops, GitOps scripts, ConfigMap snippets (IBM MQ container chart) |
| Verify only | `runmqsc -v <qmgr>` | CI dry-run before apply |
| Control command (auth) | `setmqaut` | Shell-friendly OAM grants; equivalent to `SET AUTHREC` |
| Programmatic | PCF (`MQCMD_*`) | Operators, automation; maps 1:1 to MQSC semantics |

MQSC commands are processed by the queue manager. Most `DEFINE` commands support `REPLACE` (create or overwrite) and `NOREPLACE` (fail if exists). For GitOps-style reconciliation, **`DEFINE ... REPLACE`** plus **`ALTER`** for attributes that must not be reset on every apply is a common pattern (see §8).

### 1.2 Object discovery commands

| Intent | MQSC |
|--------|------|
| Show object definition | `DISPLAY QLOCAL('MY.QUEUE') ALL` |
| Show specific attributes | `DISPLAY QLOCAL('MY.QUEUE') MAXDEPTH DEFPSIST` |
| List objects by type | `DISPLAY QLOCAL(*)`, `DISPLAY CHANNEL(*)`, `DISPLAY TOPIC(*)` |
| Show authority | `DISPLAY AUTHREC PROFILE('MY.QUEUE') OBJTYPE(QUEUE) PRINCIPAL('alice') AUTHLIST` |
| Show channel auth | `DISPLAY CHLAUTH('MY.SVRCONN')` |
| Runtime channel state | `DISPLAY CHSTATUS('MY.SVRCONN')` |

### 1.3 Deletion

| Object | MQSC |
|--------|------|
| Queue | `DELETE QLOCAL('MY.QUEUE')` |
| Channel | `DELETE CHANNEL('MY.SVRCONN')` |
| Topic | `DELETE TOPIC('MY.TOPIC')` |
| Subscription | `DELETE SUB('MY.SUB')` |
| Auth record | `DELETE AUTHREC PROFILE('MY.QUEUE') OBJTYPE(QUEUE) PRINCIPAL('alice')` |
| CHLAUTH rule | `SET CHLAUTH('MY.SVRCONN') TYPE(ADDRESSMAP) ACTION(REMOVEALL)` or `ACTION(REMOVE)` |

Queues with `CURDEPTH > 0` or open handles may block deletion or attribute changes.

---

## 2. Vanilla queue manager baseline

When a queue manager is created (`crtmqm`), IBM MQ automatically provisions:

1. **System objects** (`SYSTEM.*`) — internal queues, channels, and infrastructure required to run the queue manager. Names are reserved; attributes may be altered but names must not be renamed.
2. **Default template objects** (`SYSTEM.DEFAULT.*`, `SYSTEM.DEF.*`) — attribute templates used when `DEFINE` omits parameters.

### 2.1 Inheritance from defaults

Any omitted attribute on `DEFINE QLOCAL(MY.QUEUE)` is copied from **`SYSTEM.DEFAULT.LOCAL.QUEUE`** at creation time. Changing the default object later does **not** retroactively change existing queues.

```mqsc
* See effective defaults for local queues:
DISPLAY QLOCAL('SYSTEM.DEFAULT.LOCAL.QUEUE') ALL

* Explicit copy from template:
DEFINE QLOCAL('MY.QUEUE') LIKE('SYSTEM.DEFAULT.LOCAL.QUEUE') REPLACE
```

### 2.2 Security defaults (modern queue managers)

New multiplatform queue managers typically ship with:

| Area | Default | Notes |
|------|---------|-------|
| Connection authentication | `SYSTEM.DEFAULT.AUTHINFO.IDPWOS` via `CONNAUTH` | OS user/password checking |
| Channel authentication | `CHLAUTH(ENABLED)` | Default rules include blocking privileged users on client channels |
| Object authority (OAM) | Enabled | Grants via `SET AUTHREC` / `setmqaut` |

Default connection auth object (typical):

```mqsc
DISPLAY AUTHINFO('SYSTEM.DEFAULT.AUTHINFO.IDPWOS') ALL
* AUTHTYPE(IDPWOS) ADOPTCTX(NO) CHCKCLNT(REQDADM) CHCKLOCL(OPTIONAL) ...
```

Default CHLAUTH includes a rule blocking `*MQADMIN` on all channels (`TYPE(BLOCKUSER)`). Client connectivity almost always requires **explicit CHLAUTH rules** in addition to OAM grants.

### 2.3 What “unconfigured” means for this operator

A vanilla cluster-ready queue manager already has listeners (if configured in `crtmqm`/container), system queues, and default templates. **Application resources do not exist** until you define them: application queues, SVRCONN channels, topic objects, subscriptions, and OAM/CHLAUTH for principals.

---

## 3. Object inventory (operator priority)

Objects are ordered by expected CRD / reconciliation priority for a message-queue **resource** operator.

| Priority | MQ object | MQSC object synonym | Primary DEFINE / SET |
|----------|-----------|---------------------|----------------------|
| P0 | Local queue | `QLOCAL`, `QL` | `DEFINE QLOCAL` |
| P0 | Alias queue | `QALIAS`, `QA` | `DEFINE QALIAS` |
| P1 | Remote queue / QM alias | `QREMOTE`, `QR` | `DEFINE QREMOTE` |
| P1 | Model queue | `QMODEL`, `QM` | `DEFINE QMODEL` |
| P0 | Authority (OAM) | — | `SET AUTHREC` |
| P0 | Server-connection channel | `CHANNEL`, `CHL` | `DEFINE CHANNEL` … `CHLTYPE(SVRCONN)` |
| P0 | Channel authentication | `CHLAUTH` | `SET CHLAUTH` |
| P1 | Authentication information | `AUTHINFO` | `DEFINE AUTHINFO` |
| P1 | Administrative topic | `TOPIC`, `TPC` | `DEFINE TOPIC` |
| P2 | Durable subscription | `SUB` | `DEFINE SUB` |
| P2 | Process definition | `PROCESS`, `PROC` | `DEFINE PROCESS` |
| P3 | Listener | `LISTENER`, `LSTR` | `DEFINE LISTENER` |
| P3 | Namelist | `NAMELIST`, `NL` | `DEFINE NAMELIST` |
| — | Queue manager | `QMGR` | `ALTER QMGR` (limited) |

Message channels (`CHLTYPE(SDR|RCVR|SVR|RQSTR|CLUS*)`) are included for completeness but are usually owned by **inter-QM networking** teams rather than application teams.

---

## 4. Queues

Queues are the core application-facing objects. Four definition types exist on a queue manager; only **QLOCAL** holds messages physically on that manager.

### 4.1 Local queue (`QLOCAL`)

**Purpose:** Store messages on this queue manager. Applications use `MQPUT` / `MQGET` (or bindings equivalent).

**Common attributes (operator-relevant):**

| Attribute | Values (examples) | Meaning |
|-----------|-------------------|---------|
| `DESCR` | string | Documentation / ownership |
| `USAGE` | `NORMAL`, `XMITQ` | Application queue vs transmission queue |
| `MAXDEPTH` | integer | Max messages on queue |
| `MAXMSGL` | integer | Max message length (bytes) |
| `DEFPSIST` | `YES`, `NO` | Default persistence for messages without explicit persist flag |
| `GET` / `PUT` | `ENABLED`, `DISABLED` | Allow receive / send |
| `SHARE` | `YES`, `NO` | Multiple applications can open concurrently |
| `DEFSOPT` | `SHARED`, `EXCL` | Default open option for consumers |
| `BOTHRESH` / `BOQNAME` | int, queue name | Backout threshold and backout queue |
| `TRIGGER` / `TRIGTYPE` / `PROCESS` / `INITQ` | various | Trigger-based processing |
| `CLUSTER` / `CLUSNL` | name | Cluster membership (if using MQ clusters) |
| `ACCTQ` / `MONQ` / `STATQ` | on/off levels | Accounting and monitoring |

**Define (minimal):**

```mqsc
DEFINE QLOCAL('ORDERS.IN') REPLACE
```

**Define (production-style):**

```mqsc
DEFINE QLOCAL('ORDERS.IN') REPLACE +
  DESCR('Inbound orders - OMS') +
  USAGE(NORMAL) +
  MAXDEPTH(500000) +
  MAXMSGL(4194304) +
  DEFPSIST(YES) +
  GET(ENABLED) PUT(ENABLED) +
  SHARE +
  DEFSOPT(SHARED) +
  BOTHRESH(5) BOQNAME('ORDERS.BO')
```

**Alter (non-destructive tuning):**

```mqsc
ALTER QLOCAL('ORDERS.IN') MAXDEPTH(1000000)
```

**Verify:**

```mqsc
DISPLAY QLOCAL('ORDERS.IN') MAXDEPTH DEFPSIST GET PUT CURDEPTH
```

**Notes for operators:**

- `DEFINE ... REPLACE` retains messages but cannot run if the queue is open; plan rolling restarts.
- `CURDEPTH` is read-only runtime state — useful for status, not spec.
- `USAGE(XMITQ)` defines a transmission queue (routing), not typical application input.

---

### 4.2 Alias queue (`QALIAS`)

**Purpose:** Indirection — an alternative name resolving to a local, remote, or cluster queue (not another alias).

| Attribute | Meaning |
|-----------|---------|
| `TARGQ` / `TARGET` | Target queue or topic object name |
| `TARGTYPE` | `QUEUE` (default) or `TOPIC` |
| `DESCR` | Description |
| `CLUSTER` | Advertise alias in cluster |

**Define:**

```mqsc
DEFINE QALIAS('ORDERS') REPLACE +
  TARGQ('ORDERS.IN') +
  DESCR('Alias for order ingress')
```

**Verify:**

```mqsc
DISPLAY QALIAS('ORDERS') ALL
```

---

### 4.3 Remote queue (`QREMOTE`)

**Purpose:** Local definition pointing to a queue on another queue manager. Messages are routed via a **transmission queue** (`XMITQ`). Also used for **queue manager aliases** (`RNAME` blank).

| Attribute | Meaning |
|-----------|---------|
| `RNAME` | Queue name on remote QM (blank = QM alias) |
| `RQMNAME` | Remote queue manager name |
| `XMITQ` | Local transmission queue for routing |
| `DEFBIND` | Binding when queue is opened |
| `CLUSTER` | Cluster advertisement |

**Remote queue definition:**

```mqsc
DEFINE QREMOTE('PARTNER.ORDERS') REPLACE +
  RNAME('ORDERS.IN') +
  RQMNAME('PARTNER.QM') +
  XMITQ('SYSTEM.DEFAULT.XMIT.QUEUE') +
  DESCR('Orders on partner system')
```

**Queue manager alias (cluster routing):**

```mqsc
DEFINE QREMOTE('PARTNER.QM') REPLACE +
  RNAME(' ') +
  RQMNAME('PARTNER.QM') +
  XMITQ('SYSTEM.DEFAULT.XMIT.QUEUE')
```

---

### 4.4 Model queue (`QMODEL`)

**Purpose:** Template for **dynamic queues** created by applications via `MQOPEN` with model name. Not a physical queue.

| Attribute | Meaning |
|-----------|---------|
| Same as `QLOCAL` where applicable | Defaults for dynamic queue creation |
| `DEFTYPE` | `TEMPORARY` or `PERMANENT` dynamic queue |

**Define:**

```mqsc
DEFINE QMODEL('ORDERS.DYNAMIC') REPLACE +
  MAXDEPTH(10000) +
  DEFPSIST(NO) +
  SHARE +
  DESCR('Template for app-created temp queues')
```

---

## 5. Authorization (users and groups)

IBM MQ does not store “MQ users” as first-class objects like queues. **Principals** (OS users, LDAP users, or mapped `MCAUSER`) receive **authority records** on **profiles** (object names or generics).

### 5.1 `SET AUTHREC` (OAM)

**Command:**

```mqsc
SET AUTHREC +
  PROFILE('ORDERS.IN') +
  OBJTYPE(QUEUE) +
  PRINCIPAL('orders-app') +
  AUTHADD(GET,PUT,INQ,DSP)
```

**`OBJTYPE` values:** `QUEUE`, `QMGR`, `CHANNEL`, `TOPIC`, `PROCESS`, `NAMELIST`, `AUTHINFO`, `LISTENER`, `SERVICE`, `CLNTCONN`, `COMMINFO`, `RQMNAME`

**Common `AUTHADD` / `AUTHRMV` authorities:**

| Authority | Typical need |
|-----------|--------------|
| `CONNECT` | Attach to queue manager |
| `GET`, `PUT`, `BROWSE` | Queue messaging |
| `INQ`, `DSP` | Display attributes |
| `CRT` | Create dynamic queues (profile often `SELF` or model name) |
| `SET`, `SETALL`, `SETID` | Alter objects |
| `CTRL` | Start/stop channels, listeners (privileged) |
| `SUB`, `PUB`, `RESUME` | Pub/sub |
| `ALL`, `ALLADM`, `ALLMQI` | Bundles (see overlap rules below) |
| `PASSALL`, `PASSID` | Message identity / context passing |

**Queue manager connect (minimal app user):**

```mqsc
SET AUTHREC OBJTYPE(QMGR) PRINCIPAL('orders-app') AUTHADD(CONNECT,INQ,DSP)
SET AUTHREC PROFILE('ORDERS.IN') OBJTYPE(QUEUE) PRINCIPAL('orders-app') AUTHADD(GET,PUT,INQ,DSP)
```

**Equivalent `setmqaut`:**

```sh
setmqaut -m QM1 -t qmgr -p orders-app +connect +inq +dsp
setmqaut -m QM1 -n ORDERS.IN -t queue -p orders-app +get +put +inq +dsp
```

**Display:**

```mqsc
DISPLAY AUTHREC PROFILE('ORDERS.IN') OBJTYPE(QUEUE) PRINCIPAL('orders-app') AUTHLIST
```

**Reset and replace pattern:**

```mqsc
SET AUTHREC PROFILE('ORDERS.IN') OBJTYPE(QUEUE) PRINCIPAL('orders-app') AUTHRMV(ALL) AUTHADD(GET,PUT)
```

`AUTHADD(ALL)` / `AUTHRMV(ALL)` are processed first regardless of order on the command line; `AUTHADD` and `AUTHRMV` must not overlap (e.g. cannot `AUTHADD(DSP)` and `AUTHRMV(ALLADM)` in one command).

**Group principals:**

```mqsc
SET AUTHREC PROFILE('ORDERS.IN') OBJTYPE(QUEUE) GROUP('orders-admins') AUTHADD(GET,PUT,BROWSE,INQ,DSP)
```

### 5.2 Connection authentication (`AUTHINFO` + `CONNAUTH`)

Controls **whether** a user/password (or certificate) is accepted at connect time — separate from OAM object permissions.

**OS authentication (default):**

```mqsc
DEFINE AUTHINFO('QM1.IDPW.OS') AUTHTYPE(IDPWOS) REPLACE +
  ADOPTCTX(YES) +
  CHCKCLNT(REQUIRED) +
  CHCKLOCL(OPTIONAL)

ALTER QMGR CONNAUTH('QM1.IDPW.OS')
REFRESH SECURITY TYPE(CONNAUTH)
```

**LDAP authentication:**

```mqsc
DEFINE AUTHINFO('QM1.IDPW.LDAP') AUTHTYPE(IDPWLDAP) REPLACE +
  CONNAME('ldap.example.com(636)') +
  SECCOMM(YES) +
  LDAPUSER('cn=bind,dc=example,dc=com') +
  LDAPPWD('secret') +
  SHORTUSR('uid') USRFIELD('uid') +
  BASEDNU('ou=users,dc=example,dc=com') +
  AUTHORMD(SEARCHGRP) BASEDNG('ou=groups,dc=example,dc=com') +
  CLASSGRP('groupOfNames') GRPFIELD('cn') FINDGRP('member') +
  ADOPTCTX(YES) CHCKCLNT(REQUIRED)

ALTER QMGR CONNAUTH('QM1.IDPW.LDAP')
REFRESH SECURITY TYPE(CONNAUTH)
```

**`AUTHTYPE` values:** `IDPWOS`, `IDPWLDAP`, `CRLLDAP`, `OCSP`

For a resource operator, `AUTHINFO` is usually **platform-owned** (one per queue manager), while **`SET AUTHREC`** per application principal is the main reconciliation surface.

---

## 6. Channels

### 6.1 Server-connection channel (`CHLTYPE(SVRCONN)`)

**Purpose:** Inbound client attachment point. Clients supply QM name, channel name, host/port (or CCDT), and credentials; server binds to this definition.

**Key attributes:**

| Attribute | Meaning |
|-----------|---------|
| `TRPTYPE` | `TCP` (typical) |
| `MCAUSER` | Default user ID for authority checks if not mapped |
| `SHARECNV` | Conversations per connection |
| `MAXINST` / `MAXINSTC` | Connection limits |
| `MAXMSGL` | Max message size on channel |
| `SSLCIPH` / `SSLCAUTH` | TLS configuration |
| `DESCR` | Description |

**Not valid for SVRCONN:** `CONNAME` (clients supply connection address).

**Define:**

```mqsc
DEFINE CHANNEL('ORDERS.APP') CHLTYPE(SVRCONN) TRPTYPE(TCP) REPLACE +
  MCAUSER('orders-app') +
  SHARECNV(10) +
  MAXMSGL(4194304) +
  DESCR('Client channel for orders service')
```

**Verify:**

```mqsc
DISPLAY CHANNEL('ORDERS.APP') CHLTYPE(SVRCONN)
DISPLAY CHSTATUS('ORDERS.APP') WHERE(CHLTYPE EQ SVRCONN)
```

### 6.2 Message channels (reference)

Used between queue managers for distributed queuing / clustering.

| CHLTYPE | Role |
|---------|------|
| `SDR` | Sender — requires `CONNAME`, `XMITQ` |
| `RCVR` | Receiver |
| `SVR` | Server |
| `RQSTR` | Requester |
| `CLUSRCVR` / `CLUSSDR` | Cluster channels |

Example sender:

```mqsc
DEFINE CHANNEL('QM1.TO.QM2') CHLTYPE(SDR) TRPTYPE(TCP) REPLACE +
  CONNAME('qm2.example.com(1414)') +
  XMITQ('SYSTEM.DEFAULT.XMIT.QUEUE')
```

### 6.3 Channel authentication (`SET CHLAUTH`)

**Purpose:** Control **who may connect** on which channel and map to `MCAUSER`. Enabled by default on new queue managers.

**Common rule types:**

| TYPE | Purpose |
|------|---------|
| `BLOCKUSER` | Block user IDs (`USERLIST`) — default blocks `*MQADMIN` |
| `ADDRESSMAP` | Map client IP to `MCAUSER`, `NOACCESS`, or `CHANNEL` |
| `USERMAP` | Map asserted client user (`CLNTUSER`) |
| `SSLPEERMAP` | Map TLS DN |
| `QMGRMAP` | Map remote QM name |
| `BLOCKADDR` | Block IPs at listener (pre-channel) |

**Typical secure client pattern** (deny-all, then permit channel):

```mqsc
SET CHLAUTH('*') TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(NOACCESS) ACTION(REPLACE)
SET CHLAUTH('ORDERS.APP') TYPE(BLOCKUSER) USERLIST('nobody') ACTION(REPLACE)
SET CHLAUTH('ORDERS.APP') TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(CHANNEL) CHCKCLNT(REQUIRED) ACTION(REPLACE)
```

**Map IP to service account:**

```mqsc
SET CHLAUTH('ORDERS.APP') TYPE(ADDRESSMAP) +
  ADDRESS('10.0.0.0-10.0.255.255') +
  USERSRC(MAP) MCAUSER('orders-app') +
  ACTION(ADD)
```

**Display:**

```mqsc
DISPLAY CHLAUTH('ORDERS.APP')
```

CHLAUTH is evaluated in order; explicit match testing uses `MATCH(RUN)` style inquiry (see IBM doc [Channel authentication records](https://www.ibm.com/docs/en/ibm-mq/9.3.x?topic=mechanisms-channel-authentication-records)).

---

## 7. Publish/subscribe

### 7.1 Topic object (`DEFINE TOPIC`)

**Purpose:** Administrative node in the **topic tree** — sets defaults for `PUB`, `SUB`, scope, durability, and clustering for topic strings. Does **not** store messages.

| Attribute | Meaning |
|-----------|---------|
| `TOPSTR` | Topic string (e.g. `'retail/orders/created'`) |
| `TOPTYPE` | `TOPIC` or `ALIAS` |
| `PUB` / `SUB` | `ENABLED`, `DISABLED`, `ASPARENT` |
| `DEFPSIST` | Default persistence for publications |
| `PUBSCOPE` / `SUBSCOPE` | Hierarchy vs cluster propagation |
| `CLUSTER` | Cluster topic |

**Define:**

```mqsc
DEFINE TOPIC('RETAIL.ORDERS') REPLACE +
  TOPSTR('retail/orders') +
  DESCR('Retail order events') +
  PUB(ENABLED) SUB(ENABLED)
```

Applications may publish to a **topic string** without a topic object; objects add policy and stable names (`TOPICOBJ` in APIs).

**Base of tree:** `SYSTEM.BASE.TOPIC` — attributes apply when no administrative parent exists.

### 7.2 Subscription (`DEFINE SUB`)

**Purpose:** Administratively created **durable subscription** — copies matching publications to a **destination queue**.

| Attribute | Meaning |
|-----------|---------|
| `TOPICSTR` | Filter / topic string (wildcards `+`, `#` in app subscriptions) |
| `TOPICOBJ` | Reference to topic object |
| `DEST` | Destination queue name |
| `DESTCLAS` | `PROVIDED` or `MANAGED` (QM creates temp queue) |
| `DESTQMGR` | Remote QM if routed |
| `SUBSCOPE` | `QMGR` or `ALL` |
| `SELECTOR` | Message selector |
| `SUBUSER` | Owning user |

**Define:**

```mqsc
DEFINE SUB('ORDERS.CREATED.SUB') REPLACE +
  TOPICSTR('retail/orders/created') +
  DEST('ORDERS.EVENTS') +
  DESTCLAS(PROVIDED) +
  SUBSCOPE(QMGR)
```

**OAM for pub/sub:**

```mqsc
SET AUTHREC PROFILE('retail/orders/created') OBJTYPE(TOPIC) PRINCIPAL('orders-app') AUTHADD(SUB,DSP)
SET AUTHREC PROFILE('ORDERS.EVENTS') OBJTYPE(QUEUE) PRINCIPAL('orders-app') AUTHADD(GET,PUT)
```

---

## 8. Supporting objects

### 8.1 Process definition (`DEFINE PROCESS`)

Used with **triggering** on queues (`TRIGGER`, `PROCESS('name')`).

```mqsc
DEFINE PROCESS('ORDERS.TRIGGER') REPLACE +
  APPLICID('orders-worker') +
  ENVDATA('/opt/orders/run.sh') +
  USERDATA('start')
```

### 8.2 Listener (`DEFINE LISTENER`)

Accepts inbound TCP (or other) connections. Often provisioned by Helm/platform; still an MQ object.

```mqsc
DEFINE LISTENER('TCP.1414') TRPTYPE(TCP) PORT(1414) CONTROL(QMGR) REPLACE
START LISTENER('TCP.1414')
```

### 8.3 Namelist (`DEFINE NAMELIST`)

Named list of strings — used by channels (`CLUSNL`), clustering, or applications.

```mqsc
DEFINE NAMELIST('CLUSTER.NAMES') REPLACE +
  NAMES('QM1','QM2')
```

---

## 9. End-to-end bootstrap (vanilla → application-ready)

Minimal sequence for a client application connecting to put/get on a local queue. Adjust names for your environment.

```mqsc
* 1. Application queue
DEFINE QLOCAL('APP.IN') REPLACE +
  MAXDEPTH(100000) MAXMSGL(4194304) DEFPSIST(YES) +
  GET(ENABLED) PUT(ENABLED) SHARE +
  DESCR('Application input queue')

* 2. Client channel
DEFINE CHANNEL('APP.SVRCONN') CHLTYPE(SVRCONN) TRPTYPE(TCP) REPLACE +
  SHARECNV(10) MAXMSGL(4194304)

* 3. Channel authentication (allow this channel; require client auth)
SET CHLAUTH('*') TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(NOACCESS) ACTION(REPLACE)
SET CHLAUTH('APP.SVRCONN') TYPE(BLOCKUSER) USERLIST('nobody') ACTION(REPLACE)
SET CHLAUTH('APP.SVRCONN') TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(CHANNEL) CHCKCLNT(REQUIRED) ACTION(REPLACE)

* 4. Connection auth (if not already using defaults)
ALTER AUTHINFO(SYSTEM.DEFAULT.AUTHINFO.IDPWOS) AUTHTYPE(IDPWOS) ADOPTCTX(YES)
ALTER QMGR CONNAUTH(SYSTEM.DEFAULT.AUTHINFO.IDPWOS)
REFRESH SECURITY TYPE(CONNAUTH)

* 5. OAM for application principal (OS or LDAP user)
SET AUTHREC OBJTYPE(QMGR) PRINCIPAL('app-user') AUTHADD(CONNECT,INQ,DSP)
SET AUTHREC PROFILE('APP.IN') OBJTYPE(QUEUE) PRINCIPAL('app-user') AUTHADD(ALLMQI)

* 6. Verify
DISPLAY QLOCAL('APP.IN') GET PUT MAXDEPTH
DISPLAY CHANNEL('APP.SVRCONN') CHLTYPE MCAUSER
DISPLAY AUTHREC PROFILE('APP.IN') OBJTYPE(QUEUE) PRINCIPAL('app-user') AUTHLIST
```

**Pub/sub variant:** add `DEFINE TOPIC`, `DEFINE SUB`, and topic/queue `SET AUTHREC` with `PUB` / `SUB` / `GET` as required.

---

## 10. Implications for the Kubernetes operator

### 10.1 Natural CRD boundaries

Conceptual mapping for future design (see disclaimer at top — shipped types are
`Queue`, `Topic`, `Channel`, `QueueManagerConnection`):

| Kubernetes resource (conceptual) | MQ objects | MQSC surface |
|----------------------------------|------------|--------------|
| `MqQueue` | `QLOCAL`, optionally `QALIAS` | `DEFINE QLOCAL`, `ALTER`, `DELETE` |
| `MqRemoteQueue` | `QREMOTE` | `DEFINE QREMOTE` |
| `MqChannel` | `CHANNEL` (SVRCONN) | `DEFINE CHANNEL`, `SET CHLAUTH` |
| `MqPrincipalAccess` | OAM records | `SET AUTHREC` |
| `MqTopic` | `TOPIC` | `DEFINE TOPIC` |
| `MqSubscription` | `SUB` | `DEFINE SUB` |

Keep **`AUTHINFO` / `ALTER QMGR` / listeners** as platform concerns unless explicitly in scope later.

### 10.2 Reconciliation semantics

- Prefer **`DEFINE ... REPLACE`** for desired-state apply; use **`ALTER`** when certain attributes must not be reset (e.g. `GET`/`PUT` enabled only once in prod).
- Use **`DISPLAY`** after apply; map to status conditions (`Ready`, `Applied`, `QueueFull`).
- Handle MQ errors: object open (`MQRC_OBJECT_IN_USE`), unknown object (`MQRC_UNKNOWN_OBJECT_NAME`), authority failures (`MQRC_NOT_AUTHORIZED`).
- **Idempotency:** `SET AUTHREC` with `AUTHRMV(ALL)` then `AUTHADD(...)` matches pure desired state; document overlap rules.
- **Execution:** PCF over client connection to remote QM is the usual programmatic API; MQSC strings are fine for MVP if exec’d via `runmqsc` sidecar (less ideal for production).

### 10.3 Status vs spec

| Spec (desired) | Status (observed) |
|----------------|-------------------|
| `maxDepth`, `defPersistence`, `putEnabled` | `DISPLAY QLOCAL` attributes |
| `principal`, `authorities` | `DISPLAY AUTHREC` |
| `channelType`, `maxInstances` | `DISPLAY CHANNEL` |
| — | `CURDEPTH`, `CHSTATUS`, `QMSTATUS` for health only |

### 10.4 Testing focus

- Contract tests: MQSC generation from CRD spec (golden files).
- Integration: vanilla QM in container (IBM MQ image), apply CR, assert `DISPLAY` output.
- Security: negative tests for missing `AUTHREC` / blocked `CHLAUTH`.
- Concurrency: `REPLACE` while queue open should surface clear errors.

---

## 11. Quick reference — command cheat sheet

| Action | Command |
|--------|---------|
| Create local queue | `DEFINE QLOCAL('name') REPLACE ...` |
| Create alias | `DEFINE QALIAS('alias') TARGQ('target') REPLACE` |
| Create remote | `DEFINE QREMOTE('name') RNAME('r') RQMNAME('rqm') XMITQ('xmit') REPLACE` |
| Grant queue access | `SET AUTHREC PROFILE('q') OBJTYPE(QUEUE) PRINCIPAL('u') AUTHADD(GET,PUT)` |
| Client channel | `DEFINE CHANNEL('ch') CHLTYPE(SVRCONN) TRPTYPE(TCP) REPLACE` |
| Allow channel | `SET CHLAUTH('ch') TYPE(ADDRESSMAP) ADDRESS('*') USERSRC(CHANNEL) ACTION(REPLACE)` |
| Topic | `DEFINE TOPIC('t') TOPSTR('a/b') REPLACE` |
| Subscription | `DEFINE SUB('s') TOPICSTR('a/b') DEST('q') REPLACE` |
| Show all local queues | `DISPLAY QLOCAL(*)` |
| Delete queue | `DELETE QLOCAL('name')` |

---

## 12. Glossary

| Term | Meaning |
|------|---------|
| **MQSC** | MQ script command language administered via `runmqsc` |
| **OAM** | Object Authority Manager — `SET AUTHREC` |
| **CONNAUTH** | Connection authentication — `AUTHINFO` + `ALTER QMGR CONNAUTH` |
| **CHLAUTH** | Channel authentication records |
| **MCAUSER** | Channel connection authority ID |
| **Principal** | User or group name in OAM |
| **Profile** | Object name or generic pattern in OAM |
| **PCF** | Programmable Command Format — binary admin API |

---

*Document version: initial research (2026-06-02). Validate attribute defaults against your queue manager level before implementing CRD OpenAPI schemas.*
