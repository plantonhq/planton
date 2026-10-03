#!/usr/bin/env bash
# Write the website preview's one comment on a pull request: edit the comment that
# carries the marker <!-- website-preview --> if there is one, otherwise create it,
# so a pull request shows one line about its preview however often it is pushed.
#
#   sticky-comment.sh <pull request number> <body, starting with the marker>
#
# Needs GH_TOKEN with pull-requests: write, and GITHUB_REPOSITORY.
set -euo pipefail

number="${1:?usage: sticky-comment.sh <pull request number> <body>}"
body="${2:?usage: sticky-comment.sh <pull request number> <body>}"
marker='<!-- website-preview -->'
[[ "$body" == "$marker"* ]] || { echo "the body must start with ${marker}" >&2; exit 2; }

existing="$(gh api "repos/${GITHUB_REPOSITORY}/issues/${number}/comments" --paginate \
  --jq ".[] | select(.body | startswith(\"${marker}\")) | .id" | head -n 1)"

if [[ -n "$existing" ]]; then
  gh api -X PATCH "repos/${GITHUB_REPOSITORY}/issues/comments/${existing}" -f body="$body" >/dev/null
else
  gh api -X POST "repos/${GITHUB_REPOSITORY}/issues/${number}/comments" -f body="$body" >/dev/null
fi
