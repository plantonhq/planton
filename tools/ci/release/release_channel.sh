#!/usr/bin/env bash
# =============================================================================
# Say which channel a catalog release tag is on: `stable` or `candidate`.
#
#   release_channel.sh <tag>
#
# A vX.Y.Z-rc.N tag is a release candidate: it publishes every per-tag
# artifact a platform pins and moves no shared channel (the GitHub release is a
# pre-release that is never "latest", the BSR `main` label, the versionless
# logos and the definitions pointers all stay on the last stable release). Any
# other tag release.yaml triggers on (strict vX.Y.Z) is stable. Every lane that
# treats the two differently asks this one script, in a workflow and on a
# laptop alike.
# =============================================================================
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: release_channel.sh <tag>" >&2
  exit 2
fi
if [[ "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+-rc\.[0-9]+$ ]]; then
  echo candidate
elif [[ "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo stable
else
  echo "release_channel.sh: ${1} is not a catalog release tag (vX.Y.Z or vX.Y.Z-rc.N)" >&2
  exit 2
fi
