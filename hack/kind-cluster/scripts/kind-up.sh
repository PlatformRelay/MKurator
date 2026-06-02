#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

CLUSTER_NAME="${CLUSTER_NAME:-kurator}"
KUBECONFIG_OUT="${KUBECONFIG_OUT:-${ROOT_DIR}/.state/kubeconfig.yaml}"
KIND_CONFIG="${KIND_CONFIG:-${ROOT_DIR}/kind/cluster.yaml}"

mkdir -p "$(dirname "$KUBECONFIG_OUT")"

if ! command -v kind >/dev/null 2>&1; then
  echo "kind is required (https://kind.sigs.k8s.io/)" >&2
  exit 1
fi

if [[ -z "${KIND_EXPERIMENTAL_PROVIDER:-}" ]]; then
  # Prefer Docker, then nerdctl+containerd, then Podman.
  if command -v docker >/dev/null 2>&1; then
    : # default provider is docker
  elif command -v nerdctl >/dev/null 2>&1; then
    export KIND_EXPERIMENTAL_PROVIDER="nerdctl"
  elif command -v podman >/dev/null 2>&1; then
    export KIND_EXPERIMENTAL_PROVIDER="podman"
  else
    cat >&2 <<'EOF'
A container runtime is required.

Supported options for Kind:
  - docker (recommended on Linux)
  - nerdctl (containerd)
  - podman
EOF
    exit 1
  fi
fi

if [[ "${KIND_EXPERIMENTAL_PROVIDER:-}" == "podman" ]]; then
  # Kind + rootless Podman needs systemd cgroup delegation.
  if [[ "$(podman info --format '{{.Host.Security.Rootless}}' 2>/dev/null || echo false)" == "true" ]]; then
    cat >&2 <<'EOF'
Kind detected rootless Podman.

Running Kind with the Podman provider in rootless mode requires systemd delegation ("Delegate=yes").
See: https://kind.sigs.k8s.io/docs/user/rootless/
EOF
    exit 1
  fi
fi

if ! kind get clusters | grep -qx "$CLUSTER_NAME"; then
  kind create cluster \
    --name "$CLUSTER_NAME" \
    --config "$KIND_CONFIG" \
    --kubeconfig "$KUBECONFIG_OUT"
else
  kind export kubeconfig --name "$CLUSTER_NAME" --kubeconfig "$KUBECONFIG_OUT"
fi

echo "Kubeconfig written to: $KUBECONFIG_OUT"
