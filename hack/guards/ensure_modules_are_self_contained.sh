#!/usr/bin/env bash
set -euo pipefail

# Guard: a deployment module is its directory and reads nothing outside it.
#
# WHY THIS EXISTS
# Every module runtime hands the module exactly its own engine directory and
# nothing beside it: release packaging zips iac/tf and iac/pulumi each on
# their own (tools/ci/release), the OpenTofu runtime and the Pulumi source
# lane extract that zip into a fresh cache folder, and the Pulumi binary lane
# runs a compiled module from a generated workspace that holds only
# Pulumi.yaml. A module that reads a sibling path -- `${path.module}/..` in
# HCL, a "../" string in Go -- finds nothing there in every published form:
# `fileset()` returns empty and `for_each` plans ZERO resources, or the Go
# read fails at apply. Neither failure shows up in a working tree, where the
# sibling path happens to exist. This guard turns that class of defect into
# a static, pre-merge failure.
#
# WHAT IT CHECKS
#   1. catalog/<provider>/<kind>/iac/tf/*.tf -- no `path.module}/..` and no
#      string literal beginning with "../".
#   2. catalog/<provider>/<kind>/iac/pulumi/**/*.go, test files excepted --
#      no parent-path string literal: "../<anything>" or exactly "..". (The
#      Go ellipsis "..." in error text is not a path and is not matched.)
#
# WHAT IT DOES NOT CHECK, ON PURPOSE
# A module's _test.go files. A test runs only in the repository, where the
# sibling engine's directory exists, and never in a published form; a
# cross-engine twin test (the Pulumi module holding its OpenTofu twin's
# template text equal) has to read the other engine's directory to do its
# job.
# Reads of paths the OPERATOR supplies at apply time (a manifest annotation
# naming a docker-config file on the operator's own machine) are a
# break-glass feature for experts applying a module from a laptop. They read
# the operator's disk, never the module's directory, and are outside this
# invariant.
#
# KNOWN VIOLATIONS
# The table is empty: every module reads only its own directory. It exists
# so a module that must temporarily read outside itself can be listed while
# it is moved onto a self-contained shape. The list only shrinks: a listed
# module that no longer violates FAILS this guard until its entry is
# removed, so the table can never go stale, and an unlisted violation fails
# immediately. Set SELF_CONTAINED_GUARD_IGNORE_KNOWN=1 to see the raw result
# with the table disabled (the proof that the guard sees what it claims).

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

known_violations=()
if [[ "${SELF_CONTAINED_GUARD_IGNORE_KNOWN:-0}" == "1" ]]; then
  known_violations=()
fi

is_known() {
  local dir="$1"
  local k
  for k in "${known_violations[@]+"${known_violations[@]}"}"; do
    [[ "$k" == "$dir" ]] && return 0
  done
  return 1
}

# Violating module directories, one per line, with the first offending line
# kept for the report. Portable (macOS bash 3.2: no mapfile, no assoc arrays).
violations=""
while IFS= read -r line; do
  [[ -z "$line" ]] && continue
  file="${line%%:*}"
  dir="$(printf '%s\n' "$file" | sed -E 's#^(catalog/[^/]+/[^/]+/iac/(tf|pulumi))/.*#\1#')"
  case "$violations" in
    *"|$dir|"*) ;;
    *) violations="${violations}|$dir|"$'\n'"${line}"$'\n' ;;
  esac
done < <(
  {
    grep -rnE --include='*.tf' 'path\.module\}/\.\.|"\.\./' catalog/*/*/iac/tf 2>/dev/null || true
    grep -rnE --include='*.go' --exclude='*_test.go' '"\.\./|"\.\."' catalog/*/*/iac/pulumi 2>/dev/null || true
  } | grep -E '^catalog/[^/]+/[^/]+/iac/(tf|pulumi)/' || true
)

violating_dirs="$(printf '%s' "$violations" | grep -oE '^\|[^|]+\|$' | tr -d '|' | sort -u || true)"

unexpected=()
stale=()

while IFS= read -r dir; do
  [[ -z "$dir" ]] && continue
  if ! is_known "$dir"; then
    detail="$(printf '%s' "$violations" | grep -A1 -F "|$dir|" | tail -n 1)"
    unexpected+=("$dir -- $detail")
  fi
done <<< "$violating_dirs"

for k in "${known_violations[@]+"${known_violations[@]}"}"; do
  if ! printf '%s\n' "$violating_dirs" | grep -qx "$k"; then
    stale+=("$k")
  fi
done

failed=0

if [[ ${#unexpected[@]} -gt 0 ]]; then
  echo "ERROR: ${#unexpected[@]} module(s) read outside their own directory." >&2
  echo "Observed: a parent-path read in the module source (first offending line shown)." >&2
  echo "Meaning: every published form of the module (module.zip, source.zip, the Pulumi binary workspace) lacks that path, so the read returns nothing or fails at apply while the working tree passes." >&2
  echo "Next step: derive the payload from the pinned artifact at apply time (render the pinned chart or fetch the pinned bundle) instead of reading a sibling file; see the CRD decision tree in _rules/catalog-kind/forge/forge-catalog-kind.mdc." >&2
  printf '  - %s\n' "${unexpected[@]}" >&2
  echo >&2
  failed=1
fi

if [[ ${#stale[@]} -gt 0 ]]; then
  echo "ERROR: ${#stale[@]} known_violations entry(ies) no longer violate." >&2
  echo "Observed: the module is self-contained but still listed in this guard's known_violations table." >&2
  echo "Meaning: the table would silently excuse a future regression in that module." >&2
  echo "Next step: remove the entry from known_violations in hack/guards/ensure_modules_are_self_contained.sh." >&2
  printf '  - %s\n' "${stale[@]}" >&2
  echo >&2
  failed=1
fi

if [[ $failed -ne 0 ]]; then
  echo "Module self-containment guard FAILED." >&2
  exit 1
fi

if [[ ${#known_violations[@]} -gt 0 ]]; then
  echo "Module self-containment guard passed: no module outside the known_violations table reads beyond its own directory (${#known_violations[@]} known, shrinking)."
else
  echo "Module self-containment guard passed: every iac/tf and iac/pulumi module reads only its own directory."
fi
