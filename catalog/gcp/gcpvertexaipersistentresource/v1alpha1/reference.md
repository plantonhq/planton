# GcpVertexAiPersistentResource

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiPersistentResourceSpec defines a Vertex AI persistent
resource (`google_vertex_ai_persistent_resource`) -- a long-running
cluster of machines Vertex AI keeps provisioned so custom training jobs
(and Ray on Vertex AI) start in seconds instead of waiting minutes for
capacity, and so scarce accelerators stay reserved between jobs. A job
runs on it by naming its id in `persistent_resource_id`; the job's
network and encryption must match the resource's.

Every replica bills from the moment it is provisioned until the resource
is deleted, whether or not a job is running.

Immutable: everything except the display name, labels, and each pool's
replica_count -- a change to any other setting replaces the resource.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiPersistentResource
metadata:
  name: training-pool
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  displayName: Training pool
  labels:
    team: ml
  # Every replica bills from the moment it is provisioned until the
  # resource is deleted, whether or not a job is running.
  resourcePools:
    - id: cpu-workers
      machineSpec:
        machineType: n1-standard-4
      replicaCount: 1
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.persistentResourceId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.resourcePools` | `[]GcpVertexAiPersistentResourceResourcePool` | yes |  |  |
| `spec.resourcePools[].id` | `string` |  |  |  |
| `spec.resourcePools[].machineSpec` | `GcpVertexAiPersistentResourceMachineSpec` | yes |  |  |
| `spec.resourcePools[].machineSpec.machineType` | `string` |  |  |  |
| `spec.resourcePools[].machineSpec.acceleratorType` | `string` |  |  |  |
| `spec.resourcePools[].machineSpec.acceleratorCount` | `int32` |  |  |  |
| `spec.resourcePools[].replicaCount` | `int64` |  |  |  |
| `spec.resourcePools[].autoscalingSpec` | `GcpVertexAiPersistentResourceAutoscalingSpec` |  |  |  |
| `spec.resourcePools[].autoscalingSpec.minReplicaCount` | `int64` |  |  |  |
| `spec.resourcePools[].autoscalingSpec.maxReplicaCount` | `int64` |  |  |  |
| `spec.resourcePools[].diskSpec` | `GcpVertexAiPersistentResourceDiskSpec` |  |  |  |
| `spec.resourcePools[].diskSpec.bootDiskSizeGb` | `int32` |  |  |  |
| `spec.resourcePools[].diskSpec.bootDiskType` | `string` |  |  |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.reservedIpRanges` | `[]string` |  |  |  |
| `spec.pscInterfaceConfig` | `GcpVertexAiPersistentResourcePscInterfaceConfig` |  |  |  |
| `spec.pscInterfaceConfig.networkAttachment` | `string` |  |  |  |
| `spec.pscInterfaceConfig.dnsPeeringConfigs` | `[]GcpVertexAiPersistentResourceDnsPeeringConfig` |  |  |  |
| `spec.pscInterfaceConfig.dnsPeeringConfigs[].domain` | `string` | yes |  |  |
| `spec.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.enableCustomServiceAccount` | `bool` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the resource lives in: a literal project ID or a
GcpProject reference. If omitted, the provider's default project is
used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The Vertex AI location (region) the resource runs in, e.g.
"us-central1". Jobs that use it must run in the same location.
Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+$"}}

### spec.persistentResourceId

`string`

The resource's id -- what a job's persistent_resource_id names. Up to
63 characters: lowercase letters, digits, and hyphens, starting with a
letter and not ending with a hyphen. Defaults to metadata.name.
Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$"}}

### spec.displayName

`string`

Human-readable name shown in the console -- up to 128 UTF-8
characters. Mutable in place.

- rule: {"string":{"maxLen":"128"}}

### spec.labels

`map<string, string>`

Labels on the resource. The platform attribution labels are merged in
and win on key conflicts. Mutable in place.

### spec.resourcePools

`[]GcpVertexAiPersistentResourceResourcePool` · required

The pools of machines the resource keeps provisioned. At least one.

- rule: {"repeated":{"minItems":"1"}}

### spec.resourcePools[].id

`string`

The pool's id within the resource -- what a job's worker pool spec
refers to. Google generates one when empty. Sent only when set.
Immutable.

### spec.resourcePools[].machineSpec

`GcpVertexAiPersistentResourceMachineSpec` · required

The machine every replica runs on.

- rule: {"required":true}

### spec.resourcePools[].machineSpec.machineType

`string`

The Compute Engine machine type, e.g. "n1-standard-4", "a2-highgpu-1g",
"ct5lp-hightpu-4t" -- any type Vertex AI custom training supports.
Immutable.

### spec.resourcePools[].machineSpec.acceleratorType

`string`

The accelerator attached to every replica (NVIDIA_TESLA_T4,
NVIDIA_L4, NVIDIA_TESLA_A100, NVIDIA_A100_80GB, NVIDIA_H100_80GB,
NVIDIA_H100_MEGA_80GB, NVIDIA_H200_141GB, NVIDIA_B200, NVIDIA_GB200,
NVIDIA_RTX_PRO_6000, the older NVIDIA_TESLA_K80 / P100 / V100 / P4,
or TPU_V2 / TPU_V3 / TPU_V4_POD / TPU_V5_LITEPOD). The machine type
must be one Google pairs with it. Empty attaches none. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["NVIDIA_TESLA_K80","NVIDIA_TESLA_P100","NVIDIA_TESLA_V100","NVIDIA_TESLA_P4","NVIDIA_TESLA_T4","NVIDIA_TESLA_A100","NVIDIA_A100_80GB","NVIDIA_L4","NVIDIA_H100_80GB","NVIDIA_H100_MEGA_80GB","NVIDIA_H200_141GB","NVIDIA_B200","NVIDIA_GB200","NVIDIA_RTX_PRO_6000","TPU_V2","TPU_V3","TPU_V4_POD","TPU_V5_LITEPOD"]}}

### spec.resourcePools[].machineSpec.acceleratorCount

`int32`

How many accelerators each replica carries, set together with
accelerator_type. Sent only when set. Immutable.

- rule: {"int32":{"gte":0}}

### spec.resourcePools[].replicaCount

`int64` · optional (explicit presence)

How many replicas the pool runs (and bills) whether or not a job is
using them. Sent as a decimal string. Mutable in place.

- rule: {"int64":{"gte":"1"}}

### spec.resourcePools[].autoscalingSpec

`GcpVertexAiPersistentResourceAutoscalingSpec`

Autoscaling bounds for the pool. Omit for a fixed replica_count.

### spec.resourcePools[].autoscalingSpec.minReplicaCount

`int64` · optional (explicit presence)

Fewest replicas kept running -- at least 1 on a persistent resource
(Google rejects 0 here), and at most the pool's replica_count. Sent as
a decimal string. Immutable.

- rule: {"int64":{"gte":"1"}}

### spec.resourcePools[].autoscalingSpec.maxReplicaCount

`int64` · optional (explicit presence)

Most replicas the pool may scale to -- more than min_replica_count and
at least the pool's replica_count. Sent as a decimal string.
Immutable.

- rule: {"int64":{"gte":"1"}}

### spec.resourcePools[].diskSpec

`GcpVertexAiPersistentResourceDiskSpec`

Boot disk options. Omit for Google's defaults. Sent only when set.

### spec.resourcePools[].diskSpec.bootDiskSizeGb

`int32` · optional (explicit presence)

Boot disk size in GB. Google defaults to 100. Sent only when set.
Immutable.

- rule: {"int32":{"gte":1}}

### spec.resourcePools[].diskSpec.bootDiskType

`string`

Boot disk type: "pd-ssd" (Google's default on most machines),
"pd-standard", or "hyperdisk-balanced" (the default on A3 Ultra). Sent
only when set. Immutable.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["pd-ssd","pd-standard","hyperdisk-balanced"]}}

### spec.network

`string | valueFrom`

A VPC network to peer the resource with, so jobs reach private
services: a GcpVpcNetwork reference or a literal
projects/{project}/global/networks/{name}. Google requires the project
NUMBER in that path; both modules resolve a project ID to its number
(one project lookup at plan time). The network must already have VPC
Network Peering for Vertex AI (private services access) configured.
Omit for no peering. Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.reservedIpRanges

`[]string`

Names of the private-services-access ranges on the peered network the
resource's machines take addresses from, e.g. ["vertex-ai-range"]
(the name of a GcpGlobalAddress with purpose VPC_PEERING). Empty lets
Google use any range allocated to the peering. Requires network.
Immutable.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.pscInterfaceConfig

`GcpVertexAiPersistentResourcePscInterfaceConfig`

A Private Service Connect interface into your VPC -- the alternative
to network (peering); Google's documentation treats the two as
alternatives and its API is the authority when both are set.
Immutable.

### spec.pscInterfaceConfig.networkAttachment

`string`

The Compute Engine network attachment in the resource's region the
interface joins, as its name or its full
projects/{project}/regions/{region}/networkAttachments/{name} path.
Create the attachment first. Immutable.

### spec.pscInterfaceConfig.dnsPeeringConfigs

`[]GcpVertexAiPersistentResourceDnsPeeringConfig`

Domains Google's tenant VPC resolves through Cloud DNS zones in your
networks.

### spec.pscInterfaceConfig.dnsPeeringConfigs[].domain

`string` · required

The DNS suffix peered to, ending with a dot, e.g.
"corp.example.com.". Immutable.

- rule: {"required":true,"string":{"pattern":"^.+\\.$"}}

### spec.pscInterfaceConfig.dnsPeeringConfigs[].targetProject

`string | valueFrom` · required

The project hosting the Cloud DNS zone for the domain: a GcpProject
reference or a literal project ID. The Vertex AI Service Agent needs
roles/dns.peer on it. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork

`string | valueFrom` · required

The VPC network, in target_project, where the zone is visible: a
GcpVpcNetwork reference (its name) or a literal network name.
Immutable.

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.enableCustomServiceAccount

`bool`

True requires every job on the resource to run as a custom,
user-managed service account (the job names it); false runs jobs as
the Vertex AI Custom Code Service Agent. Immutable.

### spec.kmsKeyName

`string | valueFrom`

Customer-managed encryption key protecting the resource's disks: a
GcpKmsKey reference or a literal
projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
in the same region. Jobs on the resource must use the same key. Omit to
use Google-managed encryption. Immutable.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

What happens to the resource when this block is destroyed:
  "" / "DELETE" -- the machines are released and billing stops
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the resource leaves management and keeps running
                   (and billing)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiPersistentResource, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/persistentResources/{persistent_resource_id}. |
| `status.outputs.persistent_resource_id` | `string` | The resource's id -- what a job's persistent_resource_id takes. |
| `status.outputs.location` | `string` | The location the resource runs in. |
| `status.outputs.state` | `string` | The resource's state as last read (RUNNING once every pool is provisioned). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.pscInterfaceConfig.dnsPeeringConfigs[].targetProject` | GcpProject | `status.outputs.project_id` |
| `spec.pscInterfaceConfig.dnsPeeringConfigs[].targetNetwork` | GcpVpcNetwork | `status.outputs.network_name` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
