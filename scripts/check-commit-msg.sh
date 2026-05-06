#!/usr/bin/env bash
# Validates a commit message header against Conventional Commits 1.0.0.
# https://www.conventionalcommits.org/en/v1.0.0/
#
# Usage: check-commit-msg.sh <path-to-message-file>
#
# Exit 0: header is valid or is a kind we deliberately skip
#         (merge / revert / fixup / squash / amend autosquash messages).
# Exit 1: header is invalid; a diagnostic is printed to stderr.
set -euo pipefail

if [[ $# -ne 1 ]]; then
	echo "usage: $(basename "$0") <message-file>" >&2
	exit 2
fi

msg_file="$1"
if [[ ! -r "$msg_file" ]]; then
	echo "check-commit-msg: cannot read $msg_file" >&2
	exit 2
fi

header="$(grep -v '^#' "$msg_file" | sed -n '1p')"

case "$header" in
	"Merge "*|"Revert \""*|"fixup! "*|"squash! "*|"amend! "*)
		exit 0
		;;
esac

pattern='^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([a-z0-9._/-]+\))?!?: .+'

if [[ "$header" =~ $pattern ]]; then
	exit 0
fi

cat >&2 <<EOF
commit message header does not match Conventional Commits 1.0.0.

  got:      ${header}
  expected: <type>(<scope>)!: <subject>

  type:    build | chore | ci | docs | feat | fix | perf | refactor | revert | style | test
  scope:   optional, lower-case (e.g. (ui), (config))
  !:       optional, marks a breaking change
  subject: required, non-empty

example:
  feat(ui): add per-PR merge-blocker column

See https://www.conventionalcommits.org/en/v1.0.0/
EOF

exit 1
