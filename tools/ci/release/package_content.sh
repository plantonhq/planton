#!/usr/bin/env bash
# =============================================================================
# Package OpenMCF content for distribution as GitHub Release artifacts.
#
# Creates four versioned zip files, each scoped to a single concern:
#
#   presets_{version}.zip       -- Preset YAML + MD files, kind enum proto
#   iac-source_{version}.zip    -- IaC source (.go, .tf, .md, .yaml under iac/)
#   catalog-pages_{version}.zip -- Per-component catalog-page.md files
#   proto-source_{version}.zip  -- Raw proto source (spec, api, stack_input, stack_outputs)
#
# All zips preserve repo-relative paths so they can be extracted into a single
# directory and overlay into a virtual OpenMCF root. Consumers like the Planton
# upgrade scripts use this merged directory as --openmcf-path or OPENMCF_ROOT.
#
# Usage:
#   bash tools/ci/release/package_content.sh v0.3.50
#   bash tools/ci/release/package_content.sh v0.3.50 --dry-run
# =============================================================================

set -euo pipefail

VERSION="${1:?Usage: package_content.sh <version-tag> [--dry-run]}"
DRY_RUN="${2:-}"
V="${VERSION#v}"

REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$REPO_ROOT"

PROVIDER_BASE="apis/org/openmcf/provider"

if [ ! -d "$PROVIDER_BASE" ]; then
  echo "ERROR: Provider base directory not found: $PROVIDER_BASE"
  exit 1
fi

echo "=== Packaging OpenMCF content for ${VERSION} ==="
echo ""

create_zip() {
  local zip_name="$1"
  local description="$2"
  shift 2

  local tmp_list
  tmp_list=$(mktemp)

  # Read file paths from stdin into a sorted temp file.
  sort > "$tmp_list"

  local count
  count=$(wc -l < "$tmp_list" | tr -d ' ')

  if [ "$count" -eq 0 ]; then
    echo "  WARNING: No files found for ${description}. Skipping ${zip_name}."
    rm -f "$tmp_list"
    return
  fi

  if [ "$DRY_RUN" = "--dry-run" ]; then
    echo "  [dry-run] ${zip_name}: ${count} files"
    rm -f "$tmp_list"
    return
  fi

  zip -q -@ "$zip_name" < "$tmp_list"
  rm -f "$tmp_list"

  local size
  size=$(du -h "$zip_name" | cut -f1)
  printf "  %-30s %6s  (%s files)\n" "$zip_name" "$size" "$count"
}

# ─── 1. Presets ───────────────────────────────────────────────────────────────
echo "1/4  Presets..."
{
  find "$PROVIDER_BASE" \( -path '*/v1/presets/*.yaml' -o -path '*/v1/presets/*.md' \)
  echo "apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto"
} | create_zip "presets_${V}.zip" "presets"

# ─── 2. IaC Source ────────────────────────────────────────────────────────────
# Mirrors the ALLOWED_EXTENSIONS in iac-bundler.ts: .go, .tf, .md, .yaml
# Excludes hidden dirs, vendor, and node_modules (same as iac-bundler.ts).
echo "2/4  IaC source..."
find "$PROVIDER_BASE" -path '*/v1/iac/*' \
    \( -name '*.go' -o -name '*.tf' -o -name '*.md' -o -name '*.yaml' \) \
    ! -path '*/vendor/*' \
    ! -path '*/node_modules/*' \
    ! -path '*/.terraform/*' \
    ! -path '*/.*' \
  | create_zip "iac-source_${V}.zip" "IaC source"

# ─── 3. Catalog Pages ────────────────────────────────────────────────────────
echo "3/4  Catalog pages..."
find "$PROVIDER_BASE" -path '*/v1/catalog-page.md' \
  | create_zip "catalog-pages_${V}.zip" "catalog pages"

# ─── 4. Proto Source ──────────────────────────────────────────────────────────
echo "4/4  Proto source..."
find "$PROVIDER_BASE" \( \
    -path '*/v1/spec.proto' \
    -o -path '*/v1/api.proto' \
    -o -path '*/v1/stack_input.proto' \
    -o -path '*/v1/stack_outputs.proto' \
  \) | create_zip "proto-source_${V}.zip" "proto source"

echo ""
echo "=== Done ==="
