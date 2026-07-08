#!/usr/bin/env bash
# Assert the Helm chart renders validating admission webhook resources aligned with
# config/webhook/manifests.yaml (kubebuilder/controller-gen output).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

CHART="./charts/mkurator"
NAMESPACE="mkurator-system"
RELEASE="mkurator"

rendered="$(mktemp)"
trap 'rm -f "${rendered}"' EXIT

helm template "${RELEASE}" "${CHART}" \
  --namespace "${NAMESPACE}" \
  --set webhooks.enabled=true \
  --set webhooks.certManager.create=true \
  > "${rendered}"

require_kind() {
  local kind="$1"
  if ! grep -q "^kind: ${kind}$" "${rendered}"; then
    echo "helm-verify-admission: missing kind ${kind} in helm template output" >&2
    exit 1
  fi
}

require_kind "ValidatingWebhookConfiguration"
require_kind "Service"
require_kind "Certificate"
require_kind "Issuer"

webhook_count="$(grep -c 'path: /validate-messaging-mkurator-dev-v1' "${rendered}" || true)"
if [[ "${webhook_count}" -ne 12 ]]; then
  echo "helm-verify-admission: expected 12 validating webhook paths, found ${webhook_count}" >&2
  exit 1
fi

for path in \
  /validate-messaging-mkurator-dev-v1alpha1-authorityrecord \
  /validate-messaging-mkurator-dev-v1alpha1-channel \
  /validate-messaging-mkurator-dev-v1alpha1-channelauthrule \
  /validate-messaging-mkurator-dev-v1alpha1-queue \
  /validate-messaging-mkurator-dev-v1alpha1-queuemanagerconnection \
  /validate-messaging-mkurator-dev-v1alpha1-topic \
  /validate-messaging-mkurator-dev-v1beta1-authorityrecord \
  /validate-messaging-mkurator-dev-v1beta1-channel \
  /validate-messaging-mkurator-dev-v1beta1-channelauthrule \
  /validate-messaging-mkurator-dev-v1beta1-queue \
  /validate-messaging-mkurator-dev-v1beta1-queuemanagerconnection \
  /validate-messaging-mkurator-dev-v1beta1-topic
do
  if ! grep -q "path: ${path}" "${rendered}"; then
    echo "helm-verify-admission: missing webhook path ${path}" >&2
    exit 1
  fi
done

for name in \
  mkurator-validating-webhook-configuration \
  mkurator-webhook-service \
  mkurator-serving-cert \
  mkurator-selfsigned-issuer
do
  if ! grep -q "name: ${name}" "${rendered}"; then
    echo "helm-verify-admission: missing resource name ${name}" >&2
    exit 1
  fi
done

if ! grep -Eq 'cert-manager.io/inject-ca-from: "?mkurator-system/mkurator-serving-cert"?' "${rendered}"; then
  echo "helm-verify-admission: missing cert-manager CA injection annotation" >&2
  exit 1
fi

# Keep Helm webhook paths/names in sync with kubebuilder output.
manifests="${ROOT}/config/webhook/manifests.yaml"
for webhook_name in \
  vauthorityrecord.kb.io vchannel.kb.io vchannelauthrule.kb.io vqueue.kb.io vqueuemanagerconnection.kb.io vtopic.kb.io \
  vauthorityrecord-v1beta1.kb.io vchannel-v1beta1.kb.io vchannelauthrule-v1beta1.kb.io vqueue-v1beta1.kb.io \
  vqueuemanagerconnection-v1beta1.kb.io vtopic-v1beta1.kb.io
do
  if ! grep -q "name: ${webhook_name}" "${manifests}"; then
    echo "helm-verify-admission: ${webhook_name} missing from ${manifests}" >&2
    exit 1
  fi
  if ! grep -q "name: ${webhook_name}" "${rendered}"; then
    echo "helm-verify-admission: ${webhook_name} missing from Helm template output" >&2
    exit 1
  fi
done

disabled="$(mktemp)"
trap 'rm -f "${rendered}" "${disabled}"' EXIT
helm template "${RELEASE}" "${CHART}" \
  --namespace "${NAMESPACE}" \
  --set webhooks.enabled=false \
  > "${disabled}"

if grep -q "^kind: ValidatingWebhookConfiguration$" "${disabled}"; then
  echo "helm-verify-admission: ValidatingWebhookConfiguration rendered with webhooks.enabled=false" >&2
  exit 1
fi

echo "helm-verify-admission: ok"
