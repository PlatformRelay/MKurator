#!/usr/bin/env bash
# Sync Kubebuilder sample CRs into the Helm chart samples tree.
# Canonical source: config/samples/messaging_v1alpha1_*.yaml
# Chart output: charts/mkurator/samples/resources/ (no metadata.namespace; kustomization sets it)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SRC="${ROOT}/config/samples"
DST="${DEST_DIR:-${ROOT}/charts/mkurator/samples/resources}"

declare -A MAP=(
  [messaging_v1alpha1_queuemanagerconnection.yaml]=queuemanagerconnection.yaml
  [messaging_v1alpha1_queue.yaml]=queue.yaml
  [messaging_v1alpha1_queue_alias.yaml]=queue-alias.yaml
  [messaging_v1alpha1_queue_remote.yaml]=queue-remote.yaml
  [messaging_v1alpha1_topic.yaml]=topic.yaml
  [messaging_v1alpha1_channel.yaml]=channel.yaml
  [messaging_v1alpha1_channelauthrule.yaml]=channelauthrule.yaml
  [messaging_v1alpha1_channelauthrule_blockuser.yaml]=channelauthrule-blockuser.yaml
  [messaging_v1alpha1_channelauthrule_blockaddr.yaml]=channelauthrule-blockaddr.yaml
  [messaging_v1alpha1_authorityrecord.yaml]=authorityrecord.yaml
)

mkdir -p "${DST}"

strip_namespace() {
  awk '!/^[[:space:]]+namespace:[[:space:]]/ { print }'
}

for src_name in "${!MAP[@]}"; do
  src_path="${SRC}/${src_name}"
  if [[ ! -f "${src_path}" ]]; then
    echo "sync-samples: missing ${src_path}" >&2
    exit 1
  fi
  strip_namespace < "${src_path}" > "${DST}/${MAP[${src_name}]}"
done

# mq-credentials-secret.yaml is chart-only (not in config/samples kustomization).
KUSTOMIZE_RESOURCES=(
  mq-credentials-secret.yaml
  queuemanagerconnection.yaml
  queue.yaml
  queue-alias.yaml
  queue-remote.yaml
  topic.yaml
  channel.yaml
  channelauthrule.yaml
  channelauthrule-blockuser.yaml
  channelauthrule-blockaddr.yaml
  authorityrecord.yaml
)

{
  echo "apiVersion: kustomize.config.k8s.io/v1beta1"
  echo "kind: Kustomization"
  echo ""
  echo "namespace: mkurator-system"
  echo ""
  echo "resources:"
  for f in "${KUSTOMIZE_RESOURCES[@]}"; do
    echo "  - ${f}"
  done
} > "${DST}/kustomization.yaml"

echo "synced samples to ${DST}"
