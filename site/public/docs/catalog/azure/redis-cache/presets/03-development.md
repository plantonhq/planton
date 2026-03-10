---
title: "Development Redis Cache"
description: "This preset creates an Azure Cache for Redis with Basic tier, the smallest cache size (C0, 250 MB), and `allkeys-lru` eviction for pure caching workloads. Basic tier provides a single-node cache with..."
type: "preset"
rank: "03"
presetSlug: "03-development"
componentSlug: "redis-cache"
componentTitle: "Redis Cache"
provider: "azure"
icon: "package"
order: 3
---

# Development Redis Cache

This preset creates an Azure Cache for Redis with Basic tier, the smallest cache size (C0, 250 MB), and `allkeys-lru` eviction for pure caching workloads. Basic tier provides a single-node cache with no SLA and no replication at ~$17/month — the cheapest Azure Redis option. Designed for development, testing, and staging environments where cost matters more than availability.

## When to Use

- Local development and feature branch testing against a real Redis instance
- CI/CD pipeline caching for integration tests
- Staging environments where cache loss is acceptable (no replication)
- Learning and experimentation with Redis data structures, pub/sub, or Lua scripting
- Proof-of-concept deployments where you need a managed Redis quickly

## Key Configuration Choices

- **Basic tier** (`skuName: Basic`) -- Single-node cache with no replication and no SLA (~99% availability in practice). No support for zone redundancy, clustering, or data persistence. Upgrade to Standard for primary + replica with 99.9% SLA
- **250 MB (C0)** (`capacity: 0`) -- The smallest and cheapest cache size at ~$17/month. Supports up to 256 concurrent connections. Scale up: C1=1 GB (~$42/month), C2=2.5 GB, C3=6 GB
- **allkeys-lru eviction** (`maxmemoryPolicy: allkeys-lru`) -- Evicts least-recently-used keys when memory is full, regardless of TTL. Ideal for pure caching workloads in dev where all keys are expendable. The production presets use `volatile-lru` which only evicts keys with TTL set
- **SSL only** (`nonSslPortEnabled: false`) -- Only port 6380 (SSL) is open. The non-SSL port 6379 is disabled for security
- **TLS 1.2** (`minimumTlsVersion: "1.2"`) -- Enforces modern TLS for all connections
- **No firewall rules** -- All public IPs can connect. Acceptable for dev; add `firewallRules` to restrict access in shared environments
- **Family auto-derived** -- The cache family (C for Basic/Standard) is automatically determined from `skuName`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-redis-name>` | Globally unique cache name (1-63 chars, lowercase/numbers/hyphens) | Choose a name; becomes `{name}.redis.cache.windows.net` |

## Related Presets

- **01-standard** -- Use instead for production caching with replication and 99.9% SLA
- **02-premium-vnet** -- Use instead for private networking with VNet injection and optional clustering
