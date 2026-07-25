# Tooling setup

Maintainer setup for optional quality tools beyond default `task install`.

## go-arch-lint

Internal package layering is enforced by [go-arch-lint](https://github.com/fe3dback/go-arch-lint):

- Config: [`hack/tooling/go-arch-lint.yml`](https://github.com/platformrelay/MKurator/blob/main/hack/tooling/go-arch-lint.yml)
- Local: `task arch:lint` (also runs as part of `task lint` when wired)

Controllers must depend on `mqadmin` / adapter ports, not vice versa. See [GO_MODULE.md](../GO_MODULE.md).

## depguard / gomodguard

Configured in [`.golangci.yaml`](https://github.com/platformrelay/MKurator/blob/main/.golangci.yaml). Denies `logrus`, `pkg/errors`, and `io/ioutil` —
use `log/slog` and stdlib errors.

## SonarCloud (CI-5)

CI-based analysis runs as the advisory `sonarcloud` job in
[`.github/workflows/ci.yaml`](https://github.com/platformrelay/MKurator/blob/main/.github/workflows/ci.yaml)
(`needs: [test]`, reuses the `coverage` artifact). A
[`workflow_dispatch` shim](https://github.com/platformrelay/MKurator/blob/main/.github/workflows/sonarcloud.yaml)
remains for on-demand re-scans. **Not** a `protect-main` required check.

| Item | Status |
| --- | --- |
| Project key | `PlatformRelay_MKurator` ([`sonar-project.properties`](https://github.com/platformrelay/MKurator/blob/main/sonar-project.properties)) |
| Organization | `platformrelay` |
| Token | Repo secret `SONAR_TOKEN` (already provisioned) |
| Fork PRs | Warn-skip when the secret is withheld — job stays green |

Before the first green CI analysis with coverage:

1. On [sonarcloud.io](https://sonarcloud.io) → project **PlatformRelay_MKurator** → **Administration → Analysis Method**: **disable Automatic Analysis** (mutually exclusive with CI analysis; Automatic Analysis cannot import Go coverage).
2. Confirm `SONAR_TOKEN` is set under repo **Settings → Secrets → Actions**.
3. Land/merge CI-5; first same-repo PR or push to `main` should show non-zero Go coverage on the dashboard.

## Polaris / kubeaudit (RBAC)

RBAC audit runs in CI without local install if tools are missing — `hack/audit-rbac.sh` downloads
pinned Polaris and kubeaudit on demand.

Local: `task audit:rbac`

## Related documents

| Document | Owns |
| --- | --- |
| [coding-standards.md](coding-standards.md) | CI gate summary |
| [CICD.md](../CICD.md) | Workflow contract |
