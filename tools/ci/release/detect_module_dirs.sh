#!/usr/bin/env bash
# =============================================================================
# Module change detection for auto-tagging.
#
# Reads changed file paths on stdin and emits the unique IaC module
# directories they belong to, one per line:
#
#   --pulumi     catalog/{p}/{k}/iac/pulumi
#   --terraform  catalog/{p}/{k}/iac/tf
#
# Modules live at the kind root -- exactly two path segments between
# catalog/ and iac/ (provider, kind). The anchor is strict on BOTH sides: a
# looser pattern would match provider-infrastructure Go under aa_e2e/ or
# nested non-module iac dirs, and a stale pattern makes module auto-tagging
# silently stop firing -- no error, releases just stop (this script exists
# because that happened once). auto-tag.yaml consumes it instead of carrying
# inline patterns, and the --self-test fixtures prove the detection fires.
#
# Usage:
#   git diff --name-only A B | detect_module_dirs.sh --pulumi
#   detect_module_dirs.sh --self-test    (used by auto-tag before detection)
# =============================================================================
set -euo pipefail

detect() {
  # detect <flavor-dir>  (pulumi|tf), paths on stdin
  local flavor="$1"
  grep -E "^catalog/[^/]+/[^/]+/iac/${flavor}/" \
    | sed -E "s#(catalog/[^/]+/[^/]+/iac/${flavor})/.*#\1#" \
    | sort -u \
    || true
}

self_test() {
  local fail=0
  check() {
    # check <flavor> <input-path> <expected-output ('' for no match)>
    local got
    got=$(printf '%s\n' "$2" | detect "$1")
    if [ "$got" != "$3" ]; then
      echo "SELF-TEST FAIL: detect --${1} '${2}' = '${got}', want '${3}'"
      fail=1
    fi
  }
  # Kind-root modules fire.
  check tf "catalog/aws/awsvpc/iac/tf/main.tf" \
           "catalog/aws/awsvpc/iac/tf"
  check tf "catalog/aws/awsvpc/iac/tf/variables.tf" \
           "catalog/aws/awsvpc/iac/tf"
  check pulumi "catalog/gcp/gcpgkecluster/iac/pulumi/main.go" \
               "catalog/gcp/gcpgkecluster/iac/pulumi"
  check pulumi "catalog/gcp/gcpgkecluster/iac/pulumi/module/main.go" \
               "catalog/gcp/gcpgkecluster/iac/pulumi"
  # Non-module changes must NOT fire.
  check tf "catalog/aws/awsvpc/v1alpha1/spec.proto" ""
  check tf "catalog/aws/awsvpc/iac/pulumi/main.go" ""
  check pulumi "pkg/iac/pulumi/pulumimodule/module_directory.go" ""
  # A version-resident path (the pre-anatomy shape) must NOT fire: modules
  # have exactly provider/kind between catalog/ and iac/.
  check tf "catalog/aws/awsvpc/v1alpha1/iac/tf/main.tf" ""
  # Two files in one module collapse to one dir.
  local got
  got=$(printf '%s\n%s\n' \
    "catalog/aws/awsvpc/iac/tf/main.tf" \
    "catalog/aws/awsvpc/iac/tf/outputs.tf" | detect tf | wc -l | tr -d ' ')
  if [ "$got" != "1" ]; then
    echo "SELF-TEST FAIL: two files in one module produced ${got} dirs, want 1"
    fail=1
  fi
  if [ $fail -eq 0 ]; then
    echo "self-test: module change detection fires for kind-root modules only"
  fi
  return $fail
}

case "${1:-}" in
  --pulumi) detect pulumi ;;
  --terraform) detect tf ;;
  --self-test) self_test ;;
  *)
    echo "usage: $0 --pulumi|--terraform|--self-test (paths on stdin)" >&2
    exit 2
    ;;
esac
