#!/usr/bin/env bash
# Locks dist_* meta-tests into CI verify job.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CI_WORKFLOW="${ROOT}/.github/workflows/ci.yaml"

fail() { echo "FAIL: $*" >&2; exit 1; }
pass() { echo "ok - $*"; }

[[ -f "${CI_WORKFLOW}" ]] || fail "expected ${CI_WORKFLOW}"

grep -q 'hack/test/dist_\*_test\.sh' "${CI_WORKFLOW}" ||
  fail "ci.yaml must glob hack/test/dist_*_test.sh"

grep -q 'Hub distribution meta-tests' "${CI_WORKFLOW}" ||
  fail "ci.yaml must name the Hub distribution meta-tests step"

pass "ci.yaml globs dist_*_test.sh under Hub distribution meta-tests"
echo "OK: dist CI wiring (DIST-CI)"
