#!/usr/bin/env bash
# REQ-SONAR-MK-01/06 resolve (CI-18): sonar-project.properties suppresses the accepted/test-only
# findings via config-as-code so the SonarCloud dashboard clears on the next analysis.
#   - **/test/** excluded from main-source analysis -> clears 36x go:S4036 (test helpers)
#   - go:S2245 ignored on internal/controller/drift_resync.go -> clears 2x accepted hotspots
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
PROPS="${ROOT}/sonar-project.properties"
fail=0
require() {
  local pat="$1" msg="$2"
  if ! grep -qE "$pat" "$PROPS"; then
    echo "FAIL: $msg" >&2
    fail=1
  fi
}

# S4036: test-support code must not be graded as product source.
require '^sonar\.exclusions=.*\*\*/test/\*\*' "sonar.exclusions must include **/test/** (clears go:S4036)"

# S2245: accepted-safe requeue jitter on drift_resync.go must be ignored via multicriteria.
require '^sonar\.issue\.ignore\.multicriteria=' "must declare sonar.issue.ignore.multicriteria"
require '\.ruleKey=go:S2245' "multicriteria rule must target go:S2245"
require '\.resourceKey=.*drift_resync\.go' "multicriteria rule must scope to drift_resync.go"

[ "$fail" -eq 0 ] || exit 1
echo "OK: REQ-SONAR-MK-01/06 suppression config present in sonar-project.properties"
