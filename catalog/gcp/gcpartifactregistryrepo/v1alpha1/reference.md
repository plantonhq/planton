# GcpArtifactRegistryRepo

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpArtifactRegistryRepoSpec defines the configuration for a Google Cloud
Artifact Registry repository — the universal package store for container
images (Docker/OCI), language packages (Maven, npm, Python, Go), and OS
packages (Apt, Yum).

A repository has one of three serving modes:

  - STANDARD_REPOSITORY (default): a regular repository you push artifacts
    to. The workhorse for CI/CD pipelines publishing images and packages.
  - REMOTE_REPOSITORY: a pull-through cache of an upstream registry
    (Docker Hub, Maven Central, npmjs, PyPI, OS mirrors, or a custom
    registry). Artifacts are fetched and cached on first pull — insulating
    builds from upstream outages and rate limits.
  - VIRTUAL_REPOSITORY: a single aggregated endpoint that serves from
    multiple upstream Artifact Registry repositories by priority. Lets
    consumers use one URL while artifacts live in per-team standard repos
    and shared remote caches.

Important behavioral notes:

  - format, mode, location, project, and kms_key_name are all immutable
    after creation — changing any of them destroys and recreates the
    repository (and everything stored in it). Choose deliberately.
  - The entire remote_repository_config block is immutable EXCEPT the
    upstream credentials and disable_upstream_validation, which rotate
    in place.
  - Access control is additive IAM: each iam_members entry grants one
    role to one member on this repository and composes safely with grants
    made elsewhere. Authoritative bindings/policies (which clobber grants
    they don't list) are deliberately not modeled.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpArtifactRegistryRepo
metadata:
  name: test-artifact-repo
spec:
  location: us-central1
  format: DOCKER
  description: container images for the test service
  dockerConfig:
    immutableTags: true
  cleanupPolicies:
    - id: delete-old-untagged
      action: DELETE
      condition:
        olderThan: "2592000s"
        tagState: UNTAGGED
    - id: keep-last-10
      action: KEEP
      mostRecentVersions:
        keepCount: 10
  labels:
    team: platform
  # Destroy really destroys in E2E: the live lanes prove the full lifecycle.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.repositoryId` | `string` |  |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.format` | `string` | yes |  |  |
| `spec.mode` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`) |
| `spec.dockerConfig` | `GcpArtifactRegistryRepoDockerConfig` |  |  |  |
| `spec.dockerConfig.immutableTags` | `bool` |  |  |  |
| `spec.mavenConfig` | `GcpArtifactRegistryRepoMavenConfig` |  |  |  |
| `spec.mavenConfig.versionPolicy` | `string` |  |  |  |
| `spec.mavenConfig.allowSnapshotOverwrites` | `bool` |  |  |  |
| `spec.cleanupPolicies` | `[]GcpArtifactRegistryRepoCleanupPolicy` |  |  |  |
| `spec.cleanupPolicies[].id` | `string` | yes |  |  |
| `spec.cleanupPolicies[].action` | `string` | yes |  |  |
| `spec.cleanupPolicies[].condition` | `GcpArtifactRegistryRepoCleanupPolicyCondition` |  |  |  |
| `spec.cleanupPolicies[].condition.newerThan` | `string` |  |  |  |
| `spec.cleanupPolicies[].condition.olderThan` | `string` |  |  |  |
| `spec.cleanupPolicies[].condition.packageNamePrefixes` | `[]string` |  |  |  |
| `spec.cleanupPolicies[].condition.tagPrefixes` | `[]string` |  |  |  |
| `spec.cleanupPolicies[].condition.tagState` | `string` |  |  |  |
| `spec.cleanupPolicies[].condition.versionNamePrefixes` | `[]string` |  |  |  |
| `spec.cleanupPolicies[].mostRecentVersions` | `GcpArtifactRegistryRepoCleanupPolicyMostRecentVersions` |  |  |  |
| `spec.cleanupPolicies[].mostRecentVersions.keepCount` | `int32` |  |  |  |
| `spec.cleanupPolicies[].mostRecentVersions.packageNamePrefixes` | `[]string` |  |  |  |
| `spec.cleanupPolicyDryRun` | `bool` |  |  |  |
| `spec.remoteRepositoryConfig` | `GcpArtifactRegistryRepoRemoteConfig` |  |  |  |
| `spec.remoteRepositoryConfig.description` | `string` |  |  |  |
| `spec.remoteRepositoryConfig.dockerPublicRepository` | `string` |  |  |  |
| `spec.remoteRepositoryConfig.mavenPublicRepository` | `string` |  |  |  |
| `spec.remoteRepositoryConfig.npmPublicRepository` | `string` |  |  |  |
| `spec.remoteRepositoryConfig.pythonPublicRepository` | `string` |  |  |  |
| `spec.remoteRepositoryConfig.aptRepository` | `GcpArtifactRegistryRepoRemoteAptRepository` |  |  |  |
| `spec.remoteRepositoryConfig.aptRepository.repositoryBase` | `string` | yes |  |  |
| `spec.remoteRepositoryConfig.aptRepository.repositoryPath` | `string` | yes |  |  |
| `spec.remoteRepositoryConfig.yumRepository` | `GcpArtifactRegistryRepoRemoteYumRepository` |  |  |  |
| `spec.remoteRepositoryConfig.yumRepository.repositoryBase` | `string` | yes |  |  |
| `spec.remoteRepositoryConfig.yumRepository.repositoryPath` | `string` | yes |  |  |
| `spec.remoteRepositoryConfig.commonRepository` | `GcpArtifactRegistryRepoRemoteCommonRepository` |  |  |  |
| `spec.remoteRepositoryConfig.commonRepository.uri` | `string \| valueFrom` | yes |  | GcpArtifactRegistryRepo (`status.outputs.repository_path`) |
| `spec.remoteRepositoryConfig.upstreamCredentials` | `GcpArtifactRegistryRepoRemoteUpstreamCredentials` |  |  |  |
| `spec.remoteRepositoryConfig.upstreamCredentials.username` | `string` | yes |  |  |
| `spec.remoteRepositoryConfig.upstreamCredentials.passwordSecretVersion` | `string` | yes |  |  |
| `spec.remoteRepositoryConfig.disableUpstreamValidation` | `bool` |  |  |  |
| `spec.virtualRepositoryConfig` | `GcpArtifactRegistryRepoVirtualConfig` |  |  |  |
| `spec.virtualRepositoryConfig.upstreamPolicies` | `[]GcpArtifactRegistryRepoVirtualUpstreamPolicy` | yes |  |  |
| `spec.virtualRepositoryConfig.upstreamPolicies[].id` | `string` | yes |  |  |
| `spec.virtualRepositoryConfig.upstreamPolicies[].repository` | `string \| valueFrom` | yes |  | GcpArtifactRegistryRepo (`status.outputs.repository_path`) |
| `spec.virtualRepositoryConfig.upstreamPolicies[].priority` | `int32` |  |  |  |
| `spec.vulnerabilityScanningEnablement` | `string` |  |  |  |
| `spec.iamMembers` | `[]GcpArtifactRegistryRepoIamMember` |  |  |  |
| `spec.iamMembers[].role` | `string` | yes |  |  |
| `spec.iamMembers[].member` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.member`) |
| `spec.iamMembers[].condition` | `GcpArtifactRegistryRepoIamCondition` |  |  |  |
| `spec.iamMembers[].condition.title` | `string` | yes |  |  |
| `spec.iamMembers[].condition.expression` | `string` | yes |  |  |
| `spec.iamMembers[].condition.description` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project where the repository is created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable after creation.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.repositoryId

`string`

The last segment of the repository's resource name — the ID that appears
in registry URLs (e.g. "us-docker.pkg.dev/{project}/{repository_id}").
Must start with a letter, contain only lowercase letters, numbers, and
hyphens, and be at most 63 characters. If omitted, metadata.name is used.
Immutable after creation.

- rule: repository_id must start with a lowercase letter and contain only lowercase letters, numbers, and hyphens (max 63 characters)

### spec.location

`string` · required

The location for the repository: a region (e.g. "us-central1",
"asia-south1") or a multi-region ("us", "europe", "asia"). Regional
repositories co-located with your build/runtime infrastructure minimize
pull latency and egress cost; multi-region maximizes availability for
globally consumed artifacts. Immutable after creation.

- rule: {"required":true}

### spec.format

`string` · required

The package format stored in this repository. Common values:
  "DOCKER"  -- container images (Docker/OCI); the format behind
               us-docker.pkg.dev URLs and GKE/Cloud Run image pulls
  "MAVEN"   -- Java/JVM packages
  "NPM"     -- Node.js packages
  "PYTHON"  -- Python packages (pip/PyPI layout)
  "GO"      -- Go module proxy (remote mode caches proxy.golang.org)
  "APT"     -- Debian/Ubuntu OS packages
  "YUM"     -- RHEL/CentOS/Rocky OS packages
  "GENERIC" -- arbitrary versioned files
  "KFP"     -- Kubeflow Pipelines templates

Artifact Registry adds formats over time, so this field deliberately
accepts any string (matched case-insensitively by GCP) and lets the API
validate — see the repository format reference for the authoritative
list: https://cloud.google.com/artifact-registry/docs/supported-formats
Immutable after creation.

- rule: {"required":true}

### spec.mode

`string`

The serving mode of the repository. Defaults to STANDARD_REPOSITORY.

  "STANDARD_REPOSITORY" -- artifacts are pushed directly to this repo
  "REMOTE_REPOSITORY"   -- pull-through cache of one upstream source
                           (requires remote_repository_config)
  "VIRTUAL_REPOSITORY"  -- priority-ordered aggregation of other
                           Artifact Registry repositories (requires
                           virtual_repository_config)

Immutable after creation.

- rule: mode must be one of: STANDARD_REPOSITORY, REMOTE_REPOSITORY, VIRTUAL_REPOSITORY

### spec.description

`string`

Human-readable description of the repository's purpose, shown in the
console and API listings. Mutable in place.

### spec.labels

`map<string, string>`

User-defined labels attached to the repository, for cost attribution
and fleet queries. Merged with Planton's platform labels (which win on
key conflicts). Mutable in place.

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key (CMEK) protecting artifacts in this
repository. Accepts the fully qualified crypto key path
  projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
or a reference to a GcpKmsKey resource. The Artifact Registry service
agent needs roles/cloudkms.cryptoKeyEncrypterDecrypter on the key.
If omitted, artifacts are encrypted with Google-managed keys.
Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.dockerConfig

`GcpArtifactRegistryRepoDockerConfig`

Docker-format-specific configuration. Only meaningful when format is
DOCKER. Mutable in place.

### spec.dockerConfig.immutableTags

`bool`

If true, image tags cannot be modified, moved, or deleted once created
— a pushed tag permanently identifies one digest. New tags can still be
created. Immutable tags make deployments reproducible ("v1.2.3 today is
v1.2.3 forever") at the cost of losing mutable convenience tags like
"latest". Mutable in place.

### spec.mavenConfig

`GcpArtifactRegistryRepoMavenConfig`

Maven-format-specific configuration. Only meaningful when format is
MAVEN. Effectively immutable: GCP rejects changing the version policy
or snapshot-overwrite behavior on an existing repository.

### spec.mavenConfig.versionPolicy

`string`

The Maven version policy for this repository:
  ""         -- accept both release and snapshot versions (default)
  "RELEASE"  -- accept only release versions
  "SNAPSHOT" -- accept only snapshot versions
  "VERSION_POLICY_UNSPECIFIED" -- the API's explicit mixed-versions
               sentinel (equivalent to leaving this unset)
The conventional Maven setup is a RELEASE repository and a SNAPSHOT
repository, with builds publishing to each as appropriate.

- rule: version_policy must be one of: VERSION_POLICY_UNSPECIFIED, RELEASE, SNAPSHOT

### spec.mavenConfig.allowSnapshotOverwrites

`bool`

If true, re-publishing a non-snapshot artifact at an existing version
overwrites it. Defaults to false — GCP rejects duplicate uploads,
preserving release immutability (the safer posture for RELEASE repos).

### spec.cleanupPolicies

`[]GcpArtifactRegistryRepoCleanupPolicy`

Cleanup policies that automatically delete or protect artifact versions
based on age, tag state, or name prefixes. Policies with action DELETE
remove matching versions; policies with action KEEP protect matching
versions from every DELETE policy (KEEP always wins on overlap). Without
cleanup policies, CI-pushed repositories grow without bound — pair every
busy standard repository with at least a delete-untagged policy.
Mutable in place.

- rule: a KEEP policy needs a condition or most_recent_versions to define what it protects
- rule: most_recent_versions is only valid on a KEEP policy (it defines versions to protect, not delete)

### spec.cleanupPolicies[].id

`string` · required

Unique identifier for this policy within the repository.

- rule: {"required":true,"string":{"maxLen":"128"}}

### spec.cleanupPolicies[].action

`string` · required

What to do with versions matching this policy:
  "DELETE" -- delete matching versions
  "KEEP"   -- protect matching versions from all DELETE policies
When a version matches both a DELETE and a KEEP policy, KEEP wins.

- rule: action must be one of: DELETE, KEEP
- rule: {"required":true}

### spec.cleanupPolicies[].condition

`GcpArtifactRegistryRepoCleanupPolicyCondition`

Selects versions by age, tag state, and name prefixes. All specified
criteria must match (logical AND).

### spec.cleanupPolicies[].condition.newerThan

`string`

Match versions newer than this duration since upload, in seconds with
an "s" suffix (e.g. "2592000s" for 30 days). Typically used on KEEP
policies ("protect everything pushed in the last 30 days").

- rule: newer_than must be a duration in seconds (e.g., '2592000s' for 30 days)

### spec.cleanupPolicies[].condition.olderThan

`string`

Match versions older than this duration since upload, in seconds with
an "s" suffix (e.g. "7776000s" for 90 days). The workhorse of DELETE
policies ("delete anything older than 90 days").

- rule: older_than must be a duration in seconds (e.g., '7776000s' for 90 days)

### spec.cleanupPolicies[].condition.packageNamePrefixes

`[]string`

Match versions whose package name starts with any of these prefixes.

### spec.cleanupPolicies[].condition.tagPrefixes

`[]string`

Match versions with a tag starting with any of these prefixes
(e.g. "release-" to select release-tagged images).

### spec.cleanupPolicies[].condition.tagState

`string`

Match versions by tag status:
  "ANY"      -- tagged or untagged (default)
  "TAGGED"   -- only versions with at least one tag
  "UNTAGGED" -- only versions with no tags (superseded digests in
                Docker repos — the classic cleanup target)

- rule: tag_state must be one of: ANY, TAGGED, UNTAGGED

### spec.cleanupPolicies[].condition.versionNamePrefixes

`[]string`

Match versions whose version name starts with any of these prefixes.

### spec.cleanupPolicies[].mostRecentVersions

`GcpArtifactRegistryRepoCleanupPolicyMostRecentVersions`

Selects the N most recent versions (per package) to protect. Only valid
with action KEEP — the standard "always keep the last 10 builds" guard
paired with an age-based DELETE policy.

### spec.cleanupPolicies[].mostRecentVersions.keepCount

`int32`

Minimum number of the most recent versions to keep per package.

- rule: {"int32":{"gt":0}}

### spec.cleanupPolicies[].mostRecentVersions.packageNamePrefixes

`[]string`

Restrict the keep-count protection to packages whose name starts with
any of these prefixes. If empty, applies to all packages in the repo.

### spec.cleanupPolicyDryRun

`bool`

If true, cleanup policies run in dry-run mode: matches are logged
(visible in Cloud Audit Logs) but nothing is deleted. Use this to
validate new policies against real traffic before letting them delete.
Mutable in place.

### spec.remoteRepositoryConfig

`GcpArtifactRegistryRepoRemoteConfig`

Upstream source configuration for REMOTE_REPOSITORY mode — which
registry this repository caches. Required when (and only valid when)
mode is REMOTE_REPOSITORY. The block is immutable after creation EXCEPT
upstream_credentials and disable_upstream_validation, which update in
place (credential rotation never recreates the cache).

- rule: exactly one upstream source must be set: apt_repository, yum_repository, common_repository, or one of the docker/maven/npm/python public repository arms

### spec.remoteRepositoryConfig.description

`string`

Human-readable description of the upstream source. Immutable.

### spec.remoteRepositoryConfig.dockerPublicRepository

`string`

Well-known public Docker upstream. The only supported value is
"DOCKER_HUB". For any other Docker registry (ghcr.io, quay.io, another
AR repository), use common_repository instead. Immutable.

- rule: docker_public_repository only supports DOCKER_HUB — use common_repository for custom Docker registries

### spec.remoteRepositoryConfig.mavenPublicRepository

`string`

Well-known public Maven upstream. The only supported value is
"MAVEN_CENTRAL". For other Maven registries use common_repository.
Immutable.

- rule: maven_public_repository only supports MAVEN_CENTRAL — use common_repository for custom Maven registries

### spec.remoteRepositoryConfig.npmPublicRepository

`string`

Well-known public npm upstream. The only supported value is "NPMJS".
For other npm registries use common_repository. Immutable.

- rule: npm_public_repository only supports NPMJS — use common_repository for custom npm registries

### spec.remoteRepositoryConfig.pythonPublicRepository

`string`

Well-known public Python upstream. The only supported value is "PYPI".
For other Python registries use common_repository. Immutable.

- rule: python_public_repository only supports PYPI — use common_repository for custom Python registries

### spec.remoteRepositoryConfig.aptRepository

`GcpArtifactRegistryRepoRemoteAptRepository`

Public Apt upstream (Debian/Ubuntu mirror trees). Immutable.

### spec.remoteRepositoryConfig.aptRepository.repositoryBase

`string` · required

The mirror tree to cache from:
  "DEBIAN"          -- deb.debian.org
  "UBUNTU"          -- archive.ubuntu.com
  "DEBIAN_SNAPSHOT" -- snapshot.debian.org (point-in-time Debian
                       archives, for reproducible builds)

- rule: repository_base must be one of: DEBIAN, UBUNTU, DEBIAN_SNAPSHOT
- rule: {"required":true}

### spec.remoteRepositoryConfig.aptRepository.repositoryPath

`string` · required

The specific repository path within the base, e.g. "debian/dists/bookworm".

- rule: {"required":true}

### spec.remoteRepositoryConfig.yumRepository

`GcpArtifactRegistryRepoRemoteYumRepository`

Public Yum upstream (RHEL-family mirror trees). Immutable.

### spec.remoteRepositoryConfig.yumRepository.repositoryBase

`string` · required

The mirror tree to cache from:
  "CENTOS", "CENTOS_DEBUG", "CENTOS_VAULT", "CENTOS_STREAM",
  "ROCKY", "EPEL"

- rule: repository_base must be one of: CENTOS, CENTOS_DEBUG, CENTOS_VAULT, CENTOS_STREAM, ROCKY, EPEL
- rule: {"required":true}

### spec.remoteRepositoryConfig.yumRepository.repositoryPath

`string` · required

The specific repository path within the base,
e.g. "pub/rocky/9/BaseOS/x86_64/os".

- rule: {"required":true}

### spec.remoteRepositoryConfig.commonRepository

`GcpArtifactRegistryRepoRemoteCommonRepository`

Custom upstream: another Artifact Registry repository
("projects/{p}/locations/{l}/repositories/{r}") or a registry URI
(e.g. "https://ghcr.io", "https://registry.company.com"). The upstream
must serve the same format as this repository. Immutable.

### spec.remoteRepositoryConfig.commonRepository.uri

`string | valueFrom` · required

Another Artifact Registry repository
("projects/{p}/locations/{l}/repositories/{r}") or a registry URI
("https://registry.company.com"). Accepts a literal or a reference to
a GcpArtifactRegistryRepo resource (its repository_path output is
exactly the AR form of this value).

- references: GcpArtifactRegistryRepo (`status.outputs.repository_path`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpArtifactRegistryRepo, name: <that resource's name>, fieldPath: status.outputs.repository_path}} -- a bare string does not parse

### spec.remoteRepositoryConfig.upstreamCredentials

`GcpArtifactRegistryRepoRemoteUpstreamCredentials`

Credentials for authenticating to the upstream (private registries, or
Docker Hub authenticated pulls for higher rate limits). Mutable in
place: rotating credentials never recreates the cache.

### spec.remoteRepositoryConfig.upstreamCredentials.username

`string` · required

The username to authenticate with.

- rule: {"required":true}

### spec.remoteRepositoryConfig.upstreamCredentials.passwordSecretVersion

`string` · required

Secret Manager secret version holding the password, in the form
  projects/{project}/secrets/{secret}/versions/{version}
(use ".../versions/latest" to track rotation automatically). The
Artifact Registry service agent needs
roles/secretmanager.secretAccessor on the secret.

- rule: {"required":true,"string":{"pattern":"^projects/[^/]+/secrets/[^/]+/versions/[^/]+$"}}

### spec.remoteRepositoryConfig.disableUpstreamValidation

`bool`

If true, skip validating the upstream URL and credentials at
create/update time. Useful when the upstream is temporarily unreachable
from the control plane but will be reachable at pull time. Mutable in
place.

### spec.virtualRepositoryConfig

`GcpArtifactRegistryRepoVirtualConfig`

Upstream aggregation configuration for VIRTUAL_REPOSITORY mode — which
Artifact Registry repositories this endpoint serves from and in what
priority order. Required when (and only valid when) mode is
VIRTUAL_REPOSITORY. Mutable in place: upstreams can be added, removed,
and re-prioritized without recreating the virtual endpoint.

### spec.virtualRepositoryConfig.upstreamPolicies

`[]GcpArtifactRegistryRepoVirtualUpstreamPolicy` · required

The Artifact Registry repositories this virtual endpoint serves from.
When the same package exists in multiple upstreams, the highest
priority wins. All upstreams must serve this repository's format.

- rule: {"repeated":{"minItems":"1"}}

### spec.virtualRepositoryConfig.upstreamPolicies[].id

`string` · required

User-chosen identifier for this policy entry, unique within the
virtual repository.

- rule: {"required":true}

### spec.virtualRepositoryConfig.upstreamPolicies[].repository

`string | valueFrom` · required

The upstream Artifact Registry repository, as the full resource path
  projects/{project}/locations/{location}/repositories/{repository}
Accepts a literal or a reference to a GcpArtifactRegistryRepo resource.

- references: GcpArtifactRegistryRepo (`status.outputs.repository_path`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpArtifactRegistryRepo, name: <that resource's name>, fieldPath: status.outputs.repository_path}} -- a bare string does not parse

### spec.virtualRepositoryConfig.upstreamPolicies[].priority

`int32`

Serving priority — entries with higher values are tried first.

### spec.vulnerabilityScanningEnablement

`string`

Whether Artifact Analysis automatically scans artifacts in this
repository for vulnerabilities.

  ""          -- inherit the project-level setting (default)
  "INHERITED" -- same as empty: follow the project-level setting
  "DISABLED"  -- never scan this repository, regardless of project config

Scanning requires the Container Scanning API to be enabled on the
project and incurs per-artifact scan charges. Mutable in place.

- rule: vulnerability_scanning_enablement must be one of: INHERITED, DISABLED

### spec.iamMembers

`[]GcpArtifactRegistryRepoIamMember`

Additive IAM grants on this repository. Each entry grants one role to
one member and composes safely with grants made by other tools or
charts — removal subtracts only that exact (role, member) pair.

Common roles:
  roles/artifactregistry.reader -- pull artifacts (grant to runtime SAs)
  roles/artifactregistry.writer -- push and pull (grant to CI SAs)
  roles/artifactregistry.repoAdmin -- manage artifacts and settings

Public access: grant roles/artifactregistry.reader to the special
member "allUsers" (requires the project to allow public access).

### spec.iamMembers[].role

`string` · required

The role to grant, e.g. "roles/artifactregistry.reader",
"roles/artifactregistry.writer", "roles/artifactregistry.repoAdmin",
or a custom role's fully-qualified name.

- rule: {"required":true}

### spec.iamMembers[].member

`string | valueFrom` · required

The identity receiving the grant, in GCP IAM member format:
  serviceAccount:<email>  -- a service account (the most common in IaC;
                             reference a GcpServiceAccount resource —
                             its `member` output is exactly this value)
  user:<email> / group:<email> / domain:<domain>
  allUsers / allAuthenticatedUsers -- public access (grant with care)

- references: GcpServiceAccount (`status.outputs.member`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.member}} -- a bare string does not parse

### spec.iamMembers[].condition

`GcpArtifactRegistryRepoIamCondition`

Optional IAM Condition restricting when this grant applies. The
condition is part of the grant's identity: the same role with and
without a condition are two independent grants.

### spec.iamMembers[].condition.title

`string` · required

Short human-readable title identifying the condition's intent,
e.g. "expires-2026-12-31".

- rule: {"required":true,"string":{"maxLen":"100"}}

### spec.iamMembers[].condition.expression

`string` · required

The CEL condition expression, e.g.
request.time < timestamp("2027-01-01T00:00:00Z").

- rule: {"required":true}

### spec.iamMembers[].condition.description

`string`

Optional longer explanation of what the condition does.

- rule: {"string":{"maxLen":"256"}}

### spec.deletionPolicy

`string`

What destroying this resource does to the repository (IAM grants are
plain bindings — they are always removed with the destroy):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the repository and every artifact stored in it are
               deleted; images/packages become unpullable immediately
  "PREVENT" -- destroy FAILS; protects a registry that running
               workloads still pull from
  "ABANDON" -- the repository is removed from management but keeps
               serving artifacts in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `remote_config_requires_remote_mode`: remote_repository_config can only be set when mode is REMOTE_REPOSITORY
- `remote_mode_requires_remote_config`: mode REMOTE_REPOSITORY requires remote_repository_config to define the upstream source
- `virtual_config_requires_virtual_mode`: virtual_repository_config can only be set when mode is VIRTUAL_REPOSITORY
- `virtual_mode_requires_virtual_config`: mode VIRTUAL_REPOSITORY requires virtual_repository_config to define the upstream policies

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpArtifactRegistryRepo, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Short name of the repository (the repository ID), e.g. "app-images". |
| `status.outputs.repository_path` | `string` | Fully qualified resource path of the repository: projects/{project}/locations/{location}/repositories/{repository_id} This is the composition key other resources consume: a Cloud Function's docker_repository, a virtual repository's upstream policy, and a remote repository's common upstream all take exactly this value. |
| `status.outputs.registry_uri` | `string` | The registry endpoint clients push to and pull from, e.g. us-central1-docker.pkg.dev/my-project/app-images For Docker repositories, prefix an image name to get the full image reference ("{registry_uri}/api-server:v1.2.3"). |
| `status.outputs.location` | `string` | Location of the repository (region or multi-region), echoed from spec. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.kmsKeyName` | GcpKmsKeyHandle | `status.outputs.kms_key` |
| `spec.remoteRepositoryConfig.commonRepository.uri` | GcpArtifactRegistryRepo | `status.outputs.repository_path` |
| `spec.virtualRepositoryConfig.upstreamPolicies[].repository` | GcpArtifactRegistryRepo | `status.outputs.repository_path` |
| `spec.iamMembers[].member` | GcpServiceAccount | `status.outputs.member` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpArtifactRegistryRepo | `spec.remoteRepositoryConfig.commonRepository.uri` | `status.outputs.repository_path` |
| GcpArtifactRegistryRepo | `spec.virtualRepositoryConfig.upstreamPolicies[].repository` | `status.outputs.repository_path` |
| GcpCloudFunction | `spec.buildConfig.dockerRepository` | `status.outputs.repository_path` |

## See Also

- [Overview](../README.md)
