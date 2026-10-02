# GcpTpuVm

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpTpuVmSpec defines a Cloud TPU VM (`google_tpu_v2_vm`) -- a slice of
Google's AI accelerators (v2 through v6e / Trillium) with its host VMs,
for training and serving large models with JAX, PyTorch/XLA, or
TensorFlow. The accelerator type or topology decides the slice size
(v5litepod-8 is 8 chips on one host; larger slices span hosts that each
get their own VM and network endpoint).

Google publishes this resource only in its beta Terraform provider; the
module uses google-beta for it under the catalog's recorded admission.

Capacity is the practical constraint: TPUs live in specific zones, need
quota, and are often scarce. spot (or preemptible) capacity is far
cheaper and can be reclaimed at any time; reserved draws from a
reservation. When on-demand capacity is short, request the slice
through a GcpTpuQueuedResource instead, which waits for capacity.

Immutable: almost everything -- zone, node_id, runtime_version, the
accelerator, networking, service account, scheduling, and Secure Boot.
Mutable in place: description, labels, metadata, tags, and data_disks.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTpuVm
metadata:
  name: train-v5e
spec:
  projectId:
    value: my-gcp-project
  zone: us-west4-a
  acceleratorType: v5litepod-8
  runtimeVersion: v2-alpha-tpuv5-lite
  description: Fine-tuning slice
  schedulingConfig:
    # Far cheaper; Google may reclaim it at any time -- checkpoint often.
    spot: true
  labels:
    team: research
  metadata:
    startup-script: pip install -U "jax[tpu]"
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.zone` | `string` | yes |  |  |
| `spec.nodeId` | `string` |  |  |  |
| `spec.runtimeVersion` | `string` | yes |  |  |
| `spec.acceleratorType` | `string` |  |  |  |
| `spec.acceleratorConfig` | `GcpTpuVmAcceleratorConfig` |  |  |  |
| `spec.acceleratorConfig.type` | `string` | yes |  |  |
| `spec.acceleratorConfig.topology` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.cidrBlock` | `string` |  |  |  |
| `spec.networkConfig` | `GcpTpuVmNetworkConfig` |  |  |  |
| `spec.networkConfig.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.networkConfig.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkConfig.enableExternalIps` | `bool` |  |  |  |
| `spec.networkConfig.canIpForward` | `bool` |  |  |  |
| `spec.networkConfig.queueCount` | `int32` |  |  |  |
| `spec.networkConfigs` | `[]GcpTpuVmNetworkConfig` |  |  |  |
| `spec.networkConfigs[].network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.networkConfigs[].subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkConfigs[].enableExternalIps` | `bool` |  |  |  |
| `spec.networkConfigs[].canIpForward` | `bool` |  |  |  |
| `spec.networkConfigs[].queueCount` | `int32` |  |  |  |
| `spec.serviceAccount` | `GcpTpuVmServiceAccount` |  |  |  |
| `spec.serviceAccount.email` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.serviceAccount.scopes` | `[]string` |  |  |  |
| `spec.schedulingConfig` | `GcpTpuVmSchedulingConfig` |  |  |  |
| `spec.schedulingConfig.preemptible` | `bool` |  |  |  |
| `spec.schedulingConfig.spot` | `bool` |  |  |  |
| `spec.schedulingConfig.reserved` | `bool` |  |  |  |
| `spec.dataDisks` | `[]GcpTpuVmDataDisk` |  |  |  |
| `spec.dataDisks[].sourceDisk` | `string \| valueFrom` | yes |  | GcpComputeDisk (`status.outputs.self_link`) |
| `spec.dataDisks[].mode` | `string` |  |  |  |
| `spec.enableSecureBoot` | `bool` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |
| `spec.tags` | `[]string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the TPU lives in: a literal project ID or a GcpProject
reference. If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.zone

`string` · required

The zone the TPU runs in, e.g. "us-central1-a" or "europe-west4-a".
Only some zones offer each TPU generation; check Google's TPU regions
and zones list. Immutable.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+-[a-z]$"}}

### spec.nodeId

`string`

The TPU's id -- its name in the project. Lowercase letters, digits,
and hyphens, starting with a letter. Defaults to metadata.name.
Immutable.

- rule: node_id must be lowercase letters, digits, and hyphens, starting with a letter, up to 63 characters

### spec.runtimeVersion

`string` · required

The TPU software image, matched to the accelerator generation and
framework, e.g. "tpu-ubuntu2204-base", "v2-alpha-tpuv5-lite",
"v2-alpha-tpuv6e". Immutable.

- rule: {"required":true}

### spec.acceleratorType

`string`

The slice by name: generation and chip or core count, e.g. "v2-8",
"v3-8", "v5litepod-8", "v6e-8". Set this or accelerator_config; with
neither, Google's provider asks for "v2-8". Immutable.

### spec.acceleratorConfig

`GcpTpuVmAcceleratorConfig`

The slice by generation and topology -- how you ask for shapes the
type names do not cover. Set this or accelerator_type. Immutable.

### spec.acceleratorConfig.type

`string` · required

The TPU generation: "V2", "V3", "V4", "V5LITE_POD", "V5P", or "V6E".

- rule: {"required":true,"string":{"pattern":"^V[0-9][A-Z0-9_]*$"}}

### spec.acceleratorConfig.topology

`string` · required

The chip topology, e.g. "2x2" or "2x2x1".

- rule: {"required":true,"string":{"pattern":"^[0-9]+(x[0-9]+)+$"}}

### spec.description

`string`

What the TPU is for.

### spec.cidrBlock

`string`

A /29 CIDR block the TPU picks its address from. It must not overlap
any subnetwork of the network, any peered network, or another TPU's
block. Unset: Google chooses. Immutable.

- rule: cidr_block must be an IPv4 /29 block

### spec.networkConfig

`GcpTpuVmNetworkConfig`

The TPU's network. Set this or network_configs. Unset: the project's
default network and subnetwork. Immutable.

### spec.networkConfig.network

`string | valueFrom`

The VPC network: a GcpVpcNetwork reference or a literal
projects/{project}/global/networks/{name}. Unset: "default".

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.networkConfig.subnetwork

`string | valueFrom`

The subnetwork in the TPU's region: a GcpSubnetwork reference or a
literal projects/{project}/regions/{region}/subnetworks/{name}. Unset:
"default".

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkConfig.enableExternalIps

`bool`

Give the TPU workers external IP addresses. Without them the
subnetwork needs Private Google Access (or Cloud NAT) to reach
Google APIs and the internet.

### spec.networkConfig.canIpForward

`bool`

Let the workers send and receive packets with non-matching source or
destination addresses -- needed only when they forward routes.

### spec.networkConfig.queueCount

`int32`

The number of queues on the interface (higher network throughput on
large hosts). Unset: Google's default.

- rule: {"int32":{"gte":0}}

### spec.networkConfigs

`[]GcpTpuVmNetworkConfig`

Several network interfaces, one per entry, for multi-NIC TPU VMs.
Set this or network_config. Immutable.

### spec.networkConfigs[].network

`string | valueFrom`

The VPC network: a GcpVpcNetwork reference or a literal
projects/{project}/global/networks/{name}. Unset: "default".

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.networkConfigs[].subnetwork

`string | valueFrom`

The subnetwork in the TPU's region: a GcpSubnetwork reference or a
literal projects/{project}/regions/{region}/subnetworks/{name}. Unset:
"default".

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkConfigs[].enableExternalIps

`bool`

Give the TPU workers external IP addresses. Without them the
subnetwork needs Private Google Access (or Cloud NAT) to reach
Google APIs and the internet.

### spec.networkConfigs[].canIpForward

`bool`

Let the workers send and receive packets with non-matching source or
destination addresses -- needed only when they forward routes.

### spec.networkConfigs[].queueCount

`int32`

The number of queues on the interface (higher network throughput on
large hosts). Unset: Google's default.

- rule: {"int32":{"gte":0}}

### spec.serviceAccount

`GcpTpuVmServiceAccount`

The identity the TPU host VMs run as. Unset: the Compute Engine
default service account with access to every Cloud API. Immutable.

### spec.serviceAccount.email

`string | valueFrom`

The service account: a GcpServiceAccount reference or a literal email.
Unset: the Compute Engine default service account.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.serviceAccount.scopes

`[]string`

OAuth scopes granted to it. Unset: every Cloud API
(https://www.googleapis.com/auth/cloud-platform); access is then
governed by the account's IAM roles.

### spec.schedulingConfig

`GcpTpuVmSchedulingConfig`

Spot, preemptible, or reserved capacity. Unset: on-demand. Immutable.

### spec.schedulingConfig.preemptible

`bool`

Preemptible capacity (the older model; Google may reclaim it and ends
it after 24 hours).

### spec.schedulingConfig.spot

`bool`

Spot capacity -- the cheapest; Google may reclaim it at any time, with
no 24-hour limit. Checkpoint often.

### spec.schedulingConfig.reserved

`bool`

Draw from a reservation made for this project and zone.

### spec.dataDisks

`[]GcpTpuVmDataDisk`

Existing persistent disks to attach -- training data or checkpoints
shared across TPUs. Mutable.

### spec.dataDisks[].sourceDisk

`string | valueFrom` · required

The disk: a GcpComputeDisk reference or a literal
projects/{project}/zones/{zone}/disks/{name} in the TPU's zone.

- references: GcpComputeDisk (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeDisk, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.dataDisks[].mode

`string`

"READ_WRITE" (the default; one TPU at a time) or "READ_ONLY" (shared
read access for many TPUs).

- rule: mode must be READ_WRITE or READ_ONLY

### spec.enableSecureBoot

`bool`

Boot the host VMs with Secure Boot (Shielded VM). Sent only when true.
Immutable.

### spec.labels

`map<string, string>`

Labels on the TPU. The platform attribution labels are merged in and
win on key conflicts.

### spec.metadata

`map<string, string>`

Custom metadata on the host VMs -- for example "startup-script" (runs
on every worker at boot) and "shutdown-script". Mutable.

### spec.tags

`[]string`

Network tags on the host VMs -- the handle VPC firewall rules target.
Mutable.

- rule: {"repeated":{"unique":true}}

### spec.deletionPolicy

`string`

What happens to the TPU when this resource is destroyed:
  "" / "DELETE" -- the TPU and its host VMs are deleted (attached data
                   disks are only detached)
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the TPU leaves management and keeps running (and
                   billing)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `accelerator_type_xor_config`: set at most one of accelerator_type or accelerator_config
- `network_config_xor_configs`: set at most one of network_config or network_configs

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpTpuVm, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{zone}/nodes/{node_id}. |
| `status.outputs.node_id` | `string` | The TPU's id (the last segment of name) -- what `gcloud compute tpus tpu-vm ssh` and the TPU client libraries take. |
| `status.outputs.zone` | `string` | The zone the TPU runs in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.networkConfig.network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.networkConfig.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.networkConfigs[].network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.networkConfigs[].subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.serviceAccount.email` | GcpServiceAccount | `status.outputs.email` |
| `spec.dataDisks[].sourceDisk` | GcpComputeDisk | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
