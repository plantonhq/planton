# DigitalOceanKubernetesNodePool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanKubernetesNodePoolSpec models the full surface of the
digitalocean_kubernetes_node_pool resource: an additional worker pool
attached to an existing DOKS cluster. The cluster's own default pool is
part of the DigitalOceanKubernetesCluster kind; use this kind to grow a
cluster with separately sized, labeled, tainted, or GPU pools.

## Example

```yaml
# Example DigitalOceanKubernetesNodePool manifests.
#
# Deploy with: planton apply -f manifest.yaml
#
# The first document is the smallest real pool: a fixed one-node pool on an
# existing cluster (referenced by UUID). The second exercises the full
# surface: autoscaling bounds, Kubernetes node labels, a taint isolating
# the pool's nodes, pool-level Droplet tags, and AMD GPU partitioning.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanKubernetesNodePool
metadata:
  name: example-doknp-minimal
spec:
  nodePoolName: example-doknp-minimal
  cluster:
    value: fb7d9b81-fe06-4ee5-87f1-b9efc5af46fd
  size: s-1vcpu-2gb
  nodeCount: 1
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanKubernetesNodePool
metadata:
  name: example-doknp-full
spec:
  nodePoolName: gpu-workers
  cluster:
    value: fb7d9b81-fe06-4ee5-87f1-b9efc5af46fd
  size: gpu-mi300x1-192gb
  nodeCount: 1
  autoScale: true
  minNodes: 1
  maxNodes: 3
  labels:
    workload: ml-training
  taints:
    - key: nvidia.com/gpu
      value: "true"
      effect: NoSchedule
  tags:
    - team:ml
    - env:prod
  gpuPartitionMode: AMD_PARTITION_MODE_SPX_NPS1
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.nodePoolName` | `string` | yes |  |  |
| `spec.cluster` | `string \| valueFrom` | yes |  | DigitalOceanKubernetesCluster (`status.outputs.cluster_id`) |
| `spec.size` | `string` | yes |  |  |
| `spec.nodeCount` | `uint32` | yes |  |  |
| `spec.autoScale` | `bool` |  |  |  |
| `spec.minNodes` | `uint32` |  |  |  |
| `spec.maxNodes` | `uint32` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.taints` | `[]DigitalOceanKubernetesNodePoolTaint` |  |  |  |
| `spec.taints[].key` | `string` | yes |  |  |
| `spec.taints[].value` | `string` |  |  |  |
| `spec.taints[].effect` | `string` | yes |  |  |
| `spec.tags` | `[]string` |  |  |  |
| `spec.gpuPartitionMode` | `string` |  |  |  |

## Field Details

### spec.nodePoolName

`string` · required

A name for the node pool. Must be unique within the Kubernetes cluster.

- rule: {"required":true}

### spec.cluster

`string | valueFrom` · required

The DOKS cluster that owns this pool. Accepts the cluster UUID directly
or a reference to a DigitalOceanKubernetesCluster resource (resolved
from its cluster_id output). Changing it replaces the pool.

- references: DigitalOceanKubernetesCluster (`status.outputs.cluster_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanKubernetesCluster, name: <that resource's name>, fieldPath: status.outputs.cluster_id}} -- a bare string does not parse

### spec.size

`string` · required

The slug identifier for the Droplet size of each node (e.g.
"s-2vcpu-4gb"). Changing it replaces the pool.

- rule: {"required":true}

### spec.nodeCount

`uint32` · required

The number of nodes in the pool. With auto_scale enabled this is the
initial count; the live count then drifts freely between min_nodes and
max_nodes without producing configuration diffs.

- rule: {"required":true,"uint32":{"gt":0}}

### spec.autoScale

`bool`

Whether DigitalOcean's cluster-autoscaler manages this pool's node count
between min_nodes and max_nodes.

### spec.minNodes

`uint32`

Minimum node count when auto_scale is enabled.

### spec.maxNodes

`uint32`

Maximum node count when auto_scale is enabled.

### spec.labels

`map<string, string>`

(Optional) Kubernetes labels applied to every node in the pool, in
addition to the standard Planton labels both provisioners always apply.
Labels drive Kubernetes scheduling (nodeSelector, affinity).

### spec.taints

`[]DigitalOceanKubernetesNodePoolTaint`

(Optional) Kubernetes taints applied to every node in the pool. Taints
keep pods without a matching toleration off these nodes -- the standard
isolation mechanism for GPU or dedicated system pools.

### spec.taints[].key

`string` · required

Taint key, e.g. "dedicated".

- rule: {"required":true}

### spec.taints[].value

`string`

(Optional) Taint value, e.g. "gpu-workloads". Kubernetes allows
valueless taints, so empty is legal; the provisioners always send the
value (possibly empty), which is all the provider's required leaf asks.

### spec.taints[].effect

`string` · required

Taint effect. One of NoSchedule, PreferNoSchedule, NoExecute
(case-sensitive, exactly as Kubernetes spells them).

- rule: {"required":true,"string":{"in":["NoSchedule","PreferNoSchedule","NoExecute"]}}

### spec.tags

`[]string`

(Optional) DigitalOcean tags applied to the pool's Droplets, in addition
to the standard Planton tags both provisioners always apply. Tags drive
DigitalOcean-side grouping and billing attribution; they are unrelated
to Kubernetes labels. Tags are also the wiring surface for
Droplet-scoped resources: a DigitalOceanFirewall or load balancer that
targets a pool tag covers every current AND future node, because DOKS
applies the pool's tags to each node it creates -- the pool's node and
Droplet ids are never exported for that reason (they churn by design).
Never author tags with the `k8s:` or `terraform:` prefixes; DOKS owns
those and the provider filters them out of state.

- rule: {"repeated":{"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

### spec.gpuPartitionMode

`string`

(Optional) GPU partitioning mode for AMD GPU Droplet sizes. Changing it
replaces the pool.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AMD_PARTITION_MODE_SPX_NPS1","AMD_PARTITION_MODE_DPX_NPS2"]}}

## Validation Rules

- `autoscale_bounds`: auto_scale requires min_nodes >= 1 and max_nodes >= min_nodes

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanKubernetesNodePool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.node_pool_id` | `string` | The unique identifier (UUID) of the created node pool. |
| `status.outputs.cluster_id` | `string` | The UUID of the cluster that owns this pool. The API addresses the pool as /v2/kubernetes/clusters/{cluster_id}/node_pools/{node_pool_id}, so consumers need both ids to reach it. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.cluster` | DigitalOceanKubernetesCluster | `status.outputs.cluster_id` |

## See Also

- [Overview](../README.md)
