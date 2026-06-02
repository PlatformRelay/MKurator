# ADR-0008: Generate changelogs with git-cliff

- **Status**: Accepted
- **Date**: 2026-06-02

## Context

The project already requires [Conventional Commits](https://www.conventionalcommits.org/)
with a [gitmoji](https://gitmoji.dev/) token in every subject (see
[AGENTS.md](../../AGENTS.md)). Release artifacts are produced manually: bump
`charts/kurator/Chart.yaml`, commit [`CHANGELOG.md`](../../CHANGELOG.md), tag
`v*`, and let [`.github/workflows/release.yaml`](../../.github/workflows/release.yaml)
build/push the image and publish install manifests via
[`hack/release-assets.sh`](../../hack/release-assets.sh).

We need release notes that stay in sync with git history without hand-writing
them, and a committed changelog users can read on the default branch. The
solution must fit [ADR-0005](0005-keep-tooling-lean.md): one maintainer, pinned
tooling, CI steps that map to `task` targets, no Node-only release stack for a
Go operator.

## Decision

We will generate changelogs with **[git-cliff](https://git-cliff.org/)** (pinned
**v2.13.1**):

- **[`cliff.toml`](../../cliff.toml)** at the repo root configures parsing, grouping,
  and a Keep a Changelog–style template. A **commit preprocessor** strips the
  `:gitmoji:` token so conventional parsing works on subjects like
  `feat(queue): :sparkles: summary`.
- **[`CHANGELOG.md`](../../CHANGELOG.md)** is committed and regenerated with
  `task changelog:write` before tagging; `task changelog` previews the unreleased
  section locally.
- **Releases** stay manually versioned. On tag push, `orhun/git-cliff-action`
  runs `git cliff --latest --strip header`; [`hack/assemble-release-notes.sh`](../../hack/assemble-release-notes.sh)
  appends install/cosign instructions from
  [`.github/release-notes-install.md`](../../.github/release-notes-install.md)
  to the GitHub Release body. Checkout uses `fetch-depth: 0` so tag ranges resolve.
- **Commit filters**: user-facing types (`feat`, `fix`, `perf`, `refactor`, breaking
  `!`) appear in the changelog; `docs`, `test`, `chore`, `ci`, `build`, and `style`
  are skipped. Unconventional commits (e.g. some bot merges) are dropped via
  `filter_unconventional`.

We explicitly do **not** adopt a release-PR bot or automatic semver bumps in this
step.

## Consequences

**Positive:**

- Changelog content derives from the same commit convention contributors already
  follow; local preview matches CI output (`task changelog` / `changelog:release`).
- Single config file (`cliff.toml`), no runtime dependency on Node or npm for the
  operator build.
- Pinned binary via `task tools:git-cliff` and the same version in GitHub Actions.

**Negative / operational:**

- Commit subjects must stay parseable; gitmoji placement and breaking `!` must
  remain consistent or entries are skipped or mis-grouped.
- `cliff.toml` and the Tera template need occasional tuning (e.g. whether `docs`
  commits belong in user-facing notes).
- Full history for a release requires a non-shallow clone in CI.

**Follow-up (optional, not decided here):**

- [commitlint](https://commitlint.js.org/) or a pre-commit hook if squash merges
  or bot commits routinely break the format.
- **release-please** if we later want automated version PRs across Helm chart and
  tags (see Alternatives).

## Alternatives considered

| Option | Why not (for now) |
|--------|-------------------|
| **release-please** | Opens release PRs, bumps versions in multiple files, and owns semver. Overlaps with our existing tag + `hack/release-assets.sh` flow and adds bot PR churn for a solo maintainer. Revisit if release frequency or multi-artifact versioning grows painful. |
| **semantic-release** | Node-centric; poor fit for a Kubebuilder/Go repo and harder to align with pinned `go tool` / Task workflows. |
| **conventional-changelog** (npm) | Same ecosystem mismatch; another package manager surface for changelog-only work. |
| **GitHub `generate_release_notes: true` alone** | Ignores our gitmoji convention, groups by GitHub PR metadata not commit type, and does not maintain a repo-level `CHANGELOG.md`. Still useful as a fallback; we replaced it with git-cliff + a static install footer. |
| **Hand-written changelogs** | Accurate but drifts from git history; does not scale with agent/human commit volume. |
| **git-chglog / git-changelog** | Viable; git-cliff chosen for mature Conventional Commit support, customizable TOML + templates, active maintenance, and a well-used GitHub Action. |

## References

- Implementation: [`cliff.toml`](../../cliff.toml), [`Taskfile.yml`](../../Taskfile.yml)
  (`changelog*`, `tools:git-cliff`), [docs/CICD.md](../CICD.md) release job.
- Related: [ADR-0005](0005-keep-tooling-lean.md) (lean tooling posture).
