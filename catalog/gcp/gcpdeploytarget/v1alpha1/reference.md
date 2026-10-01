# GcpDeployTarget

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpDeployTargetSpec declares a Cloud Deploy target
(`google_clouddeploy_target`): one place a delivery pipeline's stage
deploys a release to, together with how Cloud Deploy runs the render
and deploy work for it.

A target is exactly one of five kinds:

  - gke            -- a GKE cluster, by its full cluster name;
  - anthos_cluster -- a cluster registered to a fleet (GKE Enterprise),
                      by its fleet membership;
  - run            -- a Cloud Run region in a project;
  - multi_target   -- a group of other targets in this project and
                      region, deployed together in parallel;
  - custom_target  -- a GcpDeployCustomTargetType, for anything Cloud
                      Deploy does not deploy natively.

A delivery pipeline names its targets by ID in its stages
(GcpDeliveryPipeline.serial_pipeline.stages[].target_id, a reference to
this kind's target_id output). The pipeline and its targets must live in
the same project and region.

Every render, deploy, verify, and hook job runs as a Cloud Build build.
With no execution_configs, Cloud Deploy uses Cloud Build's default pool,
the project's default compute service account
({PROJECT_NUMBER}-compute@developer.gserviceaccount.com), and a default
bucket in the target's region. Declare execution_configs to choose the
service account, the artifact bucket, a private worker pool, or a
timeout per kind of job.

Important behavioral notes:

  - location and target_id are create-time decisions: changing one
    replaces the target. The other fields update in place.
  - The execution service account needs roles/clouddeploy.jobRunner on
    the project and whatever the deployment itself touches (for Cloud
    Run, roles/run.developer and roles/iam.serviceAccountUser on the
    service's runtime service account; for GKE, roles/container.developer).
  - Destroy deletes the target. Releases and rollouts that already
    reference it keep their history; a pipeline stage that still names
    it can no longer deploy.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeployTarget
metadata:
  name: prod
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  targetId: prod
  description: Production on Cloud Run
  requireApproval: true
  labels:
    env: prod
  annotations:
    owner: platform
  deployParameters:
    min-instances: "2"
  run:
    location: projects/my-gcp-project/locations/us-central1
  executionConfigs:
    - usages:
        - RENDER
        - DEPLOY
      serviceAccount:
        value: deployer@my-gcp-project.iam.gserviceaccount.com
      artifactStorage:
        value: gs://my-gcp-project-deploy-artifacts/prod
      executionTimeout: 1800s
    - usages:
        - VERIFY
      workerPool:
        value: projects/my-gcp-project/locations/us-central1/workerPools/private-builds
      verbose: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.targetId` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.requireApproval` | `bool` |  |  |  |
| `spec.deployParameters` | `map<string, string>` |  |  |  |
| `spec.gke` | `GcpDeployTargetGke` |  |  |  |
| `spec.gke.cluster` | `string \| valueFrom` |  |  | GcpGkeCluster (`status.outputs.cluster_id`) |
| `spec.gke.internalIp` | `bool` |  |  |  |
| `spec.gke.dnsEndpoint` | `bool` |  |  |  |
| `spec.gke.proxyUrl` | `string` |  |  |  |
| `spec.anthosCluster` | `GcpDeployTargetAnthosCluster` |  |  |  |
| `spec.anthosCluster.membership` | `string \| valueFrom` |  |  | GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`) |
| `spec.run` | `GcpDeployTargetRun` |  |  |  |
| `spec.run.location` | `string` | yes |  |  |
| `spec.multiTarget` | `GcpDeployTargetMultiTarget` |  |  |  |
| `spec.multiTarget.targetIds` | `[]string \| valueFrom` | yes |  | GcpDeployTarget (`status.outputs.target_id`) |
| `spec.customTarget` | `GcpDeployTargetCustomTarget` |  |  |  |
| `spec.customTarget.customTargetType` | `string \| valueFrom` | yes |  | GcpDeployCustomTargetType (`status.outputs.name`) |
| `spec.associatedEntities` | `[]GcpDeployTargetAssociatedEntity` |  |  |  |
| `spec.associatedEntities[].entityId` | `string` | yes |  |  |
| `spec.associatedEntities[].gkeClusters` | `[]GcpDeployTargetAssociatedGkeCluster` |  |  |  |
| `spec.associatedEntities[].gkeClusters[].cluster` | `string \| valueFrom` |  |  | GcpGkeCluster (`status.outputs.cluster_id`) |
| `spec.associatedEntities[].gkeClusters[].internalIp` | `bool` |  |  |  |
| `spec.associatedEntities[].gkeClusters[].proxyUrl` | `string` |  |  |  |
| `spec.associatedEntities[].anthosClusters` | `[]GcpDeployTargetAssociatedAnthosCluster` |  |  |  |
| `spec.associatedEntities[].anthosClusters[].membership` | `string \| valueFrom` |  |  | GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`) |
| `spec.executionConfigs` | `[]GcpDeployTargetExecutionConfig` |  |  |  |
| `spec.executionConfigs[].usages` | `[]string` | yes |  |  |
| `spec.executionConfigs[].workerPool` | `string \| valueFrom` |  |  | GcpCloudBuildWorkerPool (`status.outputs.name`) |
| `spec.executionConfigs[].serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.executionConfigs[].artifactStorage` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.url`) |
| `spec.executionConfigs[].executionTimeout` | `string` |  |  |  |
| `spec.executionConfigs[].verbose` | `bool` |  |  |  |
| `spec.executionConfigs[].defaultPool` | `GcpDeployTargetDefaultPool` |  |  |  |
| `spec.executionConfigs[].defaultPool.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.executionConfigs[].defaultPool.artifactStorage` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.url`) |
| `spec.executionConfigs[].privatePool` | `GcpDeployTargetPrivatePool` |  |  |  |
| `spec.executionConfigs[].privatePool.workerPool` | `string \| valueFrom` | yes |  | GcpCloudBuildWorkerPool (`status.outputs.name`) |
| `spec.executionConfigs[].privatePool.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.executionConfigs[].privatePool.artifactStorage` | `string \| valueFrom` |  |  | GcpGcsBucket (`status.outputs.url`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project the target lives in: a literal project ID or a GcpProject
reference. Empty means the provider's default project. The delivery
pipelines that deploy to it must be in the same project. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The region the target lives in, e.g. "us-central1" -- the region of
the delivery pipelines that deploy to it. It need not be the region
the workload runs in (a run target names its own region). Required.
Immutable.

- rule: {"required":true}

### spec.targetId

`string`

The target's ID, unique in the project and region: a lowercase letter,
then up to 62 lowercase letters, digits, or hyphens, not ending with a
hyphen (Google's rule). Pipeline stages name the target by this ID.
Defaults to metadata.name. Immutable.

- rule: target_id must be a lowercase letter followed by up to 62 lowercase letters, digits, or hyphens, not ending with a hyphen

### spec.description

`string`

A description shown in the console, up to 255 characters.

- rule: {"string":{"maxLen":"255"}}

### spec.labels

`map<string, string>`

Labels on the target. Cloud Deploy also reads them: a deploy policy
or an automation can select targets by label. Keys and values are
lowercase letters, digits, underscores, and dashes, at most 64 labels.
The platform attribution labels are added on top and win on key
conflicts.

### spec.annotations

`map<string, string>`

Annotations on the target (AIP-128 key/value metadata Cloud Deploy
never reads). Only the keys declared here are managed.

### spec.requireApproval

`bool`

Require an approval before any rollout to this target proceeds: a
principal with clouddeploy.rollouts.approve (roles/clouddeploy.approver)
approves or rejects each one. The usual gate in front of production.

### spec.deployParameters

`map<string, string>`

Deploy parameters for every rollout to this target: key/value pairs
substituted into the rendered manifests wherever a
"# from-param: ${key}" marker names the key. Pipeline stages and
releases can declare parameters too.

### spec.gke

`GcpDeployTargetGke`

Deploy to a GKE cluster. Exactly one target type.

- rule: dns_endpoint and internal_ip cannot both be true

### spec.gke.cluster

`string | valueFrom`

The cluster, by full name
(projects/{project}/locations/{location}/clusters/{name}): a
GcpGkeCluster reference (its cluster_id output) or the literal name.
The cluster may be in another project; the execution service account
then needs access there.

- references: GcpGkeCluster (`status.outputs.cluster_id`)
- rule: cluster must be a full cluster name: projects/{project}/locations/{location}/clusters/{name}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.gke.internalIp

`bool`

Reach the control plane on its private IP address. Only for clusters
with a private endpoint; the default address is already private for a
cluster whose endpoint is private-only. Cloud Build must then run on a
private pool with a route to that network (execution_configs).
Cannot be combined with dns_endpoint.

### spec.gke.dnsEndpoint

`bool`

Reach the control plane through its DNS-based endpoint. Cannot be
combined with internal_ip.

### spec.gke.proxyUrl

`string`

An HTTP proxy Cloud Deploy reaches the Kubernetes API server through
(the kubeconfig proxy-url), e.g. "http://10.0.0.5:3128".

### spec.anthosCluster

`GcpDeployTargetAnthosCluster`

Deploy to a cluster registered to a fleet, through its membership
(GKE Enterprise, including attached and on-premises clusters). Exactly
one target type.

### spec.anthosCluster.membership

`string | valueFrom`

The cluster's fleet membership, by full name
(projects/{project}/locations/{location}/memberships/{id}): a
GcpGkeFleetMembership reference (its name output), a GcpGkeCluster
reference (its fleet_membership output, for a cluster that joined
through fleet_project), or the literal name.

- references: GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`)
- rule: membership must be a full membership name: projects/{project}/locations/{location}/memberships/{id}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeFleetMembership, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.run

`GcpDeployTargetRun`

Deploy Cloud Run services or jobs into a region. Exactly one target
type.

### spec.run.location

`string` · required

Where the Cloud Run services or jobs are deployed, as
projects/{project}/locations/{region}, e.g.
"projects/my-app-prod/locations/us-central1". The project may differ
from the target's own (a pipeline in a CI project deploying into
per-environment projects). Required.

- rule: run.location must be projects/{project}/locations/{location}
- rule: {"required":true}

### spec.multiTarget

`GcpDeployTargetMultiTarget`

Deploy to several other targets at once, in parallel. Exactly one
target type.

### spec.multiTarget.targetIds

`[]string | valueFrom` · required

The child targets, by ID: GcpDeployTarget references (their target_id
output) or literal IDs. Each must be in this target's project and
region. A rollout to the multi-target deploys to every child in
parallel. At least one.

- references: GcpDeployTarget (`status.outputs.target_id`)
- rule: each target ID must be a bare target ID (a lowercase letter followed by lowercase letters, digits, or hyphens), not a full resource name
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDeployTarget, name: <that resource's name>, fieldPath: status.outputs.target_id}} -- a bare string does not parse

### spec.customTarget

`GcpDeployTargetCustomTarget`

Deploy with a custom target type: your own render and deploy
containers or Skaffold modules, for platforms Cloud Deploy does not
deploy natively. Exactly one target type.

### spec.customTarget.customTargetType

`string | valueFrom` · required

The custom target type, by full name
(projects/{project}/locations/{location}/customTargetTypes/{id}): a
GcpDeployCustomTargetType reference (its name output) or the literal
name. Required.

- references: GcpDeployCustomTargetType (`status.outputs.name`)
- rule: custom_target_type must be a full name: projects/{project}/locations/{location}/customTargetTypes/{id}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDeployCustomTargetType, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.associatedEntities

`[]GcpDeployTargetAssociatedEntity`

Clusters other than the deployment target that some features deploy
to, each under an entity ID. The Gateway API canary, for example, can
deploy its HTTPRoute to a separate config cluster named here. Keyed by
entity_id.

### spec.associatedEntities[].entityId

`string` · required

The entity's ID, the key features use to find these clusters: a
lowercase letter, then up to 62 lowercase letters, digits, or hyphens,
not ending with a hyphen (Google's rule). Required.

- rule: entity_id must be a lowercase letter followed by up to 62 lowercase letters, digits, or hyphens, not ending with a hyphen
- rule: {"required":true}

### spec.associatedEntities[].gkeClusters

`[]GcpDeployTargetAssociatedGkeCluster`

GKE clusters for this entity.

### spec.associatedEntities[].gkeClusters[].cluster

`string | valueFrom`

The cluster, by full name
(projects/{project}/locations/{location}/clusters/{name}): a
GcpGkeCluster reference (its cluster_id output) or the literal name.

- references: GcpGkeCluster (`status.outputs.cluster_id`)
- rule: cluster must be a full cluster name: projects/{project}/locations/{location}/clusters/{name}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.associatedEntities[].gkeClusters[].internalIp

`bool`

Reach the control plane on its private IP address. Only for clusters
with a private endpoint.

### spec.associatedEntities[].gkeClusters[].proxyUrl

`string`

An HTTP proxy Cloud Deploy reaches the Kubernetes API server through.

### spec.associatedEntities[].anthosClusters

`[]GcpDeployTargetAssociatedAnthosCluster`

Fleet-registered clusters for this entity.

### spec.associatedEntities[].anthosClusters[].membership

`string | valueFrom`

The cluster's fleet membership, by full name
(projects/{project}/locations/{location}/memberships/{id}): a
GcpGkeFleetMembership reference (its name output), a GcpGkeCluster
reference (its fleet_membership output), or the literal name.

- references: GcpGkeFleetMembership (`status.outputs.name`), GcpGkeCluster (`status.outputs.fleet_membership`)
- rule: membership must be a full membership name: projects/{project}/locations/{location}/memberships/{id}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeFleetMembership, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.executionConfigs

`[]GcpDeployTargetExecutionConfig`

How Cloud Deploy runs this target's jobs, one configuration per group
of usages. Each usage (RENDER, DEPLOY, VERIFY, PREDEPLOY, POSTDEPLOY)
may appear in only one configuration, and when any are declared,
RENDER and DEPLOY must both be covered. With none, every job runs on
Cloud Build's default pool with the project's default compute service
account.

- rule: set at most one of default_pool or private_pool

### spec.executionConfigs[].usages

`[]string` · required

The jobs this configuration applies to, each at most once across all
of the target's configurations:
  "RENDER"     -- rendering manifests for a release
  "DEPLOY"     -- deploying, and the deployment hooks
  "VERIFY"     -- deployment verification
  "PREDEPLOY"  -- predeploy jobs
  "POSTDEPLOY" -- postdeploy jobs
Required.

- rule: usages must be RENDER, DEPLOY, VERIFY, PREDEPLOY, or POSTDEPLOY, each listed once
- rule: {"repeated":{"minItems":"1"}}

### spec.executionConfigs[].workerPool

`string | valueFrom`

The Cloud Build private pool the jobs run on: a GcpCloudBuildWorkerPool
reference (its name output) or a literal
projects/{project}/locations/{location}/workerPools/{id}. Empty means
Cloud Build's default pool. A private pool is how jobs reach a GKE
cluster's private endpoint. The top-level worker_pool,
service_account, and artifact_storage and the default_pool and
private_pool blocks express the same choice; use one form.

- references: GcpCloudBuildWorkerPool (`status.outputs.name`)
- rule: worker_pool must be a full name: projects/{project}/locations/{location}/workerPools/{id}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudBuildWorkerPool, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.executionConfigs[].serviceAccount

`string | valueFrom`

The service account the jobs run as, by email: a GcpServiceAccount
reference or a literal. Empty means the project's default compute
service account. It needs roles/clouddeploy.jobRunner, and the caller
creating releases needs roles/iam.serviceAccountUser on it.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.executionConfigs[].artifactStorage

`string | valueFrom`

Where the jobs store their outputs (rendered manifests, logs): a
bucket ("gs://my-bucket") or a path in one ("gs://my-bucket/deploy"),
a GcpGcsBucket reference (its gs:// url output) or a literal. Empty
means a default bucket Cloud Deploy creates in the target's region.
The service account needs write access to it.

- references: GcpGcsBucket (`status.outputs.url`)
- rule: artifact_storage must be a gs:// bucket or path
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.url}} -- a bare string does not parse

### spec.executionConfigs[].executionTimeout

`string`

How long one job may run, in seconds with an "s" suffix, between 10
minutes and 24 hours ("600s" to "86400s"). Empty means one hour.

- rule: execution_timeout must be a duration in seconds such as 3600s

### spec.executionConfigs[].verbose

`bool`

Turn on verbose logging in the jobs' builds, for debugging.

### spec.executionConfigs[].defaultPool

`GcpDeployTargetDefaultPool`

Use Cloud Build's default pool, with its own service account and
artifact storage (the block form of the top-level fields). Mutually
exclusive with private_pool.

### spec.executionConfigs[].defaultPool.serviceAccount

`string | valueFrom`

The service account the jobs run as, by email: a GcpServiceAccount
reference or a literal. Empty means the project's default compute
service account.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.executionConfigs[].defaultPool.artifactStorage

`string | valueFrom`

Where the jobs store their outputs: a gs:// bucket or path, a
GcpGcsBucket reference (its url output) or a literal. Empty means a
default bucket in the target's region.

- references: GcpGcsBucket (`status.outputs.url`)
- rule: artifact_storage must be a gs:// bucket or path
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.url}} -- a bare string does not parse

### spec.executionConfigs[].privatePool

`GcpDeployTargetPrivatePool`

Use a Cloud Build private pool, with its own service account and
artifact storage (the block form of the top-level fields). Mutually
exclusive with default_pool.

### spec.executionConfigs[].privatePool.workerPool

`string | valueFrom` · required

The private pool: a GcpCloudBuildWorkerPool reference (its name
output) or a literal
projects/{project}/locations/{location}/workerPools/{id}. Required.

- references: GcpCloudBuildWorkerPool (`status.outputs.name`)
- rule: worker_pool must be a full name: projects/{project}/locations/{location}/workerPools/{id}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudBuildWorkerPool, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.executionConfigs[].privatePool.serviceAccount

`string | valueFrom`

The service account the jobs run as, by email: a GcpServiceAccount
reference or a literal. Empty means the project's default compute
service account.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.executionConfigs[].privatePool.artifactStorage

`string | valueFrom`

Where the jobs store their outputs: a gs:// bucket or path, a
GcpGcsBucket reference (its url output) or a literal. Empty means a
default bucket in the target's region.

- references: GcpGcsBucket (`status.outputs.url`)
- rule: artifact_storage must be a gs:// bucket or path
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.url}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What destroy does:
  "" / "DELETE" -- the target is deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the target leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.exactly_one_target_type`: set exactly one of gke, anthos_cluster, run, multi_target, or custom_target
- `spec.unique_entity_ids`: entity_id must be unique within associated_entities
- `spec.usage_in_one_execution_config`: each usage may appear in only one execution configuration
- `spec.execution_configs_cover_render_and_deploy`: when execution_configs are set, they must include the RENDER and DEPLOY usages

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDeployTarget, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/targets/{target_id}. |
| `status.outputs.target_id` | `string` | The target's ID -- what GcpDeliveryPipeline stages and multi-targets reference. |
| `status.outputs.uid` | `string` | Google's unique identifier for the target. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.gke.cluster` | GcpGkeCluster | `status.outputs.cluster_id` |
| `spec.anthosCluster.membership` | GcpGkeFleetMembership | `status.outputs.name` |
| `spec.anthosCluster.membership` | GcpGkeCluster | `status.outputs.fleet_membership` |
| `spec.multiTarget.targetIds` | GcpDeployTarget | `status.outputs.target_id` |
| `spec.customTarget.customTargetType` | GcpDeployCustomTargetType | `status.outputs.name` |
| `spec.associatedEntities[].gkeClusters[].cluster` | GcpGkeCluster | `status.outputs.cluster_id` |
| `spec.associatedEntities[].anthosClusters[].membership` | GcpGkeFleetMembership | `status.outputs.name` |
| `spec.associatedEntities[].anthosClusters[].membership` | GcpGkeCluster | `status.outputs.fleet_membership` |
| `spec.executionConfigs[].workerPool` | GcpCloudBuildWorkerPool | `status.outputs.name` |
| `spec.executionConfigs[].serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.executionConfigs[].artifactStorage` | GcpGcsBucket | `status.outputs.url` |
| `spec.executionConfigs[].defaultPool.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.executionConfigs[].defaultPool.artifactStorage` | GcpGcsBucket | `status.outputs.url` |
| `spec.executionConfigs[].privatePool.workerPool` | GcpCloudBuildWorkerPool | `status.outputs.name` |
| `spec.executionConfigs[].privatePool.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.executionConfigs[].privatePool.artifactStorage` | GcpGcsBucket | `status.outputs.url` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDeliveryPipeline | `spec.serialPipeline.stages[].targetId` | `status.outputs.target_id` |
| GcpDeliveryPipeline | `spec.automations[].selector.targets[].id` | `status.outputs.target_id` |
| GcpDeliveryPipeline | `spec.automations[].rules[].promoteReleaseRule.destinationTargetId` | `status.outputs.target_id` |
| GcpDeliveryPipeline | `spec.automations[].rules[].timedPromoteReleaseRule.destinationTargetId` | `status.outputs.target_id` |
| GcpDeployPolicy | `spec.selectors[].target.id` | `status.outputs.target_id` |
| GcpDeployTarget | `spec.multiTarget.targetIds` | `status.outputs.target_id` |

## See Also

- [Overview](../README.md)
