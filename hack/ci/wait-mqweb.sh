#!/usr/bin/env bash
# Wait until mqweb is reachable on the kind NodePort (HAProxy ingress).
set -euo pipefail

HOST="${KURATOR_E2E_MQ_HOST:-mq.localhost}"
PORT="${KURATOR_E2E_MQ_PORT:-30443}"
URL="https://127.0.0.1:${PORT}/ibmmq/console/"
MAX_ATTEMPTS="${MQWEB_WAIT_ATTEMPTS:-60}"
SLEEP_SECS="${MQWEB_WAIT_SLEEP:-10}"

echo "Waiting for mqweb at ${URL} (Host: ${HOST}) …"

for attempt in $(seq 1 "${MAX_ATTEMPTS}"); do
  code="$(curl -sk -o /dev/null -w "%{http_code}" -H "Host: ${HOST}" "${URL}" || true)"
  if [[ "${code}" =~ ^(200|301|302|401|403)$ ]]; then
    echo "mqweb ready (HTTP ${code}) after ${attempt} attempt(s)."
    exit 0
  fi
  echo "attempt ${attempt}/${MAX_ATTEMPTS}: HTTP ${code:-000}; retrying in ${SLEEP_SECS}s …"
  sleep "${SLEEP_SECS}"
done

echo "mqweb did not become ready in time" >&2
exit 1
