#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

CLUSTER_NAME="${CLUSTER_NAME:-kurator}"
KUBECONFIG_OUT="${KUBECONFIG_OUT:-${ROOT_DIR}/.state/kubeconfig.yaml}"

if ! command -v kind >/dev/null 2>&1; then
  echo "kind is required (https://kind.sigs.k8s.io/)" >&2
  exit 1
fi

if [[ -z "${KIND_EXPERIMENTAL_PROVIDER:-}" ]]; then
  if command -v docker >/dev/null 2>&1; then
    : # default provider is docker
  elif command -v nerdctl >/dev/null 2>&1; then
    export KIND_EXPERIMENTAL_PROVIDER="nerdctl"
  elif command -v podman >/dev/null 2>&1; then
    export KIND_EXPERIMENTAL_PROVIDER="podman"
  fi
fi

kind delete cluster --name "$CLUSTER_NAME" || true
rm -f "$KUBECONFIG_OUT"
