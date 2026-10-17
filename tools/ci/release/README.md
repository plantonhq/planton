# Release Tooling

Scripts and utilities for managing Planton releases (IaC modules, content
distribution zips, the InfraChart catalog, definitions, the operator, the
Helm charts, and the third-party images Planton publishes).

## One rule for every release

Every release is a git tag named by what it releases and its version, cut by
a `make release-*` front door that pushes the tag. CI derives every
artifact's version from the tag; nothing in git changes because of a release.
Planton's own lines compute their next version; a third-party line's version
is upstream's own, declared once in a committed file.

| Tag | Releases | Front door |
|---|---|---|
| `vX.Y.Z` | the catalog: IaC modules, content zips, definitions, InfraCharts (`release.yaml`) | `make release` |
| `operator/vX.Y.Z` | the operator image and the `planton-operator` Helm chart, one version line (`release.operator.yaml` -> `release.helm.yaml`) | `make release-operator` |
| `helm/<chart>/vX.Y.Z` | one Helm chart under `helm/` (`release.helm.yaml`) | `make release-helm chart=<chart>` |
| `agent-substrate/vX.Y.Z` | Agent Substrate's images at `ghcr.io/plantonhq/planton/agent-substrate/<component>` and its install set at `downloads.planton.ai/releases/agent-substrate/vX.Y.Z/`, built from upstream's tag (`release.agent-substrate.yaml`) | `make release-agent-substrate` |

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

# Publish Agent Substrate's images and install set (triggers release.agent-substrate.yaml)
make release-agent-substrate               # tags the version tools/ci/release/agent-substrate.yaml declares
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
│  ├── Pulumi module binaries    -> Cloudflare R2 (downloads.planton.ai)      │
│  ├── Terraform module zips     -> Cloudflare R2 (downloads.planton.ai)      │
│  ├── Module manifest (after both module lanes) -> Cloudflare R2             │
│  ├── Content distribution zips -> Cloudflare R2 (downloads.planton.ai)      │
│  ├── Definitions               -> Cloudflare R2 (downloads.planton.ai)      │
│  └── InfraChart catalog zip    -> attached to the GitHub Release            │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Official modules: how they are built, published and recognised

A platform that deploys Planton's official IaC modules recognises each module by its **module digest**: sha256 over the `"<artifact> <sha256>"` lines, in name order, of every artifact a runner may execute for it (one gzipped Pulumi binary per platform, or the OpenTofu/Terraform `module.zip`). The digest is computed from the bytes a release actually published, so it is exact: two releases carry the same digest for a module only when they published byte-identical artifacts for it. A platform records the digest each resource was last applied with and re-runs a resource on a new release only when its module's digest moved; a runner refuses a download whose bytes are not the ones the release published.

Three pieces make that hold:

- **`tools/modulebuild` is the only place a module is built or packed.** The release lanes call it and upload what it produces. A `module.zip` holds the module directory's git-tracked files except Markdown (so a README edit never moves the digest), its provider lock file included, plus `LICENSE` and `NOTICE`. A Pulumi binary is compiled with the toolchain `go.mod` names, `GOWORK=off`, cgo off, default microarchitecture levels, `-trimpath -buildvcs=false -ldflags "-s -w -buildid="`, and gzipped with no file name or time. Archives are written by `pkg/detzip`: lexical entry order, no timestamps, stored entries. `modulebuild list` names the modules a release ships (every tracked `catalog/<provider>/<kind>/iac/{pulumi,tf}`, internal `_`-prefixed providers excluded).
- **The module manifest.** Each module leg records what it published -- module directory, artifact, and the checksum of the bytes the key holds, which `publish_once.sh` reports even for an object a rerun kept -- and uploads that list as a workflow artifact. Once both module lanes have finished, `release.yaml`'s `module-manifest` job merges every list with `modulebuild manifest` and publishes `releases/<tag>/modules/manifest.json`. It refuses an incomplete release (a module missing a platform, a module no leg recorded), so the manifest's presence means the release's modules are complete. Its grammar, and the reader platforms use, is `pkg/downloads/modulemanifest`.
- **Reproducibility gates.** `lint.reproducible-modules.yaml` builds representative modules twice, in two checkouts of one commit with empty build caches, and fails unless the bytes are identical (`check_reproducible_modules.sh`). Correctness never depends on it -- a module that rebuilt differently would only read as changed -- but it keeps a release from re-running resources whose modules it did not change.

What moves a module's digest, and so re-runs the resources built from it: any change to the files it ships (its lock file included), to `LICENSE` or `NOTICE`, and for a Pulumi module to anything compiled into its binary -- its own code, every package it imports (the kind registry included, so a release that adds a kind moves every Pulumi module), its dependencies' versions, and the Go toolchain.

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

### publish_once.sh

Uploads an artifact under its tag exactly once (a tag's artifacts never change: a rerun keeps what is published) and prints on stdout the sha256 of the bytes the key holds. Every upload records its sha256 as object metadata; a kept object without one is hashed. A rebuild that differs from a kept object is a warning, because the published bytes stand. Tests: `cd tools/ci/release && python3 -m unittest publish_once_test.py`.

### check_reproducible_modules.sh

The reproducibility gate above, runnable by hand: `bash tools/ci/release/check_reproducible_modules.sh`.

### package_content.sh

Packages the content distribution zips (presets, IaC source, catalog pages,
proto source, the kind reference pack) for the release's R2 upload.

## Agent Substrate: a third-party line

Agent Substrate publishes no images and no pre-built install, so Planton
publishes both for the Substrate kinds and for anyone else who installs it.
The pieces, all read by `release.agent-substrate.yaml`:

- `agent-substrate.yaml` declares the one version the lane builds. A pull
  request that touches the lane rehearses that version end to end (build both
  architectures, compose, check; nothing is published), `make
  release-agent-substrate` tags it, and the lane refuses to publish any other
  version that is not already published. Adopting a new upstream release moves
  this line in the pull request whose rehearsal proves it.
- `compose_agent_substrate_release.py` turns the build into the install set:
  upstream's `manifests/` tree with every image pinned by digest, the CRD
  bundle, `images.json`, the licenses and `SHA256SUMS`. Its docstring says why
  the install is upstream's whole tree rather than one file. Tests:
  `cd tools/ci/release && python3 -m unittest compose_agent_substrate_release_test.py`
  (the overlay checks need `kustomize` on PATH).
- `verify_agent_substrate_release.sh <version> [<composed directory>]` proves a
  published version from an adopter's side, with no credentials: every file
  matches its sum (and, given the directory a run composed, every byte), and
  every image pulls from ghcr.io and the Artifact Registry mirror by digest and
  by tag.

The lane runs upstream's code in a job that holds no credential, and publishes
only from jobs that run Planton's own code; its header explains the boundary.
A version is sealed once its `SHA256SUMS` is uploaded (last): later runs verify
it and build nothing.

## Required GitHub Secrets

The module/content/definitions release workflows and the Agent Substrate lane
upload to Cloudflare R2 and need `CLOUDFLARE_R2_ACCESS_KEY_ID`,
`CLOUDFLARE_R2_SECRET_ACCESS_KEY`, and `CLOUDFLARE_R2_ENDPOINT`. The GitHub
Release itself, the chart-catalog attachment and every ghcr.io push need only
the automatically provided `GITHUB_TOKEN`.

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
