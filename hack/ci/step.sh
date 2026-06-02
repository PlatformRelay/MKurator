#!/usr/bin/env bash
# Print a visible phase banner for CI/local Task workflows.
#
# Usage:
#   source hack/ci/step.sh && ci_step "message"
#   bash hack/ci/step.sh "message"
ci_step() {
  echo ""
  echo "==> $(date -Is) $*"
  echo ""
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  set -euo pipefail
  if [[ $# -lt 1 ]]; then
    echo "usage: $0 <message>" >&2
    exit 2
  fi
  ci_step "$*"
fi
