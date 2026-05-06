#!/usr/bin/env bash
# Rejects commits that would leave Go files unformatted, mirroring the
# CI gofmt gate. Designed to be called from .githooks/pre-commit (with
# the default "staged" mode, i.e. only files about to be committed) and
# from anywhere else with "all" to scan the entire tree.
#
# Usage:
#   check-go-fmt.sh             # checks staged .go files only
#   check-go-fmt.sh staged      # same as above
#   check-go-fmt.sh all         # checks every .go file in the repo
#
# Exit 0: nothing needs formatting (or no .go files in scope).
# Exit 1: at least one file would change under gofmt.
# Exit 2: usage error.
set -euo pipefail

mode="${1:-staged}"

files=()
case "$mode" in
	staged)
		# --diff-filter=ACMR drops deletions. mapfile would be cleaner
		# here but is bash 4+, and macOS ships /bin/bash at 3.2; the
		# while-read form below is the portable equivalent.
		while IFS= read -r f; do
			[[ -n "$f" ]] && files+=("$f")
		done < <(git diff --cached --name-only --diff-filter=ACMR -- '*.go')
		;;
	all)
		while IFS= read -r f; do
			[[ -n "$f" ]] && files+=("$f")
		done < <(git ls-files -- '*.go')
		;;
	*)
		echo "usage: $(basename "$0") [staged|all]" >&2
		exit 2
		;;
esac

if [[ ${#files[@]} -eq 0 ]]; then
	exit 0
fi

# gofmt -l prints one path per file that would be changed; empty
# output means everything is already formatted.
unformatted="$(gofmt -l "${files[@]}")"
if [[ -z "$unformatted" ]]; then
	exit 0
fi

cat >&2 <<EOF
gofmt would reformat the following file(s):

${unformatted}

Fix locally with:
  gofmt -w ${unformatted//$'\n'/ }
  git add -u
  # then re-run your commit

CI runs the same check, so leaving these unformatted will block the PR.
EOF

exit 1
