#!/usr/bin/env bash
# Regenerate committed artifacts and fail if anything drifts.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT

copy_generated() {
  for path in config/crd/bases config/rbac charts/kurator/crds; do
    if [[ -d "$path" ]]; then
      mkdir -p "$scratch/$(dirname "$path")"
      cp -a "$path" "$scratch/$path"
    fi
  done
  shopt -s nullglob
  for f in api/*/zz_generated.deepcopy.go; do
    mkdir -p "$scratch/$(dirname "$f")"
    cp -a "$f" "$scratch/$f"
  done
}

copy_generated

go tool controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/..."
go tool controller-gen \
  rbac:roleName=manager-role \
  crd \
  webhook \
  paths="./api/...;./internal/controller/...;./internal/webhook/...;./cmd/..." \
  output:crd:artifacts:config=config/crd/bases

bash hack/helm-sync-crds.sh

echo "verify: comparing generated artifacts..."

for path in config/crd/bases config/rbac charts/kurator/crds; do
  if [[ -d "$scratch/$path" ]] || [[ -d "$path" ]]; then
    if ! diff -ru "$scratch/$path" "$path"; then
      echo "verify: drift in $path — run 'task generate && task manifests'" >&2
      exit 1
    fi
  fi
done

shopt -s nullglob
for f in api/*/zz_generated.deepcopy.go; do
  if ! diff -u "$scratch/$f" "$f"; then
    echo "verify: drift in $f — run 'task generate && task manifests'" >&2
    exit 1
  fi
done

echo "verify: ok"
