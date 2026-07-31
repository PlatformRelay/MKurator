#!/usr/bin/env bash
# REQ-SEC-PIPELINE (CI-19): the govulncheck workflow gates PRs/pushes, not just the weekly cron,
# so the dependency-vuln scan actually protects merges ("the pipeline doesn't run [on PRs]").
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WF="${ROOT}/.github/workflows/vulncheck.yaml"
fail=0

# Isolate the top-level `on:` block (up to the next top-level key) and assert the triggers.
on_block="$(awk '/^on:/{f=1;next} /^[a-zA-Z]/{if(f)exit} f' "$WF")"
check() {
  local pat="$1" msg="$2"
  if ! grep -qE "$pat" <<<"$on_block"; then
    echo "FAIL: $msg" >&2
    fail=1
  fi
}
check '^\s*pull_request:' "vulncheck.yaml on: must include pull_request"
check '^\s*push:' "vulncheck.yaml on: must include push"

[ "$fail" -eq 0 ] || exit 1
echo "OK: REQ-SEC-PIPELINE vulncheck.yaml triggers on push + pull_request"
