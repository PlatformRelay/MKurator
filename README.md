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

- [AGENTS.md](AGENTS.md) — context, conventions, toolchain, and doc map.
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — components, runtime, CRDs, reconcile flow, security.
- [docs/NON_FUNCTIONAL_REQUIREMENTS.md](docs/NON_FUNCTIONAL_REQUIREMENTS.md) — quality bars.
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) — set up, build, test, run locally.
- [docs/CICD.md](docs/CICD.md) — CI/CD pipeline design.
- [docs/adr/](docs/adr/) — architecture decision records.
- [docs/ROADMAP.md](docs/ROADMAP.md) — phased delivery plan.
- [SECURITY.md](SECURITY.md) — security posture and reporting.

## Planned workflow

Development is driven by [Task](https://taskfile.dev) with a local
[kind](https://kind.sigs.k8s.io/) cluster (see [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md)):

```sh
task cluster:up   # local platform: kind + ingress + cert-manager + IBM MQ
task build        # build the manager (CGO-free, static)
task deploy       # install CRDs + operator
task test:run     # unit + envtest suites (Ginkgo)
task test:e2e     # end-to-end against the local Queue Manager
```

(Task targets land with the Phase 1 scaffold; see the roadmap. The
`hack/kind-cluster` platform works today via its scripts.)

## License

MIT — see [LICENSE](LICENSE).
