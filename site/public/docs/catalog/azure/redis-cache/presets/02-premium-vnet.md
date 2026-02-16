---
title: "Premium Redis Cache with VNet Injection"
description: "This preset creates an Azure Cache for Redis with Premium tier injected into a virtual network subnet. VNet injection provides private IP addressing and network isolation -- the cache is not..."
type: "preset"
rank: "02"
presetSlug: "02-premium-vnet"
componentSlug: "redis-cache"
componentTitle: "Redis Cache"
provider: "azure"
icon: "package"
order: 2
---

# Premium Redis Cache with VNet Injection

This preset creates an Azure Cache for Redis with Premium tier injected into a virtual network subnet. VNet injection provides private IP addressing and network isolation -- the cache is not accessible from the public internet. Premium tier also enables clustering (sharding), data persistence, and zone redundancy. This is the recommended configuration for production workloads requiring enterprise networking and the highest performance.

## When to Use

- Production caches requiring private network isolation with no public internet exposure
- Applications running on VMs or AKS clusters in the same VNet that need low-latency private connectivity
- Compliance-driven environments (PCI-DSS, HIPAA) mandating private-only cache access
- High-throughput workloads that may need clustering (sharding) or data persistence in the future

## Key Configuration Choices

- **Premium tier** (`skuName: Premium`) -- Required for VNet injection, clustering, data persistence, and zone redundancy
- **6 GB cache (P1)** (`capacity: 1`) -- 6 GB per shard. Scale up: P2=13 GB, P3=26 GB, P4=53 GB, P5=120 GB
- **VNet injection** (`subnetId`) -- Cache is deployed inside the specified subnet with a private IP. The subnet must be dedicated to Redis (no other resources). ForceNew field
- **Public access disabled** (`publicNetworkAccessEnabled: false`) -- Cache is only reachable within the VNet
- **No clustering** -- Clustering is not configured (no `shardCount`). Add `shardCount: 2` to enable data partitioning across shards for higher throughput
- **volatile-lru eviction** (`maxmemoryPolicy: volatile-lru`) -- Evicts LRU keys with TTL when full. Change to `allkeys-lru` for pure cache workloads
- **No firewall rules** -- Firewall rules are not effective when VNet-injected. Network access is controlled by VNet/NSG rules instead
- **Family auto-derived** -- The cache family (P for Premium) is automatically determined from `skuName`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-redis-name>` | Globally unique cache name (1-63 chars, lowercase/numbers/hyphens) | Choose a name; becomes `{name}.redis.cache.windows.net` |
| `<redis-subnet-resource-id>` | ARM resource ID of a subnet dedicated to Redis | `AzureSubnet` status outputs (subnet must have no other resources) |

## Related Presets

- **01-standard** -- Use instead for production caching without VNet requirements at lower cost
