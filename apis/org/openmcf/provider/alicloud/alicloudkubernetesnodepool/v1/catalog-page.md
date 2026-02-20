# AlicloudKubernetesNodePool

Deploy and manage worker node pools in an Alibaba Cloud ACK Managed Kubernetes cluster.

## What is it?

An ACK node pool is a group of ECS worker nodes within a Kubernetes cluster that share
the same instance configuration, scaling policy, and node properties. Node pools decouple
worker node management from the cluster control plane, enabling independent scaling and
heterogeneous workload support.

## When to use it

- **Workload isolation**: Separate GPU nodes from CPU nodes, or production from batch workloads
- **Cost optimization**: Run non-critical workloads on spot instances while keeping critical
  services on on-demand nodes
- **Multi-AZ resilience**: Spread nodes across availability zones with BALANCE or PRIORITY policies
- **Auto-scaling**: Let the cluster auto-scaler add and remove nodes based on pod resource demands
- **Managed lifecycle**: Enable auto-repair and auto-upgrade to reduce operational burden

## Key features

- Multiple instance types per pool for availability during spot reclamation
- ESSD system and data disks with encryption and performance level control
- Kubernetes labels and taints for fine-grained pod scheduling
- Integrated auto-scaling with configurable min/max bounds
- Managed node pool support with auto-repair, auto-upgrade, and vulnerability patching
- Spot instance strategies (market price or price-limited) for cost savings up to 90%

## Dependencies

| Dependency | Required | Description |
|-----------|----------|-------------|
| AlicloudKubernetesCluster | Yes | Parent cluster that this node pool belongs to |
| AlicloudVswitch | Yes | VSwitches for node placement (1-5, multi-AZ recommended) |
| AlicloudSecurityGroup | No | Custom security groups (defaults to cluster's security group) |

## Outputs

| Output | Description |
|--------|-------------|
| `nodePoolId` | Node pool ID assigned by Alibaba Cloud |
| `scalingGroupId` | Auto Scaling group ID for monitoring scaling activities |
