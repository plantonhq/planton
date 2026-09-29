#!/usr/bin/env bash
set -euo pipefail

# Guard: a kind's two IaC engines must tag the same Azure resource with the
# SAME identity tags -- the same keys AND the same values.
#
# The defect class this retires: the same AzureResourceGroup manifest was
# tagged `resource_kind = azure_resource_group` by OpenTofu and
# `resource_kind = azureresourcegroup` by Pulumi, and OpenTofu wrote
# `resource_id = <metadata.name>` when the resource had no id while Pulumi
# wrote no `resource_id` at all -- fleet-wide, one hand-spelled literal per
# kind. Anything that finds Planton's resources by tag (a cost report, an
# Azure Policy, a cleanup script, a portal query) found a different set
# depending on which engine ran, and switching engines re-tagged the resource.
#
# WHAT IT CHECKS (catalog/azure; AWS has ensure_cross_engine_tag_keys.sh)
# For each kind directory `catalog/azure/<kind>/` with an `iac/` tree:
#   1. The identity keys declared in `iac/tf/locals.tf` (the quoted keys of
#      the tag maps) and in `iac/pulumi/module/locals.go` (the azuretagkeys
#      constants) are the same six-key set -- or neither engine declares any
#      (a kind whose resources take no tags).
#   2. The Pulumi module names those keys through the shared azuretagkeys
#      constants, never as string literals.
#   3. The `resource_kind` value is the kind's CloudResourceKind enum name,
#      lowercased: OpenTofu writes it as that exact literal, Pulumi derives it
#      from that exact enum value. The enum is the source of truth, so a
#      hand-spelled literal (snake_case, a sibling's kind) cannot slip in.
#   4. `resource_id` is written only when metadata.id is set, never with the
#      name as a stand-in: OpenTofu as a conditional map over var.metadata.id,
#      Pulumi inside an `if ....Metadata.Id != ""` block.
#
# BOUNDARY: this guard reads the DECLARED tag maps, not rendered sends -- a
# module that builds the map and then drops it at a send site is the live
# proof's catch.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

provider_root="catalog/azure"
kind_enum_proto="shared/cloudresourcekind/cloud_resource_kind.proto"

canonical_keys="environment
organization
resource
resource_id
resource_kind
resource_name"

identity_key_pattern='resource|resource_name|resource_kind|resource_id|organization|environment'
constant_pattern='Resource|ResourceName|ResourceKind|ResourceId|Organization|Environment'

# Maps azuretagkeys constant names to the tag keys they render as
# (pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys).
constant_to_key() {
  case "$1" in
    Resource) echo "resource" ;;
    ResourceName) echo "resource_name" ;;
    ResourceKind) echo "resource_kind" ;;
    ResourceId) echo "resource_id" ;;
    Organization) echo "organization" ;;
    Environment) echo "environment" ;;
  esac
}

# Every Azure CloudResourceKind enum name, one per line.
azure_enum_names="$(grep -oE '^[[:space:]]*Azure[A-Za-z0-9]+[[:space:]]*=[[:space:]]*[0-9]+' "$kind_enum_proto" \
  | sed -E 's/^[[:space:]]*//; s/[[:space:]]*=.*$//')"

# The enum name whose lowercase is the kind directory's name.
enum_name_for_kind() {
  awk -v kind="$1" 'tolower($0) == kind { print; exit }' <<< "$azure_enum_names"
}

# OpenTofu identity keys: the quoted identity keys assigned in locals.tf,
# counted only when the file declares a resource_kind tag at all.
extract_tf_keys() {
  local f="$1"
  [[ -f "$f" ]] || return 0
  grep -qE '"resource_kind"[[:space:]]*=' "$f" || return 0
  { grep -hoE "\"(${identity_key_pattern})\"[[:space:]]*=" "$f" || true; } \
    | sed -E 's/^"//; s/"[[:space:]]*=$//' | sort -u
}

# Pulumi identity keys: the azuretagkeys constants referenced in locals.go,
# plus any literal identity keys (which check 2 then refuses), so a module
# still on literals reports its real key set rather than "none".
extract_pulumi_keys() {
  local f="$1"
  [[ -f "$f" ]] || return 0
  {
    { grep -hoE "azuretagkeys\.(${constant_pattern})\b" "$f" || true; } \
      | sed -E 's/^azuretagkeys\.//' \
      | while IFS= read -r c; do constant_to_key "$c"; done
    { grep -hoE "\"(${identity_key_pattern})\"[[:space:]]*(:|\])" "$f" || true; } \
      | sed -E 's/^"//; s/"[[:space:]]*(:|\])$//'
  } | sort -u
}

one_line() { echo $1 | tr '\n' ' ' | sed 's/ $//'; }

set_violations=()
shape_violations=()
kind_violations=()
id_violations=()
failing_kinds=()

while IFS= read -r kinddir; do
  kind="${kinddir#${provider_root}/}"
  [[ -d "${kinddir}/iac" ]] || continue
  tf_locals="${kinddir}/iac/tf/locals.tf"
  pulumi_locals="${kinddir}/iac/pulumi/module/locals.go"
  before=$(( ${#set_violations[@]} + ${#shape_violations[@]} + ${#kind_violations[@]} + ${#id_violations[@]} ))

  tf_keys="$(extract_tf_keys "$tf_locals")"
  pulumi_keys="$(extract_pulumi_keys "$pulumi_locals")"

  # 2. Shape: identity keys through the azuretagkeys constants only.
  if compgen -G "${kinddir}/iac/pulumi/module/*.go" > /dev/null; then
    if grep -qE "\"(${identity_key_pattern})\"[[:space:]]*(:|\])" "${kinddir}"/iac/pulumi/module/*.go; then
      shape_violations+=("${kind} -> literal identity tag keys in the Pulumi module; use the azuretagkeys constants")
    fi
  fi

  # 1. Same key set on both engines, or none on either.
  if [[ -z "$tf_keys" && -z "$pulumi_keys" ]]; then
    continue
  fi
  if [[ -z "$tf_keys" || -z "$pulumi_keys" ]]; then
    set_violations+=("${kind} -> identity tags declared by ONE engine only (tf: $(one_line "$tf_keys"); pulumi: $(one_line "$pulumi_keys"))")
  else
    [[ "$tf_keys" == "$canonical_keys" ]] \
      || set_violations+=("${kind} -> tf identity keys [$(one_line "$tf_keys")] are not the six-key set [$(one_line "$canonical_keys")]")
    [[ "$pulumi_keys" == "$canonical_keys" ]] \
      || set_violations+=("${kind} -> pulumi identity keys [$(one_line "$pulumi_keys")] are not the six-key set [$(one_line "$canonical_keys")]")
  fi

  # 3. resource_kind is the enum name lowercased, on both engines.
  enum_name="$(enum_name_for_kind "$kind")"
  if [[ -z "$enum_name" ]]; then
    kind_violations+=("${kind} -> no Azure CloudResourceKind enum value lowercases to the directory name")
  else
    expected_kind="$(echo "$enum_name" | tr '[:upper:]' '[:lower:]')"
    if [[ -n "$tf_keys" ]]; then
      tf_kind_values="$(grep -hE '"resource_kind"[[:space:]]*=' "$tf_locals" \
        | sed -E 's/^.*"resource_kind"[[:space:]]*=[[:space:]]*//; s/[[:space:]]*$//' | sort -u)"
      if [[ "$tf_kind_values" != "\"${expected_kind}\"" ]]; then
        kind_violations+=("${kind} -> tf resource_kind is $(one_line "$tf_kind_values"), expected the literal \"${expected_kind}\" (the enum name ${enum_name}, lowercased)")
      fi
    fi
    if [[ -n "$pulumi_keys" ]]; then
      pulumi_kind_values="$({ grep -hE "(azuretagkeys\.ResourceKind|\"resource_kind\")[[:space:]]*(:|\])" "$pulumi_locals" || true; } \
        | sed -E 's/^.*(azuretagkeys\.ResourceKind|"resource_kind")[[:space:]]*(:|\][[:space:]]*=)[[:space:]]*//; s/,?[[:space:]]*$//' | sort -u)"
      if [[ "$pulumi_kind_values" != "strings.ToLower(cloudresourcekind.CloudResourceKind_${enum_name}.String())" ]]; then
        kind_violations+=("${kind} -> pulumi resource_kind is $(one_line "$pulumi_kind_values"), expected strings.ToLower(cloudresourcekind.CloudResourceKind_${enum_name}.String())")
      fi
    fi
  fi

  # 4. resource_id only when metadata.id is set -- no name fallback.
  if [[ -n "$tf_keys" ]]; then
    # A `resource_id = ...` local is the name-fallback shape; every
    # "resource_id" entry must be the conditional map over var.metadata.id.
    if grep -qE '^[[:space:]]*resource_id[[:space:]]*=' "$tf_locals" \
        || grep -E '"resource_id"[[:space:]]*=' "$tf_locals" \
          | grep -vqE '\?[[:space:]]*\{[[:space:]]*"resource_id"[[:space:]]*=[[:space:]]*var\.metadata\.id[[:space:]]*\}[[:space:]]*:[[:space:]]*\{\}'; then
      id_violations+=("${kind} -> tf resource_id is not the conditional map ? { \"resource_id\" = var.metadata.id } : {} (a resource_id local falls back to the name)")
    fi
  fi
  if [[ -n "$pulumi_keys" ]]; then
    if grep -qE '(azuretagkeys\.ResourceId|"resource_id")[[:space:]]*:' "$pulumi_locals"; then
      id_violations+=("${kind} -> pulumi puts resource_id in the map literal; set it only when Metadata.Id is non-empty")
    fi
    unguarded="$(awk '
      /^[[:space:]]*$/ { next }
      /(azuretagkeys\.ResourceId|"resource_id")[[:space:]]*\][[:space:]]*=/ {
        guarded = (prev ~ /^[[:space:]]*if [A-Za-z0-9_.]+\.Metadata\.Id != "" \{[[:space:]]*$/) &&
                  ($0 ~ /=[[:space:]]*[A-Za-z0-9_.]+\.Metadata\.Id[[:space:]]*$/)
        if (!guarded) print NR
      }
      { prev = $0 }
    ' "$pulumi_locals")"
    if [[ -n "$unguarded" ]]; then
      id_violations+=("${kind} -> pulumi sets resource_id outside an if ...Metadata.Id != \"\" guard, or to something other than Metadata.Id (locals.go:$(one_line "$unguarded"))")
    fi
  fi

  after=$(( ${#set_violations[@]} + ${#shape_violations[@]} + ${#kind_violations[@]} + ${#id_violations[@]} ))
  [[ $after -gt $before ]] && failing_kinds+=("$kind")
done < <(find "$provider_root" -maxdepth 1 -mindepth 1 -type d | sort)

report() {
  local title="$1"; shift
  [[ $# -gt 0 ]] || return 0
  echo "ERROR: $# ${title}" >&2
  printf '  - %s\n' "$@" >&2
  echo >&2
}

report "finding(s): the engines declare different identity tag keys:" ${set_violations[@]+"${set_violations[@]}"}
report "finding(s): the Pulumi module spells identity tag keys as literals:" ${shape_violations[@]+"${shape_violations[@]}"}
report "finding(s): the resource_kind tag is not the lowercased enum name:" ${kind_violations[@]+"${kind_violations[@]}"}
report "finding(s): resource_id is written without an id, or with the name as a stand-in:" ${id_violations[@]+"${id_violations[@]}"}

if [[ ${#failing_kinds[@]} -gt 0 ]]; then
  echo "Cross-engine Azure tag guard FAILED: ${#failing_kinds[@]} kind(s) tag the same resource differently per engine." >&2
  exit 1
fi

echo "Cross-engine Azure tag guard passed: every kind's engines write the same identity tags, key for key and value for value."
