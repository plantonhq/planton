---
title: "Kapsule Pool"
description: "Kapsule Pool deployment documentation"
icon: "package"
order: 100
componentName: "scalewaykapsulepool"
---

# Scaleway Kapsule Pool

Deploys an additional node pool into an existing Scaleway Kapsule Kubernetes cluster. This is a standalone resource that creates a single `scaleway_k8s_pool` and supports autoscaling, autohealing, Kubernetes labels, taints, and custom upgrade policies.

## What Gets Created

When you deploy a ScalewayKapsulePool resource, OpenMCF provisions:

- **Kapsule Node Pool** — a `kubernetes.Pool` resource providing a group of identically configured worker nodes (same instance type, root volume, container runtime) in the referenced Kapsule cluster. Kubernetes labels and taints are applied via Scaleway's Cloud Controller Manager tag convention.

## Prerequisites

- **Scaleway credentials** configured via environment variables or OpenMCF provider config
- **An existing Kapsule cluster** — the pool attaches to a cluster referenced by `clusterId`. Can be created via a ScalewayKapsuleCluster resource.
- **A valid instance type** eligible for Kapsule (instances with insufficient memory such as DEV1-S, PLAY2-PICO, STARDUST are not eligible)

## Quick Start

Create a file `kapsule-pool.yaml`:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: my-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayKapsulePool.my-pool
spec:
  region: fr-par
  clusterId:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: my-cluster
      fieldPath: status.outputs.cluster_id
  nodeType: DEV1-M
  size: 2
```

Deploy:

```shell
openmcf apply -f kapsule-pool.yaml
```

This creates a two-node pool of `DEV1-M` instances in the `fr-par` region, attached to the `my-cluster` Kapsule cluster.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Scaleway region where the pool will be created. Must match the parent cluster's region. Cannot be changed after creation. | Required |
| `clusterId` | `StringValueOrRef` | Reference to the Kapsule cluster. Can be a literal cluster ID or a `valueFrom` reference to a ScalewayKapsuleCluster resource's output. Cannot be changed after creation. | Required |
| `nodeType` | `string` | Instance type for worker nodes (e.g., `"DEV1-M"`, `"GP1-XS"`, `"PRO2-S"`). Cannot be changed after creation. | Required |
| `size` | `int` | Number of nodes in the pool. When autoscaling is enabled, this is the initial size. | Required, minimum 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `autoScale` | `bool` | `false` | Enables the cluster autoscaler for this pool. Requires `minSize` and `maxSize`. |
| `minSize` | `int` | — | Minimum node count when autoscaling is enabled. The autoscaler will not scale below this number. |
| `maxSize` | `int` | — | Maximum node count when autoscaling is enabled. Controls cost ceiling. |
| `autohealing` | `bool` | `false` | When `true`, Scaleway automatically detects and replaces unhealthy nodes. Recommended for production. |
| `containerRuntime` | `string` | `"containerd"` | Container runtime for pool nodes. Cannot be changed after creation. |
| `rootVolumeType` | `string` | — | Root volume storage type. Depends on instance type and availability zone. Cannot be changed after creation. |
| `rootVolumeSizeInGb` | `int` | — | Root volume size in GB. If omitted, uses the instance type default. Cannot be changed after creation. |
| `publicIpDisabled` | `bool` | `false` | When `true`, nodes have only private IPs. Requires a Public Gateway or NAT for external access. Cannot be changed after creation. |
| `zone` | `string` | — | Zone within the region for node placement (e.g., `"fr-par-1"`). If omitted, Scaleway chooses automatically. Cannot be changed after creation. |
| `placementGroupId` | `string` | — | Placement group UUID for anti-affinity scheduling. Spreads nodes across different hypervisors. Cannot be changed after creation. |
| `kubernetesLabels` | `map<string, string>` | `{}` | Key-value pairs applied as Kubernetes node labels via the Scaleway CCM tag convention. Used for `nodeSelector` and affinity-based scheduling. |
| `taints` | `ScalewayKapsulePoolTaint[]` | `[]` | Kubernetes taints applied to all nodes. Each taint has `key` (required), `value`, and `effect` (required). Effect must be `"NoSchedule"`, `"PreferNoSchedule"`, or `"NoExecute"`. |
| `upgradePolicy.maxSurge` | `int` | `0` | Maximum extra nodes created during a rolling upgrade. Higher values speed up upgrades but increase cost temporarily. |
| `upgradePolicy.maxUnavailable` | `int` | `1` | Maximum nodes unavailable simultaneously during a rolling upgrade. |
| `kubeletArgs` | `map<string, string>` | `{}` | Custom kubelet arguments for pool nodes. Use with caution — incorrect values can prevent nodes from joining the cluster. |

## Examples

### Development Pool

A minimal additional pool for development workloads:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: dev-workers
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayKapsulePool.dev-workers
spec:
  region: fr-par
  clusterId:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: dev-cluster
      fieldPath: status.outputs.cluster_id
  nodeType: DEV1-M
  size: 2
  autohealing: true
  containerRuntime: containerd
```

### Production Pool with Autoscaling and Labels

A production pool with autoscaling, private nodes, Kubernetes labels for workload scheduling, and a safe upgrade policy:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: app-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayKapsulePool.app-pool
spec:
  region: fr-par
  clusterId:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: prod-cluster
      fieldPath: status.outputs.cluster_id
  nodeType: PRO2-M
  size: 3
  autoScale: true
  minSize: 3
  maxSize: 10
  autohealing: true
  publicIpDisabled: true
  containerRuntime: containerd
  rootVolumeSizeInGb: 100
  kubernetesLabels:
    workload: application
    tier: frontend
  upgradePolicy:
    maxSurge: 1
    maxUnavailable: 0
```

### GPU Pool with Taints and Zone Pinning

A dedicated pool for GPU workloads using taints to prevent non-GPU pods from being scheduled, pinned to a specific zone:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayKapsulePool
metadata:
  name: gpu-pool
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayKapsulePool.gpu-pool
spec:
  region: fr-par
  zone: fr-par-2
  clusterId:
    valueFrom:
      kind: ScalewayKapsuleCluster
      name: prod-cluster
      fieldPath: status.outputs.cluster_id
  nodeType: GP1-S
  size: 2
  autoScale: true
  minSize: 1
  maxSize: 4
  autohealing: true
  publicIpDisabled: true
  containerRuntime: containerd
  kubernetesLabels:
    accelerator: gpu
    team: ml
  taints:
    - key: nvidia.com/gpu
      value: "true"
      effect: NoSchedule
  kubeletArgs:
    maxPods: "150"
  upgradePolicy:
    maxSurge: 1
    maxUnavailable: 1
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `pool_id` | `string` | Regional ID of the created node pool (e.g., `fr-par/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`). |
| `pool_version` | `string` | Actual Kubernetes version running on pool nodes. May differ from the cluster version during rolling upgrades. |
| `current_size` | `int` | Actual number of nodes currently in the pool. When autoscaling is enabled, this may differ from the `size` field in the spec. |

## Related Components

- [ScalewayKapsuleCluster](/docs/catalog/scaleway/kapsule-cluster) — provides the Kapsule cluster that this pool attaches to
- [ScalewayPrivateNetwork](/docs/catalog/scaleway/private-network) — provides the Private Network used by the parent cluster
- [ScalewayPublicGateway](/docs/catalog/scaleway/public-gateway) — required when `publicIpDisabled` is `true` to give nodes outbound internet access
