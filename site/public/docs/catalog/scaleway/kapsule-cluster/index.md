---
title: "Kapsule Cluster"
description: "Kapsule Cluster deployment documentation"
icon: "package"
order: 100
componentName: "scalewaykapsulecluster"
---

# Scaleway Kapsule Cluster

Deploys a Scaleway Kapsule managed Kubernetes cluster with an embedded default node pool, Private Network attachment, optional auto-upgrade, and cluster-level autoscaler configuration. This is a composite resource — a single manifest produces a working cluster with compute capacity ready for workloads.

## What Gets Created

When you deploy a ScalewayKapsuleCluster resource, Planton provisions:

- **Kapsule Cluster** — a `kubernetes.Cluster` resource providing a fully managed Kubernetes control plane (API server, etcd, scheduler, controller-manager) in the specified region, attached to a Private Network
- **Default Node Pool** — a `kubernetes.Pool` resource created alongside the cluster with the specified instance type, size, and optional autoscaling, autohealing, and upgrade policy configuration

## Prerequisites

- **Scaleway credentials** configured via environment variables or Planton provider config
- **A Private Network** in the target region — required for all Kapsule clusters. Can be created via a ScalewayPrivateNetwork resource.
- **A valid Kubernetes version** available in the target region (e.g., `"1.32"` or `"1.32.3"`)

## Quick Start

Create a file `kapsule-cluster.yaml`:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayKapsuleCluster
metadata:
  name: my-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayKapsuleCluster.my-cluster
spec:
  region: fr-par
  kubernetesVersion: "1.32"
  cni: cilium
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  defaultNodePool:
    nodeType: DEV1-M
    size: 2
```

Deploy:

```shell
planton apply -f kapsule-cluster.yaml
```

This creates a Kapsule cluster with Cilium CNI in `fr-par`, attached to the specified Private Network, with a two-node default pool using `DEV1-M` instances.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Scaleway region for the cluster (e.g., `"fr-par"`, `"nl-ams"`, `"pl-waw"`). Cannot be changed after creation. | Required |
| `kubernetesVersion` | `string` | Kubernetes version. Can be minor (`"1.32"`) or patch (`"1.32.3"`). Use minor version when auto-upgrade is enabled. | Required |
| `cni` | `string` | Container Network Interface plugin. Cannot be changed after creation. Recommended: `"cilium"`. | Required |
| `privateNetworkId` | `StringValueOrRef` | Private Network UUID for cluster networking. Can reference a ScalewayPrivateNetwork resource via `valueFrom`. Cannot be changed after creation. | Required |
| `defaultNodePool.nodeType` | `string` | Instance type for worker nodes (e.g., `"DEV1-M"`, `"GP1-XS"`, `"PRO2-S"`). Cannot be changed after creation. | Required |
| `defaultNodePool.size` | `int` | Number of nodes in the default pool. When autoscaling is enabled, this is the initial size. | Required, minimum 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | `string` | `"kapsule"` | Cluster type. Options: `"kapsule"` (shared control plane), `"kapsule-dedicated-4"`, `"kapsule-dedicated-8"`, `"kapsule-dedicated-16"` (dedicated control planes with node limits). |
| `description` | `string` | `""` | Human-readable description shown in the Scaleway console. |
| `deleteAdditionalResources` | `bool` | `true` | When `true`, Scaleway cleans up LBs, volumes, and routes created by Kubernetes on cluster deletion. Set to `false` to preserve data volumes. |
| `autoUpgrade.enable` | `bool` | — | Enables automatic Kubernetes patch version upgrades during the maintenance window. |
| `autoUpgrade.maintenanceWindowStartHour` | `int` | — | UTC hour (0–23) when the maintenance window starts. Required when `autoUpgrade.enable` is `true`. |
| `autoUpgrade.maintenanceWindowDay` | `string` | — | Day of the week for maintenance. Options: `"monday"` through `"sunday"`, or `"any"`. Required when `autoUpgrade.enable` is `true`. |
| `autoscalerConfig.disableScaleDown` | `bool` | `false` | When `true`, the autoscaler only scales up, never removes nodes. |
| `autoscalerConfig.scaleDownDelayAfterAdd` | `string` | `"10m"` | Duration to wait after a scale-up before considering scale-down. |
| `autoscalerConfig.scaleDownUnneededTime` | `string` | `"10m"` | Duration a node must be underutilized before becoming a scale-down candidate. |
| `autoscalerConfig.estimator` | `string` | `"binpacking"` | Resource estimation algorithm for scheduling decisions. |
| `autoscalerConfig.expander` | `string` | `"random"` | Node group expansion strategy. Options: `"random"`, `"most-pods"`, `"least-waste"`, `"priority"`. |
| `autoscalerConfig.scaleDownUtilizationThreshold` | `double` | `0.5` | Utilization threshold (0.0–1.0) below which a node is a scale-down candidate. |
| `autoscalerConfig.maxGracefulTerminationSec` | `int` | `600` | Maximum seconds to wait for pod termination during scale-down. |
| `autoscalerConfig.ignoreDaemonsetsUtilization` | `bool` | `false` | When `true`, DaemonSet resource usage is excluded from utilization calculations. |
| `autoscalerConfig.balanceSimilarNodeGroups` | `bool` | `false` | When `true`, the autoscaler balances node counts across similar groups. |
| `autoscalerConfig.expendablePodsPriorityCutoff` | `int` | `-10` | Pods with priority below this value won't prevent scale-down. |
| `featureGates` | `string[]` | `[]` | Kubernetes feature gates to enable (e.g., `["GracefulNodeShutdown"]`). |
| `admissionPlugins` | `string[]` | `[]` | Additional Kubernetes admission plugins to enable (e.g., `["AlwaysPullImages"]`). |
| `podCidr` | `string` | `"100.64.0.0/15"` | Pod network CIDR. Cannot be changed after creation. |
| `serviceCidr` | `string` | `"10.32.0.0/20"` | Service network CIDR. Cannot be changed after creation. |
| `defaultNodePool.name` | `string` | `"{cluster-name}-default"` | Pool name. Must be unique within the cluster. Cannot be changed after creation. |
| `defaultNodePool.autoScale` | `bool` | `false` | Enables the cluster autoscaler for this pool. Requires `minSize` and `maxSize`. |
| `defaultNodePool.minSize` | `int` | — | Minimum node count when autoscaling is enabled. |
| `defaultNodePool.maxSize` | `int` | — | Maximum node count when autoscaling is enabled. |
| `defaultNodePool.autohealing` | `bool` | `false` | When `true`, Scaleway automatically replaces unhealthy nodes. |
| `defaultNodePool.containerRuntime` | `string` | `"containerd"` | Container runtime for pool nodes. Cannot be changed after creation. |
| `defaultNodePool.rootVolumeType` | `string` | — | Root volume storage type. Depends on instance type and zone. Cannot be changed after creation. |
| `defaultNodePool.rootVolumeSizeInGb` | `int` | — | Root volume size in GB. If omitted, uses the instance type default. Cannot be changed after creation. |
| `defaultNodePool.publicIpDisabled` | `bool` | `false` | When `true`, nodes have only private IPs. Requires a Public Gateway or NAT for external access. |
| `defaultNodePool.upgradePolicy.maxSurge` | `int` | `0` | Maximum extra nodes created during a rolling upgrade. |
| `defaultNodePool.upgradePolicy.maxUnavailable` | `int` | `1` | Maximum nodes unavailable simultaneously during a rolling upgrade. |

## Examples

### Development Cluster

A minimal cluster for development with a small node pool:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayKapsuleCluster
metadata:
  name: dev-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.ScalewayKapsuleCluster.dev-cluster
spec:
  region: fr-par
  kubernetesVersion: "1.32"
  cni: cilium
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  defaultNodePool:
    nodeType: DEV1-M
    size: 2
    autohealing: true
    containerRuntime: containerd
```

### Production Cluster with Autoscaling

A production-ready cluster with autoscaling, auto-upgrade, private nodes, and a dedicated control plane:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayKapsuleCluster
metadata:
  name: prod-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.ScalewayKapsuleCluster.prod-cluster
spec:
  region: fr-par
  kubernetesVersion: "1.32"
  cni: cilium
  type: kapsule-dedicated-8
  description: Production Kubernetes cluster
  deleteAdditionalResources: true
  privateNetworkId: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
  autoUpgrade:
    enable: true
    maintenanceWindowStartHour: 3
    maintenanceWindowDay: sunday
  autoscalerConfig:
    scaleDownDelayAfterAdd: "15m"
    scaleDownUnneededTime: "15m"
    expander: least-waste
    scaleDownUtilizationThreshold: 0.6
  defaultNodePool:
    name: system
    nodeType: PRO2-M
    size: 3
    autoScale: true
    minSize: 3
    maxSize: 10
    autohealing: true
    publicIpDisabled: true
    containerRuntime: containerd
    upgradePolicy:
      maxSurge: 1
      maxUnavailable: 0
```

### Cluster with Private Network Reference

Reference an Planton-managed Private Network instead of hardcoding the UUID:

```yaml
apiVersion: scaleway.planton.dev/v1
kind: ScalewayKapsuleCluster
metadata:
  name: ref-cluster
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.ScalewayKapsuleCluster.ref-cluster
spec:
  region: nl-ams
  kubernetesVersion: "1.31"
  cni: calico
  privateNetworkId:
    valueFrom:
      kind: ScalewayPrivateNetwork
      name: app-network
      fieldPath: status.outputs.private_network_id
  defaultNodePool:
    nodeType: GP1-XS
    size: 3
    autoScale: true
    minSize: 2
    maxSize: 6
    autohealing: true
    publicIpDisabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `cluster_id` | `string` | Regional ID of the created Kapsule cluster. Referenced by ScalewayKapsulePool resources. |
| `kubeconfig` | `string` | Raw kubeconfig file content for connecting to the cluster. Contains API server URL, CA certificate, and authentication token. Sensitive. |
| `apiserver_url` | `string` | URL of the Kubernetes API server (e.g., `https://<uuid>.api.k8s.fr-par.scw.cloud:6443`). |
| `cluster_ca_certificate` | `string` | Base64-encoded CA certificate of the Kubernetes API server. Used to configure Kubernetes providers in IaC tools. |
| `wildcard_dns` | `string` | DNS wildcard for ready nodes in the cluster. Can be used for DNS-based service discovery. |
| `default_pool_id` | `string` | Regional ID of the default node pool. |

## Related Components

- [ScalewayPrivateNetwork](/docs/catalog/scaleway/private-network) — provides the Private Network required by the cluster
- [ScalewayKapsulePool](/docs/catalog/scaleway/kapsule-pool) — adds additional node pools with different instance types, labels, or taints
- [ScalewayLoadBalancer](/docs/catalog/scaleway/load-balancer) — provisions load balancers for exposing services running on the cluster
