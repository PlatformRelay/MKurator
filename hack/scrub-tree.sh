#!/usr/bin/env bash
# Fail if ANY tracked file contains forbidden employer or legacy-project strings.
# Tree-wide CI net (preflight); hack/scrub.sh stays the fast staged-files
# pre-commit net. Prints offending filenames only — never matched content.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# Fail closed: a missing or empty patterns file must redden CI, not silently
# disable the gate (the patterns file is excluded from its own scan below).
if [ ! -s hack/scrub-patterns.txt ]; then
  echo "scrub-tree: hack/scrub-patterns.txt missing or empty" >&2
  exit 1
fi

# git ls-files limits the scan to tracked files; only the patterns file itself
# is excluded (it legitimately contains the forbidden strings — the scrub
# scripts are scanned like any other file). NUL-delimited to survive spaces
# and other exotic filenames.
files=()
while IFS= read -r -d '' f; do
  [ "$f" = "hack/scrub-patterns.txt" ] && continue
  files+=("$f")
done < <(git ls-files -z)

# Fail closed: an empty path list makes `rg`/`grep` search the cwd (or stdin),
# which can miss tracked leaks or scan untracked junk. Never treat "no files"
# as a clean tree.
if [ "${#files[@]}" -eq 0 ]; then
  echo "scrub-tree: no tracked files to scan (unexpected empty git ls-files)" >&2
  exit 1
fi

set +e
if command -v rg >/dev/null 2>&1; then
  hits="$(rg -il -f hack/scrub-patterns.txt -- "${files[@]}")"
else
  # ubuntu-latest runners lack ripgrep; grep -iEl preserves the
  # case-insensitive list-filenames-only semantics of rg -il.
  hits="$(grep -iEl -f hack/scrub-patterns.txt -- "${files[@]}")"
fi
status=$?
set -e

# Scanner exit codes: 0 = matches found (fail below), 1 = no matches (pass);
# anything else is a scanner error — fail closed, never report a clean tree.
if [ "$status" -gt 1 ]; then
  echo "scrub-tree: scanner error (exit $status)" >&2
  exit "$status"
fi

if [ -n "${hits}" ]; then
  echo "scrub-tree: forbidden strings found in tracked files:" >&2
  echo "${hits}" >&2
  exit 1
fi

echo "scrub-tree: ok (${#files[@]} tracked files clean)"
