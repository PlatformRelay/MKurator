#!/usr/bin/env bash
# Run the Ginkgo e2e suite with verbose progress (used by task test:e2e).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=hack/ci/step.sh
source "${ROOT}/hack/ci/step.sh"

cd "${ROOT}"

ci_step "E2E tests (build image, load kind, deploy operator — output streams below)"

echo "KUBECONFIG=${KUBECONFIG:-<unset>}"
echo "KIND_CLUSTER=${KIND_CLUSTER:-<unset>}"
echo "KURATOR_E2E_MQ=${KURATOR_E2E_MQ:-<unset>}"
echo "CERT_MANAGER_INSTALL_SKIP=${CERT_MANAGER_INSTALL_SKIP:-<unset>}"
echo ""

export CGO_ENABLED="${CGO_ENABLED:-1}"

# -ginkgo.v / -ginkgo.progress: spec and long-running step visibility
# -count=1: do not skip the suite via cached pass
go test -tags=e2e ./test/e2e/... \
  -race \
  -v \
  -count=1 \
  -timeout=90m \
  -ginkgo.v \
  -ginkgo.progress \
  -ginkgo.show-node-events
