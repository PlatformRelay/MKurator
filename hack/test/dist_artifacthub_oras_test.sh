#!/usr/bin/env bash
# DIST-AH-02: release workflow pushes Artifact Hub repository metadata via oras.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FILE="${ROOT}/.github/workflows/release.yaml"

if ! grep -qF 'oras-project/setup-oras@1d808f7d7f6995cc68b7bf507bfe5c5446e1dc9d' "$FILE"; then
  echo "FAIL: release.yaml must pin oras-project/setup-oras by SHA" >&2
  exit 1
fi

if ! grep -qF 'artifacthub-repo.yml:application/vnd.cncf.artifacthub.repository-metadata.layer.v1.yaml' "$FILE"; then
  echo "FAIL: release.yaml missing oras push layer for artifacthub-repo.yml" >&2
  exit 1
fi

if ! grep -qF ':artifacthub.io' "$FILE"; then
  echo "FAIL: release.yaml missing oras push tag :artifacthub.io" >&2
  exit 1
fi

if ! grep -qF 'oras push' "$FILE"; then
  echo "FAIL: release.yaml missing oras push step" >&2
  exit 1
fi

if ! grep -qF 'artifact-metadata: write' "$FILE"; then
  echo "FAIL: release job needs artifact-metadata: write for oras push" >&2
  exit 1
fi

echo "OK: release.yaml Artifact Hub oras push wiring (DIST-AH-02)"
