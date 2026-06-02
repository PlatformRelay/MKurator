# Release 0.5.0 — follow-ups

Tracking work to finish the **Phase 5 auth** release after CRD/reconciler code landed on
`main`. Update checkboxes as items complete.

**Related:** [ROADMAP.md](../ROADMAP.md#phase-5--user--authority-management) ·
[PHASE5_AUTH_SKETCH.md](../PHASE5_AUTH_SKETCH.md) · [RELEASE.md](../RELEASE.md)

## Documentation

- [x] README — Phase 5 CRs in “what ships”; CI tier table includes auth
- [x] [INSTALL_AND_USE.md](../INSTALL_AND_USE.md) — auth CRs, install `VERSION`, sample table
- [x] [config/samples/README.md](../../config/samples/README.md) — apply order + field notes for auth samples
- [x] [PHASE5_AUTH_SKETCH.md](../PHASE5_AUTH_SKETCH.md) — mark shipped vs planned rule types
- [x] [ATTRIBUTE_RECONCILIATION.md](../ATTRIBUTE_RECONCILIATION.md) — auth objects in scope
- [x] [DEVELOPMENT.md](../DEVELOPMENT.md) — test tiers, drop stale `make` e2e references
- [x] [CICD.md](../CICD.md) — integration tier covers auth MQSC; verify includes mocks
- [x] [ROADMAP.md](../ROADMAP.md) — Phase 5 partial exit; link here for remaining items

## E2e (required for Phase 5 exit)

- [ ] **KUBECONFIG** — force kind kubeconfig in `hack/ci/run-e2e.sh` (or BeforeSuite) so
  cert-manager install does not hit a broken default `kubectl` context
- [ ] **ChannelAuthRule e2e** — apply CR after channel fixture; assert `SET CHLAUTH` via mqweb;
  delete CR and assert rule removed (or idempotent not-found)
- [ ] **AuthorityRecord e2e** — apply CR for queue profile + principal; assert mqweb/OAM effect;
  delete and assert cleanup
- [ ] **Webhook negative** (optional) — invalid auth CR rejected on `kubectl apply`
- [ ] `task ci:e2e` green locally

## Release mechanics

- [ ] Bump `charts/kurator/Chart.yaml` `version` / `appVersion` to **0.5.0**
- [ ] `task changelog:write` and commit `CHANGELOG.md`
- [ ] `git tag v0.5.0` · `gh release create` · `git push origin main` · `git push origin v0.5.0`
- [ ] Confirm GitHub Actions **CI**, **Integration**, and **E2E** workflows green on the tag push

## CI hardening (nice-to-have, post-0.5.0)

- [ ] Add `task format:check` to `ci.yaml` (Task target already exists)
- [ ] Path filters on integration/e2e workflows to skip when only docs change
- [ ] Scheduled `govulncheck` workflow (if not already covered by Renovate weekly)

## Out of scope for 0.5.0

- Additional CHLAUTH rule types beyond `ADDRESSMAP` (schema allows them; adapter validates at MQSC apply time)
- TLS channel drift (`sslciph`, `sslcauth`) — remains Phase 4 optional item
- PCF adapter, OCI Helm registry push
