---
title: "Autoscaling Worker Pool"
description: "This preset creates an autoscaling node pool with PRO2-M instances (4 vCPU, 16 GB RAM) that scales between 1 and 8 nodes based on workload demand. The upgrade policy uses 1 surge node for..."
type: "preset"
rank: "02"
presetSlug: "02-autoscaling-workers"
componentSlug: "kapsule-pool"
componentTitle: "Kapsule Pool"
provider: "scaleway"
icon: "package"
order: 2
---

# Autoscaling Worker Pool

This preset creates an autoscaling node pool with PRO2-M instances (4 vCPU, 16 GB RAM) that scales between 1 and 8 nodes based on workload demand. The upgrade policy uses 1 surge node for zero-downtime rolling updates. This is the standard pattern for elastic production workloads.

## When to Use

- Workloads with variable traffic patterns (e.g., batch jobs, CI/CD runners, traffic spikes)
- Production pools where cost optimization matters (scale down during low demand)
- Worker pools for background processing, queue consumers, or data pipelines

## Key Configuration Choices

- **PRO2-M nodes** (`nodeType: PRO2-M`) -- 4 vCPU, 16 GB RAM; production-optimized instances with guaranteed resources
- **Autoscaling** (`autoScale: true`, 1-8 nodes) -- the cluster autoscaler adds nodes for pending pods and removes underutilized nodes
- **Autohealing** (`autohealing: true`) -- unhealthy nodes are replaced automatically
- **Private nodes** (`publicIpDisabled: true`) -- secure production posture
- **Upgrade policy** (`maxSurge: 1`, `maxUnavailable: 1`) -- during upgrades, one extra node is created before draining the old one, ensuring workloads always have capacity
- **Kubernetes label** (`pool: workers`) -- enables targeted scheduling via `nodeSelector`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-kapsule-cluster-id>` | UUID of the parent Kapsule cluster | Scaleway console or `ScalewayKapsuleCluster` status outputs |

## Related Presets

- **01-general-purpose** -- Use instead for fixed-size pools with predictable capacity needs
