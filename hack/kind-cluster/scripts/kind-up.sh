#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=kind-common.sh
source "${ROOT_DIR}/scripts/kind-common.sh"

CLUSTER_NAME="${CLUSTER_NAME:-mkurator}"
KUBECONFIG_OUT="${KUBECONFIG_OUT:-${ROOT_DIR}/.state/kubeconfig.yaml}"
KIND_CONFIG="${KIND_CONFIG:-${ROOT_DIR}/kind/cluster.yaml}"

mkdir -p "$(dirname "$KUBECONFIG_OUT")"

if ! command -v kind >/dev/null 2>&1; then
  echo "kind is required (https://kind.sigs.k8s.io/)" >&2
  exit 1
fi

_kind_detect_provider

if [[ "${KIND_EXPERIMENTAL_PROVIDER:-}" == "podman" ]]; then
  if [[ "$(podman info --format '{{.Host.Security.Rootless}}' 2>/dev/null || echo false)" == "true" ]]; then
    cat >&2 <<'EOF'
Kind detected rootless Podman. See https://kind.sigs.k8s.io/docs/user/rootless/
EOF
    exit 1
  fi
fi

if kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
  if docker inspect "${CLUSTER_NAME}-control-plane" >/dev/null 2>&1; then
    kind export kubeconfig --name "$CLUSTER_NAME" --kubeconfig "$KUBECONFIG_OUT"
    echo "Kind cluster ${CLUSTER_NAME} already exists; kubeconfig refreshed."
    echo "Kubeconfig written to: $KUBECONFIG_OUT"
    exit 0
  fi
  echo "Kind cluster ${CLUSTER_NAME} is registered but the node is missing; recreating."
  kind delete cluster --name "$CLUSTER_NAME" || true
fi

_kind_remove_orphan_node "$CLUSTER_NAME"
_kind_remove_conflicting_clusters "$CLUSTER_NAME"

if _ports_in_use_on_host; then
  _kind_fail_ports_still_blocked
fi

if ! kind create cluster \
  --name "$CLUSTER_NAME" \
  --config "$KIND_CONFIG" \
  --kubeconfig "$KUBECONFIG_OUT"; then
  echo "kind create failed; cleaning up partial node (if any)." >&2
  docker rm -f "${CLUSTER_NAME}-control-plane" >/dev/null 2>&1 || true
  exit 1
fi

echo "Kubeconfig written to: $KUBECONFIG_OUT"
