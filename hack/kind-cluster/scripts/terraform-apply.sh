#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

KUBECONFIG_PATH="${KUBECONFIG_PATH:-${ROOT_DIR}/.state/kubeconfig.yaml}"
TLS_ENV="${TLS_ENV:-${ROOT_DIR}/.state/tls.env}"
STATE_DIR="${STATE_DIR:-${ROOT_DIR}/.state}"

if [[ ! -f "$KUBECONFIG_PATH" ]]; then
  echo "Kubeconfig not found at: $KUBECONFIG_PATH" >&2
  echo "Run: task cluster:kind:up" >&2
  exit 1
fi

if [[ ! -f "$TLS_ENV" ]]; then
  echo "TLS env not found at: $TLS_ENV" >&2
  echo "Run: task cluster:tls" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "$TLS_ENV"

if [[ -z "${TLS_CERT_STRING:-}" || -z "${TLS_KEY_STRING:-}" ]]; then
  echo "TLS_CERT_STRING / TLS_KEY_STRING missing from $TLS_ENV" >&2
  exit 1
fi

mkdir -p "$STATE_DIR"

terraform -chdir="${ROOT_DIR}/terraform" init -upgrade
terraform -chdir="${ROOT_DIR}/terraform" apply -auto-approve \
  -var="kubeconfig=${KUBECONFIG_PATH}" \
  -var="state_dir=${STATE_DIR}" \
  -var="tls_cert_string=${TLS_CERT_STRING}" \
  -var="tls_key_string=${TLS_KEY_STRING}"
