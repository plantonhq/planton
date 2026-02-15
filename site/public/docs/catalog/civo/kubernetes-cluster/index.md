---
title: "Kubernetes Cluster"
description: "Kubernetes Cluster deployment documentation"
icon: "package"
order: 100
componentName: "civokubernetescluster"
---

# Civo Kubernetes Cluster

Deploys a managed K3s-based Kubernetes cluster on Civo Cloud with configurable node pools, optional high-availability control plane, and automatic patch upgrades. The cluster is attached to an existing Civo network and exposes a kubeconfig for immediate access.

## What Gets Created

When you deploy a CivoKubernetesCluster resource, OpenMCF provisions:

- **Kubernetes Cluster (K3s)** — a `civo_kubernetes_cluster` resource in the specified region, attached to the given network, running the requested Kubernetes version
- **Default Node Pool** — a node pool with the configured instance size and node count, created as part of the cluster

## Prerequisites

- **Civo credentials** configured via environment variables or OpenMCF provider config
- **An existing Civo network** in the target region (can be created with CivoVpc)
- **A supported Kubernetes version** string (e.g., `1.28.2`) — check Civo's available versions for your region

## Quick Start

Create a file `civo-k8s.yaml`:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoKubernetesCluster
metadata:
  name: my-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoKubernetesCluster.my-cluster
spec:
  clusterName: my-cluster
  region: nyc1
  kubernetesVersion: "1.28.2"
  network:
    value: network-uuid-here
  defaultNodePool:
    size: g4s.kube.medium
    nodeCount: 2
```

Deploy:

```shell
openmcf apply -f civo-k8s.yaml
```

This creates a two-node K3s cluster in New York running Kubernetes 1.28.2 on `g4s.kube.medium` instances.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `clusterName` | `string` | Name of the Kubernetes cluster. Must be unique per account. Alphanumeric and hyphens recommended. | Required |
| `region` | `enum` | Civo region where the cluster is created. Valid values: `lon1`, `lon2`, `fra1`, `nyc1`, `phx1`, `mum1`. | Required |
| `kubernetesVersion` | `string` | Kubernetes version for the cluster (e.g., `1.28.2`). Must be a Civo-supported version. | Required |
| `network` | `StringValueOrRef` | Network ID where the cluster resides. Must be an existing network in the same region. Can reference a CivoVpc resource via `valueFrom`. | Required |
| `defaultNodePool.size` | `string` | Instance size for each node in the default pool (e.g., `g4s.kube.medium`). Defines CPU and memory allocation. | Required |
| `defaultNodePool.nodeCount` | `uint32` | Number of nodes in the default pool. | Required, must be > 0 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `highlyAvailable` | `bool` | `false` | When `true`, creates the cluster with multiple control-plane nodes for increased availability. |
| `autoUpgrade` | `bool` | `false` | When `true`, the cluster automatically upgrades to new Kubernetes patch releases as they become available. |
| `disableSurgeUpgrade` | `bool` | `false` | When `true`, disables surge upgrades. By default, upgrades may temporarily provision extra nodes to minimize downtime. |
| `tags` | `string[]` | `[]` | Tags to apply to the cluster for organizational purposes within Civo. |

## Examples

### Minimal Development Cluster

A single-node cluster for development and testing:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoKubernetesCluster
metadata:
  name: dev-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.CivoKubernetesCluster.dev-cluster
spec:
  clusterName: dev-cluster
  region: fra1
  kubernetesVersion: "1.28.2"
  network:
    value: network-uuid-here
  defaultNodePool:
    size: g4s.kube.small
    nodeCount: 1
```

### Production Cluster with HA and Auto-Upgrade

A multi-node cluster with high availability and automatic patch upgrades:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoKubernetesCluster
metadata:
  name: prod-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoKubernetesCluster.prod-cluster
spec:
  clusterName: prod-cluster
  region: nyc1
  kubernetesVersion: "1.28.2"
  network:
    value: network-uuid-here
  highlyAvailable: true
  autoUpgrade: true
  tags:
    - environment:production
    - team:platform
  defaultNodePool:
    size: g4s.kube.large
    nodeCount: 3
```

### Using Foreign Key References

Reference an OpenMCF-managed CivoVpc instead of hardcoding the network ID:

```yaml
apiVersion: civo.openmcf.org/v1
kind: CivoKubernetesCluster
metadata:
  name: ref-cluster
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.CivoKubernetesCluster.ref-cluster
spec:
  clusterName: ref-cluster
  region: lon1
  kubernetesVersion: "1.28.2"
  network:
    valueFrom:
      kind: CivoVpc
      name: my-network
      field: status.outputs.network_id
  highlyAvailable: true
  autoUpgrade: true
  disableSurgeUpgrade: false
  tags:
    - environment:production
    - managed-by:openmcf
  defaultNodePool:
    size: g4s.kube.xlarge
    nodeCount: 5
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | Unique identifier of the created Kubernetes cluster |
| `kubeconfig_b64` | `string` | Base64-encoded kubeconfig for accessing the cluster |
| `api_server_endpoint` | `string` | Endpoint URL of the Kubernetes API server |
| `created_at_rfc3339` | `string` | Timestamp when the cluster was created, in RFC 3339 format |

## Related Components

- [CivoVpc](/docs/catalog/civo/vpc) — provides the network for cluster placement
- [CivoKubernetesNodePool](/docs/catalog/civo/kubernetes-node-pool) — adds additional node pools to the cluster
- [CivoFirewall](/docs/catalog/civo/firewall) — controls network access to the cluster
- [CivoVolume](/docs/catalog/civo/volume) — provides persistent block storage for cluster workloads
