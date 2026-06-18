#!/usr/bin/env bash
# Assert Helm chart CRDs serve v1alpha1 + v1beta1 with conversion webhook wiring,
# and match the kustomize-built CRD bundle from config/crd.
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
# Reuse helm-sync-crds source: kustomize build output is checked in verify.sh;
# here we only compare helm files to a fresh kustomize build via subprocess-free read.
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
    conversion = doc.get("spec", {}).get("conversion", {})
    if conversion.get("strategy") != "Webhook":
        raise SystemExit(f"helm-verify-crds: {name} conversion.strategy must be Webhook")
    webhook = conversion.get("webhook", {})
    service = webhook.get("clientConfig", {}).get("service", {})
    if service.get("name") != "mkurator-webhook-service":
        raise SystemExit(f"helm-verify-crds: {name} missing mkurator-webhook-service")
    if service.get("path") != "/convert":
        raise SystemExit(f"helm-verify-crds: {name} conversion webhook path must be /convert")

    annotations = doc.get("metadata", {}).get("annotations", {})
    if annotations.get("cert-manager.io/inject-ca-from") != "mkurator-system/mkurator-serving-cert":
        raise SystemExit(f"helm-verify-crds: {name} missing cert-manager CA injection annotation")

    versions = {v["name"]: v for v in doc["spec"]["versions"]}
    for version in ("v1alpha1", "v1beta1"):
        if version not in versions:
            raise SystemExit(f"helm-verify-crds: {name} missing version {version}")
        if not versions[version].get("served"):
            raise SystemExit(f"helm-verify-crds: {name} must serve {version}")

    alpha = versions["v1alpha1"]
    beta = versions["v1beta1"]
    if not alpha.get("storage"):
        raise SystemExit(f"helm-verify-crds: {name} must store v1alpha1")
    if beta.get("storage"):
        raise SystemExit(f"helm-verify-crds: {name} must not store v1beta1 yet")

    if name not in built_docs:
        raise SystemExit(f"helm-verify-crds: {name} missing from config/crd kustomize build")
    if helm_docs[name] != built_docs[name]:
        raise SystemExit(
            f"helm-verify-crds: {name} drift from config/crd — run 'task helm:sync-crds'"
        )

print("helm-verify-crds: ok")
PY
