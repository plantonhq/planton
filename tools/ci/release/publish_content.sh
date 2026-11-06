#!/usr/bin/env bash
# =============================================================================
# Package one content artifact of a catalog release and publish it.
#
#   publish_content.sh <tag> <artifact>
#
# artifact is one of:
#   presets | iac-source | catalog-pages | proto-source | reference-pack |
#   conversion-corpus     a source-file zip (package_content.sh), at
#                         releases/<tag>/content/<artifact>.zip
#   catalog-bundle        the catalog bundle, built and verified
#                         (make build-catalog-bundle / verify-catalog-bundle)
#   catalog-schemas       the JSON schemas, built and verified
#                         (make build-catalog-schemas / verify-catalog-schemas)
#   logos                 every kind's logo at the versionless
#                         catalog/logos/<provider>/<kind>/logo.svg, stable
#                         releases only, then each one proven to serve
#
# Every per-tag artifact goes through publish_once.sh: a tag names one set of
# bytes forever, so a rerun keeps what is already published (a platform pins
# the catalog bundle by its sha256). The logos are the one versionless
# channel, synced over on every stable release by design.
#
# Runs from the repository root at the tag. Reads R2_BUCKET (default
# planton-foundation-downloads) and the R2 credentials from AWS_ACCESS_KEY_ID,
# AWS_SECRET_ACCESS_KEY and AWS_ENDPOINT_URL, as every release lane does.
# =============================================================================
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: publish_content.sh <tag> <artifact>" >&2
  exit 2
fi
tag="$1"
artifact="$2"
release_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export R2_BUCKET="${R2_BUCKET:-planton-foundation-downloads}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-auto}"
: "${AWS_ENDPOINT_URL:?publishing needs the R2 endpoint in AWS_ENDPOINT_URL}"
immutable="public, max-age=31536000, immutable"

publish() {
  local file="$1" key="releases/${tag}/content/$2"
  bash "${release_dir}/publish_once.sh" "${file}" "${key}" --cache-control "${immutable}" > /dev/null
  echo "published: https://downloads.planton.ai/${key}"
}

case "${artifact}" in
  presets | iac-source | catalog-pages | proto-source | reference-pack | conversion-corpus)
    bash "${release_dir}/package_content.sh" "${tag}" "${artifact}"
    publish "${artifact}.zip" "${artifact}.zip"
    ;;
  catalog-bundle)
    make build-catalog-bundle tag="${tag}"
    make verify-catalog-bundle
    publish build/catalog-bundle.zip catalog-bundle.zip
    ;;
  catalog-schemas)
    make build-catalog-schemas
    make verify-catalog-schemas
    publish build/catalog-schemas.zip catalog-schemas.zip
    ;;
  logos)
    if [ "$(bash "${release_dir}/release_channel.sh" "${tag}")" != "stable" ]; then
      echo "${tag} is a release candidate: the versionless logos stay on the last stable release"
      exit 0
    fi
    # Staged as build/logos/<provider>/<kind>/logo.svg, so the sync maps 1:1
    # onto the versionless key shape. The _test refusal and the
    # empty-selection failure mirror package_content.sh's guards.
    rm -rf build/logos
    mkdir -p build/logos
    count=0
    for f in catalog/*/*/logo.svg; do
      case "${f}" in
        catalog/_test/*)
          echo "ERROR: _test provider content must never ship: ${f}" >&2
          exit 1
          ;;
      esac
      dest="build/logos/${f#catalog/}"
      mkdir -p "$(dirname "${dest}")"
      cp "${f}" "${dest}"
      count=$((count + 1))
    done
    if [ "${count}" -eq 0 ]; then
      echo "ERROR: no logos selected -- the selector no longer matches the tree" >&2
      exit 1
    fi
    aws s3 sync build/logos "s3://${R2_BUCKET}/catalog/logos" \
      --content-type image/svg+xml \
      --cache-control "public, max-age=3600" \
      --endpoint-url "${AWS_ENDPOINT_URL}"
    # Each staged key on the public URL, the path users hit, must answer 200
    # with the svg content type: the proof that what was staged is what the
    # world gets.
    fails=0
    for f in build/logos/*/*/logo.svg; do
      url="https://downloads.planton.ai/catalog/logos/${f#build/logos/}"
      read -r code ctype <<< "$(curl -sfI -o /dev/null -w '%{http_code} %{content_type}' \
        --retry 3 --retry-delay 5 --retry-all-errors "${url}" || true)"
      if [ "${code:-}" != "200" ] || [ "${ctype%%;*}" != "image/svg+xml" ]; then
        echo "FAIL: ${url} -> ${code:-} ${ctype:-}" >&2
        fails=$((fails + 1))
      fi
    done
    echo "published ${count} logos (${fails} not serving)"
    [ "${fails}" -eq 0 ]
    ;;
  *)
    echo "publish_content.sh: unknown artifact ${artifact}" >&2
    exit 2
    ;;
esac
