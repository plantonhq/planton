#!/usr/bin/env bash
set -euo pipefail

# Guard: the _test provider never reaches a user surface.
#
# WHY THIS EXISTS
# The _test provider carries the full per-kind file shape (protos, IaC
# modules, presets, docs) ON PURPOSE -- it is the standing canary that
# exercises every pipeline real kinds flow through. That same completeness is
# what makes it shippable by accident: any selector that gathers per-kind
# content by path shape will happily gather _test content too. Its absence
# from user surfaces used to rest on the kinds simply lacking files; this
# guard makes the boundary EXPLICIT and CI-enforced:
#
#   1. Release content packaging (tools/ci/release/package_content.sh)
#      excludes _test in every selector AND refuses _test paths inside
#      create_zip itself. The dry-run below exercises both.
#   2. Module auto-tagging skips _test kinds (auto-tag.yaml).
#
# What _test content legitimately reaches: the kind registry, generated
# stubs, and test/certification surfaces -- exactly what needs it.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

failures=()

# 1. The packager's own runtime boundary, exercised end to end: --dry-run
# runs every selector and create_zip's _test refusal without writing zips.
if ! bash tools/ci/release/package_content.sh vGUARD --dry-run >/dev/null 2>&1; then
  failures+=("tools/ci/release/package_content.sh --dry-run failed -- a content selector matched _test provider content (or no longer matches the tree)")
fi

# 2. Module auto-tagging must skip _test kinds.
if ! grep -q '"_test"' .github/workflows/auto-tag.yaml; then
  failures+=(".github/workflows/auto-tag.yaml no longer skips the _test provider")
fi

if [ ${#failures[@]} -gt 0 ]; then
  echo "The _test provider must never reach a user surface:"
  for f in "${failures[@]}"; do
    echo "  - $f"
  done
  exit 1
fi

echo "OK: _test provider stays internal (packaging, auto-tagging)."
