#!/usr/bin/env bash
# Sign GitHub Release artifacts with cosign keyless (Sigstore OIDC).
# Writes Sigstore bundles as *.sigstore.json — OSSF Scorecard Signed-Releases
# matches that extension (not the default .bundle from cosign sign-blob).
#
# Usage: hack/sign-release-assets.sh <dist-dir>
set -euo pipefail

DIST="${1:?dist directory required (e.g. dist)}"
cd "${DIST}"

shopt -s nullglob
files=(install-crds.yaml install.yaml mkurator-*.tgz sbom*.json checksums.txt)

signed=0
for f in "${files[@]}"; do
  [[ -f "${f}" ]] || continue
  echo "Signing ${f}"
  cosign sign-blob --yes --bundle "${f}.sigstore.json" "${f}"
  signed=$((signed + 1))
done

if [[ "${signed}" -eq 0 ]]; then
  echo "no release artifacts found to sign in ${DIST}" >&2
  exit 1
fi

echo "signed ${signed} release artifact(s) in ${DIST}/"
