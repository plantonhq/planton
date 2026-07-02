#!/usr/bin/env bash
set -euo pipefail

# Guard: every infra chart under charts/<provider>/<name>/ must be structurally
# complete so it is discoverable and loadable downstream.
#
# WHY THIS EXISTS
# A chart is discovered and loaded by the InfraChart project contract (Chart.yaml
# + values.yaml + a templates/ directory) and, once released as the catalog
# bundle, seeded into local instances by plantond. A chart missing any of those
# pieces silently drops out of discovery (or fails to load) instead of failing
# loudly, so a structurally broken chart can ship unnoticed. This guard enforces
# the contract at PR time.
#
# SCOPE
# These are grep-based PRESENCE checks (mirroring the sibling guards that grep for
# `package main`), NOT full YAML-schema or template-render validation. Deep
# render/validation is the authoritative Platform build's job and is deliberately
# not duplicated here -- this guard needs no toolchain beyond find + grep.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

charts_base="charts"

missing_chart_yaml=()
missing_values_yaml=()
missing_templates=()
bad_chart_identity=()
missing_params=()

if [[ -d "$charts_base" ]]; then
  # A chart is any directory that contains a Chart.yaml (excluding build artifacts).
  while IFS= read -r chart_yaml; do
    [[ -z "$chart_yaml" ]] && continue
    chart_dir="$(dirname "$chart_yaml")"

    # 1) values.yaml must exist (chartproject.Load reads it; its absence fails load).
    if [[ ! -f "$chart_dir/values.yaml" ]]; then
      missing_values_yaml+=("$chart_dir")
    else
      # 4) values.yaml must declare a params block.
      if ! grep -Eq '^params:' "$chart_dir/values.yaml"; then
        missing_params+=("$chart_dir")
      fi
    fi

    # 2) A non-empty templates/ directory must exist (required by IsProject discovery).
    if [[ ! -d "$chart_dir/templates" ]] || [[ -z "$(find "$chart_dir/templates" -maxdepth 1 -type f -name '*.yaml' -print -quit 2>/dev/null)" ]]; then
      missing_templates+=("$chart_dir")
    fi

    # 3) Chart.yaml must carry the InfraChart identity (apiVersion + kind).
    if ! grep -Eq '^apiVersion:[[:space:]]*infra-hub\.planton\.ai/v1[[:space:]]*$' "$chart_yaml" \
       || ! grep -Eq '^kind:[[:space:]]*InfraChart[[:space:]]*$' "$chart_yaml"; then
      bad_chart_identity+=("$chart_dir")
    fi
  done < <(find "$charts_base" -type f -name 'Chart.yaml' -not -path '*/build/*' 2>/dev/null | sort)

  # A directory that looks like a chart (has values.yaml or templates/) but has no
  # Chart.yaml is also broken -- it will never be discovered as a project.
  while IFS= read -r values_yaml; do
    [[ -z "$values_yaml" ]] && continue
    chart_dir="$(dirname "$values_yaml")"
    if [[ ! -f "$chart_dir/Chart.yaml" ]]; then
      missing_chart_yaml+=("$chart_dir")
    fi
  done < <(find "$charts_base" -type f -name 'values.yaml' -not -path '*/build/*' 2>/dev/null | sort)
fi

status=0

if [[ ${#missing_chart_yaml[@]} -gt 0 ]]; then
  status=1
  echo "ERROR: ${#missing_chart_yaml[@]} chart dir(s) have a values.yaml but NO Chart.yaml (never discovered as a project):" >&2
  printf '  - %s\n' "${missing_chart_yaml[@]}" >&2
  echo >&2
fi

if [[ ${#missing_values_yaml[@]} -gt 0 ]]; then
  status=1
  echo "ERROR: ${#missing_values_yaml[@]} chart(s) are missing values.yaml (chartproject.Load will fail):" >&2
  printf '  - %s\n' "${missing_values_yaml[@]}" >&2
  echo >&2
fi

if [[ ${#missing_templates[@]} -gt 0 ]]; then
  status=1
  echo "ERROR: ${#missing_templates[@]} chart(s) have no non-empty templates/ directory (required for discovery):" >&2
  printf '  - %s\n' "${missing_templates[@]}" >&2
  echo >&2
fi

if [[ ${#bad_chart_identity[@]} -gt 0 ]]; then
  status=1
  echo "ERROR: ${#bad_chart_identity[@]} chart(s) have a Chart.yaml missing the InfraChart identity" >&2
  echo "(expected 'apiVersion: infra-hub.planton.ai/v1' and 'kind: InfraChart'):" >&2
  printf '  - %s\n' "${bad_chart_identity[@]}" >&2
  echo >&2
fi

if [[ ${#missing_params[@]} -gt 0 ]]; then
  status=1
  echo "ERROR: ${#missing_params[@]} chart(s) have a values.yaml with no 'params:' block:" >&2
  printf '  - %s\n' "${missing_params[@]}" >&2
  echo >&2
fi

if [[ $status -ne 0 ]]; then
  echo "Chart structure guard FAILED. See errors above." >&2
  exit 1
fi

echo "Chart structure guard passed: every chart has Chart.yaml, values.yaml (with params), and a non-empty templates/ directory."
