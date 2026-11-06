#!/usr/bin/env bash
# =============================================================================
# Validate a catalog release's InfraCharts against the release's own contract
# and attach the chart catalog to its GitHub Release.
#
#   publish_charts.sh <tag>
#
# The bundle must be publishable by this tag's own contract before it ships:
# the structure guard (including the apiVersion identity every control plane's
# proto constraint enforces at apply) and the chart validator built from this
# tree, never a prebuilt one, since its bulk is the generated contract that
# changes every release. Then infra-charts.zip (`git archive` of the tracked
# charts/ tree: deterministic, build outputs excluded), its sha256 and
# version.txt are attached to the release create_github_release.sh made;
# --clobber keeps a rerun idempotent, and the archive's bytes are the same on
# every run.
#
# Runs from the repository root with HEAD at the tag.
# =============================================================================
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: publish_charts.sh <tag>" >&2
  exit 2
fi
tag="$1"
repo="${GITHUB_REPOSITORY:-plantonhq/planton}"

if [ "$(git rev-parse HEAD)" != "$(git rev-parse "${tag}^{commit}")" ]; then
  echo "publish_charts.sh: HEAD is not ${tag}; the chart catalog is archived from HEAD, so check out the tag first" >&2
  exit 1
fi

bash ./hack/guards/ensure_chart_structure.sh
validator="$(mktemp -d)/chart-validator"
go build -o "${validator}" ./cmd/chart-validator
if find charts -type f -name 'Chart.yaml' -not -path '*/build/*' | grep -q .; then
  "${validator}" --all charts/
else
  echo "no charts in the tree; nothing to validate"
fi

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
git archive --format=zip -o "${work}/infra-charts.zip" HEAD charts/
if command -v sha256sum > /dev/null 2>&1; then
  sha256sum "${work}/infra-charts.zip" | awk '{print $1}' > "${work}/infra-charts.sha256"
else
  shasum -a 256 "${work}/infra-charts.zip" | awk '{print $1}' > "${work}/infra-charts.sha256"
fi
printf '%s\n' "${tag}" > "${work}/version.txt"
echo "built infra-charts.zip ($(wc -c < "${work}/infra-charts.zip" | tr -d ' ') bytes), sha256=$(cat "${work}/infra-charts.sha256")"

gh release upload "${tag}" --repo "${repo}" \
  "${work}/infra-charts.zip" "${work}/infra-charts.sha256" "${work}/version.txt" \
  --clobber
