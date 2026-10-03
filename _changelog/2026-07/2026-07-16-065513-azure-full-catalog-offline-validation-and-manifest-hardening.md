# Azure Full-Catalog Offline Validation and Public-Manifest Hardening

**Date**: July 16, 2026
**Type**: Bug Fix
**Components**: Azure Provider, API Definitions, Manifest Processing, Testing Framework, Build System

## Summary

The first catalog-wide offline validation pass over the entire Azure surface —
all 109 kinds, 391 public manifests (282 presets + 109 hack manifests), 12
infra-charts, all repo guards, and the full Bazel spec-test tree. The pass
surfaced and fixed 55 invalid public manifests, three spec-test files that no
longer compiled after earlier foreign-key retrofits, a silently-failing repo
guard, a missing Bazel repo registration, and two stale generated artifacts.
It also confirmed the Azure Pulumi shared-builder migration is complete:
109 of 109 modules construct their provider through the shared builders, with
zero inline `NewProvider` calls remaining.

## Problem Statement / Motivation

Per-kind gates run at forge/update time, but several validation classes only
manifest when the whole catalog is checked at once:

- Presets and hack manifests are public catalog artifacts users copy-paste.
  They are schema-validated only when someone runs `planton validate-manifest`
  against them — which no per-session gate mandated. Placeholder text in
  format-validated fields, missing `StringValueOrRef` wrappers from later
  retrofits, and stale enum spellings all accumulate silently.
- Spec tests are compiled per kind at their own session's gate; a retrofit
  that changes a *shared* field type (plain string → `StringValueOrRef`) can
  break a *different* kind's test file without that kind's gate ever re-running.
- Generated artifacts (`pkg/protodocs/index.json.gz`, the site catalog mirror)
  drift when a session edits their sources without re-running the generator.

## What Was Found and Fixed

### 55 invalid public manifests (of 391 swept; now 391/391 green)

| Class | Files | Example |
| --- | --- | --- |
| Angle-bracket placeholder in a format-validated field | ~48 | `vaultName: <globally-unique-vault-name>` fails the vault-name pattern; replaced with a concrete replace-me value plus a `# Replace:` comment stating the format contract |
| Bare string where `StringValueOrRef` is required | 2 | The metric-alert web-test preset's `webTestId` (now wired by `valueFrom` to the web test's `web_test_id` output, scope included) and the network-interface hack manifest's App Gateway pool id (missing `value:` wrapper) |
| Stale enum spelling | 1 | AKS preset `networkPolicy: AZURE` → `NETWORK_POLICY_AZURE` |
| CEL-pairing violations / format examples | 4 | SSL checks paired with an `http://` URL; reserved MySQL admin login; non-IP firewall placeholders; malformed SSH-key placeholder |

Companion preset markdown placeholder tables (53 files) were aligned with the
new concrete values, and the site catalog mirror regenerated.

### Spec-test compile debt from foreign-key retrofits (3 files)

The first-ever full `bazel test //apis/dev/planton/provider/azure/...` run
exposed test files that stopped compiling when shared fields became
`StringValueOrRef`:

- `azuremonitormetricalert/v1/spec_test.go` — `WebTestId` bare strings ×2
- `azurenetworkinterface/v1/spec_test.go` — `ApplicationSecurityGroupIds` as `[]string`
- `azurenetworksecuritygroup/v1/spec_test.go` — source/destination ASG ids ×3 sites

All fixed with the files' own literal-wrapper helpers. `go vet` across all
109 Azure v1 packages is clean, and the full suite now passes 109/109.

### Infrastructure fixes

- `hack/guards/ensure_tf_provider_pins.sh` used bash-4 `mapfile` and exited 127
  on macOS's bash 3.2 — the provider-pin guard was silently not running.
  Replaced with a portable while-read loop; the guard passes.
- `MODULE.bazel` was missing the `use_repo` entries for the Key Vault
  certificates data-plane SDK (and an AWS cloudcontrol sibling); regenerated
  with `bazel mod tidy`.
- `pkg/protodocs/index.json.gz` was stale (missing the newest Azure kinds'
  documentation) and the site catalog mirror had 21 un-synced Azure pages;
  both regenerated from source.

## Validation (what ran and passed)

- `make build-go`; `buf lint`; `make e2e-build` + the Azure `aa_e2e` package
  with `-tags=e2e`
- `bazel test //apis/dev/planton/provider/azure/...` — 109/109 pass
- `planton chart validate --all charts/azure` — 12/12 charts (defaults + every
  bool toggle), tree-built CLI; all 12 chart icons return HTTP 200
- `planton validate-manifest` over all 391 Azure presets + hack manifests —
  391/391 pass
- All five `hack/guards/*.sh` green (including the repaired provider-pin guard)
- `planton secret-coverage --check`; `planton validate-refs --check`
- `go test ./pkg/outputs/... ./pkg/infrachart/... ./pkg/refcheck/...` (uncached)
- `planton tofu plan` on the changed network-interface hack manifest
- Generator freshness: kind map, proto docs, E2E matrix, and site catalog
  re-run; the remaining working-tree diff is exactly the regenerated content
- Shared-builder census: 109/109 Azure Pulumi modules import the shared
  builders; zero inline provider construction anywhere under the Azure tree

## Workflow Uplift

The gap that let 55 invalid manifests accumulate is now closed at the source:

- `_rules/catalog-kind/update/update-catalog-kind.mdc` — the
  preset-validation step now mandates mechanically running
  `planton validate-manifest` (tree-built CLI) on every preset and the hack
  manifest of a touched kind.
- `_rules/catalog-kind/forge/flow/022-presets.mdc` — the self-validate
  step now requires the same command per preset, as the authoritative check
  for its placeholder rule.

## Impact

Every Azure public artifact a user can copy from the catalog now passes the
same validation the control plane applies on submission. The full-catalog
sweep, spec-test tree, and guard set are all green on the branch, which is the
release-readiness bar for the Azure 90/10 surface.

---

**Status**: ✅ Production Ready
