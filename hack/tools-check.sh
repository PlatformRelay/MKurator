#!/usr/bin/env bash
# Verify external dev tools by tier (A/B/C). Exit 1 if any required tool is missing.
# Usage: TOOLS_TIER=C bash hack/tools-check.sh
#   TOOLS_TIER: A (default inner loop), B (+ Docker), C (+ kind stack)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TIER="${TOOLS_TIER:-C}"
TIER="$(echo "${TIER}" | tr '[:lower:]' '[:upper:]')"

# Prefer repo-local bin/ (task tools:install) after existing PATH entries.
export PATH="${PATH}:${ROOT}/bin"

missing=0
warn=0

check_cmd() {
  local label="$1"
  local cmd="$2"
  shift 2
  local hint="$*"

  if command -v "${cmd}" >/dev/null 2>&1; then
    local version=""
    case "${cmd}" in
      go) version="$(go version 2>/dev/null | awk '{print $3}')" ;;
      task) version="$(task --version 2>/dev/null | head -1)" ;;
      kind) version="$(kind version 2>/dev/null | head -1)" ;;
      docker) version="$(docker --version 2>/dev/null)" ;;
      *) version="$("${cmd}" version 2>/dev/null | head -1 || true)" ;;
    esac
    printf '  OK  %-12s %s\n' "${label}" "${version}"
  else
    printf '  MISS %-12s (need: %s)\n' "${label}" "${hint}"
    missing=$((missing + 1))
  fi
}

check_go_version() {
  if ! command -v go >/dev/null 2>&1; then
    check_cmd "go" "go" "https://go.dev/dl/ or brew install go"
    return
  fi
  local ver
  ver="$(go version | awk '{print $3}' | sed 's/^go//')"
  printf '  OK  %-12s go%s\n' "go" "${ver}"
  if [[ "${ver}" != 1.26.* ]]; then
    printf '  WARN go           expected 1.26.x (go.mod); got go%s\n' "${ver}"
    warn=$((warn + 1))
  fi
}

echo "Kurator dev tools check (tier ${TIER})"
echo "  docs/LOCAL_SETUP.md · task tools:install · brew bundle (macOS)"
echo ""

echo "Tier A — inner loop:"
check_go_version
check_cmd "task" "task" "brew install go-task/tap/go-task or task tools:install"
echo ""

if [[ "${TIER}" == "B" || "${TIER}" == "C" ]]; then
  echo "Tier B — integration tests:"
  check_cmd "docker" "docker" "brew install --cask docker or apt install docker.io"
  echo ""
fi

if [[ "${TIER}" == "C" ]]; then
  echo "Tier C — local kind stack:"
  check_cmd "kind" "kind" "brew install kind or task tools:install"
  check_cmd "kubectl" "kubectl" "brew install kubectl"
  check_cmd "helm" "helm" "brew install helm"
  check_cmd "terraform" "terraform" "brew install terraform or task tools:install"
  check_cmd "mkcert" "mkcert" "brew install mkcert or task tools:install"
  echo ""
fi

if ((missing > 0)); then
  echo "${missing} required tool(s) missing for tier ${TIER}."
  exit 1
fi

if ((warn > 0)); then
  echo "All required tools present; ${warn} version warning(s) above."
else
  echo "All required tools present for tier ${TIER}."
fi
