#!/usr/bin/env bash
# REQ-SONAR-MK-04: CI gitleaks curl enforces HTTPS.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FILE="${ROOT}/.github/workflows/ci.yaml"
# Extract the Install gitleaks run block (until next step-level key at 2-space indent under job… keep it simple: lines with curl near gitleaks).
block="$(awk '
  /name: Install gitleaks/ {grab=1; next}
  grab && /^      - name:/ {exit}
  grab {print}
' "$FILE")"
if ! printf '%s\n' "$block" | grep -q 'curl'; then
  echo "FAIL: Install gitleaks step has no curl" >&2
  exit 1
fi
if ! printf '%s\n' "$block" | grep -E 'curl' | grep -qE -- "--proto[= ]['\"]?=https"; then
  echo "FAIL: gitleaks curl missing --proto '=https'" >&2
  printf '%s\n' "$block" >&2
  exit 1
fi
if ! printf '%s\n' "$block" | grep -E 'curl' | grep -qE -- '--tlsv1\.2'; then
  echo "FAIL: gitleaks curl missing --tlsv1.2" >&2
  printf '%s\n' "$block" >&2
  exit 1
fi
echo "OK: gitleaks curl enforces HTTPS (REQ-SONAR-MK-04)"
