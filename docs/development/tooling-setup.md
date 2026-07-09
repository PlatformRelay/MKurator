# Tooling setup

Maintainer setup for optional quality tools beyond default `task install`.

## go-arch-lint

Internal package layering is enforced by [go-arch-lint](https://github.com/fe3dback/go-arch-lint):

- Config: [`hack/tooling/go-arch-lint.yml`](https://github.com/conduit-ops/MKurator/blob/main/hack/tooling/go-arch-lint.yml)
- Local: `task arch:lint` (also runs as part of `task lint` when wired)

Controllers must depend on `mqadmin` / adapter ports, not vice versa. See [GO_MODULE.md](../GO_MODULE.md).

## depguard / gomodguard

Configured in [`.golangci.yaml`](https://github.com/conduit-ops/MKurator/blob/main/.golangci.yaml). Denies `logrus`, `pkg/errors`, and `io/ioutil` —
use `log/slog` and stdlib errors.

## SonarCloud (disabled)

SonarCloud analysis is **scaffolded but disabled** until the repository moves to the **conduit-ops**
GitHub organization.

| Item | Status |
| --- | --- |
| Workflow | [`.github/workflows/sonarcloud.yaml`](https://github.com/conduit-ops/MKurator/blob/main/.github/workflows/sonarcloud.yaml) (`if: false`) |
| Token | Set `SONAR_TOKEN` in repo secrets when enabling |
| Project key | `conduit-ops_mkurator` (placeholder in workflow) |

To enable after org migration:

1. Create SonarCloud project under **conduit-ops**.
2. Add `SONAR_TOKEN` secret.
3. Remove `if: false` from the workflow job.
4. Uncomment `SONAR_PROJECT_KEY` in workflow env.

## Polaris / kubeaudit (RBAC)

RBAC audit runs in CI without local install if tools are missing — `hack/audit-rbac.sh` downloads
pinned Polaris and kubeaudit on demand.

Local: `task audit:rbac`

## Related documents

| Document | Owns |
| --- | --- |
| [coding-standards.md](coding-standards.md) | CI gate summary |
| [CICD.md](../CICD.md) | Workflow contract |
