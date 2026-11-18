#!/usr/bin/env bash
# =============================================================================
# Publish a release artifact under its tag exactly once, and say which bytes
# the tag holds.
#
# A tag names exactly one set of bytes for every artifact under it, forever:
# platforms that deploy a module record "this resource ran release vX.Y.Z" and
# trust the tag to mean those exact bytes, so an artifact that changed under
# its tag would let a resource read as unchanged while different code is what
# deploys. The first write under a key wins; a later write to the same key --
# a rerun of a failed lane, or a release re-dispatched by hand for an old tag --
# keeps what is already published and says so, which keeps reruns working
# without ever replacing a published artifact.
#
# Every upload records the bytes' sha256 as the object's metadata, and the
# script prints the sha256 of the bytes the key holds on stdout -- the uploaded
# file's, or the kept object's. Callers that describe a release (the module
# manifest) record that value, so what they describe is always what is
# published, even when a rerun rebuilt different bytes; a rebuild that differs
# from the kept object is reported as a warning, because the published bytes
# stand. Everything else the script says goes to stderr.
#
# Usage:
#   publish_once.sh <local-file> <object-key> [extra `aws s3 cp` arguments...]
#
# Reads R2_BUCKET and AWS_ENDPOINT_URL from the environment, like the
# workflows' own aws calls. Exits non-zero only when an upload fails.
# =============================================================================
set -euo pipefail

# R2 takes "auto" as its region; the runner's default (us-west-2) is refused on
# every multipart upload, which a large module needs.
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-auto}"

file="$1"
key="$2"
shift 2

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}

local_sha="$(sha256_of "${file}")"

if published_sha="$(aws s3api head-object --bucket "${R2_BUCKET}" --key "${key}" \
    --endpoint-url "${AWS_ENDPOINT_URL}" --query 'Metadata.sha256' --output text 2>/dev/null)"; then
  if [ -z "${published_sha}" ] || [ "${published_sha}" = "None" ]; then
    # Published before uploads recorded their checksum: hash the published
    # bytes themselves rather than trust the rebuild.
    kept="$(mktemp)"
    aws s3 cp "s3://${R2_BUCKET}/${key}" "${kept}" --endpoint-url "${AWS_ENDPOINT_URL}" >&2
    published_sha="$(sha256_of "${kept}")"
    rm -f "${kept}"
  fi
  echo "= Kept: https://downloads.planton.ai/${key} (already published under this tag; a tag's artifacts never change)" >&2
  if [ "${published_sha}" != "${local_sha}" ]; then
    echo "::warning::${key}: this rebuild (${local_sha}) differs from the published bytes (${published_sha}); the published bytes stand" >&2
  fi
  echo "${published_sha}"
  exit 0
fi

# R2 answers a burst of parallel uploads with a passing InternalError now and
# then (v0.6.22's logos, v0.6.23's Pulumi modules); one such answer used to
# fail the whole lane. A failed upload writes nothing under the key, so trying
# again is safe; three tries with a growing pause outlast the blip, and a real
# outage still fails here, loudly.
for attempt in 1 2 3; do
  if aws s3 cp "${file}" "s3://${R2_BUCKET}/${key}" --endpoint-url "${AWS_ENDPOINT_URL}" \
      --metadata "sha256=${local_sha}" "$@" >&2; then
    echo "✓ Uploaded: https://downloads.planton.ai/${key}" >&2
    echo "${local_sha}"
    exit 0
  fi
  if [ "${attempt}" -lt 3 ]; then
    echo "::warning::${key}: upload attempt ${attempt} of 3 failed; trying again in $((attempt * 15))s" >&2
    sleep $((attempt * 15))
  fi
done
exit 1
