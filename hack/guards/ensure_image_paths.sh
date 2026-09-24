#!/usr/bin/env bash
set -euo pipefail

# Guard: every Planton container image is named by one rule, and retired
# addresses never come back.
#
# WHY: Planton's images live at ghcr.io/plantonhq/planton/<slug>, where the
# slug names what the image is (control-plane, runner, operator, mcp,
# runner-tunnel, client-apps/web) and Planton OS images carry the os/ prefix
# (os/mcp, os/client-apps/web). Two earlier schemes -- repository-mirrored
# paths under planton/product/... and the mcp-os spelling -- once coexisted
# with this one, and the documents disagreed with the workflows about which
# image used which. A default the operator compiles in, a chart value, an
# example manifest, or a page that names a retired address sends an adopter to
# an image that will never be updated again. This guard holds every tracked
# file to the current addresses; dated records (_changelog/) are the one place
# the old ones may remain, because they describe the past.
#
# The platform repository holds the same table for its scripts and workflows
# (product/tools/ci/registry.sh) and runs the same check.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

# Retired address forms, matched literally. Each entry: pattern|replacement.
retired=(
  'ghcr.io/plantonhq/planton/product/client-apps/planton/web|ghcr.io/plantonhq/planton/client-apps/web'
  'ghcr.io/plantonhq/planton/product/client-apps/planton-os/web|ghcr.io/plantonhq/planton/os/client-apps/web'
  'ghcr.io/plantonhq/planton/product/|ghcr.io/plantonhq/planton/<slug> (no product/ prefix)'
  'ghcr.io/plantonhq/planton/mcp-os|ghcr.io/plantonhq/planton/os/mcp'
)

status=0
for entry in "${retired[@]}"; do
  pattern="${entry%%|*}"
  replacement="${entry#*|}"
  # Tracked files (what a commit would carry); dated records and this guard's
  # own table are excluded by path. git grep exits 1 when nothing matches and
  # >1 on a real error.
  set +e
  hits="$(git grep -n -F -- "$pattern" -- ':(exclude)_changelog' ':(exclude)hack/guards/ensure_image_paths.sh')"
  rc=$?
  set -e
  if [[ $rc -gt 1 ]]; then
    echo "ERROR: git grep failed (exit $rc) while checking '${pattern}'" >&2
    exit "$rc"
  fi
  if [[ -n "$hits" ]]; then
    status=1
    echo "ERROR: retired image address '${pattern}' is still referenced:" >&2
    echo "$hits" | sed 's/^/  /' >&2
    echo "  Use ${replacement} instead." >&2
    echo >&2
  fi
done

if [[ $status -ne 0 ]]; then
  echo "Image path guard FAILED. Every Planton image is ghcr.io/plantonhq/planton/<slug>; Planton OS images carry the os/ prefix." >&2
  exit 1
fi

# Rule two: code names a registry root only in its home. Every release is
# published to ghcr.io and mirrored, byte for byte, to Google Artifact
# Registry (asia-south1-docker.pkg.dev/plantonhq), and an install chooses
# between them with one setting (spec.imageRegistry, a module's
# chart_repository). A root written into any other code file is a default that
# setting cannot move. The operator's home is one package; each catalog module
# is self-contained (ensure_modules_are_self_contained.sh), so its vars.go and
# locals.tf are its own home, mirroring the proto field's default. Comments,
# tests (they pin the defaults' values), generated stubs, protos (the
# catalog's defaults live there), docs and the site are not code defaults.
homes=(
  operator/internal/plantonregistry/plantonregistry.go
  catalog/kubernetes/kubernetesplantonoperator/iac/pulumi/module/vars.go
  catalog/kubernetes/kubernetesplantonoperator/iac/tf/locals.tf
  catalog/kubernetes/kubernetesplantonrunner/iac/pulumi/module/vars.go
  catalog/kubernetes/kubernetesplantonrunner/iac/tf/locals.tf
)
home_excludes=()
for home in "${homes[@]}"; do
  home_excludes+=(":(exclude)${home}")
done

set +e
root_hits="$(git grep --untracked -n -E 'ghcr\.io/plantonhq|docker\.pkg\.dev/plantonhq' -- \
  '*.go' '*.ts' '*.tsx' '*.js' '*.mjs' '*.cjs' '*.py' '*.tf' '*.java' '*.rs' \
  ':(exclude)_changelog' ':(exclude)site' ':(exclude)*_test.go' ':(exclude)*.test.*' \
  ':(exclude)*.pb.go' ':(exclude)**/node_modules/**' "${home_excludes[@]}")"
rc=$?
set -e
if [[ $rc -gt 1 ]]; then
  echo "ERROR: git grep failed (exit $rc) while checking registry roots in code" >&2
  exit "$rc"
fi
root_hits="$(printf '%s\n' "$root_hits" | awk '
  NF == 0 { next }
  { text = $0; sub(/^[^:]*:[0-9]+:/, "", text); sub(/^[ \t]+/, "", text) }
  text ~ /^(\/\/|#|\*|\/\*)/ { next }
  { print }')"
if [[ -n "$root_hits" ]]; then
  echo "ERROR: a registry root is written into code outside its home:" >&2
  echo "$root_hits" | sed 's/^/  /' >&2
  echo "  Derive it from operator/internal/plantonregistry (the operator) or the module's own vars.go / locals.tf (a catalog module)." >&2
  exit 1
fi

echo "OK: no retired image address is referenced outside dated records, and code names a registry root only in its home."
