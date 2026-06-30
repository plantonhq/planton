# ACK Kubernetes Node Pool: From Manual Scaling to Declarative Worker Management

## Introduction

In Alibaba Cloud's managed Kubernetes architecture, the control plane and data plane have completely separate lifecycles. The control plane (API server, etcd, scheduler, controller manager) is managed by ACK as a service. The data plane — the worker nodes that actually run application workloads — is managed through **node pools**. Each node pool is a group of ECS instances that share the same instance type, disk configuration, scaling policy, and Kubernetes scheduling properties (labels, taints).

This separation is not just organizational. It is the foundation of production Kubernetes operations: different workloads need different node types (CPU-optimized for web servers, GPU for ML inference, memory-optimized for caches), different scaling policies (fixed-size for stateful, auto-scaling for stateless), and different cost models (on-demand for critical workloads, spot for batch jobs). Node pools are the mechanism that makes this heterogeneity manageable.

This document examines node pool deployment methodologies, the critical configuration decisions, and how Planton abstracts the complexity into a declarative API.

## The Node Pool Deployment Landscape

### Level 0: Manual Provisioning via ACK Console (The Anti-Pattern)

The ACK console provides a wizard for adding node pools to existing clusters. The wizard walks through instance type selection, disk configuration, scaling settings, and Kubernetes configuration.

**Common Mistakes**:

1. **Single Instance Type**: The wizard allows selecting just one instance type. In a region with limited capacity or during spot market volatility, the auto-scaler cannot provision nodes when that single type is unavailable. Production node pools should specify **multiple instance types** for availability.

2. **Undersized System Disks**: The default 40 GiB system disk fills quickly with container images, especially when running many different workloads. Container image layers accumulate on the system disk; production nodes typically need 120 GiB or more.

3. **Missing Security Groups**: If no security group is specified, the node pool inherits the cluster's default security group. This is often too permissive or too restrictive for the specific workload's requirements. Security groups are immutable after node pool creation.

4. **No Auto-Scaling Configuration**: Node pools created without auto-scaling are fixed-size. When traffic spikes, there are no nodes to schedule pending pods on. Adding auto-scaling after creation requires modifying the scaling group through the ESS (Elastic Scaling Service) console separately.

**Verdict**: Acceptable for adding ad-hoc node pools in development. Unacceptable for production where node pool configuration must be version-controlled and reproducible.

### Level 1: Scripted Provisioning with Alibaba Cloud CLI

The `aliyun cs` CLI can create node pools via the `CreateClusterNodePool` API:

```bash
aliyun cs CreateClusterNodePool --ClusterId c-xxx --body '{
  "nodepool_info": {"name": "compute-pool"},
  "scaling_group": {
    "instance_types": ["ecs.g7.xlarge", "ecs.g7.2xlarge"],
    "vswitch_ids": ["vsw-aaa", "vsw-bbb"],
    "system_disk_category": "cloud_essd",
    "system_disk_size": 120,
    "desired_size": 3
  },
  "kubernetes_config": {
    "labels": [{"key": "workload-type", "value": "compute"}]
  }
}'
```

The API payload is deeply nested and the naming conventions differ from Terraform/Pulumi (`nodepool_info` vs. `node_pool_name`, `scaling_group` vs. flat fields). This impedance mismatch makes it error-prone to translate between different management tools.

**Verdict**: Suitable for one-off automation. Not suitable for managing node pool lifecycle (scaling policy changes, image upgrades, spot configuration changes).

### Level 2: Infrastructure as Code (Terraform, Pulumi)

IaC tools provide the right abstraction for node pool management: declarative desired state with drift detection and incremental updates.

#### Terraform

The `alicloud_cs_kubernetes_node_pool` resource provides comprehensive coverage:

```hcl
resource "alicloud_cs_kubernetes_node_pool" "compute" {
  cluster_id     = alicloud_cs_managed_kubernetes.cluster.id
  node_pool_name = "compute-pool"
  vswitch_ids    = [alicloud_vswitch.a.id, alicloud_vswitch.b.id]
  instance_types = ["ecs.g7.xlarge", "ecs.g7.2xlarge"]
  desired_size   = 3

  system_disk_category = "cloud_essd"
  system_disk_size     = 120

  labels {
    key   = "workload-type"
    value = "compute"
  }

  scaling_config {
    enable   = true
    min_size = 2
    max_size = 10
  }

  management {
    enable       = true
    auto_repair  = true
    auto_upgrade = true
  }
}
```

**Immutability Constraints**: `cluster_id` and `security_group_ids` are ForceNew — changing them destroys and recreates the node pool (and all nodes in it). This is a critical consideration for production operations.

**Provider Quirks**: The `desired_size` field is a string in the Terraform schema despite representing an integer count. The `name` field is deprecated since provider v1.219.0 in favor of `node_pool_name`. The `node_count` field is deprecated since v1.158.0 in favor of `desired_size`.

#### Pulumi

Pulumi's `cs.NodePool` maps to the same underlying resource:

```go
nodePool, err := cs.NewNodePool(ctx, "compute-pool", &cs.NodePoolArgs{
    ClusterId:          cluster.ID(),
    NodePoolName:       pulumi.String("compute-pool"),
    VswitchIds:         pulumi.StringArray{vswA.ID(), vswB.ID()},
    InstanceTypes:      pulumi.StringArray{pulumi.String("ecs.g7.xlarge")},
    DesiredSize:        pulumi.String("3"),
    SystemDiskCategory: pulumi.String("cloud_essd"),
    SystemDiskSize:     pulumi.Int(120),
})
```

## Production Node Pool Architecture

### Instance Type Strategy

**Multiple Instance Types**: Always specify at least two instance types per node pool. This is not about workload requirements — it's about **availability**. The ECS fleet in any single instance type can be exhausted in a specific AZ. Multiple types give the auto-scaler fallback options.

**Instance Family Selection**:
- `ecs.g7.*` (General Purpose): Balanced CPU/memory for most web workloads
- `ecs.c7.*` (Compute Optimized): High CPU-to-memory ratio for compute-intensive tasks
- `ecs.r7.*` (Memory Optimized): High memory-to-CPU ratio for caches and in-memory databases
- `ecs.gn7i.*` (GPU): For ML training and inference workloads

### Disk Configuration

**System Disk**: The OS disk holds the operating system, container runtime, and container image layers. Production recommendations:
- **Category**: `cloud_essd` for consistent IOPS; `cloud_ssd` as a budget alternative
- **Size**: 120 GiB minimum; 200+ GiB for clusters running many distinct container images
- **Performance Level**: `PL1` (default) is sufficient for most workloads; `PL2`/`PL3` for I/O-intensive workloads
- **Encryption**: Enable for compliance requirements

**Data Disks**: Additional disks for application data, local caching, or emptyDir volumes. Each node can have multiple data disks with independent size, category, and encryption settings.

### Auto-Scaling

The cluster auto-scaler watches for pods that cannot be scheduled due to insufficient resources. When pending pods exist, it provisions new nodes up to `max_size`. When nodes are underutilized, it drains and terminates them down to `min_size`.

**Scaling Config Fields**:
- `enable`: Toggle auto-scaling on/off
- `min_size`: Floor — the auto-scaler never removes nodes below this count
- `max_size`: Ceiling — the auto-scaler never adds nodes above this count
- `type`: Instance classification: `cpu`, `gpu`, `gpushare`, `spot`

**Multi-AZ Policy**: Controls how the auto-scaler distributes nodes across availability zones:
- `PRIORITY`: Fill the first AZ, then overflow to the next
- `BALANCE`: Evenly distribute across AZs (best for HA)
- `COST_OPTIMIZED`: Prefer the cheapest AZ

### Managed Node Pool Lifecycle

When `management.enable` is true, ACK takes responsibility for node health:
- **Auto-Repair**: Automatically replaces nodes that fail Kubernetes health checks (NotReady, DiskPressure, etc.)
- **Auto-Upgrade**: Automatically upgrades kubelet when the cluster Kubernetes version is upgraded
- **Max Unavailable**: Controls the blast radius of managed operations — how many nodes can be simultaneously unavailable during repair/upgrade cycles

### Spot Instances for Cost Optimization

Spot instances can reduce compute costs by up to 90% but can be reclaimed by Alibaba Cloud when capacity is needed:

- **SpotAsPriceGo**: Pay the current market price; cheapest but highest reclamation risk
- **SpotWithPriceLimit**: Set per-instance-type price caps; instance is not created if market price exceeds the limit
- **NoSpot**: On-demand only (default)

**Production Pattern**: Use spot instances for stateless, fault-tolerant workloads (web servers, batch processors) with on-demand node pools for stateful workloads (databases, message queues). Combine with pod disruption budgets and taints to ensure graceful handling of spot reclamation.

## Production Best Practices and Anti-Patterns

| Category | Best Practice | Common Anti-Pattern | Impact |
|----------|--------------|---------------------|--------|
| **Availability** | Specify **multiple instance types** per pool | Single instance type | Auto-scaler blocked when type unavailable |
| **Disks** | System disk **120 GiB+ with cloud_essd** | Default 40 GiB cloud_efficiency | Image pulls fail when disk fills |
| **Security** | Assign **dedicated security groups** per pool | Inheriting cluster default | Overly permissive or restrictive access |
| **Scaling** | Configure **auto-scaling with min/max bounds** | Fixed-size pool with no auto-scaling | No capacity during traffic spikes |
| **Scheduling** | Use **labels and taints** for workload isolation | All workloads on all nodes | Noisy neighbor problems |
| **Cost** | Use **spot instances** for stateless workloads | On-demand for everything | 2-10x higher compute costs |
| **Lifecycle** | Enable **management** with auto-repair | Manual node health monitoring | Unhealthy nodes serve traffic |
| **Multi-AZ** | Use **BALANCE** policy for HA workloads | PRIORITY policy for critical services | Unbalanced AZ distribution |

## What Planton Supports

### Design Philosophy: 80/20 API Structure

The AliCloudKubernetesNodePool spec covers the production-critical node pool settings while excluding niche features (kubelet_configuration, instance_patterns, tee_config, eflo_node_group) that affect less than 5% of use cases.

**Core Fields**: `cluster_id`, `name`, `vswitch_ids`, `instance_types`, `desired_size`
**Disk**: `system_disk` (category, size, performance level, encryption), `data_disks`
**Networking**: `security_group_ids`, `internet_max_bandwidth_out`
**Kubernetes**: `labels`, `taints`, `cpu_policy`, `runtime_name`, `unschedulable`
**Scaling**: `scaling_config` (enable, min/max, type), `multi_az_policy`
**Management**: auto-repair, auto-upgrade, max_unavailable
**Spot**: `spot_strategy`, `spot_price_limits`
**Billing**: `instance_charge_type`, `period`, auto-renew settings

### Foreign Key References

- `cluster_id` → `AliCloudKubernetesCluster.status.outputs.cluster_id`
- `vswitch_ids` → `AliCloudVswitch.status.outputs.vswitch_id`
- `security_group_ids` → `AliCloudSecurityGroup.status.outputs.security_group_id`

### Implementation Landscape

**Pulumi Module**: A single `cs.NodePool` resource. The `locals.go` file resolves `StringValueOrRef` fields (cluster ID, VSwitch IDs, security group IDs) and merges tags. Helper functions provide defaults for image type (`AliyunLinux3`), system disk (`cloud_essd`, 120 GiB), instance charge type (`PostPaid`), and cloud monitor (`true`).

**Terraform Module**: A single `alicloud_cs_kubernetes_node_pool` resource with dynamic blocks for data disks, labels, taints, scaling config, management, and spot price limits.

## Conclusion

ACK node pools are the workhorse of Kubernetes data plane management. Getting their configuration right — instance types, disk sizing, auto-scaling bounds, spot strategy, and managed lifecycle settings — directly determines cluster availability, cost efficiency, and operational overhead.

Planton's AliCloudKubernetesNodePool component encodes these decisions into a declarative API with sensible defaults: `cloud_essd` disks at 120 GiB, `AliyunLinux3` images, `PostPaid` billing, and CloudMonitor enabled. The foreign key reference to `AliCloudKubernetesCluster` ensures the cluster-nodepool relationship is explicit and type-safe.
