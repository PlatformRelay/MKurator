#!/usr/bin/env bash
# REQ-SONAR-MK-01: drift_resync.go keeps math/rand with gosec nolint + S2245 SAFE rationale.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FILE="${ROOT}/internal/controller/drift_resync.go"
fail=0
require() {
  local pat="$1" msg="$2"
  if ! grep -qE "$pat" "$FILE"; then
    echo "FAIL: $msg" >&2
    fail=1
  fi
}
require 'math/rand' "must import/use math/rand (do not switch to crypto/rand)"
require 'nolint:gosec.*G404' "must keep gosec G404 nolint for math/rand jitter"
require 'S2245' "must document Sonar S2245 accepted SAFE for non-cryptographic jitter"
require 'SAFE|safe' "must state S2245 disposition as SAFE"
if [[ "$fail" -ne 0 ]]; then
  exit 1
fi
echo "OK: REQ-SONAR-MK-01 markers present in drift_resync.go"
