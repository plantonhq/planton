#!/usr/bin/env bash
set -euo pipefail

# Guard: the oldest platform release the operator runs is a release an
# adopter can actually pull, and no kind scenario declares one it refuses.
#
# WHY THIS EXISTS
# platformversion.MinimumSupported promises a release published whole: a
# platform declared exactly at the floor pulls the control plane, the runner
# and the console at that tag (the console's tag defaults to spec.version).
# Only the floor's FORMAT was ever tested. Moving the floor and publishing a
# release are two acts, and once the floor named a tag whose console image
# was not published yet: the operator accepted the platform and the console
# waited on ImagePullBackOff. The platform kind's own kind scenarios drift the
# other way -- a scenario below the floor is refused with
# VersionSupported=False before anything boots, so its lane proves nothing.
#
# THE RULE
# 1. For the floor tag, and for every tag a platform kind scenario declares,
#    each image the operator deploys (<registry>/<slug>:<tag>, the registry
#    root and the slugs read from the operator's own defaults) has a
#    published manifest. Each image's architectures are printed, so a reader
#    sees which nodes a release can run on without a second lookup.
# 2. Every platform kind scenario declares a version at or above the floor.
#
# WHERE IT RUNS
# lint.operator.yaml, when the floor, a platform kind scenario, or this guard
# changes. It reads the public registry anonymously (docker manifest inspect).
#
# TESTING
# MANIFEST_INSPECT replaces the manifest-inspect command (default:
# "docker manifest inspect -v"); it is called with one image reference and
# prints docker's verbose JSON (one object for a single manifest, an array for
# an index) or exits non-zero when the manifest does not exist. REPO_ROOT
# points the guard at another tree. ensure_platform_floor_is_published_test.sh
# drives both with a fake registry.

repo_root_dir="${REPO_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
cd "$repo_root_dir"

read -r -a manifest_inspect <<< "${MANIFEST_INSPECT:-docker manifest inspect -v}"

floor_file="operator/internal/platformversion/platformversion.go"
registry_file="operator/internal/plantonregistry/plantonregistry.go"
images_file="operator/internal/resources/images.go"
scenarios_dir="catalog/kubernetes/kubernetesplantonplatform/e2e/scenarios"

floor="$(sed -nE 's/^const MinimumSupported = "([^"]+)"$/\1/p' "$floor_file")"
registry="$(sed -nE 's/^const DefaultImageRegistry = "([^"]+)"$/\1/p' "$registry_file")"
# The slug constants: `ControlPlaneImageSlug = "control-plane"`, one per image
# the operator deploys, so an image added there is checked here without an edit.
slugs="$(sed -nE 's/^[[:space:]]+[A-Za-z]+ImageSlug[[:space:]]*=[[:space:]]*"([^"]+)"$/\1/p' "$images_file")"

if [[ -z "$floor" || -z "$registry" || -z "$slugs" ]]; then
  echo "GUARD BUG: could not read the floor (${floor_file}), the registry root (${registry_file}) or the image slugs (${images_file}); their declarations changed shape -- fix this guard's extraction before trusting it" >&2
  exit 1
fi

# is_below A B: true when release A sorts strictly before release B.
is_below() {
  [[ "$1" != "$2" && "$(printf '%s\n%s\n' "$1" "$2" | sort -V | head -n 1)" == "$1" ]]
}

failures=()

# Rule 2, and the tags rule 1 checks beyond the floor.
tags=("$floor")
for scenario in "$scenarios_dir"/*.yaml; do
  version="$(sed -nE 's/^  version:[[:space:]]*"?([^"[:space:]]+)"?[[:space:]]*$/\1/p' "$scenario")"
  if [[ -z "$version" ]]; then
    failures+=("${scenario} declares no spec.version this guard can read; the operator refuses a platform without one -- declare a release at or above ${floor}")
    continue
  fi
  if is_below "$version" "$floor"; then
    failures+=("${scenario} declares ${version}, below the operator's floor ${floor}, so the operator refuses it with VersionSupported=False and the scenario proves nothing -- declare a published release at or above ${floor}")
  fi
  [[ " ${tags[*]} " == *" ${version} "* ]] || tags+=("$version")
done

# Rule 1.
for tag in "${tags[@]}"; do
  while IFS= read -r slug; do
    ref="${registry}/${slug}:${tag}"
    if ! manifest="$("${manifest_inspect[@]}" "$ref" 2>&1)"; then
      failures+=("${ref} is not published (${manifest_inspect[*]} said: $(printf '%s' "$manifest" | head -n 1)); a platform at ${tag} would wait on ImagePullBackOff for it -- publish that release's ${slug} image, or point the floor and the scenarios at a release whose images are all published")
      continue
    fi
    # Attestation entries in an index carry the platform unknown/unknown.
    architectures="$(printf '%s' "$manifest" | jq -r '
      (if type == "array" then .[] else . end)
      | .Descriptor.platform // empty
      | select(.os != "unknown")
      | "\(.os)/\(.architecture)\(if .variant then "/" + .variant else "" end)"' 2>/dev/null | sort -u | paste -sd ',' - | sed 's/,/, /g')"
    echo "published: ${ref} (${architectures:-architecture not stated})"
  done <<< "$slugs"
done

if [[ ${#failures[@]} -gt 0 ]]; then
  echo "The operator's platform floor or a platform kind scenario names a release an adopter cannot run:" >&2
  for f in "${failures[@]}"; do
    echo "  - $f" >&2
  done
  exit 1
fi

echo "OK: floor ${floor} and every platform kind scenario name releases whose images are all published."
