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
#   - Every image images.json names resolves with no login at the reference
#     it names, both by its pinned digest and by its tag, and the tag names
#     exactly that digest. The reference carries its own registry: upstream's
#     ghcr.io/agent-substrate/substrate from v1.0.0, Planton's own build
#     (ghcr.io/plantonhq/planton/agent-substrate) for the releases before it.
#
# Needs curl, jq and sha256sum (or macOS's shasum); crane comes pinned from
# ensure_release_tools.sh.
set -euo pipefail

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
  echo "usage: verify_agent_substrate_release.sh <version> [<composed directory>]" >&2
  exit 2
fi
version="$1"
expect="${2:-}"

base="https://downloads.planton.ai/releases/agent-substrate/${version}"

# A separate assignment, so a refused install stops the run.
tools_env="$(bash "$(dirname "${BASH_SOURCE[0]}")/ensure_release_tools.sh")"
eval "${tools_env}"

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

# check_sums checks SHA256SUMS in the current directory, with GNU sha256sum or
# macOS's shasum, whichever the machine has.
check_sums() {
  if command -v sha256sum > /dev/null 2>&1; then
    sha256sum --quiet -c SHA256SUMS
  else
    shasum -a 256 --quiet -c SHA256SUMS
  fi
}

failed=0
fail() {
  echo "::error::$1" >&2
  failed=1
}

# The published set downloads into set/; the comparison's diff and the
# anonymous Docker config sit beside it, so the trap removes everything.
set_dir="${work}/set"
mkdir -p "${set_dir}"
if ! curl -fsSL "${base}/SHA256SUMS" -o "${set_dir}/SHA256SUMS"; then
  echo "::error::${base}/SHA256SUMS does not download; the version is not published (a published version always carries it, uploaded last)" >&2
  exit 1
fi
while read -r _ path; do
  mkdir -p "$(dirname "${set_dir}/${path}")"
  curl -fsSL "${base}/${path}" -o "${set_dir}/${path}" || fail "${base}/${path} is listed in SHA256SUMS but does not download"
done < "${set_dir}/SHA256SUMS"
(cd "${set_dir}" && check_sums) || fail "files under ${base} do not match their SHA256SUMS; an upload was interrupted or replaced"

if [ -n "${expect}" ]; then
  if ! diff -r "${expect}" "${set_dir}" > "${work}/composed.diff"; then
    sed 's/^/  /' "${work}/composed.diff" >&2
    fail "what ${base} serves differs from what this run composed; an earlier run published this version with other contents, so investigate before anyone installs it"
  fi
fi

# An empty Docker config: every pull below is anonymous, as an adopter's is.
export DOCKER_CONFIG="${work}/anonymous"
mkdir -p "${DOCKER_CONFIG}"
echo '{}' > "${DOCKER_CONFIG}/config.json"

# with_retries runs a registry read up to three times, five seconds apart:
# ghcr.io throttles a burst of anonymous reads from one address, and a
# throttled read is not an image an adopter cannot pull.
with_retries() {
  local attempt
  for attempt in 1 2 3; do
    if "$@"; then
      return 0
    fi
    [ "${attempt}" -lt 3 ] && sleep 5
  done
  return 1
}

while read -r name ref; do
  pinned="${ref##*@}"
  tagged="${ref%@*}"
  if ! with_retries crane manifest "${tagged%:*}@${pinned}" > /dev/null 2>&1; then
    fail "${tagged%:*}@${pinned} (${name}) does not pull without credentials"
    continue
  fi
  tag_digest="$(with_retries crane digest "${tagged}" 2>/dev/null || true)"
  if [ "${tag_digest}" != "${pinned}" ]; then
    fail "${tagged} names ${tag_digest:-nothing}, but the release pins ${name} to ${pinned}"
    continue
  fi
  echo "verified ${name} ${pinned} at ${tagged}"
done < <(jq -r '.images | to_entries[] | "\(.key) \(.value)"' "${set_dir}/images.json")

if [ "${failed}" -ne 0 ]; then
  exit 1
fi
echo "Agent Substrate ${version}: every published file matches, and every image pulls anonymously by digest and by tag."
