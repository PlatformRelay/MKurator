#!/usr/bin/env bash
# Run the Ginkgo e2e suite with Helm operator deploy (KURATOR_E2E_DEPLOY=helm).
set -euo pipefail

export KURATOR_E2E_DEPLOY=helm

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
exec bash "${ROOT}/hack/ci/run-e2e.sh"
