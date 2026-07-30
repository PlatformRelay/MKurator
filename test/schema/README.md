# CRD OpenAPI contract tests (no cluster)

Golden **spec** OpenAPI fragments extracted from `config/crd/bases/*.yaml` catch
kubebuilder marker drift without kind or e2e. Fragments include
`x-kubernetes-validations` (CEL) rules migrated per
[ADR-0025](../../docs/adr/0025-cel-first-admission-validation.md).

CEL acceptance/rejection parity with the prior webhook rules is covered by
`test/admission/webhook_envtest_test.go` (envtest against the committed
single-version v1beta1 CRDs).

## Enforced kinds

All **six** v1beta1 messaging kinds have checked-in goldens (see `DefaultCases` in `extract.go`):

| CRD file | Golden |
| --- | --- |
| `messaging.mkurator.dev_queues.yaml` | `queue.spec.openapi.yaml` |
| `messaging.mkurator.dev_topics.yaml` | `topic.spec.openapi.yaml` |
| `messaging.mkurator.dev_channels.yaml` | `channel.spec.openapi.yaml` |
| `messaging.mkurator.dev_channelauthrules.yaml` | `channelauthrule.spec.openapi.yaml` |
| `messaging.mkurator.dev_authorityrecords.yaml` | `authorityrecord.spec.openapi.yaml` |
| `messaging.mkurator.dev_queuemanagerconnections.yaml` | `queuemanagerconnection.spec.openapi.yaml` |

## Extend pattern

1. Add a row to `DefaultCases` in `extract.go` (CRD filename + golden filename).
2. Regenerate the golden: `task test:schema:update`
3. Commit `test/schema/golden/<kind>.spec.openapi.yaml`

## Commands

| Task | Purpose |
|------|---------|
| `task test:schema` | Run fragment tests only |
| `task test:schema:update` | Rewrite goldens from current CRDs |
| `task verify` | Includes schema check after controller-gen diff |

Sample YAML contract tests live in [`test/samples/`](../samples/) (decode + admission validation).

`kubectl explain` goldens are not implemented here; envtest CRD install is
unnecessary because fragments are derived directly from committed CRD YAML.
