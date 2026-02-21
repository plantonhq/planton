---
title: "General-Purpose Auto-Scaling Node Pool"
description: "This preset creates a production node pool with auto-scaling enabled, spanning three availability zones with balanced distribution. It uses general-purpose ECS g7 instances with managed lifecycle..."
type: "preset"
rank: "01"
presetSlug: "01-general-purpose-autoscaling"
componentSlug: "kubernetesnodepool"
componentTitle: "KubernetesNodePool"
provider: "alicloud"
icon: "package"
order: 1
---

# General-Purpose Auto-Scaling Node Pool

This preset creates a production node pool with auto-scaling enabled, spanning three availability zones with balanced distribution. It uses general-purpose ECS g7 instances with managed lifecycle features (auto-repair, auto-upgrade) and SSH key-based access. The auto-scaler maintains between 2 and 20 nodes based on pending pod resource requests.

## When to Use

- Standard production workloads (web services, APIs, microservices)
- Clusters that need to scale horizontally with demand
- Multi-AZ deployments requiring even distribution across zones
- Teams that want ACK to automatically repair unhealthy nodes and upgrade kubelet

## Key Configuration Choices

- **Two instance types** (`ecs.g7.2xlarge`, `ecs.g7.xlarge`) -- Multiple types improve scheduling success when a specific type is capacity-constrained in an AZ. The g7 family provides a balanced CPU-to-memory ratio (1:4) suitable for most workloads.
- **Auto-scaling 2-20** (`scalingConfig`) -- Floor of 2 ensures at least one node per AZ pair for resilience. Ceiling of 20 prevents runaway scaling; adjust based on your workload profile and budget.
- **BALANCE multi-AZ policy** (`multiAzPolicy: BALANCE`) -- Distributes nodes evenly across the three AZs. This maximizes resilience against single-zone failures at the cost of slightly less cost optimization compared to COST_OPTIMIZED.
- **Managed lifecycle** (`management`) -- ACK automatically repairs nodes that report NotReady conditions, upgrades kubelet when the cluster version changes, and limits disruption to 2 nodes at a time during rolling operations.
- **PL1 ESSD system disk** (`performanceLevel: PL1`) -- Provides 50,000 IOPS and 350 MB/s throughput, sufficient for container image pulls and node-level I/O. PL0 is cheaper but may bottleneck during image-heavy deployments.
- **SSH key access** (`keyName`) -- Key-based authentication is more secure and auditable than password-based access for managed node pools.
- **AliyunLinux3** (`imageType: AliyunLinux3`) -- Alibaba Cloud's optimized Linux distribution with long-term support, tuned kernel parameters for containers, and faster boot times than CentOS or Ubuntu on ECS.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code matching the parent cluster | Your cluster's region |
| `<your-cluster-id>` | ACK cluster ID | `AliCloudKubernetesCluster` stack outputs |
| `<vswitch-id-zone-a>` | VSwitch in first AZ | `AliCloudVswitch` stack outputs |
| `<vswitch-id-zone-b>` | VSwitch in second AZ | `AliCloudVswitch` stack outputs |
| `<vswitch-id-zone-c>` | VSwitch in third AZ | `AliCloudVswitch` stack outputs |
| `<your-ssh-key-pair>` | ECS SSH key pair name | ECS console or your key management system |
| `<your-team>` | Team or business unit | Your organizational structure |
| `<your-cost-center>` | Cost center code | Your finance team |

## Related Presets

- **02-fixed-size-development** -- Use for development clusters where auto-scaling overhead is unnecessary
- **03-cost-optimized-spot** -- Use for batch processing or non-critical workloads where spot instance cost savings outweigh availability guarantees
