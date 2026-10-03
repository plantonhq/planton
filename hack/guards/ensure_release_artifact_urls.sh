#!/usr/bin/env bash
set -euo pipefail

# Guard: the module artifact URLs the CLI composes must resolve (HTTP 200)
# against the live R2 release layout.
#
# WHY THIS EXISTS
# The CLI's URL builders (pkg/downloads) and the release lanes
# (release.terraform-modules.yaml, release.pulumi-modules.yaml) each own half
# of one contract: the R2 key shape
#   modules/terraform/{kind}/module.zip
#   modules/pulumi/{kind}/{platform}.gz (.exe.gz on windows)
#   modules/pulumi/{kind}/source.zip
# When the two halves drift, nothing fails loudly — every released CLI
# silently degrades each module fetch into a full git clone. Unit tests in
# pkg/downloads pin the CLI half and a post-upload probe in the release lanes
# pins the CI half; this guard is the end-to-end check that a given release
# tag actually serves the shapes the CLI will request.
#
# SCOPE
# NETWORK-DEPENDENT and therefore NOT a PR gate (unlike the sibling offline
# grep guards). Run it on demand — after a release, or when investigating a
# module-download failure:
#   ./hack/guards/ensure_release_artifact_urls.sh [release-tag]
# Without an argument it probes the newest plain v* tag known to this clone.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

base_url="https://downloads.planton.dev/releases"

tag="${1:-}"
if [ -z "$tag" ]; then
  tag=$(git tag --sort=-creatordate | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -1)
  if [ -z "$tag" ]; then
    echo "ERROR: no plain v* release tag found in this clone; pass one explicitly." >&2
    exit 1
  fi
fi

# Representative kinds, one per major provider family. The version
# segment is read from the kind's directory in the working tree — the
# same source the release lanes discover from — so the probe follows a kind
# through version graduations without edits here.
probe_kinds=(
  "aws/awss3bucket"
  "gcp/gcpgcsbucket"
)

failures=0

probe() {
  local url="$1"
  local code
  code=$(curl -s -o /dev/null -w "%{http_code}" -I "$url")
  if [ "$code" = "200" ]; then
    echo "  200  $url"
  else
    echo "  $code  $url  <-- EXPECTED 200"
    failures=$((failures + 1))
  fi
}

echo "Probing release artifact URLs at tag: $tag"

for entry in "${probe_kinds[@]}"; do
  provider="${entry%%/*}"
  kind="${entry##*/}"
  kind_path="catalog/${provider}/${kind}"

  if [ ! -d "$kind_path" ]; then
    echo "ERROR: kind directory not found: $kind_path (update probe_kinds)" >&2
    exit 1
  fi
  if [ ! -d "$kind_path/iac" ]; then
    echo "ERROR: no iac/ module set under $kind_path (modules live at the kind root)" >&2
    exit 1
  fi

  # Keys are versionless (one live module set per kind); the release tag
  # segment versions the artifact. These shapes are pinned by pkg/downloads'
  # tests and uploaded by the release lanes -- three homes, one contract.
  echo ""
  echo "${kind}:"
  probe "${base_url}/${tag}/modules/terraform/${kind}/module.zip"
  probe "${base_url}/${tag}/modules/pulumi/${kind}/darwin_arm64.gz"
  probe "${base_url}/${tag}/modules/pulumi/${kind}/linux_amd64.gz"
  probe "${base_url}/${tag}/modules/pulumi/${kind}/linux_arm64.gz"
  probe "${base_url}/${tag}/modules/pulumi/${kind}/windows_amd64.exe.gz"
  # A 404 here at an old tag is expected: source.zip ships only with releases
  # cut after the pulumi source lane joined the workflow. Probe a newer tag.
  probe "${base_url}/${tag}/modules/pulumi/${kind}/source.zip"
done

echo ""
if [ $failures -gt 0 ]; then
  echo "FAIL: ${failures} artifact URL(s) did not resolve at ${tag} — the CLI's fast download path degrades to git clone for these."
  exit 1
fi
echo "PASS: every probed artifact URL resolves at ${tag}."
