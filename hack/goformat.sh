#!/usr/bin/env bash
# Apply gofmt, goimports, and golines to project Go sources.
#
# Generated files (controller-gen deepcopy, mockery mocks) are owned by their
# generators, not by this script: reformatting them makes the next `task verify`
# report codegen drift against the generator output. They are detected by the
# standard "// Code generated ... DO NOT EDIT." header rather than by path, so
# new generators are covered without touching this script.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# The header must precede the package clause, but may sit below a build tag and
# a license block (controller-gen puts it on line 7), so scan the file preamble.
is_generated() {
	sed -n '1,/^package /p' "$1" 2>/dev/null |
		grep -qE '^// Code generated .* DO NOT EDIT\.$'
}

mapfile -t dirs < <(go list -f '{{.Dir}}' ./... 2>/dev/null || true)
if ((${#dirs[@]} == 0)); then
	echo "format: no Go packages"
	exit 0
fi

for dir in "${dirs[@]}"; do
	files=()
	for f in "${dir}"/*.go; do
		[[ -e "$f" ]] || continue
		if is_generated "$f"; then
			continue
		fi
		files+=("$f")
	done
	if ((${#files[@]} == 0)); then
		continue
	fi

	gofmt -w "${files[@]}"
	go tool goimports -local github.com/platformrelay/mkurator -w "${files[@]}"
	for f in "${files[@]}"; do
		# Kubebuilder marker comments on API types must not be re-wrapped.
		if [[ "$f" == */api/v1beta1/*_types.go ]]; then
			continue
		fi
		go tool golines -w --max-len=120 --shorten-comments "$f"
	done
done
