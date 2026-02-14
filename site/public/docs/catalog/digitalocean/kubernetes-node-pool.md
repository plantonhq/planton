---
title: "Kubernetes Node Pool"
description: "Kubernetes Node Pool deployment documentation"
icon: "package"
order: 100
componentName: "digitaloceankubernetesnodepool"
---

# DigitalOcean Kubernetes Node Pool

Adds an additional node pool to an existing DigitalOcean Kubernetes (DOKS) cluster. The component provisions a `digitalocean_kubernetes_node_pool` resource with a configurable Droplet size, fixed or auto-scaling node count, Kubernetes labels and taints for workload scheduling, and DigitalOcean tags for billing attribution. It is designed to be used alongside a DigitalOceanKubernetesCluster resource, referencing the parent cluster by name or cross-stack output.

## What Gets Created

When you deploy a DigitalOceanKubernetesNodePool resource, OpenMCF provisions:

- **Kubernetes Node Pool** -- a `digitalocean_kubernetes_node_pool` resource attached to the specified DOKS cluster, with the configured Droplet size and node count
- **Auto-Scaling Policy** -- configured only when `autoScale` is `true`, allows the cluster autoscaler to manage node count between `minNodes` and `maxNodes`
- **Kubernetes Labels** -- applied to every node in the pool, enabling pod scheduling via `nodeSelector` and node affinity rules
- **Kubernetes Taints** -- applied to every node in the pool, preventing pods without matching tolerations from being scheduled (or evicting them if already running)
- **DigitalOcean Tags** -- applied to the underlying Droplets for cost attribution and organizational grouping in DigitalOcean

## Prerequisites

- **DigitalOcean credentials** configured via environment variables or OpenMCF provider config
- **An existing DOKS cluster** -- either a cluster UUID or a reference to a DigitalOceanKubernetesCluster resource via `valueFrom`
- **A valid Droplet size slug** available in the cluster's region (e.g., `s-4vcpu-8gb`, `g-8vcpu-32gb`)

## Quick Start

Create a file `node-pool.yaml`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanKubernetesNodePool
metadata:
  name: worker-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanKubernetesNodePool.worker-pool
spec:
  nodePoolName: worker-pool
  cluster:
    value: "existing-cluster-uuid"
  size: s-4vcpu-8gb
  nodeCount: 3
```

Deploy:

```shell
openmcf apply -f node-pool.yaml
```

This creates a three-node pool of `s-4vcpu-8gb` Droplets in the specified DOKS cluster.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `nodePoolName` | `string` | Name of the node pool. Must be unique within the Kubernetes cluster. | Required |
| `cluster` | `StringValueOrRef` | Cluster UUID or reference to a DigitalOceanKubernetesCluster resource. Accepts a literal `value` or a `valueFrom` cross-stack reference. | Required |
| `size` | `string` | Droplet size slug for each node (e.g., `s-4vcpu-8gb`). Determines CPU and memory per node. | Required |
| `nodeCount` | `uint32` | Number of nodes to provision. Acts as initial desired count when auto-scaling is enabled. | Required, must be > 0 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `autoScale` | `bool` | `false` | Enables auto-scaling for the node pool. When `true`, `minNodes` and `maxNodes` must also be set. |
| `minNodes` | `uint32` | `0` | Minimum number of nodes when auto-scaling is enabled. Required if `autoScale` is `true`. |
| `maxNodes` | `uint32` | `0` | Maximum number of nodes when auto-scaling is enabled. Required if `autoScale` is `true`. |
| `labels` | `map<string,string>` | `{}` | Kubernetes labels applied to all nodes. Used for `nodeSelector` and node affinity scheduling. |
| `taints` | `Taint[]` | `[]` | Kubernetes taints applied to all nodes. Each taint has `key`, `value`, and `effect` (`NoSchedule`, `PreferNoSchedule`, or `NoExecute`). |
| `tags` | `string[]` | `[]` | DigitalOcean tags applied to the node pool Droplets for billing and organizational purposes. |

## Examples

### Fixed-Size Worker Pool

A simple, fixed-size pool for general-purpose workloads:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanKubernetesNodePool
metadata:
  name: web-workers
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.DigitalOceanKubernetesNodePool.web-workers
spec:
  nodePoolName: web-workers
  cluster:
    value: "doks-cluster-uuid"
  size: s-4vcpu-8gb
  nodeCount: 3
  tags:
    - web
    - development
```

### Auto-Scaling Pool with Labels

A production pool that scales between 2 and 10 nodes, with Kubernetes labels for workload targeting:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanKubernetesNodePool
metadata:
  name: api-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanKubernetesNodePool.api-pool
spec:
  nodePoolName: api-pool
  cluster:
    value: "prod-cluster-uuid"
  size: s-8vcpu-16gb
  nodeCount: 3
  autoScale: true
  minNodes: 2
  maxNodes: 10
  labels:
    workload: api
    env: production
  tags:
    - production
    - team-platform
```

### High-Memory Pool with Taints and Cluster Reference

A dedicated high-memory pool that uses taints to isolate workloads and references the parent cluster via `valueFrom`:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanKubernetesNodePool
metadata:
  name: ml-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.DigitalOceanKubernetesNodePool.ml-pool
spec:
  nodePoolName: ml-pool
  cluster:
    valueFrom:
      kind: DigitalOceanKubernetesCluster
      name: prod-cluster
      field: metadata.name
  size: g-8vcpu-32gb
  nodeCount: 2
  autoScale: true
  minNodes: 1
  maxNodes: 5
  labels:
    workload: ml-inference
    gpu: "true"
  taints:
    - key: dedicated
      value: ml
      effect: NoSchedule
    - key: workload-type
      value: gpu
      effect: PreferNoSchedule
  tags:
    - production
    - ml-team
```

Pods targeting this pool must include matching tolerations:

```yaml
tolerations:
  - key: dedicated
    value: ml
    effect: NoSchedule
  - key: workload-type
    value: gpu
    effect: PreferNoSchedule
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `node_pool_id` | `string` | UUID of the created node pool |
| `node_ids` | `string[]` | IDs of the individual Droplet nodes in the pool |

## Related Components

- [DigitalOceanKubernetesCluster](/docs/catalog/digitalocean/kubernetes-cluster) -- provides the parent DOKS cluster that this node pool is attached to
- [DigitalOceanVpc](/docs/catalog/digitalocean/vpc) -- provides the VPC in which the parent cluster and its node pools reside
- [DigitalOceanFirewall](/docs/catalog/digitalocean/firewall) -- controls network access to node pool Droplets
- [DigitalOceanLoadBalancer](/docs/catalog/digitalocean/load-balancer) -- provisions load balancers for exposing services running on pool nodes
