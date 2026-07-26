#!/usr/bin/env bash
# REQ-SONAR-MK-05: docs pip install refuses source builds (--only-binary=:all:).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FILE="${ROOT}/.github/workflows/docs.yaml"
if ! grep -E 'pip install' "$FILE" | grep -qF -- '--only-binary=:all:'; then
  echo "FAIL: pip install must include --only-binary=:all:" >&2
  grep -n 'pip install' "$FILE" >&2 || true
  exit 1
fi
echo "OK: docs pip uses --only-binary=:all: (REQ-SONAR-MK-05)"
