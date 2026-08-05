#!/usr/bin/env bash
# DIST-OLM-02: operatorhub-pr script and release workflow gating.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${ROOT}"
SCRIPT="${ROOT}/hack/operatorhub-pr.sh"
WORKFLOW="${ROOT}/.github/workflows/release.yaml"

if [[ ! -x "${SCRIPT}" ]]; then
  echo "FAIL: hack/operatorhub-pr.sh must be executable" >&2
  exit 1
fi

HELP_OUT="$(mktemp)"
if ! "${SCRIPT}" --help >"${HELP_OUT}" 2>&1; then
  echo "FAIL: operatorhub-pr.sh --help exited non-zero" >&2
  exit 1
fi
if ! grep -qF 'operatorhub-pr.sh' "${HELP_OUT}"; then
  echo "FAIL: operatorhub-pr.sh --help must print usage" >&2
  exit 1
fi
rm -f "${HELP_OUT}"

FAKE_DIGEST='sha256:deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef'
DRY_OUT="$(mktemp)"
trap 'rm -f "${DRY_OUT}"' EXIT
if ! DRY_RUN=1 VERSION=9.9.9-test IMAGE_DIGEST="${FAKE_DIGEST}" "${SCRIPT}" >"${DRY_OUT}" 2>&1; then
  echo "FAIL: DRY_RUN=1 operatorhub-pr.sh exited non-zero" >&2
  cat "${DRY_OUT}" >&2
  exit 1
fi
if ! grep -qF 'Bundle verified' "${DRY_OUT}"; then
  echo "FAIL: DRY_RUN=1 must generate and verify bundle without clone/push" >&2
  cat "${DRY_OUT}" >&2
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
