# GCP Artifact Registry Repo

Deploys a Google Cloud Artifact Registry repository — the universal package store for container images (Docker/OCI), language packages (Maven, npm, Python, Go), and OS packages (Apt, Yum). A repository serves in one of three modes: **standard** (a regular repository CI pushes to), **remote** (a pull-through cache of one upstream registry, insulating builds from outages and rate limits), or **virtual** (one aggregated endpoint serving from multiple Artifact Registry repositories by priority). Supports CMEK encryption, immutable Docker tags, Maven version policies, automatic cleanup policies, vulnerability scanning, and additive per-repository IAM grants.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Artifact Registry API enablement** (`artifactregistry.googleapis.com`) on the target project (never disabled on destroy)
- **Artifact Registry Repository** -- in the chosen project and location, with the declared format and serving mode
- **Remote upstream configuration** -- for `REMOTE_REPOSITORY` mode: exactly one upstream (Docker Hub, Maven Central, npmjs, PyPI, an Apt/Yum mirror, or a custom registry URI/AR repository), optionally with authenticated credentials referencing a Secret Manager version
- **Virtual upstream policies** -- for `VIRTUAL_REPOSITORY` mode: the priority-ordered Artifact Registry repositories the endpoint serves from
- **Cleanup policies** -- DELETE policies remove matching versions; KEEP policies protect matches from every DELETE (KEEP wins on overlap); dry-run rehearses policies in Cloud Audit Logs before anything deletes
- **IAM grants** -- one `google_artifact_registry_repository_iam_member` per `iamMembers` entry; additive, composing safely with grants made elsewhere

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **Cloud KMS key** (if using CMEK) with `roles/cloudkms.cryptoKeyEncrypterDecrypter` granted to the Artifact Registry service agent.
- **Secret Manager secret** (if the remote upstream needs credentials) with `roles/secretmanager.secretAccessor` granted to the Artifact Registry service agent — the spec carries only the secret VERSION PATH, never the password.

## Deploy

### Console

Open the deployment store, find **GCP Artifact Registry Repo**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Docker Repository** preset in the [Presets](#presets) tab for the CI/CD workhorse shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpArtifactRegistryRepo
metadata:
  name: team-images
  org: acme-corp
  env: prod
spec:
  location: us-central1
  format: DOCKER
  dockerConfig:
    immutableTags: true
  cleanupPolicies:
    - id: delete-untagged
      action: DELETE
      condition:
        tagState: UNTAGGED
        olderThan: 2592000s
    - id: keep-recent-10
      action: KEEP
      mostRecentVersions:
        keepCount: 10
```

```shell
planton apply -f artifact-registry-repo.yaml
```

This creates a standard Docker repository at `us-central1-docker.pkg.dev/{project}/team-images` with immutable tags and a self-cleaning storage policy. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying the production trio (per-team standard repos + a shared cache + one virtual endpoint), wire the virtual repository's upstreams to the sibling repositories deployed in the same InfraPipeline:

```yaml
spec:
  location: us-central1
  format: DOCKER
  mode: VIRTUAL_REPOSITORY
  virtualRepositoryConfig:
    upstreamPolicies:
      - id: team-images
        repository:
          valueFrom:
            kind: GcpArtifactRegistryRepo
            name: team-images
            fieldPath: status.outputs.repository_path
        priority: 100
      - id: dockerhub-cache
        repository:
          valueFrom:
            kind: GcpArtifactRegistryRepo
            name: dockerhub-cache
            fieldPath: status.outputs.repository_path
        priority: 50
```

The InfraPipeline resolves the dependency graph, deploys the upstream repositories first, then aggregates them behind the virtual endpoint.

## Key Configuration

These are the most important decisions when configuring a repository. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Format and location (immutable)** -- `format` is the package type (`DOCKER`, `MAVEN`, `NPM`, `PYTHON`, `GO`, `APT`, `YUM`, `GENERIC`, ...; GCP adds formats over time, so any API-accepted value works). `location` is a region (co-locate with builds/runtimes for pull latency and egress cost) or a multi-region (`us`, `europe`, `asia`). Both — along with the project, repository ID, and CMEK key — destroy and recreate the repository (and everything in it) if changed.

**Serving mode (immutable)** -- Leave `mode` unset for a standard push repository. `REMOTE_REPOSITORY` requires `remoteRepositoryConfig` with exactly one upstream; `VIRTUAL_REPOSITORY` requires `virtualRepositoryConfig` with at least one prioritized upstream. The production trio composes all three.

**Cleanup policies (mutable)** -- Without them, CI-pushed repositories grow without bound. The classic Docker pair: DELETE untagged versions older than 30 days + KEEP the 10 most recent versions per package. Set `cleanupPolicyDryRun: true` to rehearse against real traffic first.

**Access (mutable)** -- Each `iamMembers` entry grants one role to one member on this repository: `roles/artifactregistry.reader` for runtime service accounts that pull, `roles/artifactregistry.writer` for CI service accounts that push. Reference a `GcpServiceAccount`'s `member` output, or grant `allUsers` for public access (with care).

**Docker immutable tags (mutable)** -- With `dockerConfig.immutableTags`, a pushed tag permanently identifies one digest — reproducible deployments at the cost of mutable tags like `latest`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |
| **GcpServiceAccount** (optional) | `iamMembers[].member` | `status.outputs.member` |
| **GcpArtifactRegistryRepo** (optional) | `virtualRepositoryConfig.upstreamPolicies[].repository` | `status.outputs.repository_path` |
| **GcpArtifactRegistryRepo** (optional) | `remoteRepositoryConfig.commonRepository.uri` | `status.outputs.repository_path` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `repository_path` | Full resource path (`projects/{project}/locations/{location}/repositories/{repo}`) | Virtual upstreams, custom remote upstreams — the repo-to-repo composition key |
| `registry_uri` | The pull/push handle (`{location}-{format}.pkg.dev/{project}/{repo}`) | docker push/pull, npm config, pip index-url, service image references |
| `name` | The repository ID (last path segment) | Naming, correlation |
| `location` | The region or multi-region | Co-location checks, tooling configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard Docker Repository** -- The CI/CD workhorse: a private Docker repository with immutable tags, a self-cleaning storage policy (delete untagged + keep recent), and an additive pull grant for the runtime service account. Start from the **Standard Docker Repository** preset.

**Remote Docker Hub Cache** -- A pull-through cache of Docker Hub: first pulls fetch and cache, later pulls serve from GCP — builds survive upstream outages and per-IP rate limits. Start from the **Remote Docker Hub Cache** preset.

**Virtual Aggregation Endpoint** -- One URL for every consumer: the virtual repository serves from the team's standard repo (priority 100) and the shared Docker Hub cache (priority 50) — your own artifacts always win. Start from the **Virtual Aggregation Endpoint** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project when it differs from the connection default
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the CMEK encryption key for artifacts at rest
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- its `member` output is exactly what an IAM grant consumes
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) / [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) -- pull images from the repository's `registry_uri`
