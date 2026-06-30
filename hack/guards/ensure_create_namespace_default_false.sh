#!/usr/bin/env bash
set -euo pipefail

# Guard: a Kubernetes iac/tf module must NOT default `create_namespace` to true.
#
# WHY: the proto field is a plain proto3 `bool create_namespace` whose zero value is
# false, and the proto->tfvars converter marshals with EmitUnpopulated:false, so an
# unset/false value is OMITTED from the generated tfvars. If the TF variable then
# defaults to true (`optional(bool, true)`), the module creates the namespace on every
# unset deploy AND makes an explicit `false` unrepresentable (it is always dropped, so
# TF always sees true). That is exactly how KubernetesTemporal failed with
# 'namespaces "<ns>" already exists'. The canonical default-false shape is plain bool +
# `optional(bool, false)`: false == unset (no create), explicit true still creates.
#
# If a kind genuinely needs to default to CREATING the namespace, express that in the
# proto via `optional bool create_namespace [(dev.planton.shared.options.default) = "true"]`
# and let the manifest defaults applier (internal/manifest/protodefaults) inject it --
# the TF default must still stay false, because the applier populates the field
# regardless of the TF fallback. Never encode default-true as a TF-only fallback.
#
# This is a static check (no network, no cluster), so it covers every module including
# e2e skip/deferred components that never run a real apply.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

provider_base="apis/dev/planton/provider"

violations=()

while IFS= read -r varsfile; do
  [[ -z "$varsfile" ]] && continue
  # Match `create_namespace = optional(bool, true)` tolerant of surrounding whitespace.
  if grep -Eq 'create_namespace[[:space:]]*=[[:space:]]*optional\(bool,[[:space:]]*true\)' "$varsfile"; then
    violations+=("$varsfile")
  fi
done < <(find "$provider_base" -type f -path "*/v1/iac/tf/variables.tf" 2>/dev/null | sort)

if [[ ${#violations[@]} -gt 0 ]]; then
  echo "ERROR: ${#violations[@]} iac/tf module(s) default create_namespace to true." >&2
  echo "An unset/false proto3 bool is omitted from tfvars, so a TF default of true creates the namespace" >&2
  echo "on every deploy and makes an explicit false unrepresentable. Use optional(bool, false):" >&2
  printf '  - %s\n' "${violations[@]}" >&2
  echo >&2
  echo "create_namespace default-false guard FAILED." >&2
  exit 1
fi

echo "create_namespace default-false guard passed: no iac/tf module defaults create_namespace to true."
