#!/usr/bin/env bash
# Run IBM MQ integration tests (used by task test:integration).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
# shellcheck source=hack/ci/test-artifacts.sh
source "${ROOT}/hack/ci/test-artifacts.sh"
# shellcheck source=hack/ci/suite-lock.sh
source "${ROOT}/hack/ci/suite-lock.sh"

if [[ "${KURATOR_SUITE_LOCK_HELD:-}" != "1" ]]; then
  suite_lock_acquire "${EXCLUSIVE_TEST_LOCK_NAME}"
  trap suite_lock_release EXIT
fi

cd "${ROOT}"

ARTIFACTS_DIR="$(test_artifacts_dir "${ROOT}")"
INTEGRATION_JUNIT="${ARTIFACTS_DIR}/integration-junit.xml"
# Pinned converter for stdlib testing output (integration suite is not Ginkgo).
GO_JUNIT_REPORT_PKG="github.com/jstemmer/go-junit-report/v2@v2.1.0"

set -o pipefail
go test -tags=integration -race -count=1 -json ./test/integration/mq/... |
  go run "${GO_JUNIT_REPORT_PKG}" >"${INTEGRATION_JUNIT}"
