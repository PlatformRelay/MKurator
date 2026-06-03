# Security Policy

MKurator manages administrative objects on IBM MQ Queue
Managers and handles connection credentials. Security is a first-class concern;
the enforced requirements live in
[docs/NON_FUNCTIONAL_REQUIREMENTS.md](docs/NON_FUNCTIONAL_REQUIREMENTS.md) and the
threat-relevant design in [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md#security-model).

## Reporting a vulnerability

Please report suspected vulnerabilities **privately** — do not open a public
issue for security problems.

- Use **GitHub Security Advisories** ("Report a vulnerability") on this
  repository, or contact the maintainer privately.
- Include affected version/commit, a description, reproduction steps, and impact.
- You will receive an acknowledgement; fixes for confirmed issues are prioritised
  and disclosed once a fix is available.

This is a personal project without a formal SLA, but security reports are taken
seriously and handled promptly.

## Supported versions

The project is pre-1.0 (`v1alpha1`). Only the latest released version / default
branch receives fixes. The API contract may change between alpha releases.

## Security posture

- **No inline secrets**: credentials and CA material come from referenced
  Kubernetes `Secret`s only — never in CR specs, code, images, or logs.
- **TLS by default**: HTTPS to mqweb with certificate verification on;
  `insecureSkipVerify` is opt-in and intended for local development only.
- **Least-privilege RBAC**: scoped to the operator's own API group, referenced
  Secrets, Events, and the leader-election Lease — no wildcards, no cluster-admin.
- **Hardened runtime**: CGO-free static binary in a distroless nonroot image,
  read-only root filesystem, dropped capabilities, no privilege escalation.
- **Supply chain**: pinned tooling and CI action SHAs, committed `go.sum`,
  `govulncheck` (PR + scheduled), and Trivy image scanning. See
  [docs/CICD.md](docs/CICD.md).

## Handling credentials in development

The local environment (`hack/kind-cluster`) ships **development-only** default
passwords for the IBM MQ users and Grafana. These are for ephemeral local kind
clusters and must never be reused in any shared or production environment.
