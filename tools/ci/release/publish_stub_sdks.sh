#!/usr/bin/env bash
# =============================================================================
# Build a catalog release's prebuilt Java and TypeScript stub trees and
# publish them, sealed.
#
#   publish_stub_sdks.sh <tag>
#
# Publishes under releases/<tag>/stubs/: stubs-java.zip and stubs-ts.zip, a
# .sha256 beside each, and stub-sdks-manifest.json (the release tag, the
# generator pins and the file counts), which the platform's
# fetch_upstream_stubs.py reads and checks before it extracts anything.
#
# The zips are not byte-reproducible (zip stamps each file's time), so a
# rerun could rebuild different bytes. The manifest is therefore the seal,
# uploaded last: a run that finds it published changes nothing. Before it,
# each zip goes through publish_once.sh, each .sha256 is written from the
# bytes its key then holds, and a zip an earlier, unsealed run left with
# other bytes than this build stops the run, because this build's manifest
# would not describe it.
#
# buf is a generation input, so its version must be the lane's pin
# (BUF_VERSION, default 1.65.0). Runs from the repository root at the tag;
# reads R2_BUCKET and the R2 credentials from the environment, as every lane
# does.
# =============================================================================
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: publish_stub_sdks.sh <tag>" >&2
  exit 2
fi
tag="$1"
release_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export R2_BUCKET="${R2_BUCKET:-planton-foundation-downloads}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-auto}"
: "${AWS_ENDPOINT_URL:?publishing needs the R2 endpoint in AWS_ENDPOINT_URL}"
buf_version="${BUF_VERSION:-1.65.0}"
prefix="releases/${tag}/stubs"
immutable="public, max-age=31536000, immutable"

if aws s3api head-object --bucket "${R2_BUCKET}" --endpoint-url "${AWS_ENDPOINT_URL}" \
    --key "${prefix}/stub-sdks-manifest.json" > /dev/null 2>&1; then
  echo "the stub SDKs of ${tag} are published and sealed; nothing to do"
  exit 0
fi
if [ "$(buf --version)" != "${buf_version}" ]; then
  echo "Observed: buf is $(buf --version), but the stub lane pins ${buf_version}." >&2
  echo "Means:    buf generates the stubs, and the manifest records its version for the platform to check." >&2
  echo "Next:     put buf ${buf_version} on PATH and re-run." >&2
  exit 1
fi

make build-stub-sdks tag="${tag}"
out="generated/stub-sdks"
for zip in stubs-java.zip stubs-ts.zip; do
  local_sha="$(cut -d' ' -f1 "${out}/${zip}.sha256")"
  published_sha="$(bash "${release_dir}/publish_once.sh" "${out}/${zip}" "${prefix}/${zip}" --cache-control "${immutable}")"
  if [ "${published_sha}" != "${local_sha}" ]; then
    echo "Observed: ${prefix}/${zip} already holds other bytes (sha256 ${published_sha}) than this build (${local_sha}), and the release is not sealed." >&2
    echo "Means:    an earlier run stopped before the manifest; this build's manifest would not describe what is published." >&2
    echo "Next:     check that the earlier run's tree was this tag, then delete ${prefix}/ from the bucket and re-run." >&2
    exit 1
  fi
  printf '%s  %s\n' "${published_sha}" "${zip}" > "${out}/${zip}.sha256"
  bash "${release_dir}/publish_once.sh" "${out}/${zip}.sha256" "${prefix}/${zip}.sha256" --cache-control "${immutable}" > /dev/null
  echo "published: https://downloads.planton.ai/${prefix}/${zip}"
done
bash "${release_dir}/publish_once.sh" "${out}/stub-sdks-manifest.json" "${prefix}/stub-sdks-manifest.json" \
  --content-type application/json --cache-control "${immutable}" > /dev/null
echo "published and sealed: https://downloads.planton.ai/${prefix}/stub-sdks-manifest.json"
