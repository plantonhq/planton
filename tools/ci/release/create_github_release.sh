#!/usr/bin/env bash
# =============================================================================
# Create the GitHub Release for a catalog tag, once.
#
#   create_github_release.sh <tag>
#
# The release object is what the chart catalog's assets attach to and what
# /releases/latest/download resolves to. A release that already exists is
# kept (a rerun changes nothing); a release candidate is a pre-release that is
# never "latest", so chart and module consumers never resolve to it.
#
# Reads GITHUB_REPOSITORY (default plantonhq/planton) and authenticates as gh
# does: GH_TOKEN in a workflow, the signed-in account on a laptop.
# =============================================================================
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: create_github_release.sh <tag>" >&2
  exit 2
fi
tag="$1"
repo="${GITHUB_REPOSITORY:-plantonhq/planton}"
channel="$(bash "$(dirname "${BASH_SOURCE[0]}")/release_channel.sh" "${tag}")"

if gh release view "${tag}" --repo "${repo}" > /dev/null 2>&1; then
  echo "Release ${tag} already exists"
  exit 0
fi
if [ "${channel}" = "candidate" ]; then
  gh release create "${tag}" --repo "${repo}" --title "${tag}" --generate-notes --prerelease --latest=false
else
  gh release create "${tag}" --repo "${repo}" --title "${tag}" --generate-notes
fi
