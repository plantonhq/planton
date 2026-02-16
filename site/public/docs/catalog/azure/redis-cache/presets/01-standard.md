---
title: "Standard Redis Cache"
description: "This preset creates an Azure Cache for Redis with Standard tier (primary + replica), 1 GB cache size (C1), Redis 6, and TLS 1.2 enforcement. The Standard tier provides a replicated two-node cache..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "redis-cache"
componentTitle: "Redis Cache"
provider: "azure"
icon: "package"
order: 1
---

# Standard Redis Cache

This preset creates an Azure Cache for Redis with Standard tier (primary + replica), 1 GB cache size (C1), Redis 6, and TLS 1.2 enforcement. The Standard tier provides a replicated two-node cache with 99.9% SLA, making it the recommended entry point for production caching workloads. The `volatile-lru` eviction policy removes least-recently-used keys that have a TTL set when memory is full.

## When to Use

- Application caching for web sessions, API responses, and frequently accessed data
- Session state management for Azure App Service, ASP.NET, or other web frameworks
- Pub/sub messaging for lightweight real-time communication between services
- Production workloads needing a replicated cache with SLA guarantees without VNet requirements

## Key Configuration Choices

- **Standard tier** (`skuName: Standard`) -- Primary + replica with 99.9% SLA. Upgrade to Premium for VNet injection, clustering, or data persistence
- **1 GB cache (C1)** (`capacity: 1`) -- Suitable for small-to-medium caching workloads. Scale up: C2=2.5 GB, C3=6 GB, C4=13 GB, C5=26 GB, C6=53 GB
- **Redis 6** (`redisVersion: "6"`) -- Default and recommended. Redis 4 is end-of-life
- **SSL only** (`nonSslPortEnabled: false`) -- Only port 6380 (SSL) is open. The non-SSL port 6379 is disabled for security
- **TLS 1.2** (`minimumTlsVersion: "1.2"`) -- Enforces modern TLS for all connections
- **volatile-lru eviction** (`maxmemoryPolicy: volatile-lru`) -- Evicts LRU keys that have a TTL when memory is full. Use `allkeys-lru` for pure cache workloads where all keys are expendable
- **No firewall rules** -- All public IPs can connect. Add `firewallRules` to restrict access to specific IP ranges
- **Family auto-derived** -- The cache family (C for Basic/Standard, P for Premium) is automatically determined from `skuName`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-redis-name>` | Globally unique cache name (1-63 chars, lowercase/numbers/hyphens) | Choose a name; becomes `{name}.redis.cache.windows.net` |

## Related Presets

- **02-premium-vnet** -- Use instead for VNet-injected Redis with private networking and optional clustering
