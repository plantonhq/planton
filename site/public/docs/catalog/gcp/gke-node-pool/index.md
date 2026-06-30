---
title: "GKE Node Pool"
description: "GKE Node Pool deployment documentation"
icon: "package"
order: 100
componentName: "gcpgkenodepool"
---

# GCP GKE Node Pool

Deploys a node pool into an existing GKE cluster on Google Cloud with configurable machine types, disk options, autoscaling, and Spot VM support. This component is a companion to [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — it manages the compute capacity for workloads while the cluster component manages the control plane.

## What Gets Created

When you deploy a GcpGkeNodePool resource, Planton provisions:

- **GKE Node Pool** — a `google_container_node_pool` resource with:
  - Node configuration (machine type, disk size and type, OS image)
  - Either a fixed node count or cluster autoscaler with min/max bounds and location policy
  - Node management with auto-upgrade and auto-repair enabled by default
  - Spot (preemptible) VM support
  - Upgrade settings with max surge of 2 and max unavailable of 1
  - Network tags following the `gke-<clusterName>` convention
  - OAuth scopes for Monitoring, Logging, and Cloud Storage (read-only)
  - Legacy metadata endpoints disabled
  - GCP resource labels and optional Kubernetes node labels merged onto every node

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing GKE cluster** — deployed via a [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) resource or created externally
- **IAM permissions** to create node pools in the target GCP project and GKE cluster

## Quick Start

Create a file `node-pool.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeNodePool
metadata:
  name: my-node-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpGkeNodePool.my-node-pool
spec:
  nodePoolName: default-pool
  clusterProjectId:
    value: my-gcp-project-123
  clusterName:
    value: dev-cluster
  clusterLocation:
    value: us-central1
  nodeCount: 3
```

Deploy:

```shell
planton apply -f node-pool.yaml
```

This creates a 3-node pool named `default-pool` using `e2-medium` instances with 100 GB `pd-standard` disks and Container-Optimized OS.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `nodePoolName` | `string` | Name of the node pool in the GKE cluster. | 1-40 chars, lowercase letters/numbers/hyphens, must start with a letter and end with a letter or number |
| `clusterProjectId` | `StringValueOrRef` | GCP project ID of the parent cluster. Can reference a GcpGkeCluster resource via `valueFrom` (resolves `spec.projectId`). | Required |
| `clusterName` | `StringValueOrRef` | Name of the parent GKE cluster. Can reference a GcpGkeCluster resource via `valueFrom` (resolves `metadata.name`). | Required |
| `clusterLocation` | `StringValueOrRef` | Region or zone of the parent GKE cluster (e.g., `us-central1`). Can reference a GcpGkeCluster resource via `valueFrom` (resolves `spec.location`). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `machineType` | `string` | `e2-medium` | Compute Engine machine type for node VMs (e.g., `n1-standard-4`, `e2-standard-8`). |
| `diskSizeGb` | `uint32` | `100` | Boot disk size in GB for each node. Minimum 10 GB. |
| `diskType` | `string` | `pd-standard` | Boot disk type: `pd-standard`, `pd-ssd`, or `pd-balanced`. |
| `imageType` | `string` | `COS_CONTAINERD` | Node OS image. Common values: `COS_CONTAINERD`, `COS`, `UBUNTU`, `UBUNTU_CONTAINERD`. |
| `serviceAccount` | `string` | GKE default | GCP service account email for nodes. If omitted, GKE assigns the default node service account. |
| `spot` | `bool` | `false` | Use Spot (preemptible) VMs. Reduces cost but nodes may be reclaimed at any time. |
| `nodeLabels` | `map<string, string>` | — | Kubernetes labels applied to every node in this pool. Merged with Planton-managed resource labels. |
| `nodeCount` | `uint32` | — | Fixed number of nodes (no autoscaling). Mutually exclusive with `autoscaling`. |
| `autoscaling.minNodes` | `uint32` | — | Minimum nodes per zone when autoscaling is enabled. Set to `0` for scale-to-zero. |
| `autoscaling.maxNodes` | `uint32` | — | Maximum nodes per zone when autoscaling is enabled. |
| `autoscaling.locationPolicy` | `string` | `BALANCED` | How the autoscaler distributes nodes across zones: `BALANCED` or `ANY`. |
| `management.disableAutoUpgrade` | `bool` | `false` | Set to `true` to prevent automatic Kubernetes version upgrades on nodes. |
| `management.disableAutoRepair` | `bool` | `false` | Set to `true` to prevent automatic repair of unhealthy nodes. |

One of `nodeCount` or `autoscaling` must be provided. If neither is set, the module defaults to a single node.

## Examples

### Fixed-Size Pool with Default Settings

A simple 3-node pool using all defaults — suitable for development:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeNodePool
metadata:
  name: dev-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpGkeNodePool.dev-pool
spec:
  nodePoolName: dev-pool
  clusterProjectId:
    value: my-dev-project
  clusterName:
    value: dev-cluster
  clusterLocation:
    value: us-central1
  nodeCount: 3
```

### Autoscaling Pool with Spot VMs

A cost-optimized pool that scales between 1 and 10 nodes using Spot VMs and SSD disks:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeNodePool
metadata:
  name: spot-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.GcpGkeNodePool.spot-pool
spec:
  nodePoolName: spot-workers
  clusterProjectId:
    value: my-staging-project
  clusterName:
    value: staging-cluster
  clusterLocation:
    value: us-east1
  machineType: n1-standard-4
  diskSizeGb: 200
  diskType: pd-ssd
  spot: true
  autoscaling:
    minNodes: 1
    maxNodes: 10
    locationPolicy: BALANCED
  nodeLabels:
    workload-type: batch
    cost-tier: spot
```

### Production Pool with Foreign Key References

A production-grade pool referencing a GcpGkeCluster resource, with auto-upgrade disabled for controlled rollouts:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeNodePool
metadata:
  name: prod-pool
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpGkeNodePool.prod-pool
spec:
  nodePoolName: prod-workers
  clusterProjectId:
    valueFrom:
      kind: GcpGkeCluster
      name: prod-cluster
      field: spec.projectId
  clusterName:
    valueFrom:
      kind: GcpGkeCluster
      name: prod-cluster
      field: metadata.name
  clusterLocation:
    valueFrom:
      kind: GcpGkeCluster
      name: prod-cluster
      field: spec.location
  machineType: e2-standard-8
  diskSizeGb: 500
  diskType: pd-balanced
  imageType: COS_CONTAINERD
  serviceAccount: gke-nodes@my-prod-project.iam.gserviceaccount.com
  autoscaling:
    minNodes: 3
    maxNodes: 20
    locationPolicy: BALANCED
  management:
    disableAutoUpgrade: true
    disableAutoRepair: false
  nodeLabels:
    environment: production
    team: platform
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `nodePoolName` | `string` | Name of the node pool in the GKE cluster |
| `instanceGroupUrls` | `repeated string` | URLs of the Compute Instance Group(s) backing this node pool (one per zone for regional clusters) |
| `minNodes` | `string` | Minimum node count — equals `nodeCount` for fixed-size pools, or `autoscaling.minNodes` for autoscaling pools |
| `maxNodes` | `string` | Maximum node count — equals `nodeCount` for fixed-size pools, or `autoscaling.maxNodes` for autoscaling pools |
| `currentNodeCount` | `string` | Current number of running nodes in the pool |

## Related Components

- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — provides the parent GKE cluster that this node pool attaches to
- [GcpVpc](/docs/catalog/gcp/vpc) — provides the VPC network used by the parent cluster
- [GcpSubnetwork](/docs/catalog/gcp/subnetwork) — provides the subnetwork with IP ranges for pods and services
- [GcpRouterNat](/docs/catalog/gcp/router-nat) — provides Cloud NAT for private node outbound internet access
