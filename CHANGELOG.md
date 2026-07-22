# Changelog

All notable changes to this project are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).
Release notes are generated from [Conventional Commits](https://www.conventionalcommits.org/)
on the default branch using [git-cliff](https://git-cliff.org/).

## [0.13.0](https://github.com/platformrelay/MKurator/compare/v0.12.2..v0.13.0) - 2026-07-22

### Bug Fixes

- **helm:** Drop duplicate app.kubernetes.io/name from labels helper [d75dc3d](https://github.com/platformrelay/MKurator/commit/d75dc3d28b3a713cb4a2a54a032574e555f4c9ee)

- **release:** Pass workflow inputs via env: to neutralize script injection (CI-9) ([#120](https://github.com/platformrelay/MKurator/pull/120)) [32aed5a](https://github.com/platformrelay/MKurator/commit/32aed5a816be0bf0eb4fce78425e3cb3e7ee817a)

- **ci:** Upload Codecov coverage via GitHub OIDC ([#118](https://github.com/platformrelay/MKurator/pull/118)) [34b9102](https://github.com/platformrelay/MKurator/commit/34b9102fb6604b32c5e2f6938e6d19ed950ec314)

- **ci:** Stop linguist classifying Brewfile as Ruby ([#107](https://github.com/platformrelay/MKurator/pull/107)) [b00f7af](https://github.com/platformrelay/MKurator/commit/b00f7af1256d84a480175eaee57160092bd103bd)

- **webhook:** Dedup QMC delete-deny dependents listed under both api versions (CI-4) (#104) ([#104](https://github.com/platformrelay/MKurator/pull/104)) [1b53957](https://github.com/platformrelay/MKurator/commit/1b53957e95ab9eb2c9aebb37aa66dd39abc5205e)


### Features

- **branding:** Kollect-style horizontal social cards ([#121](https://github.com/platformrelay/MKurator/pull/121)) [e63a904](https://github.com/platformrelay/MKurator/commit/e63a904446b8d29995d59d34d0fcde0914e6aa05)

- **docs:** Add dark-theme logo and social branding assets ([#119](https://github.com/platformrelay/MKurator/pull/119)) [9dc2b93](https://github.com/platformrelay/MKurator/commit/9dc2b93a06a52b316c40acd9377573815797cc5d)

- **mqrest:** WARN when insecureSkipVerify disables TLS verification [17be989](https://github.com/platformrelay/MKurator/commit/17be9895ccd201cc716753aa9da7d32831172c1e)

- **branding:** Add favicon and social preview assets ([#112](https://github.com/platformrelay/MKurator/pull/112)) [cde281d](https://github.com/platformrelay/MKurator/commit/cde281d5fc2249993b1c33289fc360a213289ee3)

- **branding:** Add repo logo (#110) ([#110](https://github.com/platformrelay/MKurator/pull/110)) [cbc4c80](https://github.com/platformrelay/MKurator/commit/cbc4c806e92c190a2e310deade516462be533d48)

## [0.12.0](https://github.com/platformrelay/MKurator/compare/v0.11.1..v0.12.0) - 2026-07-08

### Bug Fixes

- **validation:** Satisfy lint for v1beta1 admission [94b49b6](https://github.com/platformrelay/MKurator/commit/94b49b680ca2024a8ca03a20113ef26caaa66969)


### Features

- **webhook:** Add v1beta1 validating admission [1c495bb](https://github.com/platformrelay/MKurator/commit/1c495bbe9ab06ecc6ceb6e07a5f796663d59bf5b)

- **samples:** Default sample CRs to v1beta1 (8d-4) [3065354](https://github.com/platformrelay/MKurator/commit/30653541eb30a1a854166fd16a89be07916fb6a4)

- **webhook:** Conversion webhook for v1beta1 hub (8d-2) [7ca807f](https://github.com/platformrelay/MKurator/commit/7ca807f0d2dce59b00c8fe171de898cc3b8b05e9)

- **api:** Scaffold v1beta1 types and multi-version CRDs (8d-1) (#71) [575c597](https://github.com/platformrelay/MKurator/commit/575c597b165231041d9303be406289149f77c77e)


### Refactoring

- **webhook:** Split v1beta1 envtest suite [4de93de](https://github.com/platformrelay/MKurator/commit/4de93de018ca840ba20c35c0b664c9048e239f31)

## [0.11.1](https://github.com/platformrelay/MKurator/compare/v0.11.0..v0.11.1) - 2026-06-18

### Features

- **mqrest:** Probe all QLOCAL define-only candidates (#69) [9bf02b8](https://github.com/platformrelay/MKurator/commit/9bf02b8dd94249f9f5adff77d82224dbad4c975a)

## [0.11.0](https://github.com/platformrelay/MKurator/compare/v0.10.0..v0.11.0) - 2026-06-18

### Features

- **auth:** AUTHREC channel/namelist profile parity (AUTH-9) [1d66960](https://github.com/platformrelay/MKurator/commit/1d66960f11f016a44392a515542cff329ab54765)

- **channel:** Add RCVR receiver channel type (AUTH-8b) [4805573](https://github.com/platformrelay/MKurator/commit/4805573f563341e33c8273808ab41094a6dc49d0)

- **channel:** Add SDR sender channel type (AUTH-8a) (#63) [9ff4289](https://github.com/platformrelay/MKurator/commit/9ff428963048de044da76b3535bdf1aa5017eb5c)

- **mqrest:** Wire share DISPLAY probe into queue drift (Day 27 MQ-3) (#62) [8cb7d51](https://github.com/platformrelay/MKurator/commit/8cb7d5162145a86184bfad8c81b3f17f1a7546cb)

## [0.10.0](https://github.com/platformrelay/MKurator/compare/v0.9.6..v0.10.0) - 2026-06-18

### Features

- **mqrest:** DISPLAY capability probe spike (Day 22 MQ-1) (#58) [f89e579](https://github.com/platformrelay/MKurator/commit/f89e5797560036fcc23433ca77cfe9fb36ed4889)

## [0.9.6](https://github.com/platformrelay/MKurator/compare/v0.9.5..v0.9.6) - 2026-06-17

### Bug Fixes

- **e2e:** Serialize suite teardown on process 1 (#57) [7ae8c9e](https://github.com/platformrelay/MKurator/commit/7ae8c9eb33190d5c1f3f07ea5de8af721e682ca0)

- **task:** Prefer valid kind kubeconfig path [66cf2db](https://github.com/platformrelay/MKurator/commit/66cf2dbbe1457acd11be0bcc671b32178f765ec9)


### Refactoring

- **mqrest:** Table-drive CHLAUTH auth unit tests [69c63de](https://github.com/platformrelay/MKurator/commit/69c63de07c9f1408cfc4258e95273d8a38bda7ee)

## [0.9.5](https://github.com/platformrelay/MKurator/compare/v0.9.4..v0.9.5) - 2026-06-14

### Bug Fixes

- **e2e:** Use USERMAP mcaUser drift for observe-only spec [1056bc9](https://github.com/platformrelay/MKurator/commit/1056bc95af8b67a42d51eed8866f1793c58bab69)

## [0.9.4](https://github.com/platformrelay/MKurator/compare/v0.9.3..v0.9.4) - 2026-06-14

### Bug Fixes

- **auth:** USERMAP drift polish and validation breadth (AUTH-3b) [4a3ed71](https://github.com/platformrelay/MKurator/commit/4a3ed7165b66f73d3ee22569b7e41f9ca1f03acc)


### Features

- **auth:** QMGRMAP CHLAUTH remoteQueueManager field [d3c20f1](https://github.com/platformrelay/MKurator/commit/d3c20f187dcca16d2e43c8248b80b34aaade9777)

- **channelauth:** SSLPEERMAP CHLAUTH CRD and mqrest [3d00a85](https://github.com/platformrelay/MKurator/commit/3d00a855fcf12492c8d60a875978be9ede69700a)

- **channelauthrule:** USERMAP CHLAUTH CRD and mqrest SET [edea155](https://github.com/platformrelay/MKurator/commit/edea155e3d1f201e58074e1dd4c0c0dff1caef01)

## [0.9.3](https://github.com/platformrelay/MKurator/compare/v0.9.2..v0.9.3) - 2026-06-14

### Features

- **topic:** Add publishScope and subscribeScope typed fields (#39) [2d94ff7](https://github.com/platformrelay/MKurator/commit/2d94ff74791c5ffd03c5eae687661441fcda0960)

- **topic:** Add defPersistence typed field (#38) [aefd51f](https://github.com/platformrelay/MKurator/commit/aefd51f4128aa7906d3dcf33310b2673087bb8c0)

## [0.9.2](https://github.com/platformrelay/MKurator/compare/v0.9.1..v0.9.2) - 2026-06-14

### Features

- **topic:** Add publish and subscribe typed fields (#37) [e12ac33](https://github.com/platformrelay/MKurator/commit/e12ac333e7540598bdb6764a7d4f2ed16ca71ff6)

- **channel:** Add sslCipherSpec and sslClientAuth fields (#36) [416de83](https://github.com/platformrelay/MKurator/commit/416de83528eee9a8bd9d31bdcae90be2529a1996)

- **channel:** Add maxInstances and maxInstancesClient fields (#35) [e64e085](https://github.com/platformrelay/MKurator/commit/e64e085640a740caa744d152658e9c800a5bbfbd)

- **channel:** Add Channel.spec.mcaUser typed field (#34) [accf7a7](https://github.com/platformrelay/MKurator/commit/accf7a738265a191429c85006e8b194cbecc829b)

- **api:** Add Channel.spec.shareConv typed field (#33) [b042b27](https://github.com/platformrelay/MKurator/commit/b042b2788e95d30e0ce05856b4840061700098c5)

- **api:** Add Channel.spec.transportType typed field (#32) [7fbad61](https://github.com/platformrelay/MKurator/commit/7fbad6183e856ef9ee5280ad7ffaec6df7984dce)

- **api:** Add Channel.spec.maxMsgLength typed field (#31) [19ce060](https://github.com/platformrelay/MKurator/commit/19ce060a382f44b1bc1dd169fb36b54830348237)

## [0.9.1](https://github.com/platformrelay/MKurator/compare/v0.9.0..v0.9.1) - 2026-06-14

### Bug Fixes

- **api:** Guard Queue description CEL with has() (#23) [4b9f3e8](https://github.com/platformrelay/MKurator/commit/4b9f3e88cd1d12af812ab39ea77cd71c90556e61)

- **ci:** Use gitleaks CLI for org repo secret scan [bd23f0d](https://github.com/platformrelay/MKurator/commit/bd23f0d77dee198d431ad11eb8265473647577b6)


### Features

- **api:** Add Channel.spec.description typed field (#30) [a258043](https://github.com/platformrelay/MKurator/commit/a2580433bedf012a6553de445a1e4570f9cf6b1f)

- **api:** Add Topic.spec.description typed field (#29) [fbe2e39](https://github.com/platformrelay/MKurator/commit/fbe2e3998c12b568616e77f78db2eb23e93c8a63)

- **api:** Add Topic.spec.topicString typed field (#28) [7165c1f](https://github.com/platformrelay/MKurator/commit/7165c1fc44aaf66b0ea4ce9bd24f1f0b0fa7d37a)

- **api:** Add Queue remote xmitQueue and remoteQueueManager [de3c925](https://github.com/platformrelay/MKurator/commit/de3c925a175be5c226c95613b2c8413afa631dce)

- **api:** Add Queue.spec.targetQueue for alias queues [cfc8f75](https://github.com/platformrelay/MKurator/commit/cfc8f758bc940d9bab4841cc3aadefb0272553ec)

- **api:** Add typed get/put fields to Queue spec (#25) [2735298](https://github.com/platformrelay/MKurator/commit/2735298a29e447a90c44651aa51609ee9178ccb8)

- **api:** Add typed defPersistence to Queue spec (#24) [16f444d](https://github.com/platformrelay/MKurator/commit/16f444dda9cf974078fdba93c23747e1ca3600d2)

- **api:** Add typed description field to Queue spec [5a88796](https://github.com/platformrelay/MKurator/commit/5a887967728aa2d652479843d8ca140f9e726be1)

- **api:** Add typed maxDepth field to Queue spec [3725b14](https://github.com/platformrelay/MKurator/commit/3725b14976fb5e59691c29169d2c532f9fa347f6)

## [0.9.0](https://github.com/platformrelay/MKurator/compare/v0.8.0..v0.9.0) - 2026-06-10

### Bug Fixes

- **api:** Keep MQObject interface out of api package [20559e9](https://github.com/platformrelay/MKurator/commit/20559e90a63da37b54d289afcfb5b4f2eeee973b)


### Features

- **api:** Add CEL admission rules per ADR-0025 [5d3f86b](https://github.com/platformrelay/MKurator/commit/5d3f86b16a44f342d991b2492cfe2bf4aa4125e6)


### Refactoring

- **controller:** Collapse workload type switches via MQObject [c922a80](https://github.com/platformrelay/MKurator/commit/c922a8097ff4d7f7bb95f57ec17472ce91037db3)

- **validation:** Keep webhooks stateful-only [da3dfe9](https://github.com/platformrelay/MKurator/commit/da3dfe9cf896d09be114a891ffaffd1a0534d0d3)

## [0.8.0](https://github.com/platformrelay/MKurator/compare/v0.7.1..v0.8.0) - 2026-06-10

### Bug Fixes

- **controller:** Classify events via typed wrap errors [403f8d9](https://github.com/platformrelay/MKurator/commit/403f8d9ff11c489cd7728c15795efdb97e31b4b7)

- **controller:** Observe-only auth skips SET when missing [fdb8d55](https://github.com/platformrelay/MKurator/commit/fdb8d553be5a4b50b6895983808545dd6e9f6eb8)

- **controller:** Requeue workloads after finalizer add [1f0962e](https://github.com/platformrelay/MKurator/commit/1f0962e2947f367e576948b798f8c2500e1d480a)

- **controller:** Stabilize QMC Ready under secret watch [09ec4f1](https://github.com/platformrelay/MKurator/commit/09ec4f1e0cc6a1a43c0bbfe5183f04e10d5f6406)

- **controller:** Preserve QMC Ready on transient ping [9df6b39](https://github.com/platformrelay/MKurator/commit/9df6b390372abd80f9322760945c5f6ebe703666)


### Features

- **mqrest:** Add mqweb retry and circuit breaker [f514861](https://github.com/platformrelay/MKurator/commit/f5148617e3de4e0ca0fed55a5ae827a7121aa7bc)

- **api:** Add deletion and adoption lifecycle policies [c1fb8f0](https://github.com/platformrelay/MKurator/commit/c1fb8f082f46e036e00d867c8f55231e24817993)

- **runtime:** Scope Secret cache and warn on admin default [79ca386](https://github.com/platformrelay/MKurator/commit/79ca386faad12d1c6bfcea4e1a029a561c07e40c)

- **controller:** Watch referenced Secrets for QMC recovery [414ef89](https://github.com/platformrelay/MKurator/commit/414ef899db981929fcc4aa0f383ee7b350032e79)

- **controller:** Add spec.suspend and reconcile-now [e39edef](https://github.com/platformrelay/MKurator/commit/e39edef04d247ea9288fc6bb4ec7badead856b2c)

- **controller:** Expose configurable requeue intervals [c5f9860](https://github.com/platformrelay/MKurator/commit/c5f98601850d37f6fb314d9abf26535e3e49db8d)


### Refactoring

- **controller:** Delete dead drift helpers and padding tests [a01bcf5](https://github.com/platformrelay/MKurator/commit/a01bcf50b9b22d8c71d604600dd7ee4f6c41d769)

## [0.7.1](https://github.com/platformrelay/MKurator/compare/v0.7.0..v0.7.1) - 2026-06-09

### Bug Fixes

- **mqrest:** Identity-keyed cache with replace-on-rotation [2d99fbc](https://github.com/platformrelay/MKurator/commit/2d99fbcccfd7504cc15b1d412a7206bae3c5b294)

- **validation:** Harden CHLAUTH/AUTHREC MQSC fields [43f6974](https://github.com/platformrelay/MKurator/commit/43f697472b16ed2077cf2e1f74603a1f98b13712)

- **controller:** Deletion before connection chain (ADR-0022) [c483ac0](https://github.com/platformrelay/MKurator/commit/c483ac07100b6aea5cc7ef33acebee0718c4bef1)

- **controller:** Reliability fixes for wave 1 (#7-10) [83cd131](https://github.com/platformrelay/MKurator/commit/83cd131c74df376387c1482f39b11f4867e427ed)

- **controller:** Release without Secret; stop QMC hot loop [5e24f03](https://github.com/platformrelay/MKurator/commit/5e24f0392ddf7cb2b8825a3c6306333f1bd89276)

- **ci:** Stop release tags from failing GitHub Pages deploy [c894ae2](https://github.com/platformrelay/MKurator/commit/c894ae271a548ff93bff165e1104c3bd41d9a303)


### Features

- **controller:** Periodic jittered drift resync [2322d7c](https://github.com/platformrelay/MKurator/commit/2322d7ccd80a0f423abe7e9d087f7af129ebfa40)

## [0.7.0](https://github.com/platformrelay/MKurator/compare/v0.6.0..v0.7.0) - 2026-06-06

### Bug Fixes

- **test:** Stabilize Helm e2e metrics and raise coverage floor [0f2449e](https://github.com/platformrelay/MKurator/commit/0f2449ee69b83f815360ef2c414e783c134809d6)


### Features

- **release:** Add attestations and engineering standards docs [443e1a1](https://github.com/platformrelay/MKurator/commit/443e1a1d96adad9f9f2e7575b08a153b3aaa032a)

## [0.6.0](https://github.com/platformrelay/MKurator/compare/v0.5.3..v0.6.0) - 2026-06-03

### Bug Fixes

- **mqrest:** Treat AUTHREC NONE as not found [2c1aee8](https://github.com/platformrelay/MKurator/commit/2c1aee89ee981da080ff4cafad6a88b70dac707a)

- **ci:** Expose CODECOV_TOKEN on test job env [011f545](https://github.com/platformrelay/MKurator/commit/011f545368f9f34adf18be3a0e657d5dc803428b)

- **ci:** Skip Codecov without invalid secrets if [48b89a4](https://github.com/platformrelay/MKurator/commit/48b89a4a16e6f90cd5ca2ec0127cae8900cf1df8)

- **ci:** Unblock verify, codecov, and e2e CI [75866b3](https://github.com/platformrelay/MKurator/commit/75866b38fcc671d213a03f488d9e075de945a1ff)

- **test:** Bound kubectl in MQ e2e cleanup [d298d20](https://github.com/platformrelay/MKurator/commit/d298d203b5e07ef5f9ab2d38b54a9db71739a908)


### Refactoring

- [**breaking**] Rename project Kurator to MKurator [d2c5f95](https://github.com/platformrelay/MKurator/commit/d2c5f95eeecb0228a4cf99ce1749767d4b737916)

## [0.5.3](https://github.com/platformrelay/MKurator/compare/v0.5.2..v0.5.3) - 2026-06-03

### Bug Fixes

- **test:** Define kurator-system namespace in helpers [e8e69d4](https://github.com/platformrelay/MKurator/commit/e8e69d466b67d1a477e904af33c1e2bb2f4b11dc)

- **test:** Prevent e2e AfterSuite undeploy hang [0930ebc](https://github.com/platformrelay/MKurator/commit/0930ebc093d29dce8931809f4c37eb8e9ad33694)

- **ci:** Use git-cliff-action content for release notes [9fad692](https://github.com/platformrelay/MKurator/commit/9fad6921cddc303c62f482f1c8a04df69e3e1463)

- **test:** Make AfterEach kubectl delete non-blocking [e3a926d](https://github.com/platformrelay/MKurator/commit/e3a926d8b2de297007409d9075080220a34a3c2d)

- **test:** Stabilize parallel MQ e2e lanes [ad324a0](https://github.com/platformrelay/MKurator/commit/ad324a090eb0bcfa4b348072424893d3c34db117)

- **test:** Avoid corrupt merged coverage.out [dc7add6](https://github.com/platformrelay/MKurator/commit/dc7add6c31889a8305f2091ec8908040aa7dfdde)

- **rbac:** Allow events.k8s.io for controller [be9f9db](https://github.com/platformrelay/MKurator/commit/be9f9db62e407c8b11d2e1e83ada39c910f40812)

- **test:** Let Helm e2e own kurator-system namespace [b0685b1](https://github.com/platformrelay/MKurator/commit/b0685b12e931b951dccfaeb49ff5cce962249ea8)

- **deps:** Update kubernetes packages [c1376e1](https://github.com/platformrelay/MKurator/commit/c1376e19266cb9f5ad32d49a7fd90f7e8a76c788)

- **deps:** Update go dependencies [31386bc](https://github.com/platformrelay/MKurator/commit/31386bc8bc2df5ed53f4d222d3e8a21aad55c450)

## [0.5.2](https://github.com/platformrelay/MKurator/compare/v0.5.1..v0.5.2) - 2026-06-03

### Bug Fixes

- **helm:** Add auth CR RBAC and verify in helm:lint [c9c8688](https://github.com/platformrelay/MKurator/commit/c9c8688330395efc09610747eb537ac61007442a)

- **ci:** Repair Renovate workflow permissions and token [22cd170](https://github.com/platformrelay/MKurator/commit/22cd17006d769edb89d715f2c9055befc562fa73)

- **ci:** Drop invalid workflows permission from Renovate job [9e4fe1e](https://github.com/platformrelay/MKurator/commit/9e4fe1e120f898698f26c89337aa59e5b25643cd)

- **ci:** Configure Renovate repository target and token [3d598b6](https://github.com/platformrelay/MKurator/commit/3d598b691b75d97d4e2c7b8c02e902656b33112d)

- **ci:** Migrate renovate config for v43 [1208bb1](https://github.com/platformrelay/MKurator/commit/1208bb145414c92069c8c5910b70417ff6a20052)

- **ci:** Flock mutex for e2e and integration suites [f0b4227](https://github.com/platformrelay/MKurator/commit/f0b4227f3ef76a4aab059523d9c58b2810b90230)


### Features

- **e2e:** Wire Helm admission webhook e2e path [7402c3d](https://github.com/platformrelay/MKurator/commit/7402c3d7eb34925008c91ae5e32cbf39c7612f61)

- **mqpcf:** Scaffold PCF adapter behind MQAdmin [7dbed06](https://github.com/platformrelay/MKurator/commit/7dbed06e4d045de9466a9a0d693b56daf1faaeeb)


### Refactoring

- **controller:** Migrate to events EventRecorder API [2aa19fb](https://github.com/platformrelay/MKurator/commit/2aa19fbf72e2c25aed2fa00693e31b48ccb4ceec)

## [0.5.1](https://github.com/platformrelay/MKurator/compare/v0.5.0..v0.5.1) - 2026-06-03

### Bug Fixes

- **e2e:** Drop deprecated ginkgo.progress flag [8185002](https://github.com/platformrelay/MKurator/commit/8185002576b451de64d5a6c6cb2bb3137cea0364)

- **mqrest:** Treat empty AUTHREC authorities as not found [15b8c4d](https://github.com/platformrelay/MKurator/commit/15b8c4d9d78f8c6384f11d31394d66c8ae480497)


### Features

- **status:** Expose desiredMQSC on Topic, Channel, auth CRs [840c4b7](https://github.com/platformrelay/MKurator/commit/840c4b704098a32a280b6461b85fc4aa5b1fbf22)

## [0.5.0](https://github.com/platformrelay/MKurator/compare/v0.4.0..v0.5.0) - 2026-06-03

### Bug Fixes

- **auth:** Unblock ChannelAuthRule delete and e2e waits [718dff1](https://github.com/platformrelay/MKurator/commit/718dff1332e3ff8fa7f28f183b33d3ff915b0d37)

- **ci:** Skip generated files in format:check diff [4c2053a](https://github.com/platformrelay/MKurator/commit/4c2053a8ec1b9b21b136b59ee13ac76b41075c39)

- **auth:** Parse DISPLAY text and correct SET AUTHREC MQSC [8b1075d](https://github.com/platformrelay/MKurator/commit/8b1075d9cba877d568d15f3fbf6df9a0722a5a2e)

- **samples:** Unify deploy:samples for kind [e9d6849](https://github.com/platformrelay/MKurator/commit/e9d6849988a07d12bb905d30bb850f0863f838f3)

- **e2e:** Deploy operator via task deploy [ff1091e](https://github.com/platformrelay/MKurator/commit/ff1091e4b4ca287b2e2789384e62c93274c49a4b)

- **task:** Propagate KURATOR_E2E_MQ into test:e2e task env [0df9a9a](https://github.com/platformrelay/MKurator/commit/0df9a9a6949130d447ee9feecac0b9dcd8492419)

- **e2e:** Race-safe subprocess output and webhook assertion [867a201](https://github.com/platformrelay/MKurator/commit/867a20119292a2301cbf8cac4e154479c11db1f4)

- **task:** Resolve kustomize path with go tool -n [e40a0a0](https://github.com/platformrelay/MKurator/commit/e40a0a047ed54c2359cfe85f9706f80f5642afb2)

- **samples:** Let kustomization set namespace on Helm samples [64f82b9](https://github.com/platformrelay/MKurator/commit/64f82b923975ffca4f87e9de122812c78353da8b)

- **ci:** Bump Go 1.26.4 and sync verify artifacts [adb9b87](https://github.com/platformrelay/MKurator/commit/adb9b8764d2e772cb9a81ce8b9d306bfb9e6ca39)

- **ci:** Align CRDs with go tool controller-gen [9c56489](https://github.com/platformrelay/MKurator/commit/9c5648995e0ffc56d5cf9801f43c6cbf1a7235d0)

- **makefile:** Use go tool kustomize for deploy targets [87250df](https://github.com/platformrelay/MKurator/commit/87250df5495dd279c11518fddc4b08accf8db6c6)

- **e2e:** Wait for webhook cert and rollout before MQ tests [4456aec](https://github.com/platformrelay/MKurator/commit/4456aec225074aca8f3c6cd9b1483350d2a8ae78)

- **config:** Fix webhook kustomize for e2e make deploy [9307cb9](https://github.com/platformrelay/MKurator/commit/9307cb97cce77c809e85f30c0bb87e02ac628f76)


### Features

- **auth:** Drift-aware GET reconcile for auth CRs [26a25b7](https://github.com/platformrelay/MKurator/commit/26a25b7ce48186338b9f8f4e08ff3e6baacaad82)

- **operator:** Gate readyz on QMC connectivity [c5eaf5c](https://github.com/platformrelay/MKurator/commit/c5eaf5cae3c255bfff8998f52ecb7e21cf20ed89)

- **controller:** Observe-only drift policy and Phase 4 DISPLAY [4df20cd](https://github.com/platformrelay/MKurator/commit/4df20cd46c2db1db32bad6cb36bb988af1c4f119)

- **validation:** ChannelAuthRule channel referential checks [37021c7](https://github.com/platformrelay/MKurator/commit/37021c7c875f13b6ef8c667e7bf1980c1e49ba1b)

- **validation:** Tighten MQ object name checks [6d701a4](https://github.com/platformrelay/MKurator/commit/6d701a41408384e2c45cf5494354e2808b21cce1)

- **controller:** Status UX and reconcile concurrency [0ffdcf3](https://github.com/platformrelay/MKurator/commit/0ffdcf30d7832ceeeae7cd4d41cc26a936104133)

- **webhook:** Require opt-in for insecure QMC TLS [56b7fec](https://github.com/platformrelay/MKurator/commit/56b7fec8a55bda1aa7049c1cd5a1f3aa1fdfd3f9)

- **queue:** Expose status.desiredMQSC for GitOps debug [64e479e](https://github.com/platformrelay/MKurator/commit/64e479e93c1180bf001f6803c43f810bcec82fc2)

- **auth:** Add GetChannelAuth and GetAuthority MQAdmin paths [fcd4dd8](https://github.com/platformrelay/MKurator/commit/fcd4dd85450ddc2fb253ea0896b0b7c4929543d8)

- **auth:** Add ChannelAuthRule and AuthorityRecord CRDs [b960325](https://github.com/platformrelay/MKurator/commit/b960325a769ac312d76c329099898a3beacba348)

## [0.4.0](https://github.com/platformrelay/MKurator/compare/v0.3.0..v0.4.0) - 2026-06-02

### Features

- **webhook:** Deny QMC delete when dependents exist [48b5f39](https://github.com/platformrelay/MKurator/commit/48b5f39e630c71e7b4ca57f385541fd680e03e2e)

## [0.3.0](https://github.com/platformrelay/MKurator/compare/v0.2.2..v0.3.0) - 2026-06-02

### Bug Fixes

- **webhook:** Fix unit test race under -race [86e1f8c](https://github.com/platformrelay/MKurator/commit/86e1f8c41df6f98ee44ef62c40aa50224014e63f)


### Features

- **controller:** Expand Kubernetes event emission [0c267c6](https://github.com/platformrelay/MKurator/commit/0c267c641e3b94ba32e3badd94934acf9b423c66)


### Refactoring

- [**breaking**] Konih module path, docs hub, admission webhooks [90df81f](https://github.com/platformrelay/MKurator/commit/90df81f181f9cebf96596c0be98467cebbce7321)

## [0.2.2](https://github.com/platformrelay/MKurator/compare/v0.2.1..v0.2.2) - 2026-06-02

### Bug Fixes

- **makefile:** Apply CRDs from bases on make install [f740052](https://github.com/platformrelay/MKurator/commit/f740052b6dc3b6f9468c16b49a43d742096f9983)

- **test:** Pass QueueSpec to GetQueue in MQ e2e [6bf3072](https://github.com/platformrelay/MKurator/commit/6bf3072152bc0f1eb8a2eb32e628c0c8c0f0697a)


### Refactoring

- **controller:** Shared reconcile helpers and connection fixes [85a49cc](https://github.com/platformrelay/MKurator/commit/85a49cce2487d02623907abd1770ff1e63473826)

## [0.2.1](https://github.com/platformrelay/MKurator/compare/v0.2.0..v0.2.1) - 2026-06-02

### Bug Fixes

- **mqrest:** Normalize alias/remote DISPLAY attribute names from mqweb [71880ee](https://github.com/platformrelay/MKurator/commit/71880eefa24095fe28dc77b4222ff5c966feb03f)

## [0.2.0](https://github.com/platformrelay/MKurator/compare/v0.1.0..v0.2.0) - 2026-06-02

### Bug Fixes

- **ci:** Clear lint/verify; reconcile alias and remote queues [f1a674d](https://github.com/platformrelay/MKurator/commit/f1a674ddf44a8a3c814f489a8b1f35fb2f245802)

## [0.1.0] - 2026-06-02

### Bug Fixes

- **test:** Wait for CRDs after make install in MQ e2e [47d3418](https://github.com/platformrelay/MKurator/commit/47d341832f0700e38609a927f2a8fdbeb8c8daf6)

- **test:** Restore cmd declarations in deploy_helpers [d496aff](https://github.com/platformrelay/MKurator/commit/d496affa6194485a88ab4d4ac4ef99b4193825d8)

- **test:** Serialize e2e suites and idempotent namespace create [dc0647c](https://github.com/platformrelay/MKurator/commit/dc0647c9c27afbdb3f15e68a0a9e8fc2c8583d66)

- **test,ci:** Ordered MQ e2e context; gofmt metrics imports [5831666](https://github.com/platformrelay/MKurator/commit/58316662c6d1af08dea4b3b35576199ae7d43234)

- **test,ci:** MQ e2e redeploys operator; bump otel for Trivy [933b71c](https://github.com/platformrelay/MKurator/commit/933b71c490b6bc95894e054d81da20953f92fa2d)

- **ci:** Set KIND via GITHUB_ENV in e2e install step [89ff433](https://github.com/platformrelay/MKurator/commit/89ff433e7617ab68ee260d5db28372c6a777df38)

- **ci:** E2e PATH and sync deepcopy with controller-gen [4c1e294](https://github.com/platformrelay/MKurator/commit/4c1e294120318c342104e45d671849048e460dc1)

- **ci:** Unblock CI and e2e on GitHub Actions [c29429c](https://github.com/platformrelay/MKurator/commit/c29429c324bc2cb81ed980fcff23344d5ef4892a)

- **ci:** Pin correct setup-terraform action SHA [6ce2bb6](https://github.com/platformrelay/MKurator/commit/6ce2bb6314e5d9940d908fd998b85b95f39b167b)

- **queue:** Defer MQ admin client until connection is Ready [93ae3a7](https://github.com/platformrelay/MKurator/commit/93ae3a7690715e510e8707a778501c22ee29db38)

- **mqrest:** Drop maxmsglen from queue DISPLAY on mqweb 9.4 [e947130](https://github.com/platformrelay/MKurator/commit/e94713026a179866d1068483a7f6590e829d589b)

- **logging:** Reuse err var for Setup after Load [1576255](https://github.com/platformrelay/MKurator/commit/157625566cf261e741305056d63175a541336a26)


### Features

- **messaging:** Reconcile Topic and Channel CRs via mqweb [353135d](https://github.com/platformrelay/MKurator/commit/353135db3372c261fbb055c2c41b46a0db0b6e93)

- **metrics:** Add Prometheus metrics and Helm alerts [bdc414f](https://github.com/platformrelay/MKurator/commit/bdc414fe334536573fd48eacd36c4837471931b4)

- **kind:** Add mq console URL and runmqsc CLI tasks [786dc55](https://github.com/platformrelay/MKurator/commit/786dc55915d5c8072c2bed8a35a8ae0b81070ee9)

- **chart:** Add Helm chart, reference docs, and MQ e2e fixtures [0acb510](https://github.com/platformrelay/MKurator/commit/0acb51031516846feb8803706a07a4839eb7ebe1)

- Add Queue and QueueManagerConnection reconcilers [f34f98d](https://github.com/platformrelay/MKurator/commit/f34f98d4cd32cf0de81adba4fb5bf24e3e1353c2)

- **cluster:** Haproxy ingress, Argo CD, upstream IBM MQ [6bc8ba4](https://github.com/platformrelay/MKurator/commit/6bc8ba401a64c35fc8f8e660499ba5b3bffc21ca)

- Scaffold Kurator operator (Phase 1) [b27dd50](https://github.com/platformrelay/MKurator/commit/b27dd50e8e013f5f5fea9cd8b3dd7f9d98894e7d)

- **logging:** Add configurable slog logger [f3f76bf](https://github.com/platformrelay/MKurator/commit/f3f76bfdfd48987b3d20c0f3ab7a0cbc3395b8cd)

- Add one-command kind dev cluster [74855c7](https://github.com/platformrelay/MKurator/commit/74855c7e633b2ca99e79f244b314a95b3ace029e)

