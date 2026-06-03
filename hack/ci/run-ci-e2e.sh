#!/usr/bin/env bash
# CI-parity e2e workflow (used by task ci:e2e).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=hack/ci/suite-lock.sh
source "${ROOT}/hack/ci/suite-lock.sh"

suite_lock_acquire "${EXCLUSIVE_TEST_LOCK_NAME}"
export KURATOR_SUITE_LOCK_HELD=1
trap suite_lock_release EXIT

export CERT_MANAGER_INSTALL_SKIP="${CERT_MANAGER_INSTALL_SKIP:-true}"
export KURATOR_E2E_MQ="${KURATOR_E2E_MQ:-1}"
export KIND_CLUSTER="${KIND_CLUSTER:-${CLUSTER_NAME:-mkurator}}"
export KUBECONFIG="${KUBECONFIG:-${ROOT}/hack/kind-cluster/.state/kubeconfig.yaml}"

cd "${ROOT}"

bash hack/ci/step.sh "PLATFORM UP — kind cluster + IBM MQ (task cluster:up)"
task cluster:up
bash hack/ci/step.sh "PLATFORM UP — wait for mqweb (NodePort / HAProxy)"
bash hack/ci/wait-mqweb.sh
bash hack/ci/step.sh "GINKGO E2E — image build, deploy operator, MQ scenarios (kustomize)"
task test:e2e

# Optional second pass on the same cluster (saves a second platform spin on main).
# Enable with KURATOR_CI_E2E_BOTH=1 (e.g. local full parity).
if [[ "${KURATOR_CI_E2E_BOTH:-0}" == "1" ]]; then
  bash hack/ci/step.sh "GINKGO E2E — Helm deploy path on existing cluster"
  KURATOR_E2E_DEPLOY=helm task test:e2e
fi
