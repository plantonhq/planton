#!/usr/bin/env bash
# =============================================================================
# Channel-aware breaking-change gate for the proto contract.
#
# Runs `buf breaking` against a baseline and classifies every finding by the
# maturity channel its file path declares. The version directory name IS the
# channel (v1alpha1 / v1beta1 / v1), so the policy is data-driven:
#
#   alpha    (v<N>alpha<M>)  -- ADVISORY. Alpha may break in place; the break
#                               must be visible, never blocking.
#   beta     (v<N>beta<M>)   -- BLOCKING. Beta owes conversions; breaking
#                               shape requires a NEW version, never in-place.
#   stable   (v<N>)          -- BLOCKING. Stable is frozen.
#   internal (qa/, iac/) -- ADVISORY. Internal
#                               tooling formats, not API resources on the
#                               maturity ladder.
#   shared   (everything else, e.g. shared/) -- BLOCKING.
#                               Shared types are embedded in every kind's
#                               contract; breaking them breaks everything at
#                               once, so it must be a deliberate, reviewed
#                               act (change this script's policy in the same
#                               PR if the break is intended).
#
# Usage:
#   tools/ci/proto/buf_breaking_channel_gate.sh [<against-ref>]
#       against-ref defaults to '.git#branch=main' (run from the repo root)
#   tools/ci/proto/buf_breaking_channel_gate.sh --self-test
#       verifies the classifier against fixture paths; used by CI lint.
# =============================================================================
set -uo pipefail

classify() {
  # classify <path> -> prints alpha|beta|stable|internal|shared
  local path="$1" vdir
  case "$path" in
    qa/*|iac/*)
      echo internal
      return
      ;;
  esac
  vdir=$(echo "$path" | grep -oE '/v[0-9]+((alpha|beta)[0-9]+)?/' | head -1 | tr -d '/')
  if [ -z "$vdir" ]; then
    echo shared
    return
  fi
  case "$vdir" in
    *alpha*) echo alpha ;;
    *beta*) echo beta ;;
    *) echo stable ;;
  esac
}

deleted_path_from_message() {
  # File-deletion findings (FILE_NO_DELETE) carry NO path field -- the file
  # no longer exists in the input, so buf names it only inside the message:
  #   Previously present file "catalog/.../foo.proto" was deleted.
  # Extract that path so deletions classify like every other finding; a
  # deletion the gate cannot classify would otherwise be dropped SILENTLY,
  # letting a stable/shared file deletion pass unblocked.
  echo "$1" | sed -n 's/^Previously present file "\([^"]*\)" was deleted\.$/\1/p'
}

if [ "${1:-}" = "--self-test" ]; then
  fail=0
  check() {
    local got
    got=$(classify "$1")
    if [ "$got" != "$2" ]; then
      echo "SELF-TEST FAIL: classify($1) = ${got}, want $2"
      fail=1
    fi
  }
  check "catalog/aws/awsvpc/v1alpha1/spec.proto" alpha
  check "catalog/aws/awsvpc/v1beta1/spec.proto" beta
  check "catalog/aws/awsvpc/v1/spec.proto" stable
  check "catalog/aws/awsvpc/v2/spec.proto" stable
  check "catalog/_test/testcatalogkindgeneric/v1alpha1/api.proto" alpha
  check "shared/foreignkey/v1/options.proto" stable
  check "shared/catalogkind/catalog_kind.proto" shared
  check "qa/catalogkinde2eprofile/v1/api.proto" internal
  check "iac/catalogkindimportmap/v1/api.proto" internal
  check_deleted() {
    local got
    got=$(deleted_path_from_message "$1")
    if [ "$got" != "$2" ]; then
      echo "SELF-TEST FAIL: deleted_path_from_message($1) = '${got}', want '$2'"
      fail=1
    fi
  }
  check_deleted 'Previously present file "catalog/aws/awsvpc/v1alpha1/input.proto" was deleted.' \
    "catalog/aws/awsvpc/v1alpha1/input.proto"
  check_deleted 'Previously present file "shared/metadata.proto" was deleted.' \
    "shared/metadata.proto"
  check_deleted 'Field "1" with name "x" on message "Y" moved from outside to inside a oneof.' ""
  if [ $fail -eq 0 ]; then
    echo "self-test: all classifications correct"
  fi
  exit $fail
fi

AGAINST="${1:-.git#branch=main}"

BUF_STDERR=$(mktemp)
FINDINGS_JSON=$(buf breaking --against "$AGAINST" --error-format=json 2>"$BUF_STDERR")
BUF_EXIT=$?

if [ $BUF_EXIT -eq 0 ] && [ -z "$FINDINGS_JSON" ]; then
  rm -f "$BUF_STDERR"
  echo "buf breaking: no findings."
  exit 0
fi

# A nonzero exit WITHOUT findings means buf itself failed (bad baseline ref,
# compile error, network) -- that must fail the gate loudly, never pass as
# "zero findings".
if [ -z "$FINDINGS_JSON" ]; then
  echo "ERROR: buf breaking failed (exit ${BUF_EXIT}) without producing findings:" >&2
  cat "$BUF_STDERR" >&2
  rm -f "$BUF_STDERR"
  exit 2
fi
rm -f "$BUF_STDERR"

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq is required to classify buf breaking findings." >&2
  exit 2
fi

advisory=0
blocking=0

# buf emits one JSON object per line: {"path": "...", "type": "...", "message": "..."}
while IFS= read -r line; do
  [ -z "$line" ] && continue
  path=$(echo "$line" | jq -r '.path // empty')
  msg=$(echo "$line" | jq -r '.message // empty')
  if [ -z "$path" ]; then
    path=$(deleted_path_from_message "$msg")
  fi
  if [ -z "$path" ]; then
    # A finding the gate cannot attribute to a file must never vanish
    # silently -- treat it as blocking so a human looks at it.
    blocking=$((blocking + 1))
    echo "BLOCKING  [unattributable] ${msg}"
    continue
  fi

  channel=$(classify "$path")
  case "$channel" in
    alpha|internal)
      advisory=$((advisory + 1))
      echo "ADVISORY  [${channel}] ${path}: ${msg}"
      ;;
    *)
      blocking=$((blocking + 1))
      echo "BLOCKING  [${channel}] ${path}: ${msg}"
      ;;
  esac
done <<< "$FINDINGS_JSON"

echo ""
echo "buf breaking summary: ${advisory} advisory (alpha/internal), ${blocking} blocking (beta/stable/shared)"

if [ $blocking -gt 0 ]; then
  echo ""
  echo "Breaking changes were detected in beta, stable, or shared packages."
  echo "Those channels forbid in-place breaks: ship the change as a NEW api"
  echo "version with a total conversion spec, or -- for shared packages --"
  echo "make the deliberate policy call in review."
  exit 1
fi
exit 0
