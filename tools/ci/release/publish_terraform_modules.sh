#!/usr/bin/env bash
# =============================================================================
# Archive and publish every OpenTofu module of one provider for a catalog
# release, and record what was published.
#
#   publish_terraform_modules.sh <tag> <parts-directory> <provider>
#
# Each module the release ships (`modulebuild list -engine terraform`) is
# archived reproducibly (`modulebuild terraform`) and published once under
# releases/<tag>/modules/terraform/<kind>/module.zip: a tag's artifacts never
# change, so a rerun keeps the published bytes. The sha256 of the bytes each
# key holds goes into <parts-directory>/<provider>.tsv, one
# "<module dir>\t<artifact>\t<sha256>" line per module, which
# publish_module_manifest.sh turns into the engine's manifest. One uploaded
# artifact is then probed on downloads.planton.ai: the CLI and runners compose
# these URLs from a shape pinned in pkg/downloads, and the probe pins the other
# half of that contract.
#
# Runs from the repository root at the tag. Reads R2_BUCKET and the R2
# credentials from the environment, as every lane does.
# =============================================================================
set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: publish_terraform_modules.sh <tag> <parts-directory> <provider>" >&2
  exit 2
fi
tag="$1"
parts="$2"
provider="$3"
release_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
export R2_BUCKET="${R2_BUCKET:-planton-foundation-downloads}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-auto}"
: "${AWS_ENDPOINT_URL:?publishing needs the R2 endpoint in AWS_ENDPOINT_URL}"

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
go build -o "${work}/modulebuild" ./tools/modulebuild
mkdir -p "${parts}"
part="${parts}/${provider}.tsv"
: > "${part}"

modules=()
while IFS= read -r line; do
  [ -n "${line}" ] && modules+=("${line}")
done < <("${work}/modulebuild" list -engine terraform -provider "${provider}")
echo "${provider}: ${#modules[@]} OpenTofu modules"
if [ "${#modules[@]}" -eq 0 ]; then
  echo "WARNING: no OpenTofu modules found for provider ${provider}"
  exit 0
fi

failed=0
last_key=""
for module_dir in "${modules[@]}"; do
  kind="$(echo "${module_dir}" | cut -d/ -f3)"
  key="releases/${tag}/modules/terraform/${kind}/module.zip"
  zip="${work}/${kind}.zip"
  if ! "${work}/modulebuild" terraform -module "${module_dir}" -out "${zip}"; then
    echo "ERROR: failed to archive ${kind}" >&2
    failed=$((failed + 1))
    continue
  fi
  if sha="$(bash "${release_dir}/publish_once.sh" "${zip}" "${key}" --cache-control "public, max-age=31536000, immutable")"; then
    printf '%s\t%s\t%s\n' "${module_dir}" module.zip "${sha}" >> "${part}"
    last_key="${key}"
  else
    echo "ERROR: failed to publish the module of ${kind}" >&2
    failed=$((failed + 1))
  fi
  rm -f "${zip}"
done

if [ -n "${last_key}" ]; then
  url="https://downloads.planton.ai/${last_key}"
  code="$(curl -sf -I -o /dev/null -w "%{http_code}" --retry 3 --retry-delay 5 --retry-all-errors "${url}" || true)"
  if [ "${code}" != "200" ]; then
    echo "ERROR: an uploaded module does not serve at its public URL (HTTP ${code}): ${url}" >&2
    failed=$((failed + 1))
  fi
fi
echo "${provider}: ${#modules[@]} modules, ${failed} failed"
[ "${failed}" -eq 0 ]
