#!/usr/bin/env bash
# =============================================================================
# Package Planton content for distribution via Cloudflare R2.
#
# Creates six zip files, each scoped to a single concern:
#
#   presets.zip           -- Preset YAML + MD files, kind enum proto
#   iac-source.zip        -- IaC source (.go, .tf, .md, .yaml under iac/) plus
#                            the provider import catalogs (aa_import/catalog.yaml
#                            -- read together with each kind's iac/import-map.yaml)
#   catalog-pages.zip     -- each catalog kind root's catalog.md
#   proto-source.zip      -- Raw proto source (spec, api, input, outputs)
#   reference-pack.zip    -- The kind reference pack: generated reference
#                            pages, catalog indexes, the cross-reference graph,
#                            the commons page, the authored GUIDE.md / patterns
#                            wisdom layer, and the verified fact-sheet layer
#                            (per-kind cost/controls/permissions sidecars,
#                            the generated per-preset cost estimates, and the
#                            central compliance catalog + framework crosswalks)
#   conversion-corpus.zip -- The golden conversion corpus: every authored
#                            ConversionSpec plus its testdata fixtures. Gates
#                            every conversion engine (in-repo and external) --
#                            engine certification data, not user catalog
#                            content, so its transport is download + verify
#                            like every other release artifact.
#
# All zips preserve repo-relative paths so they can be extracted into a single
# directory and overlay into a virtual Planton root. Consumers like the Planton
# upgrade scripts use this merged directory as --planton-path or PLANTON_ROOT.
#
# The version tag is accepted as an argument for logging purposes only; zip
# filenames are version-free because the version is encoded in the R2 path
# (releases/{tag}/content/{name}.zip).
#
# An optional zip name limits the run to that single artifact -- the release
# workflow builds each zip in its own job so every artifact is visible by
# name in the Actions UI. Without a target, all zips are built (local/dev
# behavior). Every guard (_test refusal, empty-selection failure) applies
# per zip either way.
#
# Usage:
#   bash tools/ci/release/package_content.sh <version-tag> [zip-name] [--dry-run]
#   bash tools/ci/release/package_content.sh v0.3.50
#   bash tools/ci/release/package_content.sh v0.3.50 --dry-run
#   bash tools/ci/release/package_content.sh v0.3.50 reference-pack
# =============================================================================

set -euo pipefail

VERSION="${1:?Usage: package_content.sh <version-tag> [zip-name] [--dry-run]}"
shift

TARGET=""
DRY_RUN=""
for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN="--dry-run" ;;
    presets|iac-source|catalog-pages|proto-source|reference-pack|conversion-corpus) TARGET="$arg" ;;
    *)
      echo "ERROR: Unknown argument: $arg"
      echo "       Valid zip names: presets, iac-source, catalog-pages, proto-source, reference-pack, conversion-corpus"
      exit 1
      ;;
  esac
done

# True when the named zip should be built in this run.
wants() { [ -z "$TARGET" ] || [ "$TARGET" = "$1" ]; }

REPO_ROOT="$(cd "$(dirname "$0")/../../.." && pwd)"
cd "$REPO_ROOT"

CATALOG_ROOT="catalog"

if [ ! -d "$CATALOG_ROOT" ]; then
  echo "ERROR: Catalog root directory not found: $CATALOG_ROOT"
  exit 1
fi

echo "=== Packaging Planton content for ${VERSION} ==="
echo ""

create_zip() {
  local zip_name="$1"
  local description="$2"
  local test_policy="${3:-}"
  shift 2

  local tmp_list
  tmp_list=$(mktemp)

  # Read file paths from stdin into a sorted temp file.
  sort > "$tmp_list"

  local count
  count=$(wc -l < "$tmp_list" | tr -d ' ')

  # The _test provider is permanent internal test infrastructure ("never
  # shipped to users"). Refusing it HERE -- not only in the selectors --
  # means a future selector edit cannot quietly start shipping it.
  # The single sanctioned exception is the conversion corpus (see its
  # section): certification fixtures whose subject IS the test kind.
  if [ "$test_policy" != "allow-test-fixtures" ] && grep -q "/_test/" "$tmp_list"; then
    echo "  ERROR: ${zip_name} selected _test provider content, which must"
    echo "         never ship to users:"
    grep "/_test/" "$tmp_list" | head -5 | sed 's/^/           /'
    rm -f "$tmp_list"
    exit 1
  fi

  if [ "$count" -eq 0 ]; then
    # An empty selection means a selector pattern no longer matches the tree
    # (e.g. an api-version directory rename) -- shipping a release with a
    # silently empty content zip is worse than failing the release.
    echo "  ERROR: No files found for ${description} (${zip_name}). A selector"
    echo "         pattern no longer matches the repository layout; fix the"
    echo "         pattern before releasing."
    rm -f "$tmp_list"
    exit 1
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

# ─── Presets ─────────────────────────────────────────────────────────────────
# Presets live at the kind root (catalog/{provider}/{kind}/presets/).
if wants presets; then
  echo "Presets..."
  {
    find "$CATALOG_ROOT" \( -path '*/presets/*.yaml' -o -path '*/presets/*.md' \) ! -path '*/_test/*'
    echo "shared/catalogkind/catalog_kind.proto"
  } | create_zip "presets.zip" "presets"
fi

# ─── IaC Source ──────────────────────────────────────────────────────────────
# IaC modules live at the kind root (catalog/{provider}/{kind}/iac/).
# Mirrors the ALLOWED_EXTENSIONS in iac-bundler.ts: .go, .tf, .md, .yaml
# Excludes hidden dirs, vendor, and node_modules (same as iac-bundler.ts).
# The provider-level import catalogs (catalog/{provider}/aa_import/catalog.yaml)
# ride along: consumers resolve each kind's iac/import-map.yaml against them,
# so an artifact root carrying one without the other cannot answer imports.
if wants iac-source; then
  echo "IaC source..."
  {
    find "$CATALOG_ROOT" -path '*/iac/*' ! -path '*/_test/*' \
        \( -name '*.go' -o -name '*.tf' -o -name '*.md' -o -name '*.yaml' \) \
        ! -path '*/vendor/*' \
        ! -path '*/node_modules/*' \
        ! -path '*/.terraform/*' \
        ! -path '*/.*'
    find "$CATALOG_ROOT" -path '*/aa_import/*' ! -path '*/_test/*' -name 'catalog.yaml'
  } | create_zip "iac-source.zip" "IaC source"
fi

# ─── Catalog Pages ───────────────────────────────────────────────────────────
# The catalog page lives at the kind root (catalog/{provider}/{kind}/catalog.md);
# the anchored grep pins exactly that depth so provider-level or nested
# markdown can never leak into the artifact.
if wants catalog-pages; then
  echo "Catalog pages..."
  find "$CATALOG_ROOT" -type f -name 'catalog.md' ! -path '*/_test/*' \
    | grep -E "^${CATALOG_ROOT}/[^/]+/[^/]+/catalog\.md$" \
    | create_zip "catalog-pages.zip" "catalog pages"
fi

# ─── Proto Source ────────────────────────────────────────────────────────────
# Protos are the versioned contract and stay in version dirs; the grammar
# matches any maturity channel (v1alpha1, v1beta1, v1, ...).
if wants proto-source; then
  echo "Proto source..."
  find "$CATALOG_ROOT" -type f \( \
      -name 'spec.proto' \
      -o -name 'api.proto' \
      -o -name 'input.proto' \
      -o -name 'outputs.proto' \
    \) ! -path '*/_test/*' \
    | grep -E "/v[0-9]+((alpha|beta)[0-9]+)?/[a-z_]+\.proto$" \
    | create_zip "proto-source.zip" "proto source"
fi

# ─── Reference Pack ──────────────────────────────────────────────────────────
# The pack is selected by file NAME, never by version-segment path: these
# names are the pack's frozen public contract (a contributor-edited file is
# the same file an agent reads -- no renames between source and artifact),
# pinned by the reference generator's contract tests. Name-based selection
# also survives api-version directory renames, which path patterns would not.
#
# The fact-sheet layer rides the same pack: the per-kind sidecar names
# (cost.yaml, controls.yaml, iac/permissions.yaml) are anatomy-enforced and
# unique to fact-sheets, and the central trees (_pricing/estimates/ for the
# generated per-preset dollar estimates, _compliance/ for the control catalog
# and framework crosswalks) are selected by path like _patterns/. Coverage is
# presence-based -- only covered kinds carry sidecars -- so absence in
# the pack means "not yet published", never zero.
#
# This selection MUST stay mirrored with the self-contained catalog skill's
# assembly (pkg/skills/defspack/catalogpack.go) -- the skill and this zip are
# the same pack on two transports.
if wants reference-pack; then
  echo "Reference pack..."
  find "$CATALOG_ROOT" \( \
      -name 'reference.md' \
      -o -name 'GUIDE.md' \
      -o -name 'reference-index.md' \
      -o -name 'reference-graph.yaml' \
      -o -name 'reference-commons.md' \
      -o -name 'cost.yaml' \
      -o -name 'controls.yaml' \
      -o -name 'permissions.yaml' \
      -o -path "$CATALOG_ROOT/_patterns/*.md" \
      -o -path "$CATALOG_ROOT/_pricing/estimates/*.yaml" \
      -o -path "$CATALOG_ROOT/_compliance/*.yaml" \
    \) ! -path '*/_test/*' | create_zip "reference-pack.zip" "reference pack"
fi

# ─── Conversion Corpus ───────────────────────────────────────────────────────
# The golden corpus lives beside each kind (catalog/{provider}/{kind}/
# conversions/): the authored specs plus their testdata fixtures. Every
# conversion engine -- in this repo or outside it -- proves itself against
# this one corpus, so it ships as a verifiable artifact instead of forcing
# consumers to clone the repository at a tag. It is engine-certification
# data, not user catalog content, which is why it is the one artifact
# allowed to carry the _test kind: the torture kind's fixtures ARE the
# corpus (the never-ships boundary keeps the kind out of every SERVING
# surface; certification data is not one).
if wants conversion-corpus; then
  echo "Conversion corpus..."
  find "$CATALOG_ROOT" -type f -path '*/conversions/*' \
    | grep -E "^${CATALOG_ROOT}/[^/]+/[^/]+/conversions/" \
    | create_zip "conversion-corpus.zip" "conversion corpus" allow-test-fixtures
fi

echo ""
echo "=== Done ==="
