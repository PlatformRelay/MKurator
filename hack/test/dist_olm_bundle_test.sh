#!/usr/bin/env bash
# DIST-OLM-01: generate-olm-bundle produces a registry+v1 bundle with digest-pinned CSV.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${ROOT}"

FAKE_DIGEST='sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef'
VERSION='9.9.9-test'
BUNDLE_DIR="dist/olm-bundle/${VERSION}"

rm -rf "${BUNDLE_DIR}"
IMAGE_DIGEST="${FAKE_DIGEST}" VERSION="${VERSION}" make generate-olm-bundle

CSV="${BUNDLE_DIR}/manifests/mkurator.clusterserviceversion.yaml"
ANNOTATIONS="${BUNDLE_DIR}/metadata/annotations.yaml"

for f in "${CSV}" "${ANNOTATIONS}"; do
  if [[ ! -f "${f}" ]]; then
    echo "FAIL: missing bundle file: ${f}" >&2
    exit 1
  fi
done

for crd in \
  messaging.mkurator.dev_authorityrecords.yaml \
  messaging.mkurator.dev_channelauthrules.yaml \
  messaging.mkurator.dev_channels.yaml \
  messaging.mkurator.dev_queuemanagerconnections.yaml \
  messaging.mkurator.dev_queues.yaml \
  messaging.mkurator.dev_topics.yaml; do
  if [[ ! -f "${BUNDLE_DIR}/manifests/${crd}" ]]; then
    echo "FAIL: missing CRD in bundle: ${crd}" >&2
    exit 1
  fi
done

if ! grep -qF 'operators.operatorframework.io.bundle.package.v1: mkurator' "${ANNOTATIONS}"; then
  echo "FAIL: bundle package name must be mkurator" >&2
  exit 1
fi

if ! grep -qF "ghcr.io/platformrelay/mkurator@${FAKE_DIGEST}" "${CSV}"; then
  echo "FAIL: CSV must reference digest-pinned controller image" >&2
  exit 1
fi

echo "OK: generate-olm-bundle output (DIST-OLM-01)"
