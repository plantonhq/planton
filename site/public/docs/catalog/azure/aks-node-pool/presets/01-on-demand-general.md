---
title: "On-Demand General Purpose Node Pool"
description: "This preset creates a general-purpose AKS user node pool with on-demand (regular) VMs, autoscaling from 2 to 10 nodes across 3 availability zones. This is the standard configuration for production..."
type: "preset"
rank: "01"
presetSlug: "01-on-demand-general"
componentSlug: "aks-node-pool"
componentTitle: "AKS Node Pool"
provider: "azure"
icon: "package"
order: 1
---

# On-Demand General Purpose Node Pool

This preset creates a general-purpose AKS user node pool with on-demand (regular) VMs, autoscaling from 2 to 10 nodes across 3 availability zones. This is the standard configuration for production application workloads that need reliable, non-preemptible compute.

## When to Use

- Production application workloads that cannot tolerate node eviction
- General-purpose services (web apps, APIs, microservices) with moderate CPU/memory needs
- Teams adding a dedicated user node pool to an existing AKS cluster
- Workloads requiring high availability across availability zones

## Key Configuration Choices

- **On-demand VMs** (`spotEnabled` not set) -- Regular pricing with no risk of eviction; suitable for all workload types
- **Standard_D4s_v5** (`vmSize`) -- 4 vCPUs, 16 GiB RAM; balanced general-purpose compute
- **Autoscaling 2-10** (`autoscaling.minNodes: 2`, `autoscaling.maxNodes: 10`) -- Always has at least 2 nodes for HA; scales up to 10 under load
- **3 availability zones** -- Distributes nodes across zones for 99.95% SLA
- **User mode** (`mode: USER`) -- Runs application workloads; separated from system components

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aks-cluster-name>` | Name of the parent AKS cluster | `AzureAksCluster` metadata.name |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |

## Related Presets

- **02-spot-cost-optimized** -- Use instead for fault-tolerant, stateless workloads that can tolerate eviction in exchange for 30-90% cost savings
