#!/usr/bin/env bash
set -euo pipefail

# Guard: a Terraform resource block attaches a SECONDARY-CHANNEL provider
# (`provider = google-beta`) only under a recorded admission, and every
# recorded admission is backed by a module that actually attaches it.
#
# WHY THIS EXISTS
# The catalog's parity baseline is the canonical (GA) provider. Some
# capability lives only in a second channel -- Google's beta provider carries
# pre-GA resources, whole product families for years -- and the doctrine
# admits it surgically: per resource, per kind, with a reason and the path by
# which its promotion to GA is watched, recorded in
# pkg/providerparity/admissions/<channel>.yaml. Before this guard the list
# was a sentence in the architecture doc, and a module could attach
# `provider = google-beta` with nothing checking. The Go accounting
# (pkg/providerparity) is the full gate -- it also knows which schema serves
# each resource, so it catches an admission the GA provider has since made
# unnecessary -- and it runs in the lint.provider-parity lane. This script is
# the static, dependency-free half that runs beside the provider-pin guard in
# the terraform-modules lane: no Go toolchain, no network, seconds.
#
# WHAT IT CHECKS
# For each `catalog/<provider>/<component>/iac/tf` module:
#   - collect every resource block that sets the `provider` meta-argument to a
#     secondary channel (a local name that is not the module's primary
#     provider prefix: `google-beta` in a module of google_* resources),
#   - fail if the (resource type, kind) pair has no entry in the admissions
#     file for that channel.
# Then, for each admissions entry, fail if no module of that kind attaches the
# resource through the channel -- an admission nobody uses is judgment nobody
# re-evaluates, and it would hide the day the resource stops needing the
# channel.
#
# The admissions file is read structurally (resource:/kind: lines under
# resources:) rather than through a YAML parser, so the guard has no
# dependencies; the Go loader enforces the file's full strict schema.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

catalog_root="catalog"
admissions_dir="pkg/providerparity/admissions"

# Secondary channels the catalog knows: "<channel local name>|<GA provider
# local name>". A resource block attaching through the channel is judged
# against <admissions_dir>/<channel>.yaml. Add a row when a catalog admits
# a new channel (the Go loader must know its schema artifact too).
channels=(
  "google-beta|google"
)

# admitted_pairs holds "<channel>|<resource>|<kind>" for every recorded
# admission; attached_pairs the same shape for every attachment found in
# modules. The two sets must be equal.
admitted_pairs=()
attached_pairs=()

# Reads "<resource>|<kind>" pairs from one admissions file: walks the
# resources: list, pairing each `- resource:` with the `kind:` that follows
# it inside the same entry. Tolerates block or flow ordering as long as
# resource precedes kind within an entry (the loader's own convention).
extract_admissions() {
  awk '
    /^resources:[[:space:]]*$/ { inres=1; next }
    inres && /^[^[:space:]-]/ { inres=0 }
    inres && /^[[:space:]]*-[[:space:]]*resource:[[:space:]]*/ {
      if (res != "") { print res "|" (kind == "" ? "?" : kind) }
      res=$0; sub(/^[[:space:]]*-[[:space:]]*resource:[[:space:]]*/, "", res); gsub(/["'"'"'[:space:]]/, "", res); kind=""; next
    }
    inres && /^[[:space:]]*kind:[[:space:]]*/ {
      kind=$0; sub(/^[[:space:]]*kind:[[:space:]]*/, "", kind); gsub(/["'"'"'[:space:]]/, "", kind); next
    }
    END { if (res != "") { print res "|" (kind == "" ? "?" : kind) } }
  ' "$1"
}

# Reads "<resource type>|<channel>" for every resource block in the given
# .tf files that sets `provider = <channel>` (root name of the traversal;
# `google-beta.alias` -> google-beta). Tracks brace depth so only the
# block's own top-level attributes count, never a nested block's.
extract_attachments() {
  awk '
    depth==0 && match($0, /^resource[[:space:]]+"[a-z0-9_]+"[[:space:]]+"[^"]+"[[:space:]]*\{/) {
      t=$0; sub(/^resource[[:space:]]+"/, "", t); sub(/".*/, "", t); rtype=t; depth=1; next
    }
    depth>=1 {
      if (depth==1 && match($0, /^[[:space:]]*provider[[:space:]]*=[[:space:]]*[A-Za-z0-9_-]+/)) {
        p=$0; sub(/^[[:space:]]*provider[[:space:]]*=[[:space:]]*/, "", p); sub(/[^A-Za-z0-9_-].*/, "", p); print rtype "|" p
      }
      o=$0; ob=gsub(/\{/,"x",o); c=$0; cb=gsub(/\}/,"x",c); depth += ob - cb
      if (depth<=0) { depth=0; rtype="" }
    }
  ' "$@"
}

for row in "${channels[@]}"; do
  channel="${row%%|*}"
  file="${admissions_dir}/${channel}.yaml"
  [[ -f "$file" ]] || continue
  while IFS= read -r pair; do
    [[ -z "$pair" ]] && continue
    admitted_pairs+=("${channel}|${pair}")
  done < <(extract_admissions "$file")
done

while IFS= read -r tfdir; do
  [[ -z "$tfdir" ]] && continue
  # Portable array fill (mapfile is bash-4+; macOS ships bash 3.2).
  tffiles=()
  while IFS= read -r tf; do
    [[ -n "$tf" ]] && tffiles+=("$tf")
  done < <(find "$tfdir" -maxdepth 1 -type f -name '*.tf' 2>/dev/null)
  [[ ${#tffiles[@]} -eq 0 ]] && continue

  # catalog/<provider>/<component>/iac/tf -> the kind is the component
  # directory in the registry's PascalCase; the admissions file names kinds
  # as the registry does, and directory names are the kind lower-cased, so
  # compare case-insensitively.
  component="${tfdir#catalog/*/}"
  component="${component%%/*}"

  while IFS= read -r att; do
    [[ -z "$att" ]] && continue
    rtype="${att%%|*}"
    channel="${att#*|}"
    for row in "${channels[@]}"; do
      [[ "${row%%|*}" == "$channel" ]] || continue
      attached_pairs+=("${channel}|${rtype}|${component}")
    done
  done < <(extract_attachments "${tffiles[@]}")
done < <(find "$catalog_root" -type d -path "*/iac/tf" 2>/dev/null | grep -E '^catalog/[^/]+/[^/]+/iac/tf$' | sort)

lower() { printf '%s' "$1" | tr '[:upper:]' '[:lower:]'; }

unadmitted=()
for att in "${attached_pairs[@]+"${attached_pairs[@]}"}"; do
  channel="${att%%|*}"; rest="${att#*|}"; rtype="${rest%%|*}"; component="${rest#*|}"
  found=0
  for adm in "${admitted_pairs[@]+"${admitted_pairs[@]}"}"; do
    achannel="${adm%%|*}"; arest="${adm#*|}"; artype="${arest%%|*}"; akind="${arest#*|}"
    if [[ "$achannel" == "$channel" && "$artype" == "$rtype" && "$(lower "$akind")" == "$component" ]]; then
      found=1; break
    fi
  done
  [[ $found -eq 1 ]] || unadmitted+=("catalog/*/${component}/iac/tf -> ${rtype} attaches provider = ${channel} with no admission for kind ${component}")
done

unused=()
for adm in "${admitted_pairs[@]+"${admitted_pairs[@]}"}"; do
  channel="${adm%%|*}"; rest="${adm#*|}"; rtype="${rest%%|*}"; kind="${rest#*|}"
  found=0
  for att in "${attached_pairs[@]+"${attached_pairs[@]}"}"; do
    achannel="${att%%|*}"; arest="${att#*|}"; artype="${arest%%|*}"; acomponent="${arest#*|}"
    if [[ "$achannel" == "$channel" && "$artype" == "$rtype" && "$acomponent" == "$(lower "$kind")" ]]; then
      found=1; break
    fi
  done
  [[ $found -eq 1 ]] || unused+=("${admissions_dir}/${channel}.yaml -> ${rtype} is admitted for ${kind}, but no module of that kind attaches it through provider = ${channel}")
done

failed=0

if [[ ${#unadmitted[@]} -gt 0 ]]; then
  echo "ERROR: ${#unadmitted[@]} resource block(s) attach a secondary-channel provider without a recorded admission." >&2
  echo "Secondary-channel capability enters the catalog per resource, per kind, by recorded decision. Either" >&2
  echo "record the admission (resource, kind, reason, promotionTracking) in ${admissions_dir}/<channel>.yaml," >&2
  echo "or model the resource through the baseline provider:" >&2
  printf '  - %s\n' "${unadmitted[@]}" >&2
  echo >&2
  failed=1
fi

if [[ ${#unused[@]} -gt 0 ]]; then
  echo "ERROR: ${#unused[@]} recorded admission(s) have no module attaching the resource through the channel." >&2
  echo "An admission exists only for a module that deploys through the channel; remove the entry, or set" >&2
  echo "'provider = <channel>' on the resource block the admission is for:" >&2
  printf '  - %s\n' "${unused[@]}" >&2
  echo >&2
  failed=1
fi

if [[ $failed -ne 0 ]]; then
  echo "Secondary-channel admission guard FAILED." >&2
  exit 1
fi

echo "Secondary-channel admission guard passed: ${#attached_pairs[@]} channel attachment(s), each admitted; ${#admitted_pairs[@]} admission(s), each attached."
