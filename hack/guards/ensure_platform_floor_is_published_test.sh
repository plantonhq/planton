#!/usr/bin/env bash
set -euo pipefail

# Test for ensure_platform_floor_is_published.sh, against a fake registry.
#
# The guard's registry call is one command (MANIFEST_INSPECT), so the test
# stands in a script that answers from a table of published references: a
# single amd64 manifest, a two-architecture index, or "no such manifest". The
# repository's own tree is checked once (the floor, the slugs and the
# scenarios exactly as committed); every other case runs on a small copied
# tree whose floor and scenarios the case writes.
#
# Run: bash hack/guards/ensure_platform_floor_is_published_test.sh

repo_root_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
guard="${repo_root_dir}/hack/guards/ensure_platform_floor_is_published.sh"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# The fake registry: FAKE_REGISTRY names a file of "<ref> <shape>" lines.
cat > "${work}/fake-inspect" <<'EOF'
#!/usr/bin/env bash
ref="$1"
shape="$(awk -v ref="$ref" '$1 == ref { print $2; exit }' "$FAKE_REGISTRY")"
case "$shape" in
  single)
    printf '{"Ref":"%s","Descriptor":{"platform":{"architecture":"amd64","os":"linux"}}}\n' "$ref" ;;
  index)
    printf '[{"Descriptor":{"platform":{"architecture":"amd64","os":"linux"}}},{"Descriptor":{"platform":{"architecture":"arm64","os":"linux","variant":"v8"}}},{"Descriptor":{"platform":{"architecture":"unknown","os":"unknown"}}}]\n' ;;
  *)
    echo "no such manifest: ${ref}" >&2; exit 1 ;;
esac
EOF
chmod +x "${work}/fake-inspect"

failed=0
fail() { echo "FAIL: $*" >&2; failed=1; }

# run_guard <tree> <registry-file>: sets rc and out (stdout and stderr).
run_guard() {
  set +e
  out="$(REPO_ROOT="$1" FAKE_REGISTRY="$2" MANIFEST_INSPECT="${work}/fake-inspect" bash "$guard" 2>&1)"
  rc=$?
  set -e
}

registry_root="$(sed -nE 's/^const DefaultImageRegistry = "([^"]+)"$/\1/p' "${repo_root_dir}/operator/internal/plantonregistry/plantonregistry.go")"
slugs=(control-plane runner client-apps/web)

# publish <registry-file> <tag> <shape> [slug...]: every slug when none named.
publish() {
  local file="$1" tag="$2" shape="$3"; shift 3
  local names=("$@"); [[ ${#names[@]} -gt 0 ]] || names=("${slugs[@]}")
  for slug in "${names[@]}"; do echo "${registry_root}/${slug}:${tag} ${shape}" >> "$file"; done
}

# A copied tree: the three files the guard reads, a floor, and scenarios.
make_tree() {
  local tree="$1" floor="$2"; shift 2
  mkdir -p "${tree}/operator/internal/platformversion" "${tree}/operator/internal/plantonregistry" \
    "${tree}/operator/internal/resources" "${tree}/catalog/kubernetes/kubernetesplantonplatform/e2e/scenarios"
  cp "${repo_root_dir}/operator/internal/plantonregistry/plantonregistry.go" "${tree}/operator/internal/plantonregistry/"
  cp "${repo_root_dir}/operator/internal/resources/images.go" "${tree}/operator/internal/resources/"
  printf 'package platformversion\n\nconst MinimumSupported = "%s"\n' "$floor" > "${tree}/operator/internal/platformversion/platformversion.go"
  local pair
  for pair in "$@"; do
    printf 'kind: KubernetesPlantonPlatform\nspec:\n  namespace:\n    value: planton\n  version: %s\n' "${pair#*=}" \
      > "${tree}/catalog/kubernetes/kubernetesplantonplatform/e2e/scenarios/${pair%%=*}.yaml"
  done
}

# 1. The repository as committed: every image of the floor and of each
#    scenario's release published, so the guard passes and names them.
registry="${work}/committed.registry"; : > "$registry"
floor="$(sed -nE 's/^const MinimumSupported = "([^"]+)"$/\1/p' "${repo_root_dir}/operator/internal/platformversion/platformversion.go")"
publish "$registry" "$floor" single
for scenario in "${repo_root_dir}"/catalog/kubernetes/kubernetesplantonplatform/e2e/scenarios/*.yaml; do
  publish "$registry" "$(sed -nE 's/^  version:[[:space:]]*([^[:space:]]+)$/\1/p' "$scenario")" single
done
run_guard "$repo_root_dir" "$registry"
[[ $rc -eq 0 ]] || fail "the committed tree should pass when its releases are published; got rc=$rc: $out"
for slug in "${slugs[@]}"; do
  [[ "$out" == *"published: ${registry_root}/${slug}:${floor} (linux/amd64)"* ]] \
    || fail "the committed tree's run should name ${slug} at the floor ${floor} with its architecture; got: $out"
done

# 2. An index prints every architecture and leaves attestations out.
tree="${work}/index"; make_tree "$tree" v1.2.0 minimal=v1.2.0
registry="${work}/index.registry"; : > "$registry"
publish "$registry" v1.2.0 index
run_guard "$tree" "$registry"
[[ $rc -eq 0 ]] || fail "an all-published floor should pass; got rc=$rc: $out"
[[ "$out" == *"published: ${registry_root}/runner:v1.2.0 (linux/amd64, linux/arm64/v8)"* ]] \
  || fail "an index should print both architectures and no unknown/unknown; got: $out"

# 3. The floor's console image missing: the sentence names the reference.
tree="${work}/missing"; make_tree "$tree" v1.2.0 minimal=v1.2.0
registry="${work}/missing.registry"; : > "$registry"
publish "$registry" v1.2.0 single control-plane runner
run_guard "$tree" "$registry"
[[ $rc -ne 0 ]] || fail "a floor whose console image is unpublished should fail; got: $out"
[[ "$out" == *"${registry_root}/client-apps/web:v1.2.0 is not published"*"publish that release's client-apps/web image"* ]] \
  || fail "the failure should name the missing console reference and the next step; got: $out"

# 4. A scenario below the floor: refused by name, even with its images published.
tree="${work}/below"; make_tree "$tree" v1.2.0 minimal=v1.1.9 minimal.to=v1.2.1
registry="${work}/below.registry"; : > "$registry"
publish "$registry" v1.2.0 single; publish "$registry" v1.1.9 single; publish "$registry" v1.2.1 single
run_guard "$tree" "$registry"
[[ $rc -ne 0 ]] || fail "a scenario below the floor should fail; got: $out"
[[ "$out" == *"scenarios/minimal.yaml declares v1.1.9, below the operator's floor v1.2.0"* ]] \
  || fail "the failure should name the scenario and both versions; got: $out"
[[ "$out" != *"minimal.to.yaml declares"* ]] || fail "a scenario above the floor should not be named; got: $out"

# 5. A scenario above the floor whose release is unpublished fails too: its
#    upgrade act would wait on the pull.
tree="${work}/ahead"; make_tree "$tree" v1.2.0 minimal=v1.2.0 minimal.to=v1.3.0
registry="${work}/ahead.registry"; : > "$registry"
publish "$registry" v1.2.0 single
run_guard "$tree" "$registry"
[[ $rc -ne 0 && "$out" == *"${registry_root}/control-plane:v1.3.0 is not published"* ]] \
  || fail "a scenario's unpublished release should fail by reference; got rc=$rc: $out"

# 6. Versions compare as releases, not as text: v0.0.100 is above v0.0.75.
tree="${work}/order"; make_tree "$tree" v0.0.75 minimal=v0.0.100
registry="${work}/order.registry"; : > "$registry"
publish "$registry" v0.0.75 single; publish "$registry" v0.0.100 single
run_guard "$tree" "$registry"
[[ $rc -eq 0 ]] || fail "v0.0.100 is above the floor v0.0.75; got rc=$rc: $out"

if [[ $failed -ne 0 ]]; then
  echo "ensure_platform_floor_is_published_test: FAILED" >&2
  exit 1
fi
echo "ensure_platform_floor_is_published_test: OK"
