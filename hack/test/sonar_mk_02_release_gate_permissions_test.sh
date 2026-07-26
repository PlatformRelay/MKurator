#!/usr/bin/env bash
# REQ-SONAR-MK-02 (release-gate): no workflow-level permissions; jobs declare what they need.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FILE="${ROOT}/.github/workflows/release-gate.yaml"

# Strip comments for structural checks.
body="$(grep -vE '^\s*#' "$FILE")"

# Workflow-level permissions sit between the top-level keys and the first `jobs:` block.
pre_jobs="$(printf '%s\n' "$body" | sed -n '1,/^jobs:/p' | sed '$d')"
if printf '%s\n' "$pre_jobs" | grep -qE '^permissions:'; then
  echo "FAIL: workflow-level permissions: still present before jobs:" >&2
  exit 1
fi

# Jobs that checkout or call the Checks API must declare permissions.
for job in resolve-sha verify test integration poll-external-checks; do
  if ! awk -v job="$job" '
    $0 ~ "^  " job ":" {injob=1; next}
    injob && /^  [a-zA-Z0-9_-]+:/ {injob=0}
    injob && /^    permissions:/ {found=1}
    END {exit found ? 0 : 1}
  ' "$FILE"; then
    echo "FAIL: job '$job' missing job-scoped permissions:" >&2
    exit 1
  fi
done

if ! awk '
  /^  poll-external-checks:/ {injob=1; next}
  injob && /^  [a-zA-Z0-9_-]+:/ {injob=0}
  injob && /checks:[[:space:]]*read/ {found=1}
  END {exit found ? 0 : 1}
' "$FILE"; then
  echo "FAIL: poll-external-checks must declare checks: read" >&2
  exit 1
fi

echo "OK: release-gate permissions are job-scoped (REQ-SONAR-MK-02)"
