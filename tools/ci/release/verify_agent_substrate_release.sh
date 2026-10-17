#!/usr/bin/env bash
# tools/ci/release/verify_agent_substrate_release.sh <version> [<composed directory>]
#
# Proves a published Agent Substrate release is what it says it is, from an
# adopter's side: nothing here holds a credential.
#
#   - Every file SHA256SUMS lists downloads from downloads.planton.ai and
#     matches its sum.
#   - Given the directory a run composed, the downloaded set equals it byte for
#     byte and file for file. publish_once.sh keeps any file already published
#     under the version, so a version half-published by an earlier run with
#     different digests fails here instead of passing as "kept".
#   - Every image images.json names resolves with no login, on ghcr.io and on
#     the Artifact Registry mirror, both by its pinned digest and by its tag,
#     and the tag names exactly that digest.
#
# Needs curl, jq, sha256sum and crane on PATH.
set -euo pipefail

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  echo "usage: verify_agent_substrate_release.sh <version> [<composed directory>]" >&2
  exit 2
fi
version="$1"
expect="${2:-}"

SOURCE_ROOT="ghcr.io/plantonhq"
MIRROR_ROOT="us-central1-docker.pkg.dev/plantonhq"
base="https://downloads.planton.ai/releases/agent-substrate/${version}"

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

failed=0
fail() {
  echo "::error::$1" >&2
  failed=1
}

if ! curl -fsSL "${base}/SHA256SUMS" -o "${work}/SHA256SUMS"; then
  echo "::error::${base}/SHA256SUMS does not download; the version is not published (a published version always carries it, uploaded last)" >&2
  exit 1
fi
while read -r _ path; do
  mkdir -p "$(dirname "${work}/${path}")"
  curl -fsSL "${base}/${path}" -o "${work}/${path}" || fail "${base}/${path} is listed in SHA256SUMS but does not download"
done < "${work}/SHA256SUMS"
(cd "${work}" && sha256sum --quiet -c SHA256SUMS) || fail "files under ${base} do not match their SHA256SUMS; an upload was interrupted or replaced"

if [ -n "${expect}" ]; then
  if ! diff -r "${expect}" "${work}" > "${work}.diff"; then
    sed 's/^/  /' "${work}.diff" >&2
    fail "what ${base} serves differs from what this run composed; an earlier run published this version with other contents, so investigate before anyone installs it"
  fi
  rm -f "${work}.diff"
fi

# An empty Docker config: every pull below is anonymous, as an adopter's is.
DOCKER_CONFIG="$(mktemp -d)"
export DOCKER_CONFIG
echo '{}' > "${DOCKER_CONFIG}/config.json"

while read -r name ref; do
  pinned="${ref##*@}"
  tagged="${ref%@*}"
  path="${tagged#"${SOURCE_ROOT}"/}"
  for root in "${SOURCE_ROOT}" "${MIRROR_ROOT}"; do
    if ! crane manifest "${root}/${path%:*}@${pinned}" > /dev/null 2>&1; then
      fail "${root}/${path%:*}@${pinned} (${name}) does not pull without credentials"
      continue
    fi
    tag_digest="$(crane digest "${root}/${path}" 2>/dev/null || true)"
    if [ "${tag_digest}" != "${pinned}" ]; then
      fail "${root}/${path} names ${tag_digest:-nothing}, but the release pins ${name} to ${pinned}"
    fi
  done
  echo "verified ${name} ${pinned} on ${SOURCE_ROOT} and ${MIRROR_ROOT}"
done < <(jq -r '.images | to_entries[] | "\(.key) \(.value)"' "${work}/images.json")

if [ "${failed}" -ne 0 ]; then
  exit 1
fi
echo "Agent Substrate ${version}: every published file matches, and every image pulls anonymously from both registries by digest and by tag."
