#!/usr/bin/env bash
# Sync kustomize-built CRDs (conversion webhook + cert-manager annotations) into the Helm chart.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="${ROOT}/config/crd"
DST="${ROOT}/charts/mkurator/crds"

if [[ ! -f "${SRC}/kustomization.yaml" ]]; then
  echo "missing ${SRC}/kustomization.yaml; run: task manifests" >&2
  exit 1
fi

mkdir -p "${DST}"
rm -f "${DST}"/*.yaml

bundle="$(mktemp)"
trap 'rm -f "${bundle}"' EXIT
go tool kustomize build "${SRC}" > "${bundle}"

python3 - "${bundle}" "${DST}" <<'PY'
import pathlib
import sys

import yaml

bundle = pathlib.Path(sys.argv[1])
dst = pathlib.Path(sys.argv[2])

for doc in yaml.safe_load_all(bundle.read_text()):
    if not doc:
        continue
    crd_name = doc["metadata"]["name"]
    plural, group = crd_name.split(".", 1)
    out = dst / f"{group}_{plural}.yaml"
    with out.open("w", encoding="utf-8") as handle:
        yaml.dump(doc, handle, default_flow_style=False, sort_keys=False)
        handle.write("\n")

print(f"synced CRDs to {dst}")
PY
