#!/usr/bin/env bash
set -euo pipefail

# Guard: text a kind ships through BOTH engines is byte-identical.
#
# WHY THIS EXISTS
# Some kinds hand a program to the cluster — a shell script a CronJob runs,
# a Python module a chart loads — or render from a data table both engines
# read, such as the images a mirror setting redirects, and both engines
# must render the same bytes from one manifest: the Pulumi module carries
# the text as a Go raw string constant, the Terraform module as an HCL
# heredoc, and a release packages each module on its own (tools/ci/release),
# so the text cannot be read from a sibling directory and lives twice on
# purpose. The only difference the two copies may have is HCL's escaping of
# `${` as `$${` and `%{` as `%%{`. Two copies drift silently: a fix to one
# engine's text is a behavior only that engine ships, and nothing in a
# working tree notices. This guard turns that drift into a static,
# pre-merge failure.
#
# HOW A KIND OPTS IN
# In the Terraform module, the line immediately above a heredoc opener
# carries a marker naming the Go file (relative to iac/pulumi/module/) and
# the constant it must equal:
#
#     # parity: scripts.go snapshotScript
#     "snapshot.sh" = <<EOT
#     ...
#     EOT
#
# The heredoc must be the plain `<<` form (an indented `<<-` heredoc strips
# leading whitespace at render time and has no byte-identical Go twin). The
# Go constant must be a raw string: `const snapshotScript = ` followed by a
# backtick, the text, and a closing backtick on its own line — the shape
# every marked kind already uses.
#
# WHAT IT CHECKS
#   1. Every marker names a Go file and constant that exist.
#   2. The heredoc body, with HCL's `$${` and `%%{` unescaped, is
#      byte-identical to the constant's text.
#   3. Every marked heredoc closes (a missing terminator is a broken file).

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

failures=()
checked=0
tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

# Every marker in every Terraform module, as file:line:go-file:const.
while IFS= read -r hit; do
  [[ -z "$hit" ]] && continue
  tf_file="${hit%%:*}"
  rest="${hit#*:}"
  marker_line="${rest%%:*}"
  marker_text="${rest#*:}"
  go_file="$(printf '%s\n' "$marker_text" | sed -E 's/^[[:space:]]*#[[:space:]]*parity:[[:space:]]*([^[:space:]]+)[[:space:]]+([^[:space:]]+)[[:space:]]*$/\1/')"
  const_name="$(printf '%s\n' "$marker_text" | sed -E 's/^[[:space:]]*#[[:space:]]*parity:[[:space:]]*([^[:space:]]+)[[:space:]]+([^[:space:]]+)[[:space:]]*$/\2/')"
  module_dir="$(printf '%s\n' "$tf_file" | sed -E 's#^(catalog/[^/]+/[^/]+)/iac/tf/.*#\1#')"
  go_path="${module_dir}/iac/pulumi/module/${go_file}"
  where="${tf_file}:${marker_line}"

  if [[ ! -f "$go_path" ]]; then
    failures+=("${where}: marker names ${go_path}, which does not exist")
    continue
  fi

  # The heredoc opener is the line after the marker.
  opener_line=$((marker_line + 1))
  opener="$(sed -n "${opener_line}p" "$tf_file")"
  if printf '%s\n' "$opener" | grep -qE '<<-'; then
    failures+=("${where}: the marked heredoc is indented (<<-); only the plain << form can be byte-identical to a Go constant")
    continue
  fi
  terminator="$(printf '%s\n' "$opener" | sed -nE 's/.*<<([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*$/\1/p')"
  if [[ -z "$terminator" ]]; then
    failures+=("${where}: the line after the marker is not a heredoc opener (expected ... = <<EOT)")
    continue
  fi

  # The heredoc body: lines after the opener up to the terminator line
  # (trailing whitespace tolerated on the terminator, as HCL does), with
  # HCL's escapes undone.
  tf_raw="${tmp_dir}/tf-raw.$checked"
  tf_body="${tmp_dir}/tf.$checked"
  awk -v start="$opener_line" -v term="$terminator" '
    NR <= start { next }
    { line = $0; sub(/[[:space:]]+$/, "", line) }
    line == term { found = 1; exit }
    { print $0 }
    END { if (!found) exit 3 }
  ' "$tf_file" > "$tf_raw" || {
    failures+=("${where}: heredoc <<${terminator} never closes")
    continue
  }
  # Portable (macOS sed has no GNU -i): unescape into a second file.
  sed -e 's/\$\${/${/g' -e 's/%%{/%{/g' "$tf_raw" > "$tf_body"

  # The Go raw string: the text after the opening backtick on the const
  # line through the line before the closing backtick.
  go_body="${tmp_dir}/go.$checked"
  awk -v name="$const_name" '
    !inside && $0 ~ "^const[[:space:]]+" name "[[:space:]]*=[[:space:]]*`" {
      inside = 1; found = 1
      sub("^const[[:space:]]+" name "[[:space:]]*=[[:space:]]*`", "", $0)
      if ($0 ~ /`/) { sub(/`.*$/, "", $0); print $0; exit }
      print $0; next
    }
    inside && $0 ~ /`/ { sub(/`.*$/, "", $0); if (length($0) > 0) print $0; exit }
    inside { print $0 }
    END { if (!found) exit 3 }
  ' "$go_path" > "$go_body" || {
    failures+=("${where}: no raw-string constant ${const_name} in ${go_path}")
    continue
  }

  if ! diff -u "$go_body" "$tf_body" > "${tmp_dir}/diff.$checked"; then
    failures+=("${where}: heredoc <<${terminator} differs from ${go_path} const ${const_name}:"$'\n'"$(sed -n '1,40p' "${tmp_dir}/diff.$checked")")
  fi
  checked=$((checked + 1))
done < <(grep -rnE --include='*.tf' '^[[:space:]]*#[[:space:]]*parity:[[:space:]]*[^[:space:]]+[[:space:]]+[^[:space:]]+[[:space:]]*$' catalog/*/*/iac/tf 2>/dev/null || true)

if [[ ${#failures[@]} -gt 0 ]]; then
  echo "ERROR: ${#failures[@]} cross-engine script pair(s) are not byte-identical or not resolvable." >&2
  echo "Observed: a Terraform heredoc marked '# parity: <go-file> <const>' does not match its Go constant (after undoing HCL's \$\${ and %%{ escapes), or the marker names something that does not exist." >&2
  echo "Meaning: the two engines would ship different program text from one manifest — a fix that reached only one of them." >&2
  echo "Next step: make the heredoc and the constant identical (edit both, or copy the Go text into the heredoc with \${ written as \$\${), and keep the marker above the opener." >&2
  for f in "${failures[@]}"; do
    printf '  - %s\n' "$f" >&2
  done
  echo "Cross-engine script parity guard FAILED." >&2
  exit 1
fi

echo "OK: ${checked} cross-engine script pair(s) are byte-identical across the Pulumi and Terraform modules."
