#!/usr/bin/env bash
# Shared test report output directory (gitignored under repo root).
#
# Usage:
#   source hack/ci/test-artifacts.sh
#   dir="$(test_artifacts_dir "${ROOT}")"
test_artifacts_dir() {
  local root="${1:?root directory required}"
  local dir="${KURATOR_TEST_ARTIFACTS_DIR:-${root}/artifacts}"
  mkdir -p "${dir}"
  printf '%s\n' "${dir}"
}
