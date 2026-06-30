# Azure AKS Node Pool

Deploys a node pool into an existing Azure Kubernetes Service (AKS) cluster with configurable VM size, node count, autoscaling, availability zones, OS type, pool mode, and Spot VM pricing. The component creates a `containerservice.AgentPool` resource attached to the referenced parent cluster.

## What Gets Created

When you deploy an AzureAksNodePool resource, Planton provisions:

- **Agent Pool** — a `containerservice.AgentPool` resource in the specified resource group and parent AKS cluster, configured with the chosen VM size, initial node count, OS type, and pool mode
- **Autoscaling** — cluster autoscaler configuration on the pool when `autoscaling` is provided, with configurable minimum and maximum node counts
- **Availability Zone Spread** — node distribution across the specified Azure availability zones for high availability
- **Spot VM Configuration** — when `spotEnabled` is true, the pool uses Spot priority with a Delete eviction policy and pay-up-to-regular-price bidding

## Prerequisites

- **Azure credentials** configured via environment variables or Planton provider config
- **An AKS cluster** that the node pool will be added to (can reference an AzureAksCluster resource)
- **An Azure Resource Group** where the parent cluster resides (can reference an AzureResourceGroup resource)

## Quick Start

Create a file `nodepool.yaml`:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureAksNodePool
metadata:
  name: worker-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureAksNodePool.worker-pool
spec:
  clusterName: my-aks-cluster
  vmSize: Standard_D4s_v3
  initialNodeCount: 3
  resourceGroup: my-rg
```

Deploy:

```shell
planton apply -f nodepool.yaml
```

This creates a 3-node User-mode Linux node pool using Standard_D4s_v3 VMs with no autoscaling and no availability zone pinning.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `clusterName` | `StringValueOrRef` | Name of the parent AKS cluster. Can reference an AzureAksCluster resource via `valueFrom`. | Required |
| `vmSize` | `string` | VM size (SKU) for nodes in this pool (e.g., `Standard_D4s_v3`). Determines the CPU and memory of each node. | Required |
| `initialNodeCount` | `int32` | Number of nodes to create initially. When autoscaling is off, this is the fixed node count. | Required, must be greater than 0 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group where the parent cluster resides. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `autoscaling` | `object` | none | Autoscaling configuration. When set, cluster autoscaler is enabled for this pool. Contains `minNodes` (uint32, >= 0) and `maxNodes` (uint32, > 0). |
| `availabilityZones` | `string[]` | `[]` | Zones to spread nodes across for high availability. Valid values: `"1"`, `"2"`, `"3"`. If specified, at least 2 zones are required. |
| `osType` | `enum` | `LINUX` | Operating system type. Values: `LINUX`, `WINDOWS`. Windows pools require a cluster with Windows support. |
| `mode` | `enum` | `USER` | Pool mode. `SYSTEM` pools host critical cluster components (CoreDNS, metrics-server), must be Linux, and cannot scale to zero. `USER` pools run application workloads, can be Linux or Windows, and can scale to zero. |
| `spotEnabled` | `bool` | `false` | Use Spot (preemptible) VMs to reduce cost. Spot pools use a Delete eviction policy and pay up to the regular on-demand price. Cannot be used with System mode pools. |

## Examples

### Basic User Pool

A simple application workload pool with a fixed node count:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureAksNodePool
metadata:
  name: app-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AzureAksNodePool.app-pool
spec:
  clusterName: my-aks-cluster
  vmSize: Standard_D2s_v3
  initialNodeCount: 2
  resourceGroup: dev-rg
```

### Autoscaling Pool with Availability Zones

A production pool that scales between 2 and 10 nodes across three availability zones:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureAksNodePool
metadata:
  name: prod-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureAksNodePool.prod-pool
spec:
  clusterName: prod-cluster
  vmSize: Standard_D4s_v3
  initialNodeCount: 3
  resourceGroup: prod-rg
  autoscaling:
    minNodes: 2
    maxNodes: 10
  availabilityZones:
    - "1"
    - "2"
    - "3"
```

### Spot Instance Pool for Batch Workloads

A cost-optimized pool using Spot VMs that can scale to zero when idle:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureAksNodePool
metadata:
  name: batch-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureAksNodePool.batch-pool
spec:
  clusterName: prod-cluster
  vmSize: Standard_D8s_v3
  initialNodeCount: 1
  resourceGroup: prod-rg
  spotEnabled: true
  autoscaling:
    minNodes: 0
    maxNodes: 20
  availabilityZones:
    - "1"
    - "2"
    - "3"
```

### System Pool

A System-mode pool to host critical cluster components. System pools must be Linux and cannot scale to zero:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureAksNodePool
metadata:
  name: system-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureAksNodePool.system-pool
spec:
  clusterName: prod-cluster
  vmSize: Standard_D2s_v3
  initialNodeCount: 3
  resourceGroup: prod-rg
  mode: SYSTEM
  autoscaling:
    minNodes: 2
    maxNodes: 5
  availabilityZones:
    - "1"
    - "2"
    - "3"
```

### Using Foreign Key References

Reference Planton-managed resources instead of hardcoding names:

```yaml
apiVersion: azure.planton.dev/v1
kind: AzureAksNodePool
metadata:
  name: ref-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AzureAksNodePool.ref-pool
spec:
  clusterName:
    valueFrom:
      kind: AzureAksCluster
      name: prod-cluster
      field: metadata.name
  vmSize: Standard_D4s_v3
  initialNodeCount: 3
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: prod-rg
      field: status.outputs.resource_group_name
  autoscaling:
    minNodes: 2
    maxNodes: 8
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `nodePoolName` | `string` | Name of the node pool in AKS. Typically matches the resource `metadata.name`. |
| `agentPoolResourceId` | `string` | Azure Resource Manager ID of the created Agent Pool resource. |
| `maxPodsPerNode` | `uint32` | Maximum number of pods that can run on each node of this pool. Determined by AKS based on network configuration and VM size. |

## Related Components

- [AzureAksCluster](/docs/catalog/azure/azureakscluster) — provides the parent cluster that this node pool is added to
- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) — provides the resource group where the parent cluster resides
- [AzureVpc](/docs/catalog/azure/azurevpc) — provides the virtual network and subnets used by the AKS cluster
- [AzureSubnet](/docs/catalog/azure/azuresubnet) — provides subnets that can be assigned to node pools for network isolation
