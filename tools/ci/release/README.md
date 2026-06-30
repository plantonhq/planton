# Release Tooling

Scripts and utilities for managing Planton CLI releases.

## Quick Start

```bash
# Show what the next version would be (defaults to patch bump)
make next-version

# Show next minor version
make next-version bump=minor

# Create a release (triggers GitHub Actions workflow)
make release                   # patch bump: v0.0.0 -> v0.0.1
make release bump=minor        # minor bump: v0.0.0 -> v0.1.0
make release bump=major        # major bump: v0.0.0 -> v1.0.0
```

## Release Flow

When you run `make release`, here's what happens:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Local (make release)                                                       │
│  ├── Calculate next version using tools/ci/release/next_version.py         │
│  ├── Create git tag (e.g., v1.0.0)                                          │
│  └── Push tag to origin                                                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  GitHub Actions (.github/workflows/release.yml)                             │
│  ├── GoReleaser v2 builds CLI binaries for all platforms:                   │
│  │   ├── darwin-amd64, darwin-arm64                                         │
│  │   ├── linux-amd64, linux-arm64                                           │
│  │   └── windows-amd64, windows-arm64                                       │
│  ├── Create GitHub Release with auto-generated notes                        │
│  └── Update Homebrew Cask (planton/homebrew-tap)                    │
└─────────────────────────────────────────────────────────────────────────────┘
```

## CLI Distribution

### Install via Homebrew (macOS)

```bash
brew install --cask plantonhq/tap/planton
```

### Install via Direct Download

Download binaries from the [GitHub Releases](https://github.com/plantonhq/planton/releases) page.

```bash
# macOS (Apple Silicon)
curl -Lo planton https://github.com/plantonhq/planton/releases/latest/download/planton_VERSION_darwin_arm64.tar.gz
tar -xzf planton_VERSION_darwin_arm64.tar.gz
chmod +x planton
sudo mv planton /usr/local/bin/

# macOS (Intel)
curl -Lo planton https://github.com/plantonhq/planton/releases/latest/download/planton_VERSION_darwin_amd64.tar.gz
tar -xzf planton_VERSION_darwin_amd64.tar.gz
chmod +x planton
sudo mv planton /usr/local/bin/

# Linux (x86_64)
curl -Lo planton https://github.com/plantonhq/planton/releases/latest/download/planton_VERSION_linux_amd64.tar.gz
tar -xzf planton_VERSION_linux_amd64.tar.gz
chmod +x planton
sudo mv planton /usr/local/bin/
```

## Scripts

### next_version.py

Calculates the next semantic version based on existing git tags.

```bash
# Usage
python3 tools/ci/release/next_version.py [patch|minor|major]

# Examples
python3 tools/ci/release/next_version.py          # patch bump (default)
python3 tools/ci/release/next_version.py minor    # minor bump
python3 tools/ci/release/next_version.py major    # major bump
```

The script:
- Finds the latest tag matching strict `vX.Y.Z` pattern
- Defaults to `v0.0.0` if no tags exist
- Outputs the next version to stdout

## Required GitHub Secrets

The release workflow requires these secrets to be configured in the repository:

| Secret | Description |
|--------|-------------|
| `HOMEBREW_TAP_GITHUB_TOKEN` | GitHub PAT with `repo` scope for pushing to homebrew-tap |

Note: `GITHUB_TOKEN` is automatically provided by GitHub Actions.

### Setting Up Homebrew Tap Token

1. Go to GitHub → Settings → Developer Settings → Personal Access Tokens
2. Create a new token (classic) with `repo` scope
3. The token needs write access to `planton/homebrew-tap`
4. Add it as a repository secret named `HOMEBREW_TAP_GITHUB_TOKEN`

## Troubleshooting

### Release workflow failed

Check the GitHub Actions logs at:
https://github.com/plantonhq/planton/actions

Common issues:
- Missing or expired `HOMEBREW_TAP_GITHUB_TOKEN` secret
- Homebrew tap push permissions

### Version not incrementing correctly

The script only recognizes strict `vX.Y.Z` tags. Tags with suffixes (like `v1.0.0-beta`) are ignored.

```bash
# Check what tag will be used as base
git tag --list 'v*' --sort=-v:refname | head -5
```

### GoReleaser errors

Run a local snapshot build to debug:

```bash
goreleaser release --snapshot --clean --skip=publish
```

