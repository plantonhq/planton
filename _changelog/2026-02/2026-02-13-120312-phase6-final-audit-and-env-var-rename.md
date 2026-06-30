# Phase 6: Final Documentation Audit and PROJECT_PLANTON Environment Variable Rename

**Date**: February 13, 2026
**Type**: Enhancement
**Components**: Documentation, CLI Configuration, Backend Configuration, Environment Variables

## Summary

Completed the final audit phase of the Planton documentation feature parity project. Renamed all `PROJECT_PLANTON_*` environment variables to `PLANTON_*` across the entire codebase (Go source, internal docs, changelogs, and public docs), fixing the last Planton naming remnants. Fixed critical manifest errors in `getting-started.md`, corrected the wrong `RedisKubernetes` kind name in `index.md`, and performed a systematic 40-page cross-reference and consistency audit.

## Problem Statement / Motivation

The Planton documentation had accumulated several issues across the 5 previous phases of the docs feature parity project:

### Pain Points

- `getting-started.md` manifest was missing required proto fields (`namespace`, `container.diskSize`), meaning a new user following the guide would hit validation errors on their first deployment
- Root `index.md` used `RedisKubernetes` kind name (proto defines `KubernetesRedis`), stale provider counts, and no links to 3 new documentation sections
- 6 environment variables still used `PROJECT_PLANTON_*` naming — a Planton branding remnant in an open-source project that should have zero commercial references
- `troubleshooting.md` (562 lines) had never been formally audited
- No pages linked to the troubleshooting guide despite it existing

## Solution / What's New

### Environment Variable Rename

All `PROJECT_PLANTON_*` environment variables renamed to `PLANTON_*`:

| Old Name | New Name |
|----------|----------|
| `PROJECT_PLANTON_BACKEND_TYPE` | `PLANTON_BACKEND_TYPE` |
| `PROJECT_PLANTON_BACKEND_BUCKET` | `PLANTON_BACKEND_BUCKET` |
| `PROJECT_PLANTON_BACKEND_REGION` | `PLANTON_BACKEND_REGION` |
| `PROJECT_PLANTON_BACKEND_ENDPOINT` | `PLANTON_BACKEND_ENDPOINT` |
| `PROJECT_PLANTON_GIT_REPO` | `PLANTON_GIT_REPO` |
| `PROJECT_PLANTON_MANIFEST` | `PLANTON_MANIFEST` |

### Documentation Fixes

- **getting-started.md**: Added required `namespace` and `diskSize` fields to KubernetesPostgres manifest, added links to tutorials and provider setup guides, added troubleshooting link
- **index.md**: Fixed `RedisKubernetes` to `KubernetesRedis`, fixed icon from raw emoji to valid key, updated all provider component counts, added Scaleway and OpenStack to provider grid, added Tutorials/Examples/Contributing sections
- **troubleshooting.md**: Fixed icon from `wrench` to `gear`, full audit confirmed clean (no Planton references, no out-of-scope commands)
- **cloud-resource-kinds.md**: Fixed Azure count from 12 to 10

## Implementation Details

### Go Source Changes

**File**: `pkg/iac/tofu/backendconfig/env_vars.go`
- 4 const definitions renamed
- Comment in `ReadFromEnv()` updated

**File**: `pkg/iac/tofu/backendconfig/validate.go`
- Comment example updated

**File**: `pkg/iac/tofu/backendconfig/build_config.go`
- Comment updated

**File**: `pkg/iac/gitrepo/local_repo.go`
- Const definition and comment updated

**File**: `pkg/iac/localmodule/local_module.go`
- User-facing error message updated

**File**: `app/backend/internal/service/stack_update_service.go`
- 2 env var usages in `fmt.Sprintf` calls updated

### Verification

- `go build ./pkg/iac/...` — compiles clean
- `go build ./app/backend/...` — compiles clean
- `rg "PROJECT_PLANTON"` across entire repo — zero results

### Documentation Cross-Reference Audit (40 pages)

- All internal links verified across 7 sections
- Troubleshooting links added to `index.md` and `getting-started.md`
- All frontmatter icons validated against `iconMap`
- Component/provider counts verified against `cloud_resource_kind.proto`
- Final Planton sweep confirmed zero commercial references (only `plantonhq` in GitHub URLs and Homebrew tap)

## Benefits

- New users following `getting-started.md` will no longer hit validation errors
- Environment variables now consistently branded as `PLANTON_*`
- Zero `PROJECT_PLANTON` remnants anywhere in the codebase
- All 40 documentation pages cross-referenced and consistent
- Troubleshooting guide now discoverable from landing page and getting-started guide

## Impact

- **CLI Users**: Environment variable names changed from `PROJECT_PLANTON_BACKEND_*` to `PLANTON_BACKEND_*`. Users with existing CI/CD pipelines using the old names will need to update.
- **Documentation Readers**: Getting-started guide now produces working deployments on first try. All section navigation complete.
- **Contributors**: Internal READMEs and architecture docs now use consistent `PLANTON_*` naming.

## Related Work

- Phase 0: Existing docs audit (2026-02-12)
- Phase 1: Concepts section complete rewrite (2026-02-13)
- Phase 2: CLI docs expansion (2026-02-13)
- Phase 3: Guides expansion (2026-02-13)
- Phase 4: Tutorials section (2026-02-13)
- Phase 5: Examples + Contributing (2026-02-13)

---

**Status**: Production Ready
**Timeline**: ~2 hours
