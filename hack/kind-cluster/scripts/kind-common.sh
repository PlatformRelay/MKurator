#!/usr/bin/env bash
# Shared helpers for kind-up / kind-down. Source this file; do not execute directly.
set -euo pipefail

# NodePorts published by kind/cluster.yaml (must stay in sync).
readonly KIND_HOST_HTTP_PORT="${KIND_HOST_HTTP_PORT:-30080}"
readonly KIND_HOST_HTTPS_PORT="${KIND_HOST_HTTPS_PORT:-30443}"

_kind_detect_provider() {
  if [[ -n "${KIND_EXPERIMENTAL_PROVIDER:-}" ]]; then
    return 0
  fi
  if command -v docker >/dev/null 2>&1; then
    return 0
  fi
  if command -v nerdctl >/dev/null 2>&1; then
    export KIND_EXPERIMENTAL_PROVIDER="nerdctl"
  elif command -v podman >/dev/null 2>&1; then
    export KIND_EXPERIMENTAL_PROVIDER="podman"
  else
    cat >&2 <<'EOF'
A container runtime is required (docker, nerdctl, or podman).
EOF
    return 1
  fi
}

# True when the control-plane container publishes our fixed NodePorts on the host.
_kind_node_publishes_ports() {
  local node_container="$1"
  local mapping
  mapping="$(docker port "$node_container" 2>/dev/null || true)"
  [[ -z "$mapping" ]] && return 1
  # docker port example: "30080/tcp -> 0.0.0.0:30080"
  grep -qE "(^|:)${KIND_HOST_HTTP_PORT}(/| |->)" <<<"$mapping" \
    && grep -qE "(^|:)${KIND_HOST_HTTPS_PORT}(/| |->)" <<<"$mapping"
}

# kind cluster name for a control-plane container name (<name>-control-plane), or empty.
_kind_cluster_from_node() {
  local node_container="$1"
  if [[ ! "$node_container" =~ -control-plane$ ]]; then
    return 0
  fi
  printf '%s' "${node_container%-control-plane}"
}

# List kind cluster names whose nodes hold our NodePorts (may be empty).
_kind_clusters_holding_ports() {
  local cluster node
  for node in $(docker ps --format '{{.Names}}' 2>/dev/null | grep -E 'control-plane$' || true); do
    if _kind_node_publishes_ports "$node"; then
      cluster="$(_kind_cluster_from_node "$node")"
      [[ -n "$cluster" ]] && printf '%s\n' "$cluster"
    fi
  done
}

# Remove other kind clusters that block our fixed NodePorts (local dev convenience).
_kind_remove_conflicting_clusters() {
  local target="$1" cluster
  while IFS= read -r cluster; do
    [[ -z "$cluster" || "$cluster" == "$target" ]] && continue
    echo "Removing kind cluster ${cluster} (NodePorts ${KIND_HOST_HTTP_PORT}/${KIND_HOST_HTTPS_PORT} in use)."
    kind delete cluster --name "$cluster" || true
  done < <(_kind_clusters_holding_ports | sort -u)
}

# Drop a broken control-plane container that is not registered with kind.
_kind_remove_orphan_node() {
  local cluster_name="$1"
  local node="${cluster_name}-control-plane"
  if docker inspect "$node" >/dev/null 2>&1; then
    if ! kind get clusters 2>/dev/null | grep -qx "$cluster_name"; then
      echo "Removing orphaned node container: ${node}"
      docker rm -f "$node" >/dev/null 2>&1 || true
    fi
  fi
}

_ports_in_use_on_host() {
  if command -v ss >/dev/null 2>&1; then
    ss -tlnH "sport = :${KIND_HOST_HTTP_PORT} or sport = :${KIND_HOST_HTTPS_PORT}" 2>/dev/null | grep -q .
    return
  fi
  # Fallback when ss is unavailable.
  local port
  for port in "$KIND_HOST_HTTP_PORT" "$KIND_HOST_HTTPS_PORT"; do
    if (echo >/dev/tcp/127.0.0.1/"$port") >/dev/null 2>&1; then
      return 0
    fi
  done
  return 1
}

_kind_fail_ports_still_blocked() {
  cat >&2 <<EOF
NodePorts ${KIND_HOST_HTTP_PORT} and/or ${KIND_HOST_HTTPS_PORT} are still in use on the host.

This usually means another process (not a kind cluster we could remove) holds the ports.
Free the ports, or run: task cluster:down

To inspect: ss -tlnp | grep -E ':${KIND_HOST_HTTP_PORT}|:${KIND_HOST_HTTPS_PORT}'
EOF
  return 1
}
