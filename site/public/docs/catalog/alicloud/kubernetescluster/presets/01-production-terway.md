---
title: "Production ACK Cluster with Terway ENI Networking"
description: "This preset creates a production-grade ACK Managed Kubernetes cluster using Terway ENI-based networking. Terway assigns VPC Elastic Network Interfaces directly to pods, giving each pod a VPC-routable..."
type: "preset"
rank: "01"
presetSlug: "01-production-terway"
componentSlug: "kubernetescluster"
componentTitle: "KubernetesCluster"
provider: "alicloud"
icon: "package"
order: 1
---

# Production ACK Cluster with Terway ENI Networking

This preset creates a production-grade ACK Managed Kubernetes cluster using Terway ENI-based networking. Terway assigns VPC Elastic Network Interfaces directly to pods, giving each pod a VPC-routable IP address with native security group isolation. The cluster runs on the professional tier (ack.pro.small) with RRSA for pod-level IAM, full control plane and audit logging, deletion protection, and a weekly maintenance window with automatic patch upgrades.

## When to Use

- Production Kubernetes workloads requiring high-performance pod networking with VPC-native IP addresses
- Environments where pods need individual security group rules or direct VPC connectivity
- Teams that manage their own NAT gateway and VPC infrastructure externally
- Clusters that require audit logging and control plane observability for compliance

## Key Configuration Choices

- **Terway ENI networking** (`terway-eniip` addon + `podVswitchIds`) -- Pods receive their own VPC ENIs, enabling per-pod security groups and eliminating overlay network overhead. Requires dedicated pod VSwitches with sufficient IP capacity separate from the node VSwitches.
- **ack.pro.small** (`clusterSpec: ack.pro.small`) -- Professional managed cluster with enhanced SLA, managed node pool support, and topology-aware scheduling. The control plane management fee is modest relative to the worker node costs.
- **Three availability zones** (3 node VSwitches + 3 pod VSwitches) -- Distributes the control plane and worker nodes across three AZs for maximum resilience against zone-level failures.
- **RRSA enabled** (`enableRrsa: true`) -- Pods assume RAM roles via OIDC federation, eliminating static access keys. Requires Kubernetes 1.22.3+ and cannot be disabled once enabled.
- **Deletion protection** (`deletionProtection: true`) -- Prevents accidental cluster deletion via API. Must be explicitly disabled before the cluster can be destroyed.
- **External NAT** (`newNatGateway: false`) -- Assumes you manage NAT via a dedicated AliCloudNatGateway component, avoiding duplicate NAT gateways and giving you full control over SNAT rules and EIP allocation.
- **Enterprise security group** (`isEnterpriseSecurity Group: true`) -- Supports up to 65,536 rules and 100,000 ENIs, required for large Terway clusters.
- **Full observability** (logtail-ds, arms-prometheus, metrics-server, ack-node-problem-detector) -- Ships control plane logs and audit events to SLS with 90-day retention, collects Prometheus metrics, and detects node-level problems automatically.
- **Patch-only auto-upgrade** (`channel: patch`) -- Applies security patches and bug fixes automatically but never performs minor version upgrades (e.g., 1.30.x to 1.30.y, never 1.30 to 1.31). Runs only during the Wednesday maintenance window.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-hangzhou`, `ap-southeast-1`) | Your deployment region strategy |
| `<your-cluster-name>` | Cluster name (1-63 chars, alphanumeric) | Your naming convention |
| `<node-vswitch-id-zone-a>` | VSwitch for nodes in AZ a | `AliCloudVswitch` stack outputs |
| `<node-vswitch-id-zone-b>` | VSwitch for nodes in AZ b | `AliCloudVswitch` stack outputs |
| `<node-vswitch-id-zone-c>` | VSwitch for nodes in AZ c | `AliCloudVswitch` stack outputs |
| `<pod-vswitch-id-zone-a>` | Dedicated pod VSwitch in AZ a (separate CIDR from node VSwitch) | `AliCloudVswitch` stack outputs |
| `<pod-vswitch-id-zone-b>` | Dedicated pod VSwitch in AZ b | `AliCloudVswitch` stack outputs |
| `<pod-vswitch-id-zone-c>` | Dedicated pod VSwitch in AZ c | `AliCloudVswitch` stack outputs |
| `<your-log-project-name>` | SLS project for cluster logs and addon dashboards | `AliCloudLogProject` stack outputs |
| `<your-team>` | Team or business unit that owns this cluster | Your organizational structure |
| `<your-cost-center>` | Cost center code for billing attribution | Your finance team |

## Related Presets

- **02-development-flannel** -- Use for development and testing where Flannel overlay networking is simpler and cheaper (no dedicated pod VSwitches required)
- **03-production-flannel** -- Use when production workloads do not require per-pod ENI isolation and simpler overlay networking is preferred
