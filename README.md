# IBM Message Queue Operator

A Kubernetes operator for declaratively managing **resources on an existing
IBM MQ Queue Manager** — queues today, users/authorities and more later.

> Status: **early / work in progress.** The design is set; implementation is
> being built out in phases. See the [roadmap](docs/ROADMAP.md).

## What it does

- Reconciles custom resources (e.g. `Queue`) into MQSC objects on a running
  Queue Manager.
- Talks to the Queue Manager through the **IBM MQ Administrative REST API**
  (`mqweb`) over HTTPS — pure Go, no CGO.
- Reports status via conditions and cleans up via finalizers.

It does **not** deploy or operate Queue Manager installations; the Queue
Manager is assumed to already exist and expose `mqweb`.

## Documentation

- [AGENTS.md](AGENTS.md) — context, conventions, and contributor/agent guide.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — components, CRDs, reconcile flow.
- [docs/ROADMAP.md](docs/ROADMAP.md) — phased delivery plan.

## Planned workflow

Development is driven by [Task](https://taskfile.dev) with a local
[kind](https://kind.sigs.k8s.io/) cluster:

```sh
task kind:up      # create a local cluster
task build        # build the manager (CGO-free, static)
task deploy       # install CRDs + operator
task test:run     # unit + envtest suites (Ginkgo)
task test:e2e     # end-to-end against a Queue Manager container
```

(Task targets land with the Phase 1 scaffold; see the roadmap.)

## License

MIT — see [LICENSE](LICENSE).
