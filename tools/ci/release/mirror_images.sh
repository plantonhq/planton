#!/usr/bin/env bash
# tools/ci/release/mirror_images.sh <path:tag> [path:tag ...]
#
# Copies released artifacts from ghcr.io to the Google Artifact Registry mirror,
# by digest. Each argument is a path under ghcr.io/plantonhq, and the mirror
# keeps the same path under its own root:
#
#   planton/operator:v0.21.0        ghcr.io/plantonhq/planton/operator:v0.21.0
#                                -> asia-south1-docker.pkg.dev/plantonhq/planton/operator:v0.21.0
#   charts/planton-runner:0.7.0     ghcr.io/plantonhq/charts/planton-runner:0.7.0 (a Helm chart, an OCI artifact)
#                                -> asia-south1-docker.pkg.dev/plantonhq/charts/planton-runner:0.7.0
#
# crane copies every architecture of a multi-arch image and any OCI artifact
# byte for byte, so a digest pinned from one registry resolves on the other; the
# digests are compared after each copy. The caller has logged crane in to the
# mirror's host; ghcr.io needs no login because the packages are public.
set -euo pipefail

SOURCE_ROOT="ghcr.io/plantonhq"
MIRROR_ROOT="asia-south1-docker.pkg.dev/plantonhq"

if [ "$#" -eq 0 ]; then
  echo "usage: mirror_images.sh <path:tag> [path:tag ...]  (paths under ${SOURCE_ROOT})" >&2
  exit 2
fi

failed=0
for ref in "$@"; do
  src="${SOURCE_ROOT}/${ref}"
  dst="${MIRROR_ROOT}/${ref}"
  if ! crane copy "${src}" "${dst}"; then
    echo "::error::could not copy ${src} to ${dst}; the artifact is on ghcr.io only until this step is re-run" >&2
    failed=1
    continue
  fi
  src_digest="$(crane digest "${src}")"
  dst_digest="$(crane digest "${dst}")"
  if [ "${src_digest}" != "${dst_digest}" ]; then
    echo "::error::${dst} has digest ${dst_digest}, but ${src} has ${src_digest}; the copy is not the same artifact" >&2
    failed=1
    continue
  fi
  echo "mirrored ${ref} ${dst_digest}"
done

exit "${failed}"
