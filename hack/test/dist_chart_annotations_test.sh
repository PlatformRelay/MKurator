#!/usr/bin/env bash
# DIST-AH-01: Chart.yaml carries Artifact Hub operator metadata.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CHART="${ROOT}/charts/mkurator/Chart.yaml"

for key in \
  'artifacthub.io/license: Apache-2.0' \
  "artifacthub.io/operator: 'true'" \
  'artifacthub.io/category: streaming-messaging' \
  'artifacthub.io/operatorCapabilities: Deep Insights'; do
  if ! grep -qF "${key}" "$CHART"; then
    echo "FAIL: missing Chart.yaml annotation: ${key}" >&2
    exit 1
  fi
done

for crd in \
  'kind: QueueManagerConnection' \
  'kind: Queue' \
  'kind: Topic' \
  'kind: Channel' \
  'kind: ChannelAuthRule' \
  'kind: AuthorityRecord'; do
  if ! grep -qF "${crd}" "$CHART"; then
    echo "FAIL: missing CRD in artifacthub.io/crds: ${crd}" >&2
    exit 1
  fi
done

if ! grep -qF 'version: v1beta1' "$CHART"; then
  echo "FAIL: artifacthub.io/crds must list v1beta1 API" >&2
  exit 1
fi

if ! grep -qF 'artifacthub.io/links:' "$CHART"; then
  echo "FAIL: missing artifacthub.io/links" >&2
  exit 1
fi

if ! grep -qF 'artifacthub.io/images:' "$CHART"; then
  echo "FAIL: missing artifacthub.io/images" >&2
  exit 1
fi

echo "OK: Chart.yaml Artifact Hub annotations (DIST-AH-01)"
