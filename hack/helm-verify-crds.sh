#!/usr/bin/env bash
# Assert Helm chart CRDs are single-version (serve + store only the v1beta1 hub),
# carry NO conversion webhook stanza, and match config/crd (8e-8a: v1alpha1 is
# no longer served).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

python3 - "${ROOT}" <<'PY'
import pathlib
import sys

import yaml

root = pathlib.Path(sys.argv[1])
crd_dir = root / "charts" / "mkurator" / "crds"
expected = {
    "authorityrecords.messaging.mkurator.dev",
    "channelauthrules.messaging.mkurator.dev",
    "channels.messaging.mkurator.dev",
    "queuemanagerconnections.messaging.mkurator.dev",
    "queues.messaging.mkurator.dev",
    "topics.messaging.mkurator.dev",
}


def load_crds(path: pathlib.Path) -> dict[str, dict]:
    docs: dict[str, dict] = {}
    files = [path] if path.is_file() else sorted(path.glob("*.yaml"))
    for file in files:
        for doc in yaml.safe_load_all(file.read_text(encoding="utf-8")):
            if doc and doc.get("kind") == "CustomResourceDefinition":
                docs[doc["metadata"]["name"]] = doc
    return docs


kustomize_bundle = root / "config" / "crd"
import subprocess

built = subprocess.check_output(
    ["go", "tool", "kustomize", "build", str(kustomize_bundle)],
    text=True,
)
built_docs = {}
for doc in yaml.safe_load_all(built):
    if doc and doc.get("kind") == "CustomResourceDefinition":
        built_docs[doc["metadata"]["name"]] = doc

helm_docs = load_crds(crd_dir)

missing = expected - set(helm_docs)
if missing:
    raise SystemExit(f"helm-verify-crds: missing CRDs in chart: {sorted(missing)}")

for name in sorted(expected):
    doc = helm_docs[name]

    # 8e-8a: no conversion webhook — a dangling conversion stanza would fail
    # every CR write on admission.
    if doc.get("spec", {}).get("conversion"):
        raise SystemExit(
            f"helm-verify-crds: {name} must NOT carry a conversion stanza (v1alpha1 dropped)"
        )

    annotations = doc.get("metadata", {}).get("annotations", {})
    if "cert-manager.io/inject-ca-from" in annotations:
        raise SystemExit(
            f"helm-verify-crds: {name} must NOT carry the conversion cert-manager CA injection annotation"
        )

    versions = {v["name"]: v for v in doc["spec"]["versions"]}
    if set(versions) != {"v1beta1"}:
        raise SystemExit(
            f"helm-verify-crds: {name} must be single-version v1beta1, got {sorted(versions)}"
        )
    beta = versions["v1beta1"]
    if not beta.get("served"):
        raise SystemExit(f"helm-verify-crds: {name} must serve v1beta1")
    if not beta.get("storage"):
        raise SystemExit(f"helm-verify-crds: {name} must store v1beta1")

    if name not in built_docs:
        raise SystemExit(f"helm-verify-crds: {name} missing from config/crd kustomize build")
    if helm_docs[name] != built_docs[name]:
        raise SystemExit(
            f"helm-verify-crds: {name} drift from config/crd — run 'task helm:sync-crds'"
        )

print("helm-verify-crds: ok")
PY
