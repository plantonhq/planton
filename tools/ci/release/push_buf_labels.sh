#!/usr/bin/env bash
# =============================================================================
# Push the proto module to the BSR under the release's labels.
#
#   push_buf_labels.sh <tag>
#
# The tag label is the consumed contract: a platform's pinned buf dependency
# (buf.build/planton/planton:<tag>). A stable release also moves `main`, so
# the BSR's default view tracks releases; a release candidate leaves it.
# Pushing the same commit under the same labels again is a no-op on the BSR.
#
# Authenticates as buf does: BUF_TOKEN in a workflow, the signed-in account on
# a laptop.
# =============================================================================
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: push_buf_labels.sh <tag>" >&2
  exit 2
fi
tag="$1"
channel="$(bash "$(dirname "${BASH_SOURCE[0]}")/release_channel.sh" "${tag}")"

if [ "${channel}" = "candidate" ]; then
  buf push --label "${tag}"
  echo "pushed buf.build/planton/planton with label ${tag} (release candidate; main unchanged)"
else
  buf push --label "${tag}" --label main
  echo "pushed buf.build/planton/planton with labels ${tag} + main"
fi
