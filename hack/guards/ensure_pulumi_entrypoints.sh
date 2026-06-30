#!/usr/bin/env bash
set -euo pipefail

# Guard: every Pulumi deployment component must have a buildable entrypoint at the
# `iac/pulumi/` ROOT (a `package main` file directly in that directory), and must NOT
# place the entrypoint in a `main/` or `entrypoint/` subdirectory.
#
# WHY THIS EXISTS
# The release pipeline (.github/workflows/release.pulumi-modules.yaml) builds each component
# NON-RECURSIVELY: `go build -o <bin> ./apis/dev/planton/provider/<p>/<c>/v1/iac/pulumi`.
# That command REQUIRES a `package main` at the directory root and fails with
# `no Go files in .../iac/pulumi` when the entrypoint is missing or misplaced. A recursive
# `go build ./.../v1/...` would mask this by compiling only the `module/` library. This guard
# enforces the exact release contract so a broken entrypoint is caught at PR time, not release.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

provider_base="apis/dev/planton/provider"

missing_root_main=()
misplaced_subdir=()

if [[ -d "$provider_base" ]]; then
  while IFS= read -r pulumi_dir; do
    [[ -z "$pulumi_dir" ]] && continue

    # 1) A `package main` Go file must exist directly at the iac/pulumi root.
    has_root_main="no"
    while IFS= read -r go_file; do
      if grep -Eq '^[[:space:]]*package[[:space:]]+main[[:space:]]*$' "$go_file"; then
        has_root_main="yes"
        break
      fi
    done < <(find "$pulumi_dir" -maxdepth 1 -type f -name '*.go' 2>/dev/null)

    if [[ "$has_root_main" == "no" ]]; then
      missing_root_main+=("$pulumi_dir")
    fi

    # 2) No divergent entrypoint subdirectory. Only flag a subdir that actually CONTAINS
    #    files -- an empty leftover dir is harmless (and git does not track empty dirs, so it
    #    would never reach CI anyway). A non-empty main/ or entrypoint/ is a misplaced entrypoint.
    for subdir in "$pulumi_dir/main" "$pulumi_dir/entrypoint"; do
      if [[ -d "$subdir" ]] && [[ -n "$(find "$subdir" -mindepth 1 -print -quit 2>/dev/null)" ]]; then
        misplaced_subdir+=("$subdir")
      fi
    done
  done < <(find "$provider_base" -type d -path "*/v1/iac/pulumi" 2>/dev/null | sort)
fi

status=0

if [[ ${#missing_root_main[@]} -gt 0 ]]; then
  status=1
  echo "ERROR: ${#missing_root_main[@]} Pulumi component(s) have NO 'package main' at the iac/pulumi root." >&2
  echo "The release build (go build ./<...>/v1/iac/pulumi) will fail with 'no Go files'." >&2
  echo "Create iac/pulumi/main.go (package main) per forge rule 010-pulumi-entrypoint:" >&2
  printf '  - %s\n' "${missing_root_main[@]}" >&2
  echo >&2
fi

if [[ ${#misplaced_subdir[@]} -gt 0 ]]; then
  status=1
  echo "ERROR: ${#misplaced_subdir[@]} Pulumi component(s) place the entrypoint in a 'main/' or 'entrypoint/' subdir." >&2
  echo "The release build is non-recursive and only consumes the iac/pulumi root package." >&2
  echo "Move the entrypoint files up to iac/pulumi/ and delete the subdir:" >&2
  printf '  - %s\n' "${misplaced_subdir[@]}" >&2
  echo >&2
fi

if [[ $status -ne 0 ]]; then
  echo "Pulumi entrypoint guard FAILED. See errors above." >&2
  exit 1
fi

echo "Pulumi entrypoint guard passed: every component has a root 'package main' and no misplaced entrypoint subdir."
