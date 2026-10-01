# GcpBigtableInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpBigtableInstanceSpec defines the configuration for a Cloud Bigtable
instance with one or more clusters.

Cloud Bigtable is a fully managed, wide-column NoSQL database designed
for large analytical and operational workloads. It provides consistent
sub-10ms latency, scales to billions of rows and thousands of columns,
and is ideal for time-series data, IoT, ad-tech, fintech, and
machine-learning feature stores.

This component bundles the instance (the logical container for data)
with one or more clusters (the physical replicas serving the data).
An instance without at least one cluster cannot store or serve data.
Bigtable tables and app profiles are application-level concerns
managed separately, not included here.

Multi-cluster instances provide automatic replication and failover.
Each cluster must be in a different zone. The Bigtable client library
handles routing and failover between clusters transparently.

Important behavioral notes:

  - instance_name is immutable after creation (6-33 characters).
  - Each cluster's zone, storage_type, kms_key_name, and
    node_scaling_factor are immutable after creation.
  - GCP labels are supported at the instance level (not per cluster).
  - The GCP distinction between DEVELOPMENT and PRODUCTION instance
    types is deprecated and being removed. All instances are effectively
    PRODUCTION. Use a single small cluster for development workloads.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBigtableInstance
metadata:
  name: test-bigtable
  org: test-org
  env: dev
spec:
  projectId:
    value: my-gcp-project
  instanceName: test-bigtable-inst
  displayName: Test Bigtable Instance
  deletionProtection: false
  forceDestroy: true
  # ENTERPRISE is the provider default; ENTERPRISE_PLUS unlocks
  # multi-location automated-backup placement on tables.
  edition: ENTERPRISE
  # Explicit DELETE (the provider default) proves the round-trip.
  deletionPolicy: DELETE
  clusters:
    - clusterId: test-cluster-c1
      zone: us-central1-a
      numNodes: 3
      storageType: SSD
      nodeScalingFactor: NodeScalingFactor1X
      kmsKeyName:
        value: projects/my-gcp-project/locations/us-central1/keyRings/test-kr/cryptoKeys/test-key
    - clusterId: test-cluster-c2
      zone: us-east1-b
      storageType: SSD
      autoscalingConfig:
        minNodes: 1
        maxNodes: 5
        cpuTarget: 60
        storageTarget: 2560
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instanceName` | `string` | yes |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.clusters` | `[]GcpBigtableInstanceCluster` | yes |  |  |
| `spec.clusters[].clusterId` | `string` | yes |  |  |
| `spec.clusters[].zone` | `string` | yes |  |  |
| `spec.clusters[].numNodes` | `int32` |  |  |  |
| `spec.clusters[].storageType` | `string` |  | `SSD` |  |
| `spec.clusters[].kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`) |
| `spec.clusters[].nodeScalingFactor` | `string` |  |  |  |
| `spec.clusters[].autoscalingConfig` | `GcpBigtableInstanceClusterAutoscalingConfig` |  |  |  |
| `spec.clusters[].autoscalingConfig.minNodes` | `int32` | yes |  |  |
| `spec.clusters[].autoscalingConfig.maxNodes` | `int32` | yes |  |  |
| `spec.clusters[].autoscalingConfig.cpuTarget` | `int32` | yes |  |  |
| `spec.clusters[].autoscalingConfig.storageTarget` | `int32` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.edition` | `string` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the Bigtable instance will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing the project destroys and recreates the instance.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instanceName

`string` · required

Name of the Bigtable instance (also called Instance ID in GCP Console).
This becomes the GCP resource name and is used by Bigtable client
libraries to connect. Must be 6-33 characters: lowercase letters,
numbers, and hyphens only. Must start with a lowercase letter and
end with a letter or number.
Immutable after creation.

- rule: {"required":true,"string":{"minLen":"6","maxLen":"33","pattern":"^[a-z][a-z0-9-]{4,31}[a-z0-9]$"}}

### spec.displayName

`string`

Human-readable display name for the instance.
If not specified, defaults to the instance_name value.

### spec.deletionProtection

`bool` · optional (explicit presence)

Whether deletion protection is enabled. When true, the instance
cannot be destroyed without first setting this to false.
Strongly recommended for production instances.
Default: true.

- default: `true`

### spec.forceDestroy

`bool`

Whether to delete all backups in the instance when destroying it.
Bigtable blocks instance deletion if backups exist unless this is
set to true. Only relevant during destroy operations.

### spec.clusters

`[]GcpBigtableInstanceCluster` · required

One or more clusters that serve as physical replicas for this instance.
Each cluster must be in a different zone within the same or different
regions. At least one cluster is required. Up to 8 clusters can be
configured across cloud regions for multi-region replication.

- rule: {"repeated":{"minItems":"1"}}
- rule: only one of num_nodes or autoscaling_config may be set

### spec.clusters[].clusterId

`string` · required

Unique identifier for this cluster within the instance.
Must be 6-30 characters: lowercase letters, numbers, and hyphens only.
Must start with a lowercase letter and end with a letter or number.

- rule: {"required":true,"string":{"minLen":"6","maxLen":"30","pattern":"^[a-z][a-z0-9-]{4,28}[a-z0-9]$"}}

### spec.clusters[].zone

`string` · required

Zone where this cluster will be deployed (e.g., "us-central1-a").
Each cluster in the instance must be in a different zone.
Zones must be Bigtable-capable (see GCP Bigtable locations).
Immutable after creation.

- rule: {"required":true}

### spec.clusters[].numNodes

`int32`

Fixed number of nodes in this cluster. Mutually exclusive with
autoscaling_config. If neither is set, Bigtable auto-allocates
nodes based on the data footprint.

### spec.clusters[].storageType

`string` · optional (explicit presence)

Storage type for this cluster.
SSD: lower latency, recommended for most workloads.
HDD: lower cost, suitable for large batch-analytics workloads
where latency is less critical.
Default: SSD.
Immutable after creation.

- default: `SSD`
- rule: storage_type must be SSD or HDD

### spec.clusters[].kmsKeyName

`string | valueFrom`

Cloud KMS encryption key to protect data in this cluster (CMEK).
Format: projects/{project}/locations/{location}/keyRings/{keyring}/cryptoKeys/{key}
The key region must match the cluster zone's region. All clusters
within an instance should use the same CMEK key.
Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.clusters[].nodeScalingFactor

`string`

Node scaling factor for this cluster. Controls the granularity of
node scaling: 1X scales in increments of 1 node, 2X scales in
increments of 2 nodes (for larger workloads that benefit from
coarser scaling steps). When using 2X, num_nodes, min_nodes, and
max_nodes must all be specified in increments of 2.
If not set, GCP defaults to NodeScalingFactor1X.
Immutable after creation.

- rule: node_scaling_factor must be NodeScalingFactor1X or NodeScalingFactor2X

### spec.clusters[].autoscalingConfig

`GcpBigtableInstanceClusterAutoscalingConfig`

Autoscaling configuration for this cluster. Mutually exclusive with
num_nodes. When set, Bigtable dynamically adjusts the number of
nodes based on CPU and storage utilization targets.

- rule: max_nodes must be greater than or equal to min_nodes

### spec.clusters[].autoscalingConfig.minNodes

`int32` · required

Minimum number of nodes for autoscaling. Must be at least 1.

- rule: {"required":true,"int32":{"gte":1}}

### spec.clusters[].autoscalingConfig.maxNodes

`int32` · required

Maximum number of nodes for autoscaling. Must be at least 1 and
greater than or equal to min_nodes.

- rule: {"required":true,"int32":{"gte":1}}

### spec.clusters[].autoscalingConfig.cpuTarget

`int32` · required

Target CPU utilization percentage for autoscaling. Bigtable adds nodes
when average CPU utilization exceeds this target and removes nodes
when utilization drops sufficiently below it. Must be between 10 and 80.

- rule: {"required":true,"int32":{"lte":80,"gte":10}}

### spec.clusters[].autoscalingConfig.storageTarget

`int32`

Target storage utilization per node in GB. When total storage per node
exceeds this target, Bigtable adds nodes. The valid range depends on
the cluster's storage_type:
  - SSD: 2560 to 5120 (2.5 TiB to 5 TiB per node)
  - HDD: 8192 to 16384 (8 TiB to 16 TiB per node)
If not set, Bigtable uses the default for the storage type
(2560 for SSD, 8192 for HDD).

### spec.labels

`map<string, string>`

User-defined labels to organize and track the instance (instance-level
only — GCP has no per-cluster labels). Merged beneath Planton's
platform attribution labels (platform keys win on conflict).

### spec.edition

`string`

Edition of the instance, gating feature availability. ENTERPRISE
(the GCP default when unset) is the standard production edition;
ENTERPRISE_PLUS adds enterprise capabilities such as multi-location
automated-backup placement for the instance's tables. Upgrading
ENTERPRISE -> ENTERPRISE_PLUS applies in place; there is no
downgrade path.

- rule: edition must be ENTERPRISE or ENTERPRISE_PLUS

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the instance for org-policy and IAM
conditions. Keys in the form "tagKeys/{id}" (or the namespaced
"project/tag-key"), values "tagValues/{id}" (or
"project/tag-key/tag-value"). Create-time only: changing them later
replaces the instance — plan tag changes deliberately.

### spec.deletionPolicy

`string`

Deletion policy for the instance — what happens when this resource
is destroyed. Applies only once deletion_protection (above) permits
a destroy at all:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the instance is deleted with every cluster and table
               on it (backups additionally gated by force_destroy)
  "PREVENT" -- destroy FAILS; a second wall for the instance a
               data platform depends on
  "ABANDON" -- the instance is removed from management but left
               running (and billing) in GCP with its data intact

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpBigtableInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_id` | `string` | Fully qualified instance resource name. Format: projects/{project}/instances/{instance} Used by Bigtable client libraries and downstream tools that reference this instance. |
| `status.outputs.instance_name` | `string` | Short name of the instance (same as the spec's instance_name input). This is the value passed to Bigtable client libraries along with the project ID to establish connections. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.clusters[].kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.clusters[].kmsKeyName` | GcpKmsKeyHandle | `status.outputs.kms_key` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpBigtableTable | `spec.instance` | `status.outputs.instance_name` |

## See Also

- [Overview](../README.md)
