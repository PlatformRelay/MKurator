#!/usr/bin/env bash
# REQ-SEC-2026-08: docs lockfiles pin fixed versions of Scorecard OSV findings.
#   - js-yaml >= 5.2.2 (GHSA-724g-mxrg-4qvm, GHSA-pm4m-ph32-ghv5)
#   - click >= 8.3.3 (PYSEC-2026-2132)
#   - pymdown-extensions >= 11.0.1 (PYSEC-2026-3609 path traversal; PYSEC-2026-3654 ReDoS)
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fail=0

semver_ge() {
  [ "$1" = "$2" ] && return 0
  [ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | tail -n1)" = "$1" ]
}

js_yaml_ver="$(python3 -c '
import json,sys
lock=json.load(open(sys.argv[1]))
print(lock["packages"]["node_modules/js-yaml"]["version"])
' "${ROOT}/docs/package-lock.json")"
if semver_ge "$js_yaml_ver" "5.2.2"; then
  echo "OK: js-yaml ${js_yaml_ver} >= 5.2.2 (GHSA-724g-mxrg-4qvm / GHSA-pm4m-ph32-ghv5)"
else
  echo "FAIL: js-yaml ${js_yaml_ver} < 5.2.2 — bump docs lockfile" >&2
  fail=1
fi

click_ver="$(awk '/^click==/{sub(/^click==/,""); print $1; exit}' "${ROOT}/docs/requirements-docs.txt")"
if semver_ge "$click_ver" "8.3.3"; then
  echo "OK: click ${click_ver} >= 8.3.3 (PYSEC-2026-2132)"
else
  echo "FAIL: click ${click_ver} < 8.3.3 — bump docs/requirements-docs.txt" >&2
  fail=1
fi

pymdown_ver="$(awk '/^pymdown-extensions==/{sub(/^pymdown-extensions==/,""); print $1; exit}' "${ROOT}/docs/requirements-docs.txt")"
if semver_ge "$pymdown_ver" "11.0.1"; then
  echo "OK: pymdown-extensions ${pymdown_ver} >= 11.0.1 (PYSEC-2026-3609 / PYSEC-2026-3654)"
else
  echo "FAIL: pymdown-extensions ${pymdown_ver} < 11.0.1 — bump docs/requirements-docs.txt" >&2
  fail=1
fi

[ "$fail" -eq 0 ] || exit 1
echo "OK: REQ-SEC-2026-08 docs OSV pins present"
