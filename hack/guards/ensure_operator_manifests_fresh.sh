#!/usr/bin/env bash
set -euo pipefail

# Guard: everything controller-gen derives from the operator's Go sources is
# committed fresh -- in the operator's own config/ tree AND in the Helm chart
# that packages it.
#
# WHY: the operator's CRDs and its manager ClusterRole are generated from
# +kubebuilder markers, and `make -C operator manifests` writes them to two
# homes: config/crd/bases + config/rbac/role.yaml (what kubebuilder's own
# install path reads) and helm/planton-operator/templates/crds + rbac/ (what
# the chart ships; the CRD templates are controller-gen's output wrapped by
# operator/hack/chartcrds so the chart owns their lifecycle). Two homes
# written by one generator run is the kubebuilder-with-Helm shape; the
# invariant that makes it safe is that nobody edits either home by hand and
# the committed copies are never stale against the markers. A
# regenerate-then-diff proves exactly that, and also covers the deepcopy code
# `make generate` owns. A hand-edited CRD, a forgotten regeneration, or a
# chart copy that drifted from the operator's code all fail here with the one
# command that fixes them.
#
# Needs the Go toolchain (controller-gen is installed into operator/bin by the
# operator's Makefile); no cluster, no network beyond module downloads.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

# Only the generator-owned paths are compared, so a developer's in-progress edit
# to a hand-written source file does not read as a stale manifest.
generated_paths=(
  operator/config/crd/bases
  operator/config/rbac/role.yaml
  ':(glob)operator/api/**/zz_generated.*'
  operator/api/v1/component_sizing.json
  helm/planton-operator/templates/crds
  helm/planton-operator/rbac
)

make -C operator manifests generate >/dev/null

if ! git diff --quiet --exit-code -- "${generated_paths[@]}"; then
  echo "ERROR: generated operator manifests are stale against the Go sources." >&2
  echo "" >&2
  git --no-pager diff --stat -- "${generated_paths[@]}" >&2
  echo "" >&2
  echo "Fix: run 'make operator-manifests' and commit the result." >&2
  exit 1
fi

untracked="$(git ls-files --others --exclude-standard -- "${generated_paths[@]}")"
if [[ -n "$untracked" ]]; then
  echo "ERROR: generation produced files that are not committed:" >&2
  echo "$untracked" >&2
  echo "" >&2
  echo "Fix: run 'make operator-manifests' and commit the result." >&2
  exit 1
fi

echo "OK: operator CRDs, RBAC, deepcopy code, and the sizing registry are fresh in config/, api/v1/, and helm/planton-operator."
