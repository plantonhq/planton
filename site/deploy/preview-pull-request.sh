#!/usr/bin/env bash
# Name the open pull request a preview.site run built, from GitHub's own record of it and never
# from anything the run uploaded: a fork's pull request can rewrite preview.site.yaml, so what
# its build says about itself is the contributor's word.
#
#   preview-pull-request.sh <head repository> <head branch> <head commit>
#
# Prints the pull request's number when an open pull request in this repository has exactly that
# head repository, branch and commit. Prints nothing when none does (a newer push moved the
# branch on, or the pull request closed): a newer run, or the close, owns the preview then.
#
# Needs GH_TOKEN with pull-requests: read, and GITHUB_REPOSITORY.
set -euo pipefail

usage='usage: preview-pull-request.sh <head repository> <head branch> <head commit>'
repo="${1:?$usage}"
branch="${2:?$usage}"
sha="${3:?$usage}"
# Both go into the jq filter below; the branch, the one name a contributor picks, only ever travels as a parameter.
[[ "$repo" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] || { echo "not a repository: $repo" >&2; exit 2; }
[[ "$sha" =~ ^[0-9a-f]{40}$ ]] || { echo "not a commit: $sha" >&2; exit 2; }

gh api -X GET "repos/${GITHUB_REPOSITORY}/pulls" --paginate \
  -f state=open -f head="${repo%%/*}:${branch}" -F per_page=100 \
  --jq ".[] | select(.head.sha == \"${sha}\" and .head.repo.full_name == \"${repo}\") | .number" | head -n 1
