# Release Tooling

Scripts and utilities for managing Planton releases (IaC modules, content
distribution zips, the InfraChart catalog, definitions, the operator, and the
Helm charts).

## One rule for every release

Every release is a git tag named by the artifact's directory and its version,
cut by a `make release-*` front door that computes the next version and
pushes the tag. CI derives every artifact's version from the tag; nothing in
git changes because of a release.

| Tag | Releases | Front door |
|---|---|---|
| `vX.Y.Z` | the catalog: IaC modules, content zips, definitions, InfraCharts (`release.yaml`) | `make release` |
| `operator/vX.Y.Z` | the operator image and the `planton-operator` Helm chart, one version line (`release.operator.yaml` -> `release.helm.yaml`) | `make release-operator` |
| `helm/<chart>/vX.Y.Z` | one Helm chart under `helm/` (`release.helm.yaml`) | `make release-helm chart=<chart>` |

## Quick Start

```bash
# Show what the next catalog version would be (defaults to patch bump)
make next-version
make next-version bump=minor

# Release the catalog (triggers release.yaml)
make release                   # patch bump: v0.5.26 -> v0.5.27
make release bump=minor        # minor bump
make release version=v1.0.0    # an explicit version

# Release the operator image + chart (triggers release.operator.yaml)
make release-operator                      # patch bump in the operator/ namespace
make release-operator version=v0.8.0       # an explicit version (the first tag of a namespace)

# Release one Helm chart (triggers release.helm.yaml)
make release-helm chart=planton            # patch bump in the helm/planton/ namespace
make release-helm chart=planton version=v0.4.0
```

A namespace with no tag yet auto-bumps to `v0.0.1`, so a line that must start
elsewhere (a chart that already has published versions from before it was
tagged) cuts its first tag with an explicit `version=`.

## Release Flow

When you run `make release`, here's what happens:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│  Local (make release)                                                       │
│  ├── Calculate next version using tools/ci/release/next_version.py          │
│  ├── Create git tag (e.g., v1.0.0)                                          │
│  └── Push tag to origin                                                     │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  GitHub Actions (.github/workflows/release.yaml)                            │
│  ├── Create the GitHub Release with auto-generated notes                    │
│  ├── Pulumi module binaries    -> Cloudflare R2 (downloads.planton.dev)     │
│  ├── Terraform module zips     -> Cloudflare R2 (downloads.planton.dev)     │
│  ├── Content distribution zips -> Cloudflare R2 (downloads.planton.dev)     │
│  ├── Definitions               -> Cloudflare R2 (downloads.planton.dev)     │
│  └── InfraChart catalog zip    -> attached to the GitHub Release            │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Scripts

### next_version.py

Calculates the next semantic version based on existing git tags, in one tag
namespace at a time.

```bash
# Usage
python3 tools/ci/release/next_version.py [patch|minor|major] [--prefix <namespace>/]

# Examples
python3 tools/ci/release/next_version.py                          # catalog, patch bump (default)
python3 tools/ci/release/next_version.py minor                    # catalog, minor bump
python3 tools/ci/release/next_version.py --prefix operator/       # operator, patch bump
python3 tools/ci/release/next_version.py --prefix helm/planton/   # planton chart, patch bump
```

The script:
- Finds the newest tag matching `<prefix>vX.Y.Z` (strict semver after the prefix; pre-release and metadata tags never match)
- Defaults to `v0.0.0` if the namespace has no tag
- Prints the bare next version (`vX.Y.Z`); the caller prepends the prefix when it tags

Tests: `cd tools/ci/release && python3 -m unittest next_version_test.py`.

### detect_module_dirs.sh

Which changed files count as Pulumi/Terraform module changes (maturity-grammar-aware), with a `--self-test` that proves the patterns fire for every version channel.

Not wired to CI since 2026-09-04: module auto-tags are dispatch-only by founder direction (`force_*_all` tags every module, so nothing detects changes), and `auto-tag.yaml` no longer calls this script or its self-test. It stays because it is what push-time module tagging needs and that may come back; re-enabling restores the two `iac/` push paths, the detector call, and the self-test step in `auto-tag.yaml` (see `.github/workflows/docs/auto-tags.md`). Run `bash tools/ci/release/detect_module_dirs.sh --self-test` by hand to prove it still works.

### package_content.sh

Packages the content distribution zips (presets, IaC source, catalog pages,
proto source, the kind reference pack) for the release's R2 upload.

## Required GitHub Secrets

The module/content/definitions release workflows upload to Cloudflare R2 and
need `CLOUDFLARE_R2_ACCESS_KEY_ID`, `CLOUDFLARE_R2_SECRET_ACCESS_KEY`, and
`CLOUDFLARE_R2_ENDPOINT`. The GitHub Release itself and the chart-catalog
attachment need only the automatically provided `GITHUB_TOKEN`.

## Troubleshooting

### Release workflow failed

Check the GitHub Actions logs at:
https://github.com/plantonhq/planton/actions

### Version not incrementing correctly

The script only recognizes strict `vX.Y.Z` tags. Tags with suffixes (like `v1.0.0-beta`) are ignored.

```bash
# Check what tag will be used as base
git tag --list 'v*' --sort=-v:refname | head -5
```
