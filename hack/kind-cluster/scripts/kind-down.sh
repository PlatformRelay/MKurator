#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck source=kind-common.sh
source "${ROOT_DIR}/scripts/kind-common.sh"

CLUSTER_NAME="${CLUSTER_NAME:-mkurator}"
KUBECONFIG_OUT="${KUBECONFIG_OUT:-${ROOT_DIR}/.state/kubeconfig.yaml}"

if ! command -v kind >/dev/null 2>&1; then
  echo "kind is required (https://kind.sigs.k8s.io/)" >&2
  exit 1
fi

_kind_detect_provider

if kind get clusters 2>/dev/null | grep -qx "$CLUSTER_NAME"; then
  kind delete cluster --name "$CLUSTER_NAME" || true
fi

_kind_remove_orphan_node "$CLUSTER_NAME"

# Also drop any other kind clusters still holding our fixed NodePorts.
_kind_remove_conflicting_clusters "__none__"

rm -f "$KUBECONFIG_OUT"
