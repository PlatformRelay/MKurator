# IBM MQ 101 — local kind cluster and MKurator

A short guide for checking that IBM MQ and the MKurator operator are working on the
[`hack/kind-cluster`](https://github.com/conduit-ops/MKurator/blob/main/hack/kind-cluster/README.md) platform. For MQSC object
details see [IBM_MQ_OBJECTS.md](IBM_MQ_OBJECTS.md); for mqweb see
[IBM_MQ_REST_API.md](IBM_MQ_REST_API.md).

## Does IBM MQ include a UI?

Yes. The kind stack installs IBM MQ with **`web.enable: true`**, which starts
**mqweb** and the **IBM MQ Console** (browser UI) on the queue manager pod.

| Access | How |
|--------|-----|
| **Web console** | https://mq.localhost:30443/ibmmq/console/ — user `admin`, password `passw0rd` (local default) |
| **MQSC CLI** | `runmqsc` inside the MQ pod — `task mq:cli` (see below) |
| **Admin REST** | https://mq.localhost:30443/ibmmq/rest/v3/... — same credentials; used by MKurator |

Run `task mq:console` to print the console URL and login hints.

You also get **Grafana** and **Argo CD** on the same ingress; those are cluster
tooling, not part of IBM MQ.

## Concepts (30 seconds)

| Term | Meaning |
|------|---------|
| **Queue manager (QM)** | The MQ “server” process. Local name: **`QM1`**. |
| **Queue** | Named message destination, e.g. **`APP.ORDERS`**. |
| **MQSC** | Text commands (`DEFINE QLOCAL`, `DISPLAY QLOCAL`, …) applied with `runmqsc`. |
| **MKurator `Queue` CR** | Desired queue on an existing QM; reconciler runs MQSC via **mqweb**. |
| **`QueueManagerConnection`** | How MKurator reaches mqweb (URL + Secret). |

MKurator does **not** install the queue manager — only objects **on** `QM1`.

## One-shot local stack

From the repository root:

```sh
task local:up
```

This runs `cluster:up` (kind + IBM MQ + ingress), installs the operator, applies
sample CRs (`APP.ORDERS` queue), and prints status.

## Verify MKurator is working

### 1. Kubernetes — CR status

```sh
task local:info
# or:
kubectl get qmc,queue -n mkurator-system
kubectl describe queue orders -n mkurator-system
```

Look for **`Synced=True`** on the `Queue` and **`Ready=True`** on the
`QueueManagerConnection`.

### 2. MQSC — queue exists on the QM

```sh
task mq:runmqsc -- "DISPLAY QLOCAL('APP.ORDERS') MAXDEPTH DESCR"
```

You should see `MAXDEPTH(5000)` and `DESCR(Orders intake queue)` matching
[`charts/mkurator/samples/resources/queue.yaml`](https://github.com/conduit-ops/MKurator/blob/main/charts/mkurator/samples/resources/queue.yaml).

### 3. Web console

1. Open https://mq.localhost:30443/ibmmq/console/
2. Log in as `admin` / `passw0rd`
3. Browse to queue manager **QM1** → **Queues** → find **APP.ORDERS**

### 4. Drift test (optional)

Edit the `Queue` CR (e.g. change `maxdepth` to `6000`), apply, wait for
`Synced=True`, then:

```sh
task mq:runmqsc -- "DISPLAY QLOCAL('APP.ORDERS') MAXDEPTH"
```

### 5. Delete test (optional)

```sh
kubectl delete queue orders -n mkurator-system --wait
task mq:runmqsc -- "DISPLAY QLOCAL('APP.ORDERS')"
```

Expect “not found” (or completion code indicating the object is gone). MKurator
uses a finalizer to delete the queue on the QM before removing the CR.

## Useful tasks

| Task | Purpose |
|------|---------|
| `task cluster:up` | kind + IBM MQ only |
| `task cluster:info` | URLs (console, Grafana, Argo CD) |
| `task mq:console` | IBM MQ console URL + credentials |
| `task mq:cli` | Interactive `runmqsc` session on `QM1` |
| `task mq:runmqsc -- "<mqsc>"` | One-shot MQSC command |
| `task local:up` | Cluster + operator + samples |
| `task local:info` | Cluster URLs + MKurator CR status |

## Troubleshooting

| Symptom | Check |
|---------|--------|
| Console TLS warning | Run `mkcert -install` once, then `task cluster:tls` |
| `mq:cli` — no pod | `kubectl get pods -n ibm-mq`; wait for MQ ready after `cluster:up` |
| Queue `Synced=False` | Operator logs: `kubectl logs -n mkurator-system -l control-plane=controller-manager` |
| mqweb 401 | Secret `mq-credentials` in `mkurator-system`; keys `username` + `mqAdminPassword` |
| Wrong console path | Use `/ibmmq/console/` (9.4 mqweb), not legacy `/ibm/mq/console/` |

## Next steps

- Change samples: edit files under `charts/mkurator/samples/resources/`
- E2E against live MQ: `KURATOR_E2E_MQ=1 task test:e2e` (see [DEVELOPMENT.md](DEVELOPMENT.md))
- Deeper MQSC: [IBM_MQ_OBJECTS.md](IBM_MQ_OBJECTS.md)
