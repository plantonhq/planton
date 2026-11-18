#!/usr/bin/env bash
set -euo pipefail

# Guard: every kind README (catalog/<provider>/<kind>/README.md) and every
# helm chart README (helm/<chart>/README.md) must end with the canonical
# license footer.
#
# WHY THIS EXISTS
# The realistic unit of copying for this catalog is one kind directory,
# not the whole repository. The repo-root LICENSE does not travel with a copied
# directory -- the kind's README does. The footer is the attribution that
# travels: it states the license and links to it with an absolute URL so the
# pointer survives copies, zips, and forks. Kinds are forged continuously;
# this guard makes it impossible for a new one to ship without its attribution.
#
# SCOPE
# A last-line PRESENCE check (mirroring the sibling grep-based guards, no
# toolchain beyond find + awk): README files that are the DIRECT child of a
# kind root, plus helm chart READMEs. Deeper READMEs (iac/**) are inner
# documentation whose shipped artifacts carry LICENSE and NOTICE files instead
# of footers; provider-infrastructure dirs (aa_*) are not kinds.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

footer='© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).'

missing=()

while IFS= read -r readme; do
  [[ -z "$readme" ]] && continue
  # The last non-empty line must be exactly the canonical footer.
  last_line="$(awk 'NF {line=$0} END {print line}' "$readme")"
  if [[ "$last_line" != "$footer" ]]; then
    missing+=("$readme")
  fi
done < <(
  # Kind READMEs: the direct child of a kind root -- exactly
  # catalog/<provider>/<kind>/README.md, depth-anchored so it can never
  # match iac/**/README.md or deeper docs. Provider-infrastructure dirs
  # (aa_e2e/, aa_eval/, aa_import/) are excluded by name: they sit at the
  # same depth but are not kinds.
  find catalog -mindepth 3 -maxdepth 3 -name 'README.md' 2>/dev/null | grep -Ev '^catalog/[^/]+/aa_[^/]+/'
  # Helm chart READMEs: helm/<chart>/README.md — published distribution surfaces.
  find helm -mindepth 2 -maxdepth 2 -name README.md 2>/dev/null
)

if (( ${#missing[@]} > 0 )); then
  {
    echo "License-footer guard FAILED: ${#missing[@]} README(s) missing the canonical footer."
    printf '  %s\n' "${missing[@]}"
    echo
    echo "Fix: end each file with a separator and the canonical footer:"
    echo
    echo "---"
    echo
    echo "${footer}"
  } >&2
  exit 1
fi

echo "License-footer guard passed: every kind and helm chart README carries the canonical footer."
