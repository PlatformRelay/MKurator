#!/usr/bin/env bash
# REQ-SEC-2026-08: production and fixture Dockerfiles pin base images by digest
# (OpenSSF Scorecard Pinned-Dependencies).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fail=0

require_digest() {
  local file="$1" image="$2"
  if ! grep -E "^FROM ${image}[^ ]*@sha256:[a-f0-9]{64}" "$file" >/dev/null; then
    echo "FAIL: ${file} must pin ${image} with @sha256:<64 hex>" >&2
    grep -n "^FROM " "$file" >&2 || true
    fail=1
  else
    echo "OK: ${file} pins ${image} by digest"
  fi
}

require_digest "${ROOT}/Dockerfile" "golang:"
require_digest "${ROOT}/Dockerfile" "gcr.io/distroless/static:"
require_digest "${ROOT}/test/e2e/fixtures/metrics-curl/Dockerfile" "alpine:"

# ARG expansion must be quoted so a malicious/empty TARGETOS cannot word-split.
if ! grep -E 'GOOS="\$\{TARGETOS:-linux\}"' "${ROOT}/Dockerfile" >/dev/null; then
  echo "FAIL: Dockerfile must quote GOOS=\${TARGETOS:-linux} (docker:S6570)" >&2
  fail=1
fi
if ! grep -E 'GOARCH="\$\{TARGETARCH\}"' "${ROOT}/Dockerfile" >/dev/null; then
  echo "FAIL: Dockerfile must quote GOARCH=\${TARGETARCH} (docker:S6570)" >&2
  fail=1
fi

[ "$fail" -eq 0 ] || exit 1
echo "OK: REQ-SEC-2026-08 Dockerfile digest pins present"
