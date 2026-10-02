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
# For each `catalog/<provider>/<component>/iac/tf` module (a dir containing *.tf files):
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

catalog_root="catalog"

# Canonical catalog-wide pins.
#
# A cloud provider's modules unify on ONE pessimistic constraint, advanced
# only by the monthly pin sweep. Once unified, per-module drift is a defect:
# a lower floor can lie about what a module's arguments need, an unbounded
# floor lets upgrades flow in uncontrolled, and an off-style pin fragments
# the catalog's provider resolution. Each row is
# "<catalog dir>|<provider local name>|<exact required constraint>"; a cloud
# appends its row in the change that lands its unification sweep, and the
# monthly sweep updates the constraint here and in every module together.
canonical_pins=(
  "aws|aws|~> 6.58"
  "azure|azurerm|~> 5.0"
  # azapi is admission-only (raw ARM at a pinned type@api-version, per kind,
  # by recorded decision) and pins EXACT -- a floating raw-API provider would
  # move the ARM client under every admitted kind at once.
  "azure|azapi|2.11.0"
  "cloudflare|cloudflare|~> 5.23"
  "digitalocean|digitalocean|~> 2.99"
  "gcp|google|~> 8.3"
  # google-beta enters a module only under a recorded admission
  # (pkg/providerparity/admissions/google-beta.yaml, guarded by
  # ensure_beta_admissions.sh) and rides the same line as google so both
  # channels resolve one release.
  "gcp|google-beta|~> 8.3"
)

violations=()
pin_violations=()

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

# Extracts the version constraint declared for one provider local name inside
# required_providers blocks (same depth-tracking as extract_declared).
extract_version() {
  local want="$1"
  shift
  awk -v want="$want" '
    inrp==0 && /required_providers[ \t]*\{/ { inrp=1; depth=1; next }
    inrp==1 {
      if (depth==1 && match($0, /^[ \t]*[A-Za-z0-9_-]+[ \t]*=[ \t]*\{/)) {
        k=$0; sub(/^[ \t]*/,"",k); sub(/[ \t]*=.*/,"",k); inentry=(k==want)
      }
      if (inentry && match($0, /^[ \t]*version[ \t]*=[ \t]*"/)) {
        v=$0; sub(/^[ \t]*version[ \t]*=[ \t]*"/,"",v); sub(/".*/,"",v); print v
      }
      o=$0; ob=gsub(/\{/,"x",o); c=$0; cb=gsub(/\}/,"x",c); depth += ob - cb
      if (depth<=0) { inrp=0; inentry=0 }
    }
  ' "$@"
}

# Collects referenced provider local names from resource/data block headers.
extract_referenced() {
  grep -hoE '(resource|data)[[:space:]]+"[a-z0-9]+_' "$@" 2>/dev/null \
    | sed -E 's/.*"([a-z0-9]+)_/\1/' \
    | sort -u || true
}

while IFS= read -r tfdir; do
  [[ -z "$tfdir" ]] && continue
  # Portable array fill (mapfile is bash-4+; macOS ships bash 3.2).
  tffiles=()
  while IFS= read -r tf; do
    [[ -n "$tf" ]] && tffiles+=("$tf")
  done < <(find "$tfdir" -maxdepth 1 -type f -name '*.tf' 2>/dev/null)
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

  # Canonical-pin conformance (see the canonical_pins table above).
  cloud_dir="${tfdir#catalog/}"
  cloud_dir="${cloud_dir%%/*}"
  for rule in "${canonical_pins[@]}"; do
    rule_cloud="${rule%%|*}"
    rest="${rule#*|}"
    rule_name="${rest%%|*}"
    rule_pin="${rest#*|}"
    [[ "$cloud_dir" == "$rule_cloud" ]] || continue
    if printf '%s\n' "$declared" | grep -qx "$rule_name"; then
      actual="$(extract_version "$rule_name" "${tffiles[@]}" | head -n 1)"
      if [[ "$actual" != "$rule_pin" ]]; then
        pin_violations+=("${tfdir} -> ${rule_name} = \"${actual:-<no version>}\" (canonical: \"${rule_pin}\")")
      fi
    fi
  done
done < <(find "$catalog_root" -type d -path "*/iac/tf" 2>/dev/null | grep -E '^catalog/[^/]+/[^/]+/iac/tf$' | sort)

failed=0

if [[ ${#violations[@]} -gt 0 ]]; then
  echo "ERROR: ${#violations[@]} Terraform module(s) reference a provider without pinning it in required_providers." >&2
  echo "Unpinned providers let 'tofu init' float to the registry's latest major and can break on a provider release." >&2
  echo "Add the provider to a 'terraform { required_providers { ... } }' block (pin the major, e.g. helm \"~> 3.0\"):" >&2
  printf '  - %s\n' "${violations[@]}" >&2
  echo >&2
  failed=1
fi

if [[ ${#pin_violations[@]} -gt 0 ]]; then
  echo "ERROR: ${#pin_violations[@]} Terraform module(s) deviate from their cloud's canonical provider pin." >&2
  echo "Unified clouds carry ONE constraint in every module; only the monthly pin sweep changes it" >&2
  echo "(update the canonical_pins table in this guard and every module together):" >&2
  printf '  - %s\n' "${pin_violations[@]}" >&2
  echo >&2
  failed=1
fi

if [[ $failed -ne 0 ]]; then
  echo "Terraform provider-pin guard FAILED." >&2
  exit 1
fi

echo "Terraform provider-pin guard passed: every iac/tf module pins all referenced providers, and unified clouds carry their canonical pin."
