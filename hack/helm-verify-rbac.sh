#!/usr/bin/env bash
# Assert the Helm chart renders manager ClusterRole rules aligned with
# config/rbac/role.yaml (kubebuilder/controller-gen output) for messaging.mkurator.dev.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

CHART="./charts/mkurator"
NAMESPACE="mkurator-system"
RELEASE="mkurator"
ROLE_YAML="${ROOT}/config/rbac/role.yaml"

rendered="$(mktemp)"
normalized_helm="$(mktemp)"
normalized_kubebuilder="$(mktemp)"
trap 'rm -f "${rendered}" "${normalized_helm}" "${normalized_kubebuilder}"' EXIT

helm template "${RELEASE}" "${CHART}" \
  --namespace "${NAMESPACE}" \
  > "${rendered}"

if ! grep -q "name: ${RELEASE}-manager" "${rendered}"; then
  echo "helm-verify-rbac: missing manager ClusterRole (${RELEASE}-manager) in helm template output" >&2
  exit 1
fi

# Normalize Helm inline rules: "resources|verbs" per messaging.mkurator.dev block.
extract_helm_messaging_rules() {
  awk '
    /apiGroups: \[messaging.mkurator.dev\]/ { capture=1; resources=""; verbs=""; next }
    capture && /resources:/ {
      line=$0
      sub(/^.*resources: \[/, "", line)
      sub(/\].*$/, "", line)
      gsub(/, /, ",", line)
      resources=line
      next
    }
    capture && /verbs:/ {
      line=$0
      sub(/^.*verbs: \[/, "", line)
      sub(/\].*$/, "", line)
      gsub(/, /, ",", line)
      print resources "|" line
      capture=0
    }
  ' "${rendered}" | sort
}

# Normalize kubebuilder role.yaml rules: "resources|verbs" per messaging.mkurator.dev block.
extract_kubebuilder_messaging_rules() {
  awk '
    /- messaging.mkurator.dev/ {
      in_messaging=1
      resources=""
      verbs=""
      phase=""
      next
    }
    !in_messaging { next }
    /^  resources:$/ { phase="resources"; next }
    /^  verbs:$/ { phase="verbs"; next }
    /^- apiGroups:/ {
      if (resources != "" && verbs != "") {
        print resources "|" verbs
      }
      in_messaging=0
      next
    }
    phase == "resources" && /^  - / {
      item=$2
      resources=(resources == "" ? item : resources "," item)
      next
    }
    phase == "verbs" && /^  - / {
      item=$2
      verbs=(verbs == "" ? item : verbs "," item)
      next
    }
    END {
      if (in_messaging && resources != "" && verbs != "") {
        print resources "|" verbs
      }
    }
  ' "${ROLE_YAML}" | sort
}

extract_helm_messaging_rules > "${normalized_helm}"
extract_kubebuilder_messaging_rules > "${normalized_kubebuilder}"

if ! diff -u "${normalized_kubebuilder}" "${normalized_helm}"; then
  echo "helm-verify-rbac: manager ClusterRole messaging.mkurator.dev rules differ from ${ROLE_YAML}" >&2
  exit 1
fi

# Spot-check Phase 5 auth resources appear in both sources (not only queuemanagerconnections drift).
for resource in \
  authorityrecords \
  channelauthrules \
  channels \
  queues \
  topics \
  authorityrecords/finalizers \
  channelauthrules/finalizers \
  authorityrecords/status \
  channelauthrules/status
do
  if ! grep -q "${resource}" "${ROLE_YAML}"; then
    echo "helm-verify-rbac: ${resource} missing from ${ROLE_YAML}" >&2
    exit 1
  fi
  if ! grep -q "${resource}" "${rendered}"; then
    echo "helm-verify-rbac: ${resource} missing from Helm template output" >&2
    exit 1
  fi
done

disabled="$(mktemp)"
trap 'rm -f "${rendered}" "${normalized_helm}" "${normalized_kubebuilder}" "${disabled}"' EXIT
helm template "${RELEASE}" "${CHART}" \
  --namespace "${NAMESPACE}" \
  --set rbac.create=false \
  > "${disabled}"

if grep -q "^kind: ClusterRole$" "${disabled}"; then
  echo "helm-verify-rbac: ClusterRole rendered with rbac.create=false" >&2
  exit 1
fi

echo "helm-verify-rbac: ok"
