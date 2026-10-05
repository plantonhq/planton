#!/usr/bin/env bash
set -euo pipefail

# Guard: every Kubernetes kind that ships an E2E profile must be fully
# wired into the E2E tiers — BOTH engine test entrypoints present in
# e2e/kubernetes_test.go, and each entrypoint reachable from a Makefile
# e2e-test-kubernetes-tier* regex.
#
# WHY THIS EXISTS
# The tier wiring is three hand-maintained copies of one fact: the per-kind
# test entrypoints, the Makefile tier regexes, and the kind's own E2E
# profile. Nothing structural kept them converged, and the drift is SILENT:
# a missing entrypoint makes `go test -run <regex>` match nothing (a kind's
# lane can never run, while a green profile claims it does), and a missing
# Makefile regex entry silently drops the kind from the local tier sweeps.
# Both drifts shipped in the wild before this guard existed.

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

provider_dir="catalog/kubernetes"
entrypoints_file="e2e/kubernetes_test.go"
failures=0

# func-name → kind pairs from the entrypoint file. Every entrypoint is
# the two-line shape:
#   func TestKubernetes<Name>_<Engine>(t *testing.T) {
#       runAllScenariosForKind(t, "<kind>", "<engine>")
kinds_with_funcs="$(awk '
  /^func Test[A-Za-z0-9]+_(Pulumi|Terraform)\(t \*testing\.T\) \{/ {
    fn=$2; sub(/\(.*/, "", fn); next
  }
  /runAllScenariosForKind\(t, "/ && fn != "" {
    split($0, parts, "\"");
    print parts[2] " " fn;
    fn=""
  }
' "$entrypoints_file")"

# The union of every tier target's -run regex bodies, for reachability checks.
makefile_regexes="$(grep -o 'run "Test([^"]*' Makefile || true)"
if [[ -z "$makefile_regexes" ]]; then
  echo "GUARD BUG: no e2e tier -run regexes extracted from the Makefile — fix the extraction before trusting this guard" >&2
  exit 1
fi

# The Terraform-only sweeps (`..._Terraform"`) are a SECOND hand-maintained
# copy of each tier's member list, and they drift independently: a kind can
# be reachable through the generic tier row (which matches both engines'
# funcs) while the Terraform-only twin row silently skips it. Terraform
# funcs are therefore checked against the Terraform-suffixed rows
# specifically, never just the union (a real drift shipped exactly this
# way).
makefile_terraform_regexes="$(printf '%s\n' "$makefile_regexes" | grep '_Terraform$' || true)"
if [[ -z "$makefile_terraform_regexes" ]]; then
  echo "GUARD BUG: no Terraform-only e2e tier -run regexes extracted from the Makefile — fix the extraction before trusting this guard" >&2
  exit 1
fi

# Kind e2e assets live at the kind root: {kind}/e2e/profile.yaml.
for profile in "$provider_dir"/*/e2e/profile.yaml; do
  kind="$(basename "$(dirname "$(dirname "$profile")")")"

  # Profiles that can never RUN need no tier wiring: deferred/skip/stub
  # kinds are honest placeholders (their wiring lands with the session
  # that makes them runnable). Everything runnable — green,
  # pending_proof, real_cluster — must be fully wired.
  status="$(awk '/^  status:/ {print $2; exit}' "$profile")"
  case "$status" in
    deferred|skip|stub) continue ;;
  esac

  for engine_suffix in Pulumi Terraform; do
    func_name="$(printf '%s\n' "$kinds_with_funcs" |
      awk -v c="$kind" -v s="_$engine_suffix" '$1 == c && index($2, s) {print $2}')"

    if [[ -z "$func_name" ]]; then
      echo "MISSING ENTRYPOINT: kind '$kind' has an E2E profile but no ${engine_suffix} test func in $entrypoints_file" >&2
      failures=$((failures + 1))
      continue
    fi

    # The Makefile regexes are alternations wrapped by the Test prefix:
    # `Test(A|B|...)_`. Reachable = the func's base name (without the
    # Test prefix) appears as an exact alternation member in at least
    # one tier regex — exact-member anchoring keeps prefix collisions
    # honest (KubernetesTekton never matches via KubernetesTektonOperator).
    member="${func_name%_$engine_suffix}"
    member="${member#Test}"
    if ! printf '%s\n' "$makefile_regexes" | grep -q "[|(]${member}[|)]"; then
      echo "MISSING TIER REGEX: $func_name exists but '$member' appears in no Makefile e2e tier -run regex" >&2
      failures=$((failures + 1))
    elif [[ "$engine_suffix" == "Terraform" ]] && ! printf '%s\n' "$makefile_terraform_regexes" | grep -q "[|(]${member}[|)]"; then
      echo "MISSING TERRAFORM TIER REGEX: '$member' rides a generic tier row but appears in no Terraform-only (_Terraform) tier regex — the two row sets are copies and this one drifted" >&2
      failures=$((failures + 1))
    fi
  done
done

if [[ "$failures" -gt 0 ]]; then
  echo "" >&2
  echo "$failures E2E tier-wiring gap(s). Add the missing entrypoint(s) to $entrypoints_file and/or the kind's base name to its tier regex in the Makefile." >&2
  exit 1
fi

echo "OK: every profiled Kubernetes kind has both engine entrypoints and tier-regex coverage."
