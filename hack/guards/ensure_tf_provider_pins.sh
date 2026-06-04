#!/usr/bin/env bash
set -euo pipefail

# Guard: every OpenTofu/Terraform deployment module must PIN every provider it uses
# via a `required_providers` block. A module that references a provider's resources/data
# sources without declaring that provider lets `tofu init` resolve it to the registry's
# latest MAJOR -- which is exactly how the helm-provider v3 break reached production
# (KubernetesExternalDns had no required_providers block, so `init` floated hashicorp/helm
# to v3, whose schema rejects the v2 `set {}` block). See
# _changelog/2026-06/2026-06-04-191500-helm-provider-v3-migration-and-externaldns-parity.md.
#
# WHAT IT CHECKS
# For each `apis/**/v1/iac/tf` module (a dir containing *.tf files):
#   - collect the provider local names referenced by `resource "<name>_..."` /
#     `data "<name>_..."` (the prefix before the first underscore; e.g. helm_release ->
#     helm, kubernetes_manifest -> kubernetes, random_password -> random),
#   - collect the provider local names declared inside `required_providers { ... }`,
#   - fail if any referenced provider is not declared.
# The builtin `terraform_*` data sources (terraform_remote_state) need no pin and are ignored.
#
# This is a static check (no network, no cluster, no credentials), so it covers every
# module -- including e2e `skip`/`deferred` components that never run a real apply.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

provider_base="apis/org/openmcf/provider"

violations=()

# Extracts the provider local-name keys declared inside required_providers blocks,
# tracking brace depth so only top-level entries ("<name> = {") at the providers level
# are captured (not nested source/version lines or sibling terraform{} settings).
extract_declared() {
  awk '
    inrp==0 && /required_providers[[:space:]]*\{/ { inrp=1; depth=1; next }
    inrp==1 {
      if (depth==1 && match($0, /^[[:space:]]*[A-Za-z0-9_-]+[[:space:]]*=[[:space:]]*\{/)) {
        k=$0; sub(/^[[:space:]]*/,"",k); sub(/[[:space:]]*=.*/,"",k); print k
      }
      o=$0; ob=gsub(/\{/,"x",o); c=$0; cb=gsub(/\}/,"x",c); depth += ob - cb
      if (depth<=0) inrp=0
    }
  ' "$@" | sort -u
}

# Collects referenced provider local names from resource/data block headers.
extract_referenced() {
  grep -hoE '(resource|data)[[:space:]]+"[a-z0-9]+_' "$@" 2>/dev/null \
    | sed -E 's/.*"([a-z0-9]+)_/\1/' \
    | sort -u || true
}

while IFS= read -r tfdir; do
  [[ -z "$tfdir" ]] && continue
  mapfile -t tffiles < <(find "$tfdir" -maxdepth 1 -type f -name '*.tf' 2>/dev/null)
  [[ ${#tffiles[@]} -eq 0 ]] && continue

  declared="$(extract_declared "${tffiles[@]}")"
  referenced="$(extract_referenced "${tffiles[@]}")"

  missing=()
  while IFS= read -r p; do
    [[ -z "$p" || "$p" == "terraform" ]] && continue
    if ! printf '%s\n' "$declared" | grep -qx "$p"; then
      missing+=("$p")
    fi
  done <<< "$referenced"

  if [[ ${#missing[@]} -gt 0 ]]; then
    violations+=("${tfdir} -> unpinned: ${missing[*]}")
  fi
done < <(find "$provider_base" -type d -path "*/v1/iac/tf" 2>/dev/null | sort)

if [[ ${#violations[@]} -gt 0 ]]; then
  echo "ERROR: ${#violations[@]} Terraform module(s) reference a provider without pinning it in required_providers." >&2
  echo "Unpinned providers let 'tofu init' float to the registry's latest major and can break on a provider release." >&2
  echo "Add the provider to a 'terraform { required_providers { ... } }' block (pin the major, e.g. helm \"~> 3.0\"):" >&2
  printf '  - %s\n' "${violations[@]}" >&2
  echo >&2
  echo "Terraform provider-pin guard FAILED." >&2
  exit 1
fi

echo "Terraform provider-pin guard passed: every iac/tf module pins all referenced providers."
