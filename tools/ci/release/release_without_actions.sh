#!/usr/bin/env bash
# =============================================================================
# Run a pushed release tag's lanes from this machine, for when GitHub Actions
# cannot run them (front door: make release-without-actions tag=<tag>).
#
#   release_without_actions.sh <tag>
#
# A release is still a tag cut by its `make release-*` front door; this is the
# CI that tag would have triggered, run here. It clones the tag into a clean
# scratch tree, so what it publishes is the tag's tree and the tag's own lane
# scripts, never this checkout's working copy, and it runs the same
# tools/ci/release scripts the workflows run, in their order:
#
#   vX.Y.Z, vX.Y.Z-rc.N (release.yaml), the lanes a platform pin reads:
#     1. the GitHub Release            create_github_release.sh
#     2. the InfraChart catalog        publish_charts.sh
#     3. the BSR labels                push_buf_labels.sh
#     4. the content artifacts         publish_content.sh (six zips, the
#                                      catalog bundle and schemas, the logos)
#     5. the Java and TypeScript stubs publish_stub_sdks.sh
#     6. the OpenTofu modules          publish_terraform_modules.sh, one
#                                      provider at a time in parallel
#     7. the OpenTofu module manifest  publish_module_manifest.sh
#
# Every lane is rerun-safe: per-tag artifacts are published once
# (publish_once.sh), and the GitHub Release, the BSR push and the chart
# assets are idempotent. So a failed run is re-run as it is.
#
# Not run here, and left to the tag's own release run once Actions can run it
# (re-run that run; the lanes above then keep what this published): the
# Pulumi modules and their manifest (about 4,200 Go builds; no pin reads
# them, and Pulumi jobs on the release are refused until they exist), the
# definitions and their stable pointer, the plantonhq/skills distribution,
# and the Stigmer skills.
#
# Needs the R2 key pair in AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY and the
# endpoint in AWS_ENDPOINT_URL, gh and buf signed in, Go, git, curl, jq, zip,
# python3 and the aws CLI.
# =============================================================================
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: release_without_actions.sh <tag>   (vX.Y.Z or vX.Y.Z-rc.N)" >&2
  exit 2
fi
tag="$1"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
: "${AWS_ACCESS_KEY_ID:?the R2 key pair goes in AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY}"
: "${AWS_SECRET_ACCESS_KEY:?the R2 key pair goes in AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY}"
: "${AWS_ENDPOINT_URL:?the R2 endpoint goes in AWS_ENDPOINT_URL}"
export R2_BUCKET="${R2_BUCKET:-planton-foundation-downloads}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-auto}"

# The tag must be on the remote: a release is what the tag names there.
if ! git -C "${repo_root}" ls-remote --exit-code --tags origin "refs/tags/${tag}" > /dev/null; then
  echo "Observed: origin has no tag ${tag}." >&2
  echo "Means:    a release is a pushed tag; this runs the lanes that tag triggers." >&2
  echo "Next:     cut it with its front door (make release), then re-run." >&2
  exit 1
fi

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
git -C "${repo_root}" fetch --quiet origin "refs/tags/${tag}:refs/tags/${tag}"
git -c advice.detachedHead=false clone --quiet --branch "${tag}" "${repo_root}" "${work}/src"
src="${work}/src"
# The pinned proto toolchain is versioned and stamp-checked by make, so the
# tag's tree reuses this checkout's copy instead of downloading it again.
if [ -d "${repo_root}/.tools" ]; then
  ln -s "${repo_root}/.tools" "${src}/.tools"
fi
cd "${src}"
lanes="${src}/tools/ci/release"

lane() {
  echo ""
  echo "=== ${tag}: $1"
}

bash "${lanes}/release_channel.sh" "${tag}" > /dev/null

lane "the GitHub Release"
bash "${lanes}/create_github_release.sh" "${tag}"

lane "the InfraChart catalog"
bash "${lanes}/publish_charts.sh" "${tag}"

lane "the BSR labels"
bash "${lanes}/push_buf_labels.sh" "${tag}"

for artifact in presets iac-source catalog-pages proto-source reference-pack conversion-corpus catalog-bundle catalog-schemas logos; do
  lane "content: ${artifact}"
  bash "${lanes}/publish_content.sh" "${tag}" "${artifact}"
done

lane "the stub SDKs"
bash "${lanes}/publish_stub_sdks.sh" "${tag}"

lane "the OpenTofu modules"
go build -o "${work}/modulebuild" ./tools/modulebuild
providers="$("${work}/modulebuild" list -engine terraform | cut -d/ -f2 | sort -u)"
parts="${work}/parts-terraform"
pids=()
for provider in ${providers}; do
  bash "${lanes}/publish_terraform_modules.sh" "${tag}" "${parts}" "${provider}" > "${work}/terraform-${provider}.log" 2>&1 &
  pids+=("$!:${provider}")
done
failed=""
for entry in "${pids[@]}"; do
  if ! wait "${entry%%:*}"; then
    failed="${failed} ${entry#*:}"
  fi
  tail -1 "${work}/terraform-${entry#*:}.log"
done
if [ -n "${failed}" ]; then
  for provider in ${failed}; do
    echo "--- ${provider}" >&2
    tail -20 "${work}/terraform-${provider}.log" >&2
  done
  echo "Observed: the OpenTofu modules of${failed} did not all publish." >&2
  echo "Means:    the OpenTofu manifest is published only once every module is, so the release is not pinnable yet." >&2
  echo "Next:     read the lines above, fix the cause, and re-run; what was published is kept." >&2
  exit 1
fi

lane "the OpenTofu module manifest"
bash "${lanes}/publish_module_manifest.sh" "${tag}" terraform "${parts}"

echo ""
echo "${tag}: every lane a platform pin reads is published. Left to the tag's own"
echo "release run once Actions can run it (re-run it; it keeps what this published):"
echo "the Pulumi modules and their manifest, the definitions, the plantonhq/skills"
echo "distribution and the Stigmer skills."
