# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Release notes are generated from [Conventional Commits](https://www.conventionalcommits.org/)
on the default branch using [git-cliff](https://git-cliff.org/).

## [Unreleased]

### Bug Fixes

- **task:** Propagate KURATOR_E2E_MQ into test:e2e task env [eaa4300](https://github.com/konih/kurator/commit/eaa4300a0a6c89d35e6a11c5214d83745583c55a)

- **e2e:** Race-safe subprocess output and webhook assertion [46e9cde](https://github.com/konih/kurator/commit/46e9cdef52f041bac8236b42b3dff6a94d122c59)

- **task:** Resolve kustomize path with go tool -n [bd4bd49](https://github.com/konih/kurator/commit/bd4bd495c502944a8b25b0b6c315ba01d9f94146)

- **samples:** Let kustomization set namespace on Helm samples [2fa4097](https://github.com/konih/kurator/commit/2fa409725905068b095813830676c3bdee39db7b)

- **ci:** Bump Go 1.26.4 and sync verify artifacts [98116c6](https://github.com/konih/kurator/commit/98116c6fd14ebf8bf3807d4d9ce3c4027fb53b04)

- **ci:** Align CRDs with go tool controller-gen [513094f](https://github.com/konih/kurator/commit/513094ffd71895622bc5b96a12c58a5c5198d56b)

- **makefile:** Use go tool kustomize for deploy targets [cf78511](https://github.com/konih/kurator/commit/cf78511fce8fc8bc6a3eecf0a67a668badf5b961)

- **e2e:** Wait for webhook cert and rollout before MQ tests [0e51d30](https://github.com/konih/kurator/commit/0e51d30334969b3cae9e34dccdd4121e8a554407)

- **config:** Fix webhook kustomize for e2e make deploy [7243b13](https://github.com/konih/kurator/commit/7243b136cd0093b01aa5841ef76b9c06865dcddc)


### Features

- **auth:** Add GetChannelAuth and GetAuthority MQAdmin paths [32720e9](https://github.com/konih/kurator/commit/32720e9bf55462aa3223939918f25fb1a3cd062c)

- **auth:** Add ChannelAuthRule and AuthorityRecord CRDs [13c842e](https://github.com/konih/kurator/commit/13c842e7ab41f7a4968d45c8baefc9fb2239b13e)

## [0.4.0](https://github.com/konih/kurator/compare/v0.3.0..v0.4.0) - 2026-06-02

### Features

- **webhook:** Deny QMC delete when dependents exist [a8fc034](https://github.com/konih/kurator/commit/a8fc034fea91bab5f9cc5401a4abef8801786c61)

## [0.3.0](https://github.com/konih/kurator/compare/v0.2.2..v0.3.0) - 2026-06-02

### Bug Fixes

- **webhook:** Fix unit test race under -race [cbf16da](https://github.com/konih/kurator/commit/cbf16da462b2e7095fe1a35b65ca7c49a6f217cf)


### Features

- **controller:** Expand Kubernetes event emission [5472e56](https://github.com/konih/kurator/commit/5472e561013c310b0097becfbc0a6636ffa87536)


### Refactoring

- [**breaking**] Konih module path, docs hub, admission webhooks [f527ba3](https://github.com/konih/kurator/commit/f527ba30a2af695fa303ac8f88423a13ede8c21d)

## [0.2.2](https://github.com/konih/kurator/compare/v0.2.1..v0.2.2) - 2026-06-02

### Bug Fixes

- **makefile:** Apply CRDs from bases on make install [2f73e84](https://github.com/konih/kurator/commit/2f73e841ed2b78cca354354daf568827e2f50022)

- **test:** Pass QueueSpec to GetQueue in MQ e2e [d56c5f6](https://github.com/konih/kurator/commit/d56c5f6ba8f1f252141c2a2d40dc70a2e366d309)


### Refactoring

- **controller:** Shared reconcile helpers and connection fixes [7a66789](https://github.com/konih/kurator/commit/7a6678996084595e82a790e9b9b67c4634d345f9)

## [0.2.1](https://github.com/konih/kurator/compare/v0.2.0..v0.2.1) - 2026-06-02

### Bug Fixes

- **mqrest:** Normalize alias/remote DISPLAY attribute names from mqweb [aaf47df](https://github.com/konih/kurator/commit/aaf47df932229ce836c4d2530860a8e6a6840172)

## [0.2.0](https://github.com/konih/kurator/compare/v0.1.0..v0.2.0) - 2026-06-02

### Bug Fixes

- **ci:** Clear lint/verify; reconcile alias and remote queues [d48f7bf](https://github.com/konih/kurator/commit/d48f7bf9e8b10a29a8d0bb6dc92680ebfb468737)

## [0.1.0] - 2026-06-02

### Bug Fixes

- **test:** Wait for CRDs after make install in MQ e2e [c199052](https://github.com/konih/kurator/commit/c1990528e96c6d80c32411513f93210444f02e34)

- **test:** Restore cmd declarations in deploy_helpers [4553d9b](https://github.com/konih/kurator/commit/4553d9bb83d055227a8c60dd03d33688bd3ecccf)

- **test:** Serialize e2e suites and idempotent namespace create [8967b4c](https://github.com/konih/kurator/commit/8967b4c9185b574831a0cdb8fda61a25c58af98d)

- **test,ci:** Ordered MQ e2e context; gofmt metrics imports [6111051](https://github.com/konih/kurator/commit/61110510b36f866ff8d9c5dc859af638b2bca63b)

- **test,ci:** MQ e2e redeploys operator; bump otel for Trivy [f2fd0db](https://github.com/konih/kurator/commit/f2fd0db0e08e04c2092fcb4a36813862b85a7796)

- **ci:** Set KIND via GITHUB_ENV in e2e install step [b7f6e3a](https://github.com/konih/kurator/commit/b7f6e3ae03229bef3c9eadb82443a078eb6d2ea7)

- **ci:** E2e PATH and sync deepcopy with controller-gen [bfc0c20](https://github.com/konih/kurator/commit/bfc0c20221156f786a36332c065a6e683eb800b4)

- **ci:** Unblock CI and e2e on GitHub Actions [94ee861](https://github.com/konih/kurator/commit/94ee8611faa2e3be59b7d1dda4e1b78694d0042f)

- **ci:** Pin correct setup-terraform action SHA [5c037ac](https://github.com/konih/kurator/commit/5c037ac20ca3729f975c4e3630c49153e0cc2706)

- **queue:** Defer MQ admin client until connection is Ready [5baf674](https://github.com/konih/kurator/commit/5baf674a171e3b04d9a518d0fd83186863ec5596)

- **mqrest:** Drop maxmsglen from queue DISPLAY on mqweb 9.4 [c4f8a08](https://github.com/konih/kurator/commit/c4f8a083a559b91884f31aa5a19e595b88b98165)

- **logging:** Reuse err var for Setup after Load [1d71167](https://github.com/konih/kurator/commit/1d7116781ce9d3d3685385652efa4fc4e4c1a4eb)


### Features

- **messaging:** Reconcile Topic and Channel CRs via mqweb [3ff3463](https://github.com/konih/kurator/commit/3ff3463df697a19a625025280cefd496f981d761)

- **metrics:** Add Prometheus metrics and Helm alerts [a87d16b](https://github.com/konih/kurator/commit/a87d16b3400c698d5eb33ce8087728c4f871a08c)

- **kind:** Add mq console URL and runmqsc CLI tasks [7cf8a30](https://github.com/konih/kurator/commit/7cf8a304c73cc1425a05d4bfde6c4d632825b37b)

- **chart:** Add Helm chart, reference docs, and MQ e2e fixtures [aca907a](https://github.com/konih/kurator/commit/aca907acc16bb3667e81325a6b49bc4f600fb99d)

- Add Queue and QueueManagerConnection reconcilers [08d7a92](https://github.com/konih/kurator/commit/08d7a9261d7d7449180f0c580d0c0fded37724df)

- **cluster:** Haproxy ingress, Argo CD, upstream IBM MQ [214e048](https://github.com/konih/kurator/commit/214e048e5d274add7124f347ba11ee79fa13a3dd)

- Scaffold Kurator operator (Phase 1) [3083f03](https://github.com/konih/kurator/commit/3083f0339bd999343f6d061f483601a5ee6e690d)

- **logging:** Add configurable slog logger [f251a03](https://github.com/konih/kurator/commit/f251a03a3e025e93dd44ebe5a973d5c3df2890f7)

- Add one-command kind dev cluster [74855c7](https://github.com/konih/kurator/commit/74855c7e633b2ca99e79f244b314a95b3ace029e)

