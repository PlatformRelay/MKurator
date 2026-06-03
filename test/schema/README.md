# CRD OpenAPI contract tests (no cluster)

Golden **spec** OpenAPI fragments extracted from `config/crd/bases/*.yaml` catch
kubebuilder marker drift without kind or e2e.

## Smoke + extend pattern

1. Add a row to `DefaultCases` in `extract.go` (CRD filename + golden filename).
2. Regenerate the golden: `task test:schema:update`
3. Commit `test/schema/golden/<kind>.spec.openapi.yaml`

Enforced kinds: **Queue**, **Topic**, **Channel**, **ChannelAuthRule**, **AuthorityRecord**,
**QueueManagerConnection**. Add a `DefaultCases` row and golden to extend coverage.

## Commands

| Task | Purpose |
|------|---------|
| `task test:schema` | Run fragment tests only |
| `task test:schema:update` | Rewrite goldens from current CRDs |
| `task verify` | Includes schema check after controller-gen diff |

`kubectl explain` goldens (Option B) are not implemented here; envtest CRD install is
unnecessary because fragments are derived directly from committed CRD YAML.
