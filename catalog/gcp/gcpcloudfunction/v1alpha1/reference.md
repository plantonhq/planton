# GcpCloudFunction

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpCloudFunctionSpec defines a Cloud Functions (Gen 2) function
(`google_cloudfunctions2_function`) — source-based serverless compute
built on Cloud Run and Eventarc. You ship a source archive; Cloud Build
containerizes it with buildpacks and Cloud Run serves it, so every Gen 2
function is backed by a real Cloud Run service (the
`cloud_run_service_id` output).

The spec mirrors the API's two-config split: build_config owns HOW the
source becomes a container (runtime, entry point, source location, build
identity), service_config owns HOW it runs (resources, environment,
secrets, networking, scaling, invocation policy). The trigger decides
what invokes it — HTTPS requests, or a CloudEvent delivered by Eventarc.

Private VPC resources (Cloud SQL private IP, Memorystore) are reached
through a Serverless VPC Access connector (GcpServerlessVpcConnector),
attached by reference via service_config.vpc_connector.

## Example

```yaml
# Development manifest for GcpCloudFunction — exercises the full service
# surface (secret env + volumes, VPC connector egress, scaling, labels,
# build identity) for offline plan verification.
#
# Usage: planton tofu plan --manifest catalog/gcp/gcpcloudfunction/e2e/manifest.yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudFunction
metadata:
  name: hack-cloud-function
  id: cldfunc-hack-001
  org: planton-dev
  env: dev
spec:
  projectId:
    value: hack-project
  region: us-central1
  functionName: hack-api
  description: Development function exercising the deep service surface
  labels:
    team: platform
  buildConfig:
    runtime: python312
    entryPoint: handle_request
    source:
      storageSource:
        bucket:
          value: hack-source-bucket
        object: functions/hack-api-v1.zip
    buildEnvironmentVariables:
      GOOGLE_VENDOR_PIP_DEPENDENCIES: "1"
    serviceAccount:
      value: projects/hack-project/serviceAccounts/build-sa@hack-project.iam.gserviceaccount.com
    updatePolicy: ON_DEPLOY
  serviceConfig:
    serviceAccountEmail:
      value: fn-runtime@hack-project.iam.gserviceaccount.com
    availableMemory: 512M
    availableCpu: "1"
    timeoutSeconds: 120
    maxInstanceRequestConcurrency: 20
    environmentVariables:
      LOG_LEVEL: info
    secretEnvironmentVariables:
      - key: DATABASE_PASSWORD
        secret: hack-db-password
        version: latest
    secretVolumes:
      - mountPath: /etc/secrets
        secret: hack-tls-cert
        versions:
          - version: latest
            path: cert.pem
    vpcConnector:
      value: projects/hack-project/locations/us-central1/connectors/hack-egress
    vpcConnectorEgressSettings: ALL_TRAFFIC
    ingressSettings: ALLOW_INTERNAL_AND_GCLB
    scaling:
      minInstanceCount: 1
      maxInstanceCount: 50
    allowUnauthenticated: false
  trigger:
    triggerType: HTTP
  # Destroy really destroys in E2E: the live lanes prove the full lifecycle.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` | yes |  |  |
| `spec.functionName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.buildConfig` | `GcpCloudFunctionBuildConfig` | yes |  |  |
| `spec.buildConfig.runtime` | `string` | yes |  |  |
| `spec.buildConfig.entryPoint` | `string` | yes |  |  |
| `spec.buildConfig.source` | `GcpCloudFunctionSource` | yes |  |  |
| `spec.buildConfig.source.storageSource` | `GcpCloudFunctionStorageSource` |  |  |  |
| `spec.buildConfig.source.storageSource.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.buildConfig.source.storageSource.object` | `string` | yes |  |  |
| `spec.buildConfig.source.storageSource.generation` | `int64` |  |  |  |
| `spec.buildConfig.source.repoSource` | `GcpCloudFunctionRepoSource` |  |  |  |
| `spec.buildConfig.source.repoSource.repoName` | `string` | yes |  |  |
| `spec.buildConfig.source.repoSource.branchName` | `string` |  |  |  |
| `spec.buildConfig.source.repoSource.tagName` | `string` |  |  |  |
| `spec.buildConfig.source.repoSource.commitSha` | `string` |  |  |  |
| `spec.buildConfig.source.repoSource.dir` | `string` |  |  |  |
| `spec.buildConfig.source.repoSource.invertRegex` | `bool` |  |  |  |
| `spec.buildConfig.source.repoSource.projectId` | `string` |  |  |  |
| `spec.buildConfig.buildEnvironmentVariables` | `map<string, string>` |  |  |  |
| `spec.buildConfig.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.name`) |
| `spec.buildConfig.workerPool` | `string \| valueFrom` |  |  | GcpCloudBuildWorkerPool (`status.outputs.name`) |
| `spec.buildConfig.dockerRepository` | `string \| valueFrom` |  |  | GcpArtifactRegistryRepo (`status.outputs.repository_path`) |
| `spec.buildConfig.updatePolicy` | `enum` |  | `AUTOMATIC` |  |
| `spec.serviceConfig` | `GcpCloudFunctionServiceConfig` |  |  |  |
| `spec.serviceConfig.serviceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.serviceConfig.availableMemory` | `string` |  | `256M` |  |
| `spec.serviceConfig.availableCpu` | `string` |  |  |  |
| `spec.serviceConfig.timeoutSeconds` | `int32` |  | `60` |  |
| `spec.serviceConfig.maxInstanceRequestConcurrency` | `int32` |  | `1` |  |
| `spec.serviceConfig.environmentVariables` | `map<string, string>` (no secrets: use `secretEnvironmentVariables`) |  |  |  |
| `spec.serviceConfig.secretEnvironmentVariables` | `[]GcpCloudFunctionSecretEnvVar` |  |  |  |
| `spec.serviceConfig.secretEnvironmentVariables[].key` | `string` | yes |  |  |
| `spec.serviceConfig.secretEnvironmentVariables[].secret` | `string` |  |  |  |
| `spec.serviceConfig.secretEnvironmentVariables[].version` | `string` |  | `latest` |  |
| `spec.serviceConfig.secretEnvironmentVariables[].projectId` | `string` |  |  |  |
| `spec.serviceConfig.secretEnvironmentVariables[].value` | `string` (sensitive) |  |  |  |
| `spec.serviceConfig.secretVolumes` | `[]GcpCloudFunctionSecretVolume` |  |  |  |
| `spec.serviceConfig.secretVolumes[].mountPath` | `string` | yes |  |  |
| `spec.serviceConfig.secretVolumes[].secret` | `string` | yes |  |  |
| `spec.serviceConfig.secretVolumes[].projectId` | `string` |  |  |  |
| `spec.serviceConfig.secretVolumes[].versions` | `[]GcpCloudFunctionSecretVolumeVersion` |  |  |  |
| `spec.serviceConfig.secretVolumes[].versions[].version` | `string` | yes |  |  |
| `spec.serviceConfig.secretVolumes[].versions[].path` | `string` | yes |  |  |
| `spec.serviceConfig.vpcConnector` | `string \| valueFrom` |  |  | GcpServerlessVpcConnector (`status.outputs.self_link`) |
| `spec.serviceConfig.vpcConnectorEgressSettings` | `enum` |  | `PRIVATE_RANGES_ONLY` |  |
| `spec.serviceConfig.ingressSettings` | `enum` |  | `ALLOW_ALL` |  |
| `spec.serviceConfig.scaling` | `GcpCloudFunctionScalingConfig` |  |  |  |
| `spec.serviceConfig.scaling.minInstanceCount` | `int32` |  | `0` |  |
| `spec.serviceConfig.scaling.maxInstanceCount` | `int32` |  | `100` |  |
| `spec.serviceConfig.allTrafficOnLatestRevision` | `bool` |  | `true` |  |
| `spec.serviceConfig.binaryAuthorizationPolicy` | `string` |  |  |  |
| `spec.serviceConfig.allowUnauthenticated` | `bool` |  | `false` |  |
| `spec.serviceConfig.directVpcNetworkInterface` | `GcpCloudFunctionDirectVpcNetworkInterface` |  |  |  |
| `spec.serviceConfig.directVpcNetworkInterface.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.serviceConfig.directVpcNetworkInterface.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_name`) |
| `spec.serviceConfig.directVpcNetworkInterface.tags` | `[]string` |  |  |  |
| `spec.serviceConfig.directVpcEgress` | `enum` |  | `PRIVATE_RANGES_ONLY` |  |
| `spec.trigger` | `GcpCloudFunctionTrigger` |  |  |  |
| `spec.trigger.triggerType` | `enum` |  | `HTTP` |  |
| `spec.trigger.eventTrigger` | `GcpCloudFunctionEventTrigger` |  |  |  |
| `spec.trigger.eventTrigger.eventType` | `string` | yes |  |  |
| `spec.trigger.eventTrigger.pubsubTopic` | `string \| valueFrom` |  |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.trigger.eventTrigger.eventFilters` | `[]GcpCloudFunctionEventFilter` |  |  |  |
| `spec.trigger.eventTrigger.eventFilters[].attribute` | `string` | yes |  |  |
| `spec.trigger.eventTrigger.eventFilters[].value` | `string` | yes |  |  |
| `spec.trigger.eventTrigger.eventFilters[].operator` | `string` |  |  |  |
| `spec.trigger.eventTrigger.triggerRegion` | `string` |  |  |  |
| `spec.trigger.eventTrigger.retryPolicy` | `enum` |  | `RETRY_POLICY_DO_NOT_RETRY` |  |
| `spec.trigger.eventTrigger.serviceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the function is created in. Accepts a literal project
ID or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string` · required

Region the function is deployed in, e.g. "us-central1". Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.functionName

`string`

Name of the function in GCP. Immutable. If not specified, defaults to
metadata.name. Must be 1-63 characters: lowercase letters, digits, and
hyphens; starting with a letter.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.description

`string`

Human-readable description of what the function does.

### spec.labels

`map<string, string>`

Labels applied to the function object. User labels are merged beneath
Planton's attribution labels and shared with Google's billing system.

### spec.kmsKeyName

`string | valueFrom`

Cloud KMS key encrypting the function's resources (CMEK) — the
container image and source artifacts. The Cloud Functions and Artifact
Registry service agents must hold cryptoKeyEncrypterDecrypter on it,
and CMEK deployments require a customer-managed docker_repository.
Accepts a full crypto-key path or a reference to a GcpKmsKey resource.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.buildConfig

`GcpCloudFunctionBuildConfig` · required

How the source becomes a runnable container: runtime, entry point,
source location, and build identity. Required.

- rule: {"required":true}

### spec.buildConfig.runtime

`string` · required

Runtime the function executes in, e.g. "python312", "nodejs22",
"go123", "java21". Any current Gen 2 runtime GCP publishes is valid —
run `gcloud functions runtimes list` for the live set; deprecated
runtimes are rejected by the API at deploy time.

- rule: {"required":true,"string":{"maxLen":"32","pattern":"^[a-z][a-z0-9]*$"}}

### spec.buildConfig.entryPoint

`string` · required

Name of the function in source code that will be executed (the entry
point). For example: "hello_http" in Python, "helloHttp" in Node.js.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"128"}}

### spec.buildConfig.source

`GcpCloudFunctionSource` · required

Where the source code lives. Required.

- rule: {"required":true}
- rule: choose exactly one source: storageSource (a GCS zip archive) or repoSource (a Cloud Source Repositories revision)

### spec.buildConfig.source.storageSource

`GcpCloudFunctionStorageSource`

Source archive in Google Cloud Storage — a .zip of the function code
and dependency manifest. The standard path for CI/CD-shipped source.

### spec.buildConfig.source.storageSource.bucket

`string | valueFrom` · required

GCS bucket holding the source archive. The build service account needs
read access. Accepts a literal bucket name or a reference to a
GcpGcsBucket resource.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.buildConfig.source.storageSource.object

`string` · required

Object name (path) of the source archive in the bucket, e.g.
"functions/my-function-v1.2.3.zip". Version the object name per
release — a changed object name is what makes the deploy roll.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.buildConfig.source.storageSource.generation

`int64` · optional (explicit presence)

Generation number of the object to pin an exact object version even
if the path is overwritten. If unset, the current generation is used.

### spec.buildConfig.source.repoSource

`GcpCloudFunctionRepoSource`

Source in Cloud Source Repositories. Note GCP deprecated CSR for new
customers in June 2024 — existing repositories keep working, but new
integrations should ship archives to GCS instead.

- rule: pin the source to exactly one revision: branchName, tagName, or commitSha

### spec.buildConfig.source.repoSource.repoName

`string` · required

Name of the Cloud Source Repository, e.g. "my-repo".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.buildConfig.source.repoSource.branchName

`string`

Branch to build from (the branch's current HEAD at deploy time).

### spec.buildConfig.source.repoSource.tagName

`string`

Tag to build from.

### spec.buildConfig.source.repoSource.commitSha

`string`

Exact commit to build from — the only fully reproducible pin.

### spec.buildConfig.source.repoSource.dir

`string`

Directory within the repository containing the function source. If
unset, the repository root is used.

### spec.buildConfig.source.repoSource.invertRegex

`bool`

Invert the revision match: build from revisions that do NOT match
the configured branch/tag regex.

### spec.buildConfig.source.repoSource.projectId

`string`

Project that owns the repository, when it lives outside the
function's project. Immutable.

### spec.buildConfig.buildEnvironmentVariables

`map<string, string>`

Environment variables available at build time (e.g. buildpack knobs
like GOOGLE_ENTRYPOINT). Not injected into the runtime — use
service_config.environment_variables for that. Values are stored in
plain text on the function and must never be secrets: Google's build
API has no Secret Manager path for build-time variables. A build that
needs a private package index reaches it through the build service
account's own access (an Artifact Registry repository it can read) or
a private worker pool (worker_pool), never a token here.

### spec.buildConfig.serviceAccount

`string | valueFrom`

Service account Cloud Build runs the build as — the identity that
reads the source and pushes the image. FULLY-QUALIFIED resource name
(projects/{project}/serviceAccounts/{email}), not a bare email.
Accepts a literal or a reference to a GcpServiceAccount resource. If
omitted, GCP uses its default build identity.

- references: GcpServiceAccount (`status.outputs.name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.buildConfig.workerPool

`string | valueFrom`

The Cloud Build private worker pool the build runs in, as
projects/{project}/locations/{location}/workerPools/{pool}: a
GcpCloudBuildWorkerPool reference (its name output), or the literal
name. Use one when the build must reach a private network (a private
package index, an internal artifact store). Empty runs the build on
Google's default pool.

- references: GcpCloudBuildWorkerPool (`status.outputs.name`)
- rule: worker_pool must be a full worker pool name: projects/{project}/locations/{location}/workerPools/{pool}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudBuildWorkerPool, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.buildConfig.dockerRepository

`string | valueFrom`

User-managed Artifact Registry repository the built container is
stored in, optionally CMEK-protected (required when kms_key_name is
set). FULLY-QUALIFIED path
(projects/{project}/locations/{location}/repositories/{name}).
Accepts a literal path or a reference to a GcpArtifactRegistryRepo
resource (its repository_path output is exactly this value). If
omitted, GCP manages a default repository.

- references: GcpArtifactRegistryRepo (`status.outputs.repository_path`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpArtifactRegistryRepo, name: <that resource's name>, fieldPath: status.outputs.repository_path}} -- a bare string does not parse

### spec.buildConfig.updatePolicy

`enum`

How the runtime base image is patched. AUTOMATIC (the API default when
unset) applies security updates continuously; ON_DEPLOY pins the
runtime version at deploy time so instances never change under you
between deploys.

- default: `AUTOMATIC`

Allowed values (use exactly as shown):

- `AUTOMATIC` -- Runtime security updates apply automatically (the API default).
- `ON_DEPLOY` -- The runtime version is pinned at deploy time; updates arrive only on the next deploy.

### spec.serviceConfig

`GcpCloudFunctionServiceConfig`

How the function runs: compute resources, environment, secrets,
networking, scaling, and invocation policy.

- rule: use direct VPC egress (direct_vpc_network_interface) or a Serverless VPC Access connector (vpc_connector), not both

### spec.serviceConfig.serviceAccountEmail

`string | valueFrom`

Email of the IAM service account the function runs as — the identity
whose permissions the code exercises when calling other GCP APIs.
Accepts a literal email or a reference to a GcpServiceAccount
resource. If omitted, the project's Compute Engine default service
account is used — fine for experiments, too broad for production.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.serviceConfig.availableMemory

`string`

Memory available to each instance, as a quantity string: "256M",
"512M", "1Gi", "16Gi". CPU scales with memory unless available_cpu is
set explicitly. If unset, GCP defaults to 256M.

- default: `256M`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?(k|M|G|Mi|Gi)?$"}}

### spec.serviceConfig.availableCpu

`string`

CPUs available to each instance, e.g. "1", "2", "0.5". If unset, GCP
derives CPU from memory. Concurrency above 1 requires at least 1 CPU.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[0-9]+(\\.[0-9]+)?$"}}

### spec.serviceConfig.timeoutSeconds

`int32`

Per-request timeout in seconds. HTTP functions support up to 3600
(60 minutes); event-driven functions are capped at 540 by Eventarc's
delivery timeout. If unset, GCP defaults to 60.

- default: `60`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":3600,"gte":1}}

### spec.serviceConfig.maxInstanceRequestConcurrency

`int32`

Concurrent requests each instance handles (1-1000). GCP defaults to 1
— every request gets its own instance, safe for any runtime. Raising
it cuts instance count and cold starts for I/O-bound code, but needs
at least 1 CPU and thread-safe code.

- default: `1`
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":1000,"gte":1}}

### spec.serviceConfig.environmentVariables

`map<string, string>` · no secrets

Environment variables injected into the runtime as plain-text
KEY=VALUE pairs, written into the function where anyone who can view
it reads them. Configuration only — a credential goes in
secret_environment_variables, as a value this component stores in
Secret Manager or a secret you already own.

- secrets: this value is stored where anyone who can view the resource reads it, so a secret reference (`$secret/...`) here is refused -- put a secret in `secretEnvironmentVariables`, which keeps it in a secret store the workload reads by reference
### spec.serviceConfig.secretEnvironmentVariables

`[]GcpCloudFunctionSecretEnvVar`

Environment variables whose value comes from Secret Manager, resolved
at instance start. Each entry either names a secret you already own
(secret, version, project_id) or carries a value this component
stores for you (value). The function holds only the reference, never
the material. A secret you own needs roles/secretmanager.secretAccessor
granted to the runtime service account; a stored value gets that grant
from the component.

- rule: a secret environment variable takes exactly one of secret (a Secret Manager secret you own) or value (a secret value this component stores)
- rule: version and project_id address a secret you own; they apply only with secret, never with value

### spec.serviceConfig.secretEnvironmentVariables[].key

`string` · required

Environment variable name, e.g. "DATABASE_PASSWORD".

- rule: {"required":true,"string":{"pattern":"^[A-Za-z_][A-Za-z0-9_]*$"}}

### spec.serviceConfig.secretEnvironmentVariables[].secret

`string`

A Secret Manager secret you already own: its short name in the
function's project ("my-secret"). Cross-project secrets set
project_id. The runtime service account needs
roles/secretmanager.secretAccessor on it.

### spec.serviceConfig.secretEnvironmentVariables[].version

`string`

Version of the secret you own to resolve: a version number or
"latest" — the common choice, at the cost of new instances silently
picking up rotations. Only with secret.

- default: `latest`

### spec.serviceConfig.secretEnvironmentVariables[].projectId

`string`

Project the secret you own lives in, when it is not the function's
project. Only with secret.

### spec.serviceConfig.secretEnvironmentVariables[].value

`string` · sensitive

A secret value this component keeps in Secret Manager for you. It
creates one secret for this variable, replicated only in the
function's region, stores the value as a version, grants the
function's runtime identity (service_account_email, or the project's
Compute Engine default service account when unset) secretAccessor on
that secret alone, and points the variable at that exact version —
the function carries a reference, never the value. A changed value
adds a version and redeploys the function, so rotation is a deploy;
destroying the function removes the secret. The secret's id is
function_<region>_<function name>_<key>.

### spec.serviceConfig.secretVolumes

`[]GcpCloudFunctionSecretVolume`

Secret Manager secret versions projected as files under a mount path
— for consumers that read credentials from disk (certificates, config
files). Same accessor-role requirement as secret env vars.

### spec.serviceConfig.secretVolumes[].mountPath

`string` · required

Absolute path the volume is mounted at, e.g. "/etc/secrets". Each
configured version appears as a file under it.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.serviceConfig.secretVolumes[].secret

`string` · required

The secret: a short name for a secret in the function's project.
Cross-project secrets set project_id.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.serviceConfig.secretVolumes[].projectId

`string`

Project the secret lives in, when it is not the function's project.

### spec.serviceConfig.secretVolumes[].versions

`[]GcpCloudFunctionSecretVolumeVersion`

Which versions land at which relative paths. If empty, the "latest"
version is projected at a file named after the secret.

### spec.serviceConfig.secretVolumes[].versions[].version

`string` · required

Secret version to project: a version number or "latest".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.serviceConfig.secretVolumes[].versions[].path

`string` · required

Relative path of the file under the volume's mount path.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.serviceConfig.vpcConnector

`string | valueFrom`

Serverless VPC Access connector routing the function's egress into a
VPC — how the function reaches private IPs (Cloud SQL private IP,
Memorystore, internal load balancers). Accepts the connector's full
resource name (projects/*/locations/*/connectors/*) or a reference to
a GcpServerlessVpcConnector resource.

- references: GcpServerlessVpcConnector (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServerlessVpcConnector, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.serviceConfig.vpcConnectorEgressSettings

`enum`

Which egress traffic uses the connector: only RFC1918/private
destinations (the default; public egress keeps the normal path), or
everything (enables static egress IPs via Cloud NAT).

- default: `PRIVATE_RANGES_ONLY`

Allowed values (use exactly as shown):

- `PRIVATE_RANGES_ONLY` -- Only RFC1918/private destinations route through the connector; public egress keeps the normal path.
- `ALL_TRAFFIC` -- All outbound traffic routes through the connector — enables static egress IPs via Cloud NAT.

### spec.serviceConfig.ingressSettings

`enum`

Who can reach the function's endpoint at the network level. Use
ALLOW_INTERNAL_ONLY for private functions, ALLOW_INTERNAL_AND_GCLB
when fronting with an external Application Load Balancer.

- default: `ALLOW_ALL`

Allowed values (use exactly as shown):

- `ALLOW_ALL` -- Reachable from the public internet (invocation still subject to IAM).
- `ALLOW_INTERNAL_ONLY` -- Reachable only from within the project's VPC networks and internal GCP services.
- `ALLOW_INTERNAL_AND_GCLB` -- Reachable from internal sources and through Cloud Load Balancing — the setting for functions fronted by an external Application Load Balancer.

### spec.serviceConfig.scaling

`GcpCloudFunctionScalingConfig`

Instance scaling bounds.

- rule: minInstanceCount cannot exceed maxInstanceCount

### spec.serviceConfig.scaling.minInstanceCount

`int32`

Minimum instances kept warm. Above 0 eliminates cold starts at idle
compute cost — for latency-sensitive production endpoints.

- default: `0`
- rule: {"int32":{"lte":100,"gte":0}}

### spec.serviceConfig.scaling.maxInstanceCount

`int32`

Maximum instances the function scales to — the cost and downstream-
pressure ceiling (a runaway event storm stops here).

- default: `100`
- rule: {"int32":{"lte":3000,"gte":1}}

### spec.serviceConfig.allTrafficOnLatestRevision

`bool` · optional (explicit presence)

Whether 100% of traffic goes to the latest revision as soon as it is
ready (the API default, true). Set false to hold traffic on the
previous revision — the lever for manual canary/rollback via the
underlying Cloud Run service.

- default: `true`

### spec.serviceConfig.binaryAuthorizationPolicy

`string`

Binary Authorization policy checked before instances start, e.g.
"default" or "projects/{project}/platforms/gae/policies/{policy}".

### spec.serviceConfig.allowUnauthenticated

`bool`

Makes the function publicly invokable by unauthenticated callers by
granting run.invoker to allUsers on the underlying Cloud Run service.
Leave false for private functions and grant invoker to specific
identities instead.

- default: `false`

### spec.serviceConfig.directVpcNetworkInterface

`GcpCloudFunctionDirectVpcNetworkInterface`

Direct VPC egress: attach the function straight to a VPC network or
subnet — no Serverless VPC Access connector to size, pay for, or
saturate. The modern alternative to vpc_connector (mutually
exclusive with it); instances get IPs from the subnet, so size its
range for the scaling ceiling.

- rule: set at least one of network or subnetwork on the direct VPC interface

### spec.serviceConfig.directVpcNetworkInterface.network

`string | valueFrom`

The VPC network to attach to. Accepts a literal network name or a
reference to a GcpVpcNetwork resource.

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.serviceConfig.directVpcNetworkInterface.subnetwork

`string | valueFrom`

The subnetwork instances draw their IPs from. Accepts a literal
subnetwork name or a reference to a GcpSubnetwork resource. The
subnet's free range caps how far the function can scale.

- references: GcpSubnetwork (`status.outputs.subnetwork_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_name}} -- a bare string does not parse

### spec.serviceConfig.directVpcNetworkInterface.tags

`[]string`

Network tags applied to the function's instances — how VPC firewall
rules select their egress.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","repeated":{"unique":true,"items":{"string":{"pattern":"^[a-z]([-a-z0-9]*[a-z0-9])?$"}}}}

### spec.serviceConfig.directVpcEgress

`enum`

Which egress traffic takes the direct-VPC path: only RFC1918/private
destinations (the API default; public egress keeps the normal path),
or everything (enables static egress IPs via Cloud NAT). Only
meaningful with direct_vpc_network_interface — the connector path is
steered by vpc_connector_egress_settings instead.

- default: `PRIVATE_RANGES_ONLY`

Allowed values (use exactly as shown):

- `PRIVATE_RANGES_ONLY` -- Only RFC1918/private destinations route through the connector; public egress keeps the normal path.
- `ALL_TRAFFIC` -- All outbound traffic routes through the connector — enables static egress IPs via Cloud NAT.

### spec.trigger

`GcpCloudFunctionTrigger`

What invokes the function. If not specified, defaults to HTTP.

- rule: an EVENT_TRIGGER function needs the eventTrigger block (eventType and its filters)

### spec.trigger.triggerType

`enum`

Type of trigger. Defaults to HTTP if not specified.

- default: `HTTP`

Allowed values (use exactly as shown):

- `HTTP` -- Invoked via HTTPS requests at the function's URL.
- `EVENT_TRIGGER` -- Invoked when a CloudEvent is delivered by Eventarc.

### spec.trigger.eventTrigger

`GcpCloudFunctionEventTrigger`

Event trigger configuration. Required when trigger_type is
EVENT_TRIGGER.

### spec.trigger.eventTrigger.eventType

`string` · required

Event type that triggers the function, in CloudEvents format:
- "google.cloud.pubsub.topic.v1.messagePublished" (Pub/Sub)
- "google.cloud.storage.object.v1.finalized" (object created)
- "google.cloud.storage.object.v1.deleted" (object deleted)
- "google.cloud.firestore.document.v1.written" (document write)

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.trigger.eventTrigger.pubsubTopic

`string | valueFrom`

Pub/Sub topic for messagePublished triggers. Accepts the full
resource name (projects/{project}/topics/{name}) or a reference to a
GcpPubSubTopic resource.

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.trigger.eventTrigger.eventFilters

`[]GcpCloudFunctionEventFilter`

Event filters narrowing which events invoke the function. For Storage
triggers, filter by bucket: attribute="bucket" value="my-bucket". For
Firestore, filter by document path pattern.

### spec.trigger.eventTrigger.eventFilters[].attribute

`string` · required

Attribute to filter on (e.g. "bucket" for Storage events).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.trigger.eventTrigger.eventFilters[].value

`string` · required

Value to match.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.trigger.eventTrigger.eventFilters[].operator

`string` · optional (explicit presence)

Matching operator. Unset means exact match; "match-path-pattern"
(the only other value GCP accepts) enables path-pattern wildcards for
Firestore/audit-log filters.

### spec.trigger.eventTrigger.triggerRegion

`string`

Region the trigger listens in. Storage/audit-log sources fire in the
bucket's region (multi-region sources use "us"/"eu"); if unset, GCP
uses the function's region.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z]+(-[a-z]+[0-9]+)?$"}}

### spec.trigger.eventTrigger.retryPolicy

`enum`

Retry policy for failed deliveries. RETRY_POLICY_RETRY redelivers
with exponential backoff (at-least-once — handlers must be
idempotent); RETRY_POLICY_DO_NOT_RETRY delivers at most once.

- default: `RETRY_POLICY_DO_NOT_RETRY`

Allowed values (use exactly as shown):

- `RETRY_POLICY_DO_NOT_RETRY` -- Deliver at most once; failed invocations are not retried.
- `RETRY_POLICY_RETRY` -- Redeliver with exponential backoff (at-least-once) — handlers must be idempotent.

### spec.trigger.eventTrigger.serviceAccountEmail

`string | valueFrom`

Email of the service account Eventarc uses to invoke the function —
it needs run.invoker on the underlying service. Accepts a literal
email or a reference to a GcpServiceAccount resource. If omitted, the
default compute service account is used.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What destroying this resource does to the function:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the function (and the Cloud Run service serving it) is
               deleted; event triggers stop firing
  "PREVENT" -- destroy FAILS; protects a function other systems invoke
  "ABANDON" -- the function is removed from management but keeps
               serving and consuming events in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpCloudFunction, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.function_id` | `string` | The fully qualified resource name of the deployed function. Format: projects/{project}/locations/{region}/functions/{function-name} Example: "projects/my-project/locations/us-central1/functions/my-function" |
| `status.outputs.function_url` | `string` | The HTTPS URL of the function. Only populated for HTTP-triggered functions. For event-driven functions, this field is empty. Example: "https://us-central1-my-project.cloudfunctions.net/my-function" |
| `status.outputs.service_account_email` | `string` | The email of the service account that the function runs as. This is the runtime identity of the function. Format: "{service-account-id}@{project}.iam.gserviceaccount.com" |
| `status.outputs.state` | `string` | The current state of the function. Possible values: "ACTIVE", "OFFLINE", "DEPLOY_IN_PROGRESS", "DELETE_IN_PROGRESS", "UNKNOWN" |
| `status.outputs.cloud_run_service_id` | `string` | The Cloud Run service name (Gen 2 functions are deployed as Cloud Run services). This can be used to access the function via Cloud Run APIs if needed. Format: "projects/{project}/locations/{region}/services/{service-name}" |
| `status.outputs.eventarc_trigger_id` | `string` | The Eventarc trigger ID for event-driven functions. Only populated for functions with EVENT_TRIGGER type. Format: "projects/{project}/locations/{region}/triggers/{trigger-name}" |
| `status.outputs.name` | `string` | The bare function name (the last segment of function_id) — the composition key serverless network endpoint groups and gcloud commands reference the function by. |
| `status.outputs.uri` | `string` | The URI of the underlying Cloud Run service serving the function — the *.run.app endpoint. Populated for every Gen 2 function. |
| `status.outputs.environment` | `string` | The environment the function runs in (e.g. "GEN_2"). |
| `status.outputs.update_time` | `string` | Timestamp of the last update to the function, in RFC 3339 format. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.buildConfig.source.storageSource.bucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.buildConfig.serviceAccount` | GcpServiceAccount | `status.outputs.name` |
| `spec.buildConfig.workerPool` | GcpCloudBuildWorkerPool | `status.outputs.name` |
| `spec.buildConfig.dockerRepository` | GcpArtifactRegistryRepo | `status.outputs.repository_path` |
| `spec.serviceConfig.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.serviceConfig.vpcConnector` | GcpServerlessVpcConnector | `status.outputs.self_link` |
| `spec.serviceConfig.directVpcNetworkInterface.network` | GcpVpcNetwork | `status.outputs.network_name` |
| `spec.serviceConfig.directVpcNetworkInterface.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_name` |
| `spec.trigger.eventTrigger.pubsubTopic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.trigger.eventTrigger.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpIdentityPlatformConfig | `spec.blockingFunctions.triggers[].functionUri` | `status.outputs.function_url` |
| GcpMonitoringUptimeCheck | `spec.syntheticMonitor.cloudFunction` | `status.outputs.function_id` |
| GcpRegionNetworkEndpointGroup | `spec.cloudFunction.function` | `status.outputs.name` |

## See Also

- [Overview](../README.md)
