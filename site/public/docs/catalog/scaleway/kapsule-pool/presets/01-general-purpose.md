---
title: "General-Purpose Node Pool"
description: "This preset creates a fixed-size node pool with GP1-XS instances (4 vCPU, 16 GB RAM) for general Kubernetes workloads. Autohealing is enabled and nodes have no public IPs. This is the standard..."
type: "preset"
rank: "01"
presetSlug: "01-general-purpose"
componentSlug: "kapsule-pool"
componentTitle: "Kapsule Pool"
provider: "scaleway"
icon: "package"
order: 1
---

# General-Purpose Node Pool

This preset creates a fixed-size node pool with GP1-XS instances (4 vCPU, 16 GB RAM) for general Kubernetes workloads. Autohealing is enabled and nodes have no public IPs. This is the standard additional pool for extending a Kapsule cluster beyond its default node pool.

## When to Use

- Adding compute capacity to an existing Kapsule cluster with a different instance type than the default pool
- Running general-purpose workloads (web servers, APIs, background workers)
- Workloads with predictable resource needs that do not require autoscaling

## Key Configuration Choices

- **GP1-XS nodes** (`nodeType: GP1-XS`) -- 4 vCPU, 16 GB RAM; balanced CPU-to-memory ratio suitable for most workloads
- **Fixed 3-node pool** (`size: 3`) -- provides redundancy for pod scheduling; adjust to match capacity needs
- **Autohealing enabled** (`autohealing: true`) -- unhealthy nodes are automatically detected and replaced
- **Private nodes** (`publicIpDisabled: true`) -- nodes have no public IPs for a secure production posture
- **Kubernetes label** (`pool: general`) -- allows pod scheduling via `nodeSelector` or node affinity
- **No autoscaling** -- fixed size for predictable cost and capacity

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-kapsule-cluster-id>` | UUID of the parent Kapsule cluster | Scaleway console or `ScalewayKapsuleCluster` status outputs |

## Related Presets

- **02-autoscaling-workers** -- Use instead for elastic workloads that need automatic node scaling
