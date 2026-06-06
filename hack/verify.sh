#!/usr/bin/env bash
# Regenerate committed artifacts and fail if anything drifts.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

scratch="$(mktemp -d)"
trap 'rm -rf "$scratch"' EXIT

copy_generated() {
  for path in config/crd/bases config/rbac charts/mkurator/crds; do
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
  if [[ -d test/mocks ]]; then
    mkdir -p "$scratch/test"
    cp -a test/mocks "$scratch/test/mocks"
  fi
}

copy_generated

go tool controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/..."
go tool controller-gen \
  rbac:roleName=manager-role \
  crd \
  webhook \
  paths="./api/...;./internal/controller/...;./internal/webhook/...;./cmd/..." \
  output:crd:artifacts:config=config/crd/bases

if grep -q 'packages:' .mockery.yaml 2>/dev/null && ! grep -q 'packages: {}' .mockery.yaml; then
  go tool mockery
fi

bash hack/helm-sync-crds.sh

samples_scratch="${scratch}/charts/mkurator/samples/resources"
mkdir -p "$(dirname "${samples_scratch}")"
DEST_DIR="${samples_scratch}" bash hack/sync-samples.sh
# Chart-only files not produced by sync-samples.sh; copy so verify compares like-for-like.
for chart_only in mq-credentials-secret.yaml README.md; do
  if [[ -f "charts/mkurator/samples/resources/${chart_only}" ]]; then
    cp -a "charts/mkurator/samples/resources/${chart_only}" "${samples_scratch}/"
  fi
done

echo "verify: comparing generated artifacts..."

for path in config/crd/bases config/rbac charts/mkurator/crds; do
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

if [[ -d "$scratch/test/mocks" ]] || [[ -d test/mocks ]]; then
  if ! diff -ru "$scratch/test/mocks" test/mocks; then
    echo "verify: drift in test/mocks — run 'task generate'" >&2
    exit 1
  fi
fi

if ! diff -ru "${samples_scratch}" charts/mkurator/samples/resources; then
  echo "verify: drift in charts/mkurator/samples/resources — run 'task samples:sync'" >&2
  exit 1
fi

echo "verify: CRD OpenAPI spec fragments..."
go test ./test/schema/ -count=1

echo "verify: sample YAML contract tests..."
go test ./test/samples/ -count=1

echo "verify: ok"
