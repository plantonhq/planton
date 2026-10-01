#!/usr/bin/env bash
set -euo pipefail

# Guard: what a customer installs from oci://ghcr.io/plantonhq/charts is what
# helm/ says.
#
# WHY THIS EXISTS
# A chart under helm/ publishes only from its tag (helm/<chart>/vX.Y.Z,
# release.helm.yaml), and it releases only when its shape changes -- so a
# shape change that nobody tags ships to no one, while every README, install
# note and values recipe in this repository already teaches the new shape.
# Both ways it hurts are real: a chart whose source stopped pinning a
# platform release while the published one still declares a release the
# operator refuses, so the documented install fails before anything boots;
# and a build Role that gained the verb the runner needs while customers kept
# installing the Role without it, so their build cluster reads not ready.
#
# THE RULE
# Every chart except planton-operator (which releases with the operator, from
# the operator's tag) has a release tag, and every file it ships -- Chart.yaml,
# templates/ (NOTES.txt included), values*.yaml, crds/ -- equals that tag's.
# Markdown is left out: a README edit changes no install. Chart.yaml carries
# the development placeholder in git and in every tag, so no line is masked.
#
# WHERE IT RUNS
# lint.helm.yaml, on pushes to main only: a pull request that changes a chart
# cannot have its release tag yet, and the tag is cut after the merge. Red on
# main means published and source disagree, and the sentence says which
# command makes them agree again; the tag changes no commit, so the check is
# re-run (workflow_dispatch) once the release is out.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

failures=()

for chart_dir in helm/*/; do
  chart="$(basename "${chart_dir}")"
  [ "${chart}" = "planton-operator" ] && continue

  tag="$(git tag --list "helm/${chart}/v*" --sort=-v:refname | head -n 1)"
  if [ -z "${tag}" ]; then
    failures+=("helm/${chart} has never been released from its tag, so customers install whatever was published before it -- release it: make release-helm chart=${chart} version=vX.Y.Z (above the newest version on oci://ghcr.io/plantonhq/charts/${chart}), then re-run this check")
    continue
  fi

  changed="$(git diff --name-only "${tag}" -- "helm/${chart}" ':(exclude)*.md' | sed "s|^helm/${chart}/||" | paste -sd ',' - | sed 's/,/, /g')"
  if [ -n "${changed}" ]; then
    failures+=("helm/${chart} changed since ${tag} (${changed}); a customer still installs ${tag#helm/"${chart}"/} -- release it: make release-helm chart=${chart} bump=minor, then re-run this check")
  fi
done

if [ ${#failures[@]} -gt 0 ]; then
  echo "A published Helm chart differs from its source:"
  for f in "${failures[@]}"; do
    echo "  - $f"
  done
  exit 1
fi

echo "OK: every published Helm chart equals its source (planton-operator releases with the operator)."
