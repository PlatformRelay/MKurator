#!/usr/bin/env bash
# Run the Ginkgo e2e suite with verbose progress (used by task test:e2e).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=hack/ci/step.sh
source "${ROOT}/hack/ci/step.sh"
# shellcheck source=hack/ci/suite-lock.sh
source "${ROOT}/hack/ci/suite-lock.sh"

if [[ "${KURATOR_SUITE_LOCK_HELD:-}" != "1" ]]; then
  suite_lock_acquire "${EXCLUSIVE_TEST_LOCK_NAME}"
  trap suite_lock_release EXIT
fi

cd "${ROOT}"

# Prefer the kind cluster kubeconfig when present so kubectl never falls back to
# a stale default context (e.g. localhost:8080) during BeforeSuite cert-manager install.
KIND_KUBECONFIG="${ROOT}/hack/kind-cluster/.state/kubeconfig.yaml"
if [[ -f "${KIND_KUBECONFIG}" ]]; then
  export KUBECONFIG="${KIND_KUBECONFIG}"
fi

GINKGO_NODES="${KURATOR_E2E_NODES:-3}"

ci_step "GINKGO E2E — image build, kind load, deploy (platform must already be up; task ci:e2e runs PLATFORM UP first)"

echo "KUBECONFIG=${KUBECONFIG:-<unset>}"
echo "KIND_CLUSTER=${KIND_CLUSTER:-<unset>}"
echo "KURATOR_E2E_DEPLOY=${KURATOR_E2E_DEPLOY:-kustomize}"
echo "KURATOR_E2E_MQ=${KURATOR_E2E_MQ:-<unset>}"
echo "KURATOR_E2E_NODES=${GINKGO_NODES}"
echo "KURATOR_E2E_LABEL_FILTER=${KURATOR_E2E_LABEL_FILTER:-<unset>}"
echo "CERT_MANAGER_INSTALL_SKIP=${CERT_MANAGER_INSTALL_SKIP:-<unset>}"
echo "KURATOR_E2E_VERBOSE_LOGS=${KURATOR_E2E_VERBOSE_LOGS:-0}"
echo ""
echo "Note: go test uses -race (CGO_ENABLED=1). With parallel nodes, use fewer KURATOR_E2E_NODES on small hosts if flaky."
echo ""

export CGO_ENABLED="${CGO_ENABLED:-1}"

GINKGO_FLAGS=(
  -ginkgo.vv
  -ginkgo.show-node-events
  -ginkgo.procs="${GINKGO_NODES}"
)
if [[ -n "${KURATOR_E2E_LABEL_FILTER:-}" ]]; then
  GINKGO_FLAGS+=(-ginkgo.label-filter="${KURATOR_E2E_LABEL_FILTER}")
fi
if [[ "${GITHUB_ACTIONS:-}" == "true" ]]; then
  GINKGO_FLAGS+=(-ginkgo.github-output)
fi

ci_step "GINKGO SUITE — look for [e2e] SPEC START/PASS lines and ==> stage banners"

# -count=1: do not skip the suite via cached pass
go test -tags=e2e ./test/e2e/... \
  -race \
  -v \
  -count=1 \
  -timeout=90m \
  "${GINKGO_FLAGS[@]}"
