#!/usr/bin/env bash
# REQ-SONAR-MK-03: every curl in post-install.sh enforces HTTPS; no get-helm-3 | bash.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FILE="${ROOT}/.devcontainer/post-install.sh"
fail=0

if grep -qE 'get-helm-3' "$FILE"; then
  echo "FAIL: get-helm-3 pipe-to-bash must be replaced with a pinned helm asset download" >&2
  fail=1
fi

# Every curl line must include proto=https and tlsv1.2.
while IFS= read -r line; do
  if [[ "$line" =~ curl ]]; then
    if [[ ! "$line" =~ --proto[[:space:]]*=?[\'\"]?=https ]] && [[ ! "$line" =~ --proto[[:space:]]+\'=https\' ]] && [[ ! "$line" =~ --proto\'=https\' ]]; then
      # Accept both --proto '=https' and --proto "=https"
      if ! printf '%s\n' "$line" | grep -qE -- "--proto[= ]['\"]?=https"; then
        echo "FAIL: curl missing --proto '=https': $line" >&2
        fail=1
      fi
    fi
    if ! printf '%s\n' "$line" | grep -qE -- '--tlsv1\.2'; then
      echo "FAIL: curl missing --tlsv1.2: $line" >&2
      fail=1
    fi
  fi
done < <(grep -E 'curl' "$FILE" || true)

if ! grep -qE 'get\.helm\.sh/' "$FILE" || ! grep -qE 'HELM_VERSION=|"helm-\$\{HELM_VERSION\}|helm-v[0-9]' "$FILE"; then
  echo "FAIL: expected pinned helm tarball download from get.helm.sh" >&2
  fail=1
fi

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi
echo "OK: REQ-SONAR-MK-03 HTTPS curls + pinned helm"
