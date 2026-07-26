#!/usr/bin/env bash
# REQ-SONAR-MK-02 (docs): no workflow-level permissions; build job declares contents:read.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FILE="${ROOT}/.github/workflows/docs.yaml"
body="$(grep -vE '^\s*#' "$FILE")"
pre_jobs="$(printf '%s\n' "$body" | sed -n '1,/^jobs:/p' | sed '$d')"
if printf '%s\n' "$pre_jobs" | grep -qE '^permissions:'; then
  echo "FAIL: workflow-level permissions: still present before jobs:" >&2
  exit 1
fi
if ! awk '
  /^  build:/ {injob=1; next}
  injob && /^  [a-zA-Z0-9_-]+:/ {injob=0}
  injob && /^    permissions:/ {found=1}
  END {exit found ? 0 : 1}
' "$FILE"; then
  echo "FAIL: build job missing job-scoped permissions:" >&2
  exit 1
fi
if ! awk '
  /^  build:/ {injob=1; next}
  injob && /^  [a-zA-Z0-9_-]+:/ {injob=0}
  injob && /contents:[[:space:]]*read/ {found=1}
  END {exit found ? 0 : 1}
' "$FILE"; then
  echo "FAIL: build job must declare contents: read" >&2
  exit 1
fi
echo "OK: docs.yaml permissions are job-scoped (REQ-SONAR-MK-02 docs)"
