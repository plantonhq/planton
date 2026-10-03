# Semver Build Metadata Tags for Auto-Releases

**Date**: January 8, 2026
**Type**: Fix / Enhancement
**Components**: Build System, GitHub Actions, Release Management

## Summary

Migrated all auto-release tag formats from invalid or inconsistent patterns to valid semver build metadata format (`+`). This fixes GoReleaser failures for CLI releases and unifies the tagging convention across all kind types.

## Problem Statement / Motivation

### GoReleaser Failure

The CLI auto-release workflow was failing with:

```
error: failed to parse tag 'v0.3.2.20260107.1' as semver: invalid semantic version
```

The tag `v0.3.2.20260107.1` has 5 dot-separated segments, but semver only allows 3 (`MAJOR.MINOR.PATCH`).

### Inconsistent Patterns

Before this change, each kind used a different tagging convention:

| Kind | Before | Valid Semver? |
|-----------|--------|---------------|
| CLI | `v0.3.2.20260107.1` | No (5 segments) |
| App | `v0.3.2-app-20260107.1` | Yes (pre-release) |
| Website | `v0.3.2-website-20260107.1` | Yes (pre-release) |
| Pulumi | `v0.3.2-pulumi-awsecsservice-20260107.1` | Yes (pre-release) |
| Terraform | `v0.3.2-terraform-awsecsservice-20260107.1` | Yes (pre-release) |

While the `-` pre-release format was technically valid, it was semantically incorrect. Pre-release versions indicate versions that come *before* a final release (e.g., alpha, beta, rc). Auto-releases are builds *of* a released version, not pre-releases.

## Solution / What's New

### Semver Build Metadata Format

Adopted the semver build metadata format (`+`) for all auto-release tags:

```
v{MAJOR}.{MINOR}.{PATCH}+{kind}.{YYYYMMDD}.{N}
```

Build metadata is the correct semantic for auto-releases because:
1. It indicates additional build information for a version
2. It's ignored for version precedence (all builds of v0.3.2 are equal)
3. It's valid semver that GoReleaser accepts

### New Tag Formats

| Kind | Before | After |
|-----------|--------|-------|
| CLI | `v0.3.2.20260107.1` | `v0.3.2+cli.20260107.1` |
| App | `v0.3.2-app-20260107.1` | `v0.3.2+app.20260107.1` |
| Website | `v0.3.2-website-20260107.1` | `v0.3.2+website.20260107.1` |
| Pulumi | `v0.3.2-pulumi-awsecsservice-20260107.1` | `v0.3.2+pulumi.awsecsservice.20260107.1` |
| Terraform | `v0.3.2-terraform-awsecsservice-20260107.1` | `v0.3.2+terraform.awsecsservice.20260107.1` |

## Implementation Details

### Files Changed

| File | Changes |
|------|---------|
| `.github/workflows/auto-release.yaml` | Updated tag generation for all 5 kinds |
| `.github/workflows/auto-release.cli.yaml` | Updated header comments |
| `.github/workflows/auto-release.app.yaml` | Updated header comments and examples |
| `.github/workflows/auto-release.website.yaml` | Updated header comments and examples |
| `.github/workflows/auto-release.pulumi-modules.yaml` | Updated tag format in detect-changes and comments |
| `.github/workflows/auto-release.terraform-modules.yaml` | Updated header comments and examples |

### Tag Prefix Changes

**CLI** (in `auto-release.yaml`):
```bash
# Before
CLI_PREFIX="${LATEST_SEMVER}.${TODAY}"

# After
CLI_PREFIX="${LATEST_SEMVER}+cli.${TODAY}"
```

**Pulumi** (in `auto-release.yaml` and `auto-release.pulumi-modules.yaml`):
```bash
# Before
TAG_PREFIX="${LATEST_SEMVER}-pulumi-${KIND}-${TODAY}"

# After
TAG_PREFIX="${LATEST_SEMVER}+pulumi.${KIND}.${TODAY}"
```

## Benefits

### Valid Semver Compliance

- All tags now pass GoReleaser's semver validation
- CLI auto-releases work correctly again
- No more build failures due to invalid version parsing

### Semantic Correctness

- Build metadata (`+`) correctly indicates "builds of version X"
- Pre-release (`-`) now reserved for actual pre-releases (alpha, beta, rc)
- Clear separation between semantic releases and auto-releases

### Consistent Format

- Same pattern (`v{semver}+{type}.{details}.{date}.{seq}`) across all kinds
- Tags = Release names (exact match, no transformation)
- Easy filtering: `git tag -l 'v0.3.2+*'` shows all auto-releases for v0.3.2

### Release Page Clarity

Tags now clearly indicate what triggered them:

```
v0.3.2+cli.20260108.0          - CLI change
v0.3.2+app.20260108.0          - App change
v0.3.2+pulumi.awsecsservice.20260108.0  - Pulumi module change
v0.3.2                         - Semantic release
```

## Semver Specification Reference

From [semver.org](https://semver.org/#spec-item-10):

> Build metadata MAY be denoted by appending a plus sign and a series of dot separated identifiers immediately following the patch or pre-release version. Identifiers MUST comprise only ASCII alphanumerics and hyphens [0-9A-Za-z-].

Build metadata is ignored when determining version precedence, making it perfect for tracking builds without affecting version comparison logic.

## Related Work

- **Prior**: Multi-Platform Pulumi Binaries (`2026-01-08-063000`)
- **Prior**: Unified Auto-Release System (`2026-01-07-200000`)
- **Prior**: Unified Release Workflow Architecture (`2026-01-07-161545`)

---

**Status**: ✅ Ready for Review
**Timeline**: ~30 minutes implementation

