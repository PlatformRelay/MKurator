# Coding standards

Go *how* — formatting, lint, modules, and CI gates. Process lives in
[CONTRIBUTING.md](../../CONTRIBUTING.md); operator quality bar in [guidelines.md](guidelines.md).

## Formatting

- `gofmt`, `goimports`, `golines` (max line length **120**).
- Run `task format` locally; CI runs `task format:check`.

## Linting

**golangci-lint v2** with `default: none` (explicit opt-in). Enabled linters include:
`errcheck`, `gosec`, `govet`, `staticcheck`, `revive`, `ginkgolinter`, and others — see
[`.golangci.yaml`](../../.golangci.yaml).

- `task lint` runs golangci-lint and (when configured) `go-arch-lint` / depguard.
- Generated code (`zz_generated.*`, mocks) is excluded or lax.

## Modules and build

- Go version pinned in `go.mod`; tool directives pin controller-gen, golangci-lint, govulncheck, etc.
- `go mod verify` in preflight; committed `go.sum`.
- **CGO-free** static builds: `CGO_ENABLED=0`.
- Race detector on unit/envtest CI: `task test:run`.

## Security linting

- **gosec** enabled via golangci-lint — no credentials in code paths.
- **depguard** / **gomodguard** block deprecated logging/error libs (`logrus`, `pkg/errors`, `io/ioutil`).
- Never commit secrets; pre-commit runs gitleaks with [`.github/gitleaks.toml`](../../.github/gitleaks.toml).

## Pull request and CI gates

| Gate | Local | CI |
| --- | --- | --- |
| Codegen drift | `task verify` | `verify`, `preflight` |
| Format | `task format:check` | `lint` |
| Lint | `task lint` | `lint` |
| Unit + envtest | `task test:run` | `test` |
| Vulnerabilities | `task vuln:check` | `test`, `vulncheck` (weekly) |
| RBAC audit | `task audit:rbac` | `audit-rbac` |
| Markdown | `task lint:markdown` | `preflight` |
| Shell | `task lint:shell` | `preflight` |

Branch protection should require green **`preflight`** and **`CI`** on `main` — see [CICD.md](../CICD.md).

## Headers and style

- Apache 2.0 boilerplate on Go source files (`hack/boilerplate.go.txt`).
- Follow [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) conventions.
- Keep reconcilers thin; push MQ I/O behind the `MQAdmin` port ([ADR-0002](../adr/0002-manage-mq-via-mqweb-rest.md)).

## Related documents

| Document | Owns |
| --- | --- |
| [tooling-setup.md](tooling-setup.md) | arch-lint, SonarCloud (disabled) |
| [testing.md](testing.md) | Test tiers and coverage floor |
| [CICD.md](../CICD.md) | Full workflow matrix |
