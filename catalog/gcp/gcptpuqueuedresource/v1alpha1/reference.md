# GcpTpuQueuedResource

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpTpuQueuedResourceSpec defines a Cloud TPU queued resource
(`google_tpu_v2_queued_resource`) -- a request for TPU capacity that
waits in Google's queue until the capacity exists, then provisions the
TPU nodes it describes. It is Google's recommended way to get scarce TPU
capacity: instead of failing when a zone is out of chips (as a direct
GcpTpuVm create does), the request stays WAITING_FOR_RESOURCES and turns
into running TPUs when Google can place them.

One request can ask for several nodes at once (one node_specs entry per
node), which Google provisions together. The nodes belong to the
request: destroying the request deletes them.

Google publishes this resource only in its beta Terraform provider; the
module uses google-beta for it under the catalog's recorded admission.

Immutable: everything. Any change replaces the request and the nodes it
created.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpTpuQueuedResource
metadata:
  name: train-request
spec:
  projectId:
    value: my-gcp-project
  zone: us-west4-a
  nodeSpecs:
    - nodeId: train-v5e
      node:
        runtimeVersion: v2-alpha-tpuv5-lite
        acceleratorType: v5litepod-8
        description: Fine-tuning slice
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.zone` | `string` | yes |  |  |
| `spec.queuedResourceId` | `string` |  |  |  |
| `spec.nodeSpecs` | `[]GcpTpuQueuedResourceNodeSpec` | yes |  |  |
| `spec.nodeSpecs[].nodeId` | `string` |  |  |  |
| `spec.nodeSpecs[].node` | `GcpTpuQueuedResourceNode` | yes |  |  |
| `spec.nodeSpecs[].node.runtimeVersion` | `string` | yes |  |  |
| `spec.nodeSpecs[].node.acceleratorType` | `string` |  |  |  |
| `spec.nodeSpecs[].node.description` | `string` |  |  |  |
| `spec.nodeSpecs[].node.networkConfig` | `GcpTpuQueuedResourceNetworkConfig` |  |  |  |
| `spec.nodeSpecs[].node.networkConfig.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_id`) |
| `spec.nodeSpecs[].node.networkConfig.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.nodeSpecs[].node.networkConfig.enableExternalIps` | `bool` |  |  |  |
| `spec.nodeSpecs[].node.networkConfig.canIpForward` | `bool` |  |  |  |
| `spec.nodeSpecs[].node.networkConfig.queueCount` | `int32` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project the request and its nodes live in: a literal project
ID or a GcpProject reference. If omitted, the provider's default
project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.zone

`string` · required

The zone the capacity is requested in, e.g. "us-central1-a". Every
node is created here.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+-[a-z]$"}}

### spec.queuedResourceId

`string`

The request's id. Lowercase letters, digits, and hyphens, starting
with a letter. Defaults to metadata.name.

- rule: queued_resource_id must be lowercase letters, digits, and hyphens, starting with a letter, up to 63 characters

### spec.nodeSpecs

`[]GcpTpuQueuedResourceNodeSpec` · required

The TPU nodes requested, provisioned together once capacity exists.

- rule: node_id must be unique across node_specs
- rule: {"repeated":{"minItems":"1"}}

### spec.nodeSpecs[].nodeId

`string`

The node's id once provisioned -- its name in the project. Lowercase
letters, digits, and hyphens, starting with a letter. Unset: Google
generates one.

- rule: node_id must be lowercase letters, digits, and hyphens, starting with a letter, up to 63 characters

### spec.nodeSpecs[].node

`GcpTpuQueuedResourceNode` · required

The node itself.

- rule: {"required":true}

### spec.nodeSpecs[].node.runtimeVersion

`string` · required

The TPU software image, matched to the accelerator generation, e.g.
"tpu-ubuntu2204-base" or "v2-alpha-tpuv5-lite".

- rule: {"required":true}

### spec.nodeSpecs[].node.acceleratorType

`string`

The slice: generation and chip or core count, e.g. "v2-8",
"v5litepod-8", "v6e-16". Unset: Google's default of "v2-8".

### spec.nodeSpecs[].node.description

`string`

What the node is for.

### spec.nodeSpecs[].node.networkConfig

`GcpTpuQueuedResourceNetworkConfig`

The node's network. Unset: the project's default network and
subnetwork.

### spec.nodeSpecs[].node.networkConfig.network

`string | valueFrom`

The VPC network: a GcpVpcNetwork reference or a literal
projects/{project}/global/networks/{name}. Unset: "default".

- references: GcpVpcNetwork (`status.outputs.network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_id}} -- a bare string does not parse

### spec.nodeSpecs[].node.networkConfig.subnetwork

`string | valueFrom`

The subnetwork in the zone's region: a GcpSubnetwork reference or a
literal projects/{project}/regions/{region}/subnetworks/{name}. Unset:
"default".

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.nodeSpecs[].node.networkConfig.enableExternalIps

`bool`

Give the node's workers external IP addresses. Without them the
subnetwork needs Private Google Access (or Cloud NAT).

### spec.nodeSpecs[].node.networkConfig.canIpForward

`bool`

Let the workers forward packets with non-matching addresses.

### spec.nodeSpecs[].node.networkConfig.queueCount

`int32`

The number of queues on the interface. Unset: Google's default.

- rule: {"int32":{"gte":0}}

### spec.deletionPolicy

`string`

What happens to the request when this resource is destroyed:
  "" / "DELETE" -- the request is deleted, and with it every node it
                   provisioned
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the request leaves management; its nodes keep
                   running (and billing)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpTpuQueuedResource, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{zone}/queuedResources/{queued_resource_id}. |
| `status.outputs.queued_resource_id` | `string` | The request's id (the last segment of name). |
| `status.outputs.zone` | `string` | The zone the capacity is requested in. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.nodeSpecs[].node.networkConfig.network` | GcpVpcNetwork | `status.outputs.network_id` |
| `spec.nodeSpecs[].node.networkConfig.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |

## See Also

- [Overview](../README.md)
