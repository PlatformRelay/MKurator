#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

KUBECONFIG_PATH="${KUBECONFIG_PATH:-${ROOT_DIR}/.state/kubeconfig.yaml}"
TLS_ENV="${TLS_ENV:-${ROOT_DIR}/.state/tls.env}"
CLUSTER_NAME="${CLUSTER_NAME:-kurator}"

echo "This removes resources created by Terraform (ingress, cert-manager, monitoring, IBM MQ)."
echo "Set DELETE_CLUSTER=true to also delete the kind cluster, DELETE_STATE=true to wipe .state."
echo ""

if [[ -f "$TLS_ENV" ]]; then
  # shellcheck disable=SC1090
  source "$TLS_ENV"
fi

if [[ -d "${ROOT_DIR}/terraform" && -f "$KUBECONFIG_PATH" ]]; then
  echo "Running terraform destroy..."
  terraform -chdir="${ROOT_DIR}/terraform" init -upgrade
  terraform -chdir="${ROOT_DIR}/terraform" destroy -auto-approve \
    -var="kubeconfig=${KUBECONFIG_PATH}" \
    -var="tls_cert_string=${TLS_CERT_STRING:-}" \
    -var="tls_key_string=${TLS_KEY_STRING:-}" || true
else
  echo "Kubeconfig or terraform dir missing; skipping terraform destroy."
fi

if [[ "${DELETE_CLUSTER:-false}" == "true" ]]; then
  echo "Deleting kind cluster: ${CLUSTER_NAME}"
  kind delete cluster --name "${CLUSTER_NAME}" || true
  rm -f "$KUBECONFIG_PATH"
fi

if [[ "${DELETE_STATE:-false}" == "true" ]]; then
  echo "Removing local state: ${ROOT_DIR}/.state"
  rm -rf "${ROOT_DIR}/.state"
fi

echo "Cleanup complete."
