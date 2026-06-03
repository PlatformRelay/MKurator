#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE_DIR="${STATE_DIR:-${ROOT_DIR}/.state}"

if ! terraform -chdir="${ROOT_DIR}/terraform" output >/dev/null 2>&1; then
  echo "No Terraform outputs found. Run: task cluster:up" >&2
  exit 1
fi

echo "=== MKurator local kind cluster ==="
terraform -chdir="${ROOT_DIR}/terraform" output

ARGOCD_ENABLED="$(terraform -chdir="${ROOT_DIR}/terraform" output -raw enable_argocd 2>/dev/null || echo false)"
if [[ "${ARGOCD_ENABLED}" == "true" && -f "${STATE_DIR}/argocd.env" ]]; then
  echo ""
  echo "Argo CD credentials are in ${STATE_DIR}/argocd.env (source it to load ARGOCD_*)."
fi

cat <<EOF

Host access (via HAProxy ingress NodePort 30443, mkcert TLS):
EOF

if [[ "${ARGOCD_ENABLED}" == "true" ]]; then
  cat <<'EOF'
  - Argo CD           : https://argocd.localhost:30443/  (admin; password in .state/argocd.env)
EOF
fi

cat <<'EOF'
  - IBM MQ console    : https://mq.localhost:30443/ibmmq/console/
  - IBM MQ admin REST : https://mq.localhost:30443/ibmmq/rest/v2/admin/qmgr
  - Grafana           : https://grafana.localhost:30443/

In-cluster (QueueManagerConnection.endpoint):
  - https://ibm-mq.ibm-mq.svc:9443

MQSC CLI (from repo root):
  - task mq:cli
  - task mq:runmqsc -- "DISPLAY QLOCAL('APP.ORDERS')"
  - docs/IBM_MQ_101.md — verify MKurator + IBM MQ

Enable Argo CD on the next apply: ENABLE_ARGOCD=true task cluster:apply
EOF
