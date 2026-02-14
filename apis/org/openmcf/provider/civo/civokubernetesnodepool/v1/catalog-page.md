# Civo Kubernetes Node Pool

Adds a node pool to an existing Civo Kubernetes cluster, allowing you to scale compute capacity independently of the cluster's default pool. Each node pool runs a specified instance size and node count, with optional auto-scaling between configurable bounds.

## What Gets Created

When you deploy a CivoKubernetesNodePool resource, OpenMCF provisions:

- **Kubernetes Node Pool** — a `civo_kubernetes_node_pool` resource attached to the referenced Civo Kubernetes cluster, with the specified instance size and labels derived from the resource metadata

## Prerequisites

- **Civo credentials** configured via environment variables or OpenMCF provider config
- **An existing Civo Kubernetes cluster** — either created manually or managed via a CivoKubernetesCluster resource
- **A valid instance size slug** (e.g., `g4s.kube.medium`) — check Civo's available sizes for your region

## Quick Start

Create a file `civo-node-pool.yaml`:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoKubernetesNodePool
metadata:
  name: my-node-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoKubernetesNodePool.my-node-pool
spec:
  nodePoolName: my-node-pool
  cluster:
    value: my-cluster-id
  size: g4s.kube.medium
  nodeCount: 2
```

Deploy:

```shell
openmcf apply -f civo-node-pool.yaml
```

This creates a two-node pool using `g4s.kube.medium` instances in the specified cluster.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `nodePoolName` | `string` | Name of the node pool. Must be unique within the Civo Kubernetes cluster. | Required |
| `cluster` | `StringValueOrRef` | Reference to the Civo Kubernetes cluster. Accepts a literal cluster name/ID via `value`, or a cross-resource reference via `valueFrom` pointing to a CivoKubernetesCluster resource. | Required |
| `size` | `string` | Instance size slug for each node in the pool (e.g., `g4s.kube.medium`). Defines CPU and memory allocation. | Required |
| `nodeCount` | `uint32` | Number of nodes to provision in the pool. | Required, must be > 0 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `autoScale` | `bool` | `false` | Enable auto-scaling for this node pool. When `true`, node count is managed automatically between `minNodes` and `maxNodes`. |
| `minNodes` | `uint32` | `0` | Minimum number of nodes when auto-scaling is enabled. Should be set when `autoScale` is `true`. |
| `maxNodes` | `uint32` | `0` | Maximum number of nodes when auto-scaling is enabled. Should be set when `autoScale` is `true`. |
| `tags` | `string[]` | `[]` | Tags to apply to the node pool for organizational purposes within Civo. |

## Examples

### Basic Node Pool

A fixed-size pool added to an existing cluster by ID:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoKubernetesNodePool
metadata:
  name: workers
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoKubernetesNodePool.workers
spec:
  nodePoolName: workers
  cluster:
    value: abc123-cluster-id
  size: g4s.kube.small
  nodeCount: 1
```

### Auto-Scaling Node Pool

A pool that scales between 2 and 8 nodes based on demand:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoKubernetesNodePool
metadata:
  name: autoscale-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoKubernetesNodePool.autoscale-pool
spec:
  nodePoolName: autoscale-pool
  cluster:
    value: abc123-cluster-id
  size: g4s.kube.large
  nodeCount: 3
  autoScale: true
  minNodes: 2
  maxNodes: 8
  tags:
    - environment:production
    - team:platform
```

### Using Foreign Key References

Reference an OpenMCF-managed CivoKubernetesCluster instead of hardcoding the cluster ID:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoKubernetesNodePool
metadata:
  name: ref-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoKubernetesNodePool.ref-pool
spec:
  nodePoolName: ref-pool
  cluster:
    valueFrom:
      kind: CivoKubernetesCluster
      name: my-cluster
      field: status.outputs.cluster_id
  size: g4s.kube.xlarge
  nodeCount: 5
  autoScale: true
  minNodes: 3
  maxNodes: 10
  tags:
    - environment:production
    - managed-by:openmcf
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `node_pool_id` | `string` | Unique identifier of the created node pool |
| `node_ids` | `string[]` | IDs of the individual nodes provisioned in the pool |

## Related Components

- [CivoKubernetesCluster](/docs/catalog/civo/civokubernetescluster) — the parent cluster to which the node pool is added
- [CivoVpc](/docs/catalog/civo/civovpc) — provides the network used by the cluster
- [CivoFirewall](/docs/catalog/civo/civofirewall) — controls network access to cluster nodes
- [CivoVolume](/docs/catalog/civo/civovolume) — provides persistent block storage for workloads running on the node pool
