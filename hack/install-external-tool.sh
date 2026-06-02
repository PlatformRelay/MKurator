#!/usr/bin/env bash
# Download a pinned external tool binary.
# Usage: hack/install-external-tool.sh <tool> <version> <output-path>
# Tools: kind, mkcert, task, terraform
# Example: hack/install-external-tool.sh kind v0.27.0 bin/kind
set -euo pipefail

TOOL="${1:?tool required (kind|mkcert|task|terraform)}"
VERSION="${2:?version required}"
OUT="${3:?output path required}"

case "$(uname -m)" in
  x86_64) arch=amd64 ;;
  aarch64 | arm64) arch=arm64 ;;
  *)
    echo "unsupported architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *)
    echo "unsupported OS: $(uname -s)" >&2
    exit 1
    ;;
esac

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out_path="${root}/${OUT}"
mkdir -p "$(dirname "${out_path}")"
tmpdir="$(mktemp -d)"
trap 'rm -rf "${tmpdir}"' EXIT

ver="${VERSION#v}"

case "${TOOL}" in
  kind)
    url="https://kind.sigs.k8s.io/dl/${VERSION}/kind-${os}-${arch}"
    curl -fsSL "${url}" -o "${tmpdir}/bin"
    install -m 0755 "${tmpdir}/bin" "${out_path}"
    ;;
  mkcert)
    url="https://github.com/FiloSottile/mkcert/releases/download/${VERSION}/mkcert-${VERSION}-${os}-${arch}"
    curl -fsSL "${url}" -o "${tmpdir}/bin"
    install -m 0755 "${tmpdir}/bin" "${out_path}"
    ;;
  task)
    asset="task_${os}_${arch}.tar.gz"
    url="https://github.com/go-task/task/releases/download/v${ver}/${asset}"
    curl -fsSL "${url}" -o "${tmpdir}/task.tgz"
    tar -xzf "${tmpdir}/task.tgz" -C "${tmpdir}"
    install -m 0755 "${tmpdir}/task" "${out_path}"
    ;;
  terraform)
    zip="terraform_${ver}_${os}_${arch}.zip"
    url="https://releases.hashicorp.com/terraform/${ver}/${zip}"
    curl -fsSL "${url}" -o "${tmpdir}/terraform.zip"
    unzip -q "${tmpdir}/terraform.zip" -d "${tmpdir}"
    install -m 0755 "${tmpdir}/terraform" "${out_path}"
    ;;
  *)
    echo "unsupported tool: ${TOOL} (expected kind, mkcert, task, terraform)" >&2
    exit 1
    ;;
esac

echo "installed ${OUT} (${TOOL} ${VERSION} ${os}/${arch})"
