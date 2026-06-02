#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! terraform -chdir="${ROOT_DIR}/terraform" output >/dev/null 2>&1; then
  echo "No Terraform outputs found. Run the bring-up first (task cluster:up)." >&2
  exit 1
fi

echo "=== Kurator local dev cluster ==="
terraform -chdir="${ROOT_DIR}/terraform" output
cat <<'EOF'

Host access (via ingress-nginx NodePort 30443, mkcert TLS):
  - IBM MQ web console : https://mq.localhost:30443/ibm/mq/console/
  - IBM MQ admin REST  : https://mq.localhost:30443/ibm/mq/rest/v2/admin/qmgr
  - Grafana            : https://grafana.localhost:30443/

In-cluster (for the operator's QueueManagerConnection.endpoint):
  - https://ibm-mq.ibm-mq.svc:9443
EOF
