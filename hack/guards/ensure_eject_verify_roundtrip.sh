#!/usr/bin/env bash
set -euo pipefail

# Guard: the eject -> verify roundtrip works end to end, and verify can fail.
#
# WHY THIS EXISTS
# `planton module eject` copies an official module into a user-owned
# directory, and `planton module verify` proves a customization still honors
# the kind's contract. Two promises must hold together, and only an
# end-to-end run proves them:
#   1. An UNMODIFIED ejected module verifies green — verify must never fail
#      the module it just handed the user.
#   2. A module that breaks the contract verifies RED — a gate that cannot
#      fail teaches false confidence.
# For pulumi, the roundtrip additionally proves the ejected copy is a real
# standalone Go module: eject resolves its dependencies and verify compiles
# it with the workspace disabled, exactly as deployments build it.
#
# SCOPE
# NETWORK-DEPENDENT and therefore NOT a PR gate (unlike the sibling offline
# grep guards): a dev-build eject falls back to the git staging clone
# (~/.planton/staging, cloned once), and the pulumi dependency resolution
# reaches the Go module proxy. Run it on demand — after touching the eject
# or verify machinery, or before a release:
#   ./hack/guards/ensure_eject_verify_roundtrip.sh

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root_dir"

kind="AwsS3Bucket"
failures=0

work_dir="$(mktemp -d "${TMPDIR:-/tmp}/eject-verify-roundtrip.XXXXXX")"
cleanup() { rm -rf "$work_dir"; }
trap cleanup EXIT

echo "Building the planton CLI..."
planton_bin="${work_dir}/planton"
go build -o "$planton_bin" .

pass() { echo "  PASS  $1"; }
fail() { echo "  FAIL  $1"; failures=$((failures + 1)); }

# expect_ok / expect_fail run a command and assert its exit status.
expect_ok() {
  local description="$1"; shift
  if "$@" >"${work_dir}/last.log" 2>&1; then
    pass "$description"
  else
    fail "$description"
    sed 's/^/        /' "${work_dir}/last.log" | tail -20
  fi
}
expect_fail() {
  local description="$1"; shift
  if "$@" >"${work_dir}/last.log" 2>&1; then
    fail "$description (expected a non-zero exit)"
  else
    pass "$description"
  fi
}

echo ""
echo "=== OpenTofu module roundtrip (${kind}) ==="
tofu_dir="${work_dir}/tofu-module"
expect_ok "eject the official tofu module" \
  "$planton_bin" module eject "$kind" --provisioner tofu --output-dir "$tofu_dir"

for required_file in variables.tf LICENSE NOTICE CONTRACT.md; do
  if [ -f "${tofu_dir}/${required_file}" ]; then
    pass "ejected copy carries ${required_file}"
  else
    fail "ejected copy carries ${required_file}"
  fi
done

expect_ok "an unmodified ejected module verifies green" \
  "$planton_bin" module verify --kind "$kind" --module-dir "$tofu_dir" --skip-build-checks

# Break the contract: a required variable deployments can never satisfy.
cat >>"${tofu_dir}/variables.tf" <<'EOF'

variable "roundtrip_breaker" {
  type = string
}
EOF
expect_fail "a contract-breaking change verifies red" \
  "$planton_bin" module verify --kind "$kind" --module-dir "$tofu_dir" --skip-build-checks

echo ""
echo "=== Pulumi module roundtrip (${kind}) ==="
pulumi_dir="${work_dir}/pulumi-module"
expect_ok "eject the official pulumi module (resolves dependencies)" \
  "$planton_bin" module eject "$kind" --provisioner pulumi --output-dir "$pulumi_dir" \
  --go-module "example.com/roundtrip/$(echo "$kind" | tr '[:upper:]' '[:lower:]')"

for required_file in main.go go.mod Pulumi.yaml LICENSE NOTICE CONTRACT.md; do
  if [ -f "${pulumi_dir}/${required_file}" ]; then
    pass "ejected copy carries ${required_file}"
  else
    fail "ejected copy carries ${required_file}"
  fi
done

# Verify WITH build checks: compiles the copy standalone (GOWORK=off),
# proving the ejected module builds outside this repository.
expect_ok "an unmodified ejected module verifies green (including go build)" \
  "$planton_bin" module verify --kind "$kind" --module-dir "$pulumi_dir"

# Break the entrypoint contract: no typed IaC input.
cat >"${pulumi_dir}/main.go" <<'EOF'
package main

func main() {}
EOF
expect_fail "a contract-breaking change verifies red" \
  "$planton_bin" module verify --kind "$kind" --module-dir "$pulumi_dir" --skip-build-checks

echo ""
if [ $failures -gt 0 ]; then
  echo "FAIL: ${failures} roundtrip check(s) failed."
  exit 1
fi
echo "PASS: the eject -> verify roundtrip holds for both engines."
