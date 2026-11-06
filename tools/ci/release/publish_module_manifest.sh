#!/usr/bin/env bash
# =============================================================================
# Assemble one engine's module manifest for a catalog release and publish it.
#
#   publish_module_manifest.sh <tag> <engine> <parts-directory>
#
# engine is terraform or pulumi. The manifest names every module of that
# engine with its published artifacts' sha256 and its module digest, read from
# the parts the engine's lane recorded (`modulebuild manifest` searches the
# directory's subdirectories too, so a workflow's per-leg artifact folders and
# a laptop's flat <provider>.tsv files both read). modulebuild refuses a
# partial manifest, so a manifest is published only once every module of the
# engine is: a platform pins a release once the manifest of the engine it runs
# by default exists, and a runner runs a module only when its bytes match.
# publish_once.sh keeps a manifest already published, and the probe proves it
# serves at its public URL.
#
# Runs from the repository root at the tag. Reads R2_BUCKET and the R2
# credentials from the environment, as every lane does.
# =============================================================================
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: publish_module_manifest.sh <tag> <engine> <parts-directory>" >&2
  exit 2
fi
tag="$1"
engine="$2"
parts="$3"
release_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export R2_BUCKET="${R2_BUCKET:-planton-foundation-downloads}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-auto}"
: "${AWS_ENDPOINT_URL:?publishing needs the R2 endpoint in AWS_ENDPOINT_URL}"
case "${engine}" in
  terraform) noun="OpenTofu" ;;
  pulumi) noun="Pulumi" ;;
  *) echo "publish_module_manifest.sh: the engine is terraform or pulumi (got ${engine})" >&2; exit 2 ;;
esac

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
go run ./tools/modulebuild manifest -engine "${engine}" -tag "${tag}" -parts "${parts}" -out "${work}/manifest.json"

key="releases/${tag}/modules/${engine}/manifest.json"
bash "${release_dir}/publish_once.sh" "${work}/manifest.json" "${key}" \
  --content-type application/json --cache-control "public, max-age=31536000, immutable" > /dev/null
url="https://downloads.planton.ai/${key}"
code="$(curl -sf -I -o /dev/null -w "%{http_code}" --retry 3 --retry-delay 5 --retry-all-errors "${url}" || true)"
if [ "${code}" != "200" ]; then
  echo "ERROR: the ${noun} module manifest does not serve at its public URL (HTTP ${code}): ${url}" >&2
  exit 1
fi
echo "published the ${noun} module manifest: ${url}"
