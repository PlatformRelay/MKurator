#!/usr/bin/env bash
# DIST-OLM-02: operatorhub-pr script and release workflow gating.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="${ROOT}/hack/operatorhub-pr.sh"
WORKFLOW="${ROOT}/.github/workflows/release.yaml"

if [[ ! -x "${SCRIPT}" ]]; then
  echo "FAIL: hack/operatorhub-pr.sh must be executable" >&2
  exit 1
fi

if ! "${SCRIPT}" --help 2>&1 | grep -qF 'operatorhub-pr.sh'; then
  echo "FAIL: operatorhub-pr.sh --help must print usage" >&2
  exit 1
fi

FAKE_DIGEST='sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef'
if ! DRY_RUN=1 VERSION=9.9.9-test IMAGE_DIGEST="${FAKE_DIGEST}" "${SCRIPT}" 2>&1 | grep -qF 'Bundle verified'; then
  echo "FAIL: DRY_RUN=1 must generate and verify bundle without clone/push" >&2
  exit 1
fi

if ! grep -qF 'operatorhub-pr:' "${WORKFLOW}"; then
  echo "FAIL: release.yaml missing operatorhub-pr job" >&2
  exit 1
fi

if ! grep -qF 'OPERATORHUB_PAT' "${WORKFLOW}"; then
  echo "FAIL: operatorhub-pr job must gate on OPERATORHUB_PAT" >&2
  exit 1
fi

if ! grep -qF 'continue-on-error: true' "${WORKFLOW}"; then
  echo "FAIL: operatorhub-pr step must soft-fail (continue-on-error)" >&2
  exit 1
fi

if ! grep -qF 'hack/operatorhub-pr.sh' "${WORKFLOW}"; then
  echo "FAIL: operatorhub-pr job must invoke hack/operatorhub-pr.sh" >&2
  exit 1
fi

echo "OK: operatorhub-pr script and workflow (DIST-OLM-02)"
