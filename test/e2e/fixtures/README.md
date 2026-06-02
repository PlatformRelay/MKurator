# E2e MQ fixtures

MQSC snippets used when `KURATOR_E2E_MQ=1` (see [DEVELOPMENT.md](../../../docs/DEVELOPMENT.md)).

## Files

| File | Source | Purpose |
|------|--------|---------|
| `channel-auth-prereq.mqsc` | [mq-gitops-samples `qmdemo-mqsc-config-map.yaml`](../../../references/mq-gitops-samples/queue-manager-basic-deployment/qmdemo-mqsc-config-map.yaml) | SVRCONN channel + ADDRESSMAP CHLAUTH for fixture and Phase 4 planning |

## Environment variables

| Variable | Default | Meaning |
|----------|---------|---------|
| `KURATOR_E2E_MQ` | unset | Set to `1` to run IBM MQ e2e tests |
| `KURATOR_E2E_MQ_ENDPOINT` | `https://127.0.0.1:30443` | mqweb base URL from test runner (ingress / port-forward) |
| `KURATOR_E2E_MQ_HOST` | `mq.localhost` | HTTP `Host` header when using HAProxy ingress on kind |
| `KURATOR_E2E_MQ_QMGR` | `QM1` | Queue manager name |
| `KURATOR_E2E_MQ_USER` | `admin` | mqweb user |
| `KURATOR_E2E_MQ_PASSWORD` | `passw0rd` | mqweb password (`hack/kind-cluster` terraform default) |
| `KURATOR_E2E_MQ_INSECURE_TLS` | `true` | Skip TLS verify for local mkcert |

Operator reconcile tests use the in-cluster endpoint `https://ibm-mq.ibm-mq.svc:9443`
via `QueueManagerConnection` CRs, not these host-facing variables.
