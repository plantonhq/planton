# Automated Release Pipeline with GoReleaser and Homebrew Casks

**Date**: December 29, 2025
**Type**: Feature
**Components**: Build System, CI/CD, CLI Distribution

## Summary

Implemented a fully automated release pipeline for the Planton CLI using GoReleaser v2 and GitHub Actions. Running `make release` now auto-calculates the next semantic version, creates a git tag, and triggers a workflow that builds cross-platform binaries, creates a GitHub release, and auto-updates the Homebrew Cask.

## Problem Statement / Motivation

The release process required significant manual intervention and was error-prone:

### Pain Points

- **Manual version specification**: Every release required explicitly passing `version=vX.Y.Z`, prone to typos
- **Local tooling dependency**: Required `gh` CLI and authentication on developer machines
- **No Windows builds**: CLI was only available for darwin and linux
- **GCS-based distribution**: Binary downloads required Google Cloud Storage access
- **Manual Homebrew updates**: Formula had to be manually edited after each release with new version and checksums
- **macOS Gatekeeper warnings**: Users saw security warnings when running unsigned binaries

## Solution / What's New

A three-part release automation system:

### 1. Version Calculation Script

Created `tools/ci/release/next_version.py` with semver logic:
- Strict pattern matching for `vX.Y.Z` tags only
- Ignores tags with suffixes (like `v1.0.0-beta`)
- Supports `patch`, `minor`, `major` bump types
- Defaults to `v0.0.0` if no tags exist

### 2. GoReleaser Configuration

Created `.goreleaser.yaml` with GoReleaser v2 features:
- Cross-platform builds: darwin, linux, windows (amd64 + arm64)
- tar.gz archives (zip for Windows)
- Automatic checksums generation
- Homebrew Cask integration with auto-update
- macOS Gatekeeper fix via post-install hook

### 3. GitHub Actions Workflow

Created `.github/workflows/release.yml`:
- Triggers on `v*` tag pushes
- Uses GoReleaser v2 action
- Creates GitHub releases with auto-generated notes
- Pushes Cask updates to homebrew-tap repository

## Implementation Details

### GoReleaser Configuration

```yaml
version: 2
project_name: planton

builds:
  - ldflags:
      - -s -w -X github.com/plantonhq/planton/internal/cli/version.Version={{.Version}}
    goos: [darwin, linux, windows]
    goarch: [amd64, arm64]

homebrew_casks:
  - name: planton
    repository:
      owner: planton
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    binaries: [planton]
    hooks:
      post:
        install: |
          if OS.mac?
            system_command "/usr/bin/xattr", args: ["-dr", "com.apple.quarantine", "#{staged_path}/planton"]
          end
```

### Makefile Updates

Simplified release targets:

```makefile
bump ?= patch

next-version:
	@python3 tools/ci/release/next_version.py $(bump)

release: test
	@version=$$(python3 tools/ci/release/next_version.py $(bump)); \
	echo "Releasing: $$version ($(bump) bump)"; \
	git tag -a $$version -m "$$version"; \
	git push origin $$version
```

### Homebrew Tap Migration

Created `tap_migrations.json` for seamless migration from Formula to Cask:

```json
{
  "planton": "plantonhq/tap/planton"
}
```

## Benefits

### For Maintainers

- **Zero-friction releases**: `make release` with no arguments needed
- **No local tooling**: No `gh` CLI or cloud credentials required
- **Version preview**: `make next-version` shows what would be released
- **Consistent versioning**: No typos or version confusion

### For Users

- **Windows support**: CLI now available for Windows (amd64/arm64)
- **No security warnings**: macOS Gatekeeper quarantine auto-removed
- **Easy installation**: `brew install --cask plantonhq/tap/planton`
- **GitHub Releases**: Direct downloads without cloud authentication

### For CI/CD

- **Parallel builds**: 6 platform builds run via GoReleaser
- **Automatic checksums**: SHA256 checksums in release artifacts
- **Idempotent**: Homebrew update skips if no changes needed

## Impact

### Installation Changes

| Before | After |
|--------|-------|
| GCS downloads | GitHub Releases |
| Formula (`brew install plantonhq/tap/planton`) | Cask (`brew install --cask plantonhq/tap/planton`) |
| Manual version argument | Auto-calculated via Python script |
| Darwin + Linux only | Darwin + Linux + Windows |
| Gatekeeper warnings | No warnings (quarantine removed) |

### Files Changed

| File | Action | Description |
|------|--------|-------------|
| `tools/ci/release/next_version.py` | Created | Version calculation script |
| `tools/ci/release/README.md` | Created | Release tooling documentation |
| `.goreleaser.yaml` | Created | GoReleaser v2 configuration |
| `.github/workflows/release.yml` | Created | GitHub Actions workflow |
| `Makefile` | Updated | Simplified release targets |
| `homebrew-tap/tap_migrations.json` | Created | Cask migration file |

## Related Work

- Inspired by gitr release pipeline automation
- Uses same GoReleaser v2 patterns and Homebrew Cask approach
- Part of broader effort to standardize release processes across projects

---

**Status**: ✅ Production Ready (pending `HOMEBREW_TAP_GITHUB_TOKEN` secret setup)
**Timeline**: Single session implementation

