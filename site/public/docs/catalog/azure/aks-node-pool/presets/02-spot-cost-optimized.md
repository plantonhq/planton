---
title: "Spot Cost-Optimized Node Pool"
description: "This preset creates a cost-optimized AKS user node pool using Azure Spot VMs, which provide 30-90% savings over on-demand pricing. The pool scales to zero when idle and up to 10 nodes under load...."
type: "preset"
rank: "02"
presetSlug: "02-spot-cost-optimized"
componentSlug: "aks-node-pool"
componentTitle: "AKS Node Pool"
provider: "azure"
icon: "package"
order: 2
---

# Spot Cost-Optimized Node Pool

This preset creates a cost-optimized AKS user node pool using Azure Spot VMs, which provide 30-90% savings over on-demand pricing. The pool scales to zero when idle and up to 10 nodes under load. Spot VMs can be evicted when Azure needs capacity, so this pool is only suitable for fault-tolerant, stateless workloads.

## When to Use

- Batch processing, CI/CD runners, and other interruptible workloads
- Dev/test environments where occasional eviction is acceptable
- Stateless services with proper retry logic and graceful shutdown handling
- Cost-sensitive workloads that can be rescheduled to on-demand pools during eviction

## Key Configuration Choices

- **Spot VMs** (`spotEnabled: true`) -- 30-90% cost savings over on-demand; nodes can be evicted at any time
- **Scale-to-zero** (`autoscaling.minNodes: 0`) -- No cost when there are no pods to schedule; nodes are provisioned on demand
- **Standard_D4s_v5** (`vmSize`) -- 4 vCPUs, 16 GiB RAM; good balance of cost and capability for Spot
- **3 availability zones** -- Spreads across zones for better Spot capacity availability
- **User mode** (`mode: USER`) -- Application workloads only; Spot cannot be used for system pools

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aks-cluster-name>` | Name of the parent AKS cluster | `AzureAksCluster` metadata.name |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |

## Related Presets

- **01-on-demand-general** -- Use instead for workloads that cannot tolerate eviction (production APIs, stateful services)
