#!/usr/bin/env bash
# =============================================================================
# Install the pinned third-party tools the release scripts run, for this host.
#
#   source <(bash tools/ci/release/ensure_release_tools.sh)    # or
#   eval "$(bash tools/ci/release/ensure_release_tools.sh)"
#
# Prints one `export PATH=...` line naming the directory the tools are in
# (.tools/release/<os>_<arch>/bin under the repository root), so a workflow step
# and a laptop run the same binaries. Each tool is one exact release, checked
# against the sha256 its own checksums.txt publishes before it is unpacked; a
# tool already in place at its pinned version is reused, so a second call costs
# nothing.
#
# The versions are the release lanes' own: crane is what every lane copies and
# reads images with, and kustomize is the release built from the
# sigs.k8s.io/kustomize/api version upstream's Agent Substrate installer embeds
# (v0.21.1 in its go.mod at v0.6.0 and v1.0.0; moving the declared Substrate
# version re-checks it). Moving a version means moving every sha256 below,
# copied from that release's checksums.txt.
# =============================================================================
set -euo pipefail

CRANE_VERSION="v0.22.1"
KUSTOMIZE_VERSION="v5.8.1"

case "$(uname -s)" in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) echo "ensure_release_tools.sh: no pinned tools for $(uname -s)" >&2; exit 1 ;;
esac
case "$(uname -m)" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "ensure_release_tools.sh: no pinned tools for $(uname -m)" >&2; exit 1 ;;
esac

# go-containerregistry names its archives by Go-style OS and uname-style arch.
crane_archive() {
  local o a
  o="$([ "${os}" = linux ] && echo Linux || echo Darwin)"
  a="$([ "${arch}" = amd64 ] && echo x86_64 || echo arm64)"
  echo "go-containerregistry_${o}_${a}.tar.gz"
}
crane_sha256() {
  case "${os}_${arch}" in
    linux_amd64) echo 0ab7a1d6932a213aed964ce97666c3077fe691c8606413674a8b3e0b9ec4cda0 ;;
    linux_arm64) echo 898c0cff975f898a33e8c4580bdafb0e7c02c7faa33374e946762f97c4ab7110 ;;
    darwin_amd64) echo 6fedd06a648c11335f0e8b9547e4783002c30e9336047fec292942cd518bb799 ;;
    darwin_arm64) echo 2231fc8df8806d20d680ff1225db44e095a55dd6ac1ae8eced4faf4b278b78fb ;;
  esac
}
kustomize_sha256() {
  case "${os}_${arch}" in
    linux_amd64) echo 029a7f0f4e1932c52a0476cf02a0fd855c0bb85694b82c338fc648dcb53a819d ;;
    linux_arm64) echo 0953ea3e476f66d6ddfcd911d750f5167b9365aa9491b2326398e289fef2c142 ;;
    darwin_amd64) echo ee7cf0c1e3592aa7bb66ba82b359933a95e7f2e0b36e5f53ed0a4535b017f2f8 ;;
    darwin_arm64) echo 8886f8a78474e608cc81234f729fda188a9767da23e28925802f00ece2bab288 ;;
  esac
}

sha256_of() {
  if command -v sha256sum > /dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    shasum -a 256 "$1" | cut -d' ' -f1
  fi
}

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
bin="${repo_root}/.tools/release/${os}_${arch}/bin"
mkdir -p "${bin}"

# install <name> <version> <url> <sha256>: unpack <name> from the archive at
# url into bin, once per version.
install() {
  local name="$1" version="$2" url="$3" want="$4" stamp archive got
  stamp="${bin}/.${name}.version"
  if [ -x "${bin}/${name}" ] && [ "$(cat "${stamp}" 2>/dev/null)" = "${version}" ]; then
    return
  fi
  archive="$(mktemp)"
  curl -sSfL -o "${archive}" "${url}"
  got="$(sha256_of "${archive}")"
  if [ "${got}" != "${want}" ]; then
    rm -f "${archive}"
    echo "ensure_release_tools.sh: ${url} has sha256 ${got}, not the pinned ${want}; refusing to install it" >&2
    exit 1
  fi
  tar -xzf "${archive}" -C "${bin}" "${name}"
  rm -f "${archive}"
  chmod 0755 "${bin}/${name}"
  echo "${version}" > "${stamp}"
  echo "installed ${name} ${version} for ${os}/${arch}" >&2
}

install crane "${CRANE_VERSION}" \
  "https://github.com/google/go-containerregistry/releases/download/${CRANE_VERSION}/$(crane_archive)" \
  "$(crane_sha256)"
install kustomize "${KUSTOMIZE_VERSION}" \
  "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_${os}_${arch}.tar.gz" \
  "$(kustomize_sha256)"

echo "export PATH=\"${bin}:\${PATH}\""
