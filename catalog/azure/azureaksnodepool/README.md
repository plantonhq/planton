# AzureAksNodePool

## Overview

`AzureAksNodePool` provisions an AKS node pool: a scale set of worker
nodes attached to an existing cluster by ARM ID. Node pools are the unit
of compute shape — general, memory-optimized, GPU, spot, or Windows pools
each live as their own resource with an independent lifecycle.

## Key Features

- **Full ~49-field azurerm surface** — mode, os_type/os_sku, spot
  priority/eviction/max-price, autoscaling, subnet FKs, disk/GPU options,
  kubelet/linux_os/sysctl config, upgrade settings, Windows outbound NAT,
  and more
- **Single parent FK** — `kubernetes_cluster_id` references the cluster's
  `cluster_id` output; resource group and cluster name are derived from it
- **Spot economics** — real `priority`/`eviction_policy`/`spot_max_price`
  modeling (replaces the old `spot_enabled` bool)
- **Composable** — prerequisite is `AzureAksCluster` (which chains to
  `AzureResourceGroup` transitively)

## Prerequisites

| Kind | Why |
|------|-----|
| `AzureAksCluster` | The pool attaches to an existing cluster by ARM ID |

## Outputs

| Output | Description |
|--------|-------------|
| `node_pool_id` | ARM ID of the agent pool |
| `node_pool_name` | Pool name (used in Kubernetes node labels) |
| `node_image_version` | Node OS image version actually running |

## Related Resources

- **`AzureAksCluster`** — parent cluster (carries only the mandatory default pool)
- **`AzureSubnet`** — optional dedicated subnet for this pool's nodes

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
