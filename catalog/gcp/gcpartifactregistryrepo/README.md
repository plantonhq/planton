# GCP Artifact Registry Repository

Deploys a Google Cloud Artifact Registry repository (`google_artifact_registry_repository`) — the universal package store behind container delivery (Docker/OCI), language packages (Maven, npm, Python, Go), and OS packages (Apt, Yum), with additive per-repository IAM grants.

## Overview

Three serving modes define the operational model:

- **STANDARD_REPOSITORY** (default) — a regular repository CI pushes artifacts to; the store behind every `{location}-docker.pkg.dev` image reference.
- **REMOTE_REPOSITORY** — a pull-through cache of one upstream registry (Docker Hub, Maven Central, npmjs, PyPI, OS mirrors, or any custom registry): artifacts are fetched on first pull and served locally thereafter, insulating builds from upstream outages and rate limits.
- **VIRTUAL_REPOSITORY** — one aggregated endpoint serving from multiple Artifact Registry repositories by priority, so consumers use a single URL while artifacts live in per-team standard repos and shared caches.

`format`, `mode`, `location`, `project`, and the CMEK key are all immutable — changing any of them replaces the repository and everything stored in it. Cleanup policies keep busy repositories from growing without bound: DELETE policies remove matching versions, KEEP policies protect them (KEEP always wins on overlap).

## When to Use

- **Container delivery** — the image store for GKE, Cloud Run, and Cloud Functions workloads
- **Language package hosting** — private Maven/npm/Python/Go repositories
- **Hermetic builds** — remote caches of public registries so builds never depend on upstream availability
- **One-endpoint consumption** — virtual repositories aggregating team repos and caches

## Prerequisites

- **Deployer permissions**: the least-privilege permission set the IaC runner's principal needs lives in [`iac/permissions.yaml`](iac/permissions.yaml) (the Artifact Registry API is enabled automatically)
- For CMEK: a `GcpKmsKey` whose key the Artifact Registry service agent can use (`roles/cloudkms.cryptoKeyEncrypterDecrypter`)

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpArtifactRegistryRepo
metadata:
  name: app-images
spec:
  location: us-central1
  format: DOCKER
  cleanupPolicies:
    - id: delete-old-untagged
      action: DELETE
      condition:
        olderThan: "2592000s"
        tagState: UNTAGGED
```

This creates a private Docker repository at `us-central1-docker.pkg.dev/{project}/app-images` that cleans up superseded digests after 30 days.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | No | Target project; empty uses the provider default. Immutable |
| `repositoryId` | string | No | URL path segment; defaults to `metadata.name`. Immutable |
| `location` | string | Yes | Region or multi-region (`us`, `europe`, `asia`). Immutable |
| `format` | string | Yes | `DOCKER`, `MAVEN`, `NPM`, `PYTHON`, `GO`, `APT`, `YUM`, `GENERIC`, `KFP`, or any newer API format. Immutable |
| `mode` | string | No | `STANDARD_REPOSITORY` (default), `REMOTE_REPOSITORY`, `VIRTUAL_REPOSITORY`. Immutable |
| `description` | string | No | Human-readable description. Mutable |
| `labels` | map | No | User labels, merged beneath platform labels. Mutable |
| `kmsKeyName` | StringValueOrRef | No | CMEK key (reference a `GcpKmsKey`). Immutable |
| `dockerConfig.immutableTags` | bool | No | Tags permanently identify one digest. Mutable |
| `mavenConfig` | object | No | `versionPolicy` (`RELEASE`/`SNAPSHOT`) + `allowSnapshotOverwrites`. Effectively immutable |
| `cleanupPolicies` | list | No | DELETE/KEEP policies over age, tag state, and prefixes. Mutable |
| `cleanupPolicyDryRun` | bool | No | Log matches without deleting. Mutable |
| `remoteRepositoryConfig` | object | No | Upstream source for REMOTE mode (one arm exactly); credentials rotate in place, everything else immutable |
| `virtualRepositoryConfig` | object | No | Priority-ordered upstream policies for VIRTUAL mode. Mutable |
| `vulnerabilityScanningEnablement` | string | No | `INHERITED` (default) or `DISABLED`. Mutable |
| `iamMembers` | list | No | Additive grants: `role` + `member` (+ optional IAM condition) |
| `deletionPolicy` | string | No | `DELETE` (default: destroy removes the repository and every artifact), `PREVENT` (destroy fails), `ABANDON` (leaves the repository serving unmanaged). Mutable |

## Outputs

| Output | Description |
|--------|-------------|
| `name` | Short repository name (the repository ID) |
| `repository_path` | `projects/{p}/locations/{l}/repositories/{r}` — the composition key consumers reference |
| `registry_uri` | The push/pull endpoint, e.g. `us-central1-docker.pkg.dev/{project}/{repo}` |
| `location` | Repository location |

See the [presets](presets/) for remixable starting points and GUIDE.md for the deep dive.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
