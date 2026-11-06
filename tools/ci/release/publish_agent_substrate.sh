#!/usr/bin/env bash
# =============================================================================
# Publish (or rehearse) Planton's install set for one Agent Substrate release.
#
#   publish_agent_substrate.sh [--rehearse] [vX.Y.Z]
#
# The one code path of the Agent Substrate lane: the workflow
# (.github/workflows/release.agent-substrate.yaml) runs it, and so does a
# laptop when Actions cannot. The version defaults to the one
# tools/ci/release/agent-substrate.yaml declares.
#
#   1. Resolve upstream's tag to its commit once, so every read below sees the
#      same bytes even if the tag moves mid-run.
#   2. Publishing only: if the version's SHA256SUMS is already on the bucket,
#      the version is sealed; verify what is published and stop. Otherwise
#      refuse any version but the declared one: only the declared version has
#      been rehearsed by the change that declared it.
#   3. Check out upstream at the commit (and, for a release claiming an earlier
#      shape, at the shape's tag, with the shape's published tree).
#   4. Fetch each released image's multi-platform index anonymously from the
#      declared repository, exactly as served, so the composer derives every
#      digest from the bytes.
#   5. Compose the install set and check it (every kustomization builds, every
#      reference is one the release publishes).
#   6. Publishing only: upload every file with publish_once.sh, SHA256SUMS
#      last (its presence is the seal), then verify what an adopter downloads
#      and pulls against what this run composed.
#
# A rehearsal holds no credential and publishes nothing. Publishing reads the
# R2 key pair from AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY and the endpoint
# from AWS_ENDPOINT_URL, as every release lane does.
#
# Needs git, curl, jq, python3 and aws; crane and kustomize come pinned from
# ensure_release_tools.sh.
# =============================================================================
set -euo pipefail

usage() {
  echo "usage: publish_agent_substrate.sh [--rehearse] [vX.Y.Z]" >&2
  exit 2
}

rehearse=false
version=""
for arg in "$@"; do
  case "${arg}" in
    --rehearse) rehearse=true ;;
    v[0-9]*.[0-9]*.[0-9]*) version="${arg}" ;;
    *) usage ;;
  esac
done

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
release_dir="${repo_root}/tools/ci/release"
config="${release_dir}/agent-substrate.yaml"
UPSTREAM="https://github.com/agent-substrate/substrate"
export R2_BUCKET="${R2_BUCKET:-planton-foundation-downloads}"

declared="$(sed -n 's/^version: *//p' "${config}")"
declared_shape="$(sed -n 's/^shape: *//p' "${config}")"
repository="$(sed -n 's/^repository: *//p' "${config}")"
version="${version:-${declared}}"
case "${version}" in
  v[0-9]*.[0-9]*.[0-9]*) ;;
  *) echo "Error: the version must look like vX.Y.Z (got '${version}')" >&2; exit 1 ;;
esac
if [ -z "${repository}" ]; then
  echo "Error: ${config} declares no repository: where upstream publishes the release's images" >&2
  exit 1
fi
# The declared version carries its declared shape; any other version this run
# rehearses is its own.
shape="${version}"
if [ "${version}" = "${declared}" ]; then
  shape="${declared_shape}"
fi
case "${shape}" in
  v[0-9]*.[0-9]*.[0-9]*) ;;
  *) echo "Error: ${config} must declare the shape as vX.Y.Z (got '${shape}')" >&2; exit 1 ;;
esac

# A separate assignment, so a refused install stops the run: `eval "$(...)"`
# succeeds whatever the substitution exits with.
tools_env="$(bash "${release_dir}/ensure_release_tools.sh")"
eval "${tools_env}"
kustomize_bin="$(command -v kustomize)" || {
  echo "Error: ensure_release_tools.sh put no kustomize on PATH, and the composed set is checked with it" >&2
  exit 1
}

# 1. The commit the tag names. An annotated tag's peeled entry (^{}) is the
# commit it names.
refs="$(git ls-remote --tags "${UPSTREAM}" "refs/tags/${version}" "refs/tags/${version}^{}")"
commit="$(printf '%s\n' "${refs}" | awk '/\^\{\}$/ {print $1}')"
[ -n "${commit}" ] || commit="$(printf '%s\n' "${refs}" | awk 'NF {print $1; exit}')"
if [ -z "${commit}" ]; then
  echo "Error: ${UPSTREAM} has no tag ${version}. Planton publishes only upstream's own releases (${UPSTREAM}/tags)" >&2
  exit 1
fi
echo "Agent Substrate ${version} (shape ${shape}) at ${commit}; images from ${repository}; $([ "${rehearse}" = true ] && echo rehearsal || echo publish)"

# 2. Sealed, or the declared version.
if [ "${rehearse}" = false ]; then
  : "${AWS_ENDPOINT_URL:?publishing needs the R2 endpoint in AWS_ENDPOINT_URL}"
  # The S3 API, not downloads.planton.ai: a CDN answer can be a cached miss,
  # and this decides whether the run composes at all.
  if aws s3api head-object --bucket "${R2_BUCKET}" --endpoint-url "${AWS_ENDPOINT_URL}" \
      --key "releases/agent-substrate/${version}/SHA256SUMS" > /dev/null 2>&1; then
    echo "Agent Substrate ${version} is published and sealed; verifying it, composing nothing."
    exec bash "${release_dir}/verify_agent_substrate_release.sh" "${version}"
  fi
  if [ "${version}" != "${declared}" ]; then
    echo "Observed: a publish of Agent Substrate ${version}, which is not published yet, while ${config} declares ${declared}." >&2
    echo "Means:    only the declared version has been rehearsed by the change that declared it, so ${version} would publish unproven." >&2
    echo "Next:     declare ${version} in ${config} in a change whose rehearsal passes, then publish it." >&2
    exit 1
  fi
fi

work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT

# 3. Upstream at the commit, with no credential left on disk.
checkout() {
  local ref="$1" dir="$2"
  git init -q "${dir}"
  git -C "${dir}" fetch -q --depth 1 "${UPSTREAM}" "${ref}"
  git -C "${dir}" checkout -q FETCH_HEAD
}
checkout "${commit}" "${work}/upstream"

shape_args=()
if [ "${shape}" != "${version}" ]; then
  # A release claiming an earlier shape is compared with that shape: its
  # upstream code at the shape's tag, and its install set as published, every
  # file checked against the seal.
  checkout "refs/tags/${shape}" "${work}/upstream-shape"
  base="https://downloads.planton.ai/releases/agent-substrate/${shape}"
  mkdir -p "${work}/shape-release"
  if ! curl -fsSL "${base}/SHA256SUMS" -o "${work}/shape-release/SHA256SUMS"; then
    echo "Observed: ${base}/SHA256SUMS does not download." >&2
    echo "Means:    ${version} claims the install shape of ${shape}, which Planton has not published." >&2
    echo "Next:     declare a published shape, or ${version} itself, in ${config}." >&2
    exit 1
  fi
  while read -r _ path; do
    mkdir -p "$(dirname "${work}/shape-release/${path}")"
    curl -fsSL "${base}/${path}" -o "${work}/shape-release/${path}"
  done < "${work}/shape-release/SHA256SUMS"
  if command -v sha256sum > /dev/null 2>&1; then
    (cd "${work}/shape-release" && sha256sum --quiet -c SHA256SUMS)
  else
    (cd "${work}/shape-release" && shasum -a 256 --quiet -c SHA256SUMS)
  fi
  shape_args=(--shape-release "${work}/shape-release" --shape-source "${work}/upstream-shape")
fi

# 4. Each released image's index, anonymously, exactly as the registry serves
# it for the tag.
composer="${release_dir}/compose_agent_substrate_release.py"
mkdir -p "${work}/indexes"
export DOCKER_CONFIG="${work}/anonymous"
mkdir -p "${DOCKER_CONFIG}" && echo '{}' > "${DOCKER_CONFIG}/config.json"
images="$(python3 "${composer}" images --source "${work}/upstream")"
for image in ${images}; do
  if ! crane manifest "${repository}/${image}:${version}" > "${work}/indexes/${image}.json"; then
    echo "Observed: ${repository}/${image}:${version} does not pull without credentials." >&2
    echo "Means:    upstream's release names ${image}, and an adopter's cluster could not pull it." >&2
    echo "Next:     check upstream's release workflow and the package's visibility, then re-run." >&2
    exit 1
  fi
done

# 5. Compose and check.
(cd "${release_dir}" && python3 -m unittest -q compose_agent_substrate_release_test.py)
python3 "${composer}" \
  --source "${work}/upstream" \
  --index-manifests "${work}/indexes" \
  --repository "${repository}" \
  --version "${version}" \
  --shape "${shape}" \
  "${shape_args[@]+"${shape_args[@]}"}" \
  --commit "${commit}" \
  --out "${work}/release" \
  --kustomize "${kustomize_bin}"

if [ "${rehearse}" = true ]; then
  if [ -n "${AGENT_SUBSTRATE_RELEASE_OUT:-}" ]; then
    # Keep the composed set, to refresh the module's testdata from it.
    rm -rf "${AGENT_SUBSTRATE_RELEASE_OUT}" && cp -R "${work}/release" "${AGENT_SUBSTRATE_RELEASE_OUT}"
    echo "composed set kept at ${AGENT_SUBSTRATE_RELEASE_OUT}"
  fi
  echo "Rehearsed Agent Substrate ${version}: composed and checked; nothing published."
  exit 0
fi

# 6. Upload, seal, verify.
prefix="releases/agent-substrate/${version}"
(
  cd "${work}/release"
  find . -type f ! -name SHA256SUMS | sed 's|^\./||' | sort | while read -r path; do
    bash "${release_dir}/publish_once.sh" "${path}" "${prefix}/${path}" > /dev/null
  done
  bash "${release_dir}/publish_once.sh" SHA256SUMS "${prefix}/SHA256SUMS" > /dev/null
)
bash "${release_dir}/verify_agent_substrate_release.sh" "${version}" "${work}/release"
