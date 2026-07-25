# ADR-0028: Keep `integration` advisory (not a `protect-main` required check)

- **Status**: Accepted
- **Date**: 2026-07-25
- **Clarifies**: [ADR-0020](0020-merge-gate-matrix.md) L3 row; [CICD.md](../CICD.md) branch protection
- **Closes**: TESTQ-4 / HA-S5 remainder (interval flake already fixed in #111)

## Context

`integration.yaml` runs Docker IBM MQ + `task test:integration` on pushes/PRs that are
not docs-only (`paths-ignore`: `**.md`, `docs/**`, `charts/**/README.md`). It is **not**
listed in the live `protect-main` ruleset required checks (11 contexts as of 2026-07-25,
including `Build MkDocs site`).

TESTQ-4 asked whether to promote `integration` into the ruleset. Promoting a path-filtered
job without an always-report short-circuit recreates the GATE-1 hole: docs-only PRs sit on
"Expected — waiting for status" forever, so merges need `--admin` and the net is theater.
GATE-1 already taught this lesson for `Build MkDocs site` (always-report + then require).

ADR-0020's L3 cell ("When path filter runs") was easy to read as "required when it runs."
This ADR makes the ruleset posture explicit.

## Decision

We will **keep `integration` advisory**: it must keep running and stay visible on
code-touching PRs/pushes, but it is **not** added to `protect-main` required status checks
until a MkDocs-style always-report short-circuit exists (green when the path filter would
have skipped the real suite).

Release tagging still waits on green Integration via [release-gate](../RELEASE.md) — that
path is unchanged.

## Consequences

- Code PRs can still merge while Integration is red if the maintainer uses admin bypass
  without waiting — mitigated by AGENTS.md wait-for-fresh-head discipline for required
  checks; Integration remains a strong signal, not a hard merge gate.
- Docs-only PRs stay unblocked without phantom contexts.
- Follow-up story (optional): always-report shim for `integration`, then ruleset promote.
- HA-S5 / TESTQ-4 acceptance: interval race done (#111); required-check half = this ADR.

## Alternatives considered

- **Require `integration` now, no shim**: rejected — docs-only PRs freeze on Expected /
  need `--admin` forever.
- **Always-report shim, then require**: deferred — correct long-term path; not blocking
  today's posture; schedule when merge latency budget accepts ~minutes on every PR.
- **Drop the Integration workflow**: rejected — L3 coverage against real mqweb stays valuable.
