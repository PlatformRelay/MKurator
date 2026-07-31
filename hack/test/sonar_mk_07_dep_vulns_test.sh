#!/usr/bin/env bash
# REQ-SEC-VULN (CI-17): go.mod pins the fixed versions of the two actionable OSV findings.
#   - go.opentelemetry.io/otel >= v1.44.0 (GO-2026-5158)
#   - github.com/google/cel-go  >= v0.29.0 (GHSA-gcjh-h69q-9w9g, CVSS 6.3)
# The unfixable x/crypto/openpgp advisory (GO-2026-5932) stays an osv-scanner.toml IgnoredVuln.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
GOMOD="${ROOT}/go.mod"
fail=0

# semver_ge A B  -> exit 0 if A >= B (major.minor.patch, no pre-release handling needed here)
semver_ge() {
  [ "$1" = "$2" ] && return 0
  [ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | tail -n1)" = "$1" ]
}

check() {
  local module="$1" min="$2" id="$3"
  local have
  have="$(grep -E "^\s*${module//./\\.} v[0-9]" "$GOMOD" | head -n1 | awk '{print $2}' | sed 's/^v//')"
  if [ -z "$have" ]; then
    echo "FAIL: ${module} not found in go.mod ($id)" >&2
    fail=1
    return
  fi
  if semver_ge "$have" "${min#v}"; then
    echo "OK: ${module} v${have} >= ${min} ($id)"
  else
    echo "FAIL: ${module} v${have} < ${min} — bump to clear ${id}" >&2
    fail=1
  fi
}

check "go.opentelemetry.io/otel" "1.44.0" "GO-2026-5158"
check "github.com/google/cel-go" "0.29.0" "GHSA-gcjh-h69q-9w9g"

[ "$fail" -eq 0 ] || exit 1
echo "OK: REQ-SEC-VULN dependency pins present"
