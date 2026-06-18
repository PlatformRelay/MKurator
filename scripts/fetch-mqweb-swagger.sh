#!/usr/bin/env bash
# Fetch IBM MQ mqweb Swagger 2.0 JSON from a running queue manager (GET /ibm/api/docs).
# Requires apiDiscovery-1.0 enabled in mqwebuser.xml — see docs/schemas/README.md.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'EOF'
Usage: fetch-mqweb-swagger.sh [OPTIONS] <mqweb-base-url> [output-file]

Fetch the mqweb OpenAPI (Swagger 2.0) discovery document from GET /ibm/api/docs.

Arguments:
  mqweb-base-url   Base URL, e.g. https://localhost:9443
  output-file      Destination path (default: docs/schemas/mqweb-swagger.json)

Environment:
  MQWEB_USER       Basic auth username (default: admin)
  MQWEB_PASS       Basic auth password (default: passw0rd — Docker integration MQ)
  MQWEB_INSECURE   Set to 1 to skip TLS verification (dev/local only)

Examples:
  MQWEB_USER=admin MQWEB_PASS=changeme \
    ./scripts/fetch-mqweb-swagger.sh https://localhost:9443

  task mq:integration:up && task mq:integration:wait && task mq:swagger:fetch
EOF
}

MQWEB_USER="${MQWEB_USER:-admin}"
MQWEB_PASS="${MQWEB_PASS:-passw0rd}"
MQWEB_INSECURE="${MQWEB_INSECURE:-1}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h | --help)
      usage
      exit 0
      ;;
    --)
      shift
      break
      ;;
    -*)
      echo "unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
    *)
      break
      ;;
  esac
done

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage >&2
  exit 2
fi

BASE_URL="${1%/}"
OUT="${2:-${ROOT}/docs/schemas/mqweb-swagger.json}"
DOCS_URL="${BASE_URL}/ibm/api/docs"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required" >&2
  exit 1
fi

curl_args=(--fail --silent --show-error --location)
if [[ "${MQWEB_INSECURE}" == "1" ]]; then
  curl_args+=(-k)
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT
raw="${tmpdir}/swagger.json"

echo "Fetching ${DOCS_URL} → ${OUT}"

http_code="$(
  curl "${curl_args[@]}" \
    -u "${MQWEB_USER}:${MQWEB_PASS}" \
    -H "Accept: application/json" \
    -o "${raw}" \
    -w "%{http_code}" \
    "${DOCS_URL}" || true
)"

if [[ "${http_code}" != "200" ]]; then
  echo "fetch failed: HTTP ${http_code} from ${DOCS_URL}" >&2
  echo "Ensure apiDiscovery-1.0 is enabled in mqwebuser.xml and mqweb is running." >&2
  exit 1
fi

if ! grep -q '"swagger"[[:space:]]*:[[:space:]]*"2.0"' "${raw}"; then
  echo "response from ${DOCS_URL} does not look like Swagger 2.0 JSON" >&2
  exit 1
fi

mkdir -p "$(dirname "${OUT}")"
if command -v python3 >/dev/null 2>&1; then
  python3 -m json.tool "${raw}" >"${OUT}"
else
  cp "${raw}" "${OUT}"
fi

bytes="$(wc -c <"${OUT}" | tr -d ' ')"
echo "Wrote ${bytes} bytes to ${OUT}"
