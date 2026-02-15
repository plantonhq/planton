---
title: "Redis Cache Cluster"
description: "This preset creates a managed Redis cache cluster with one node, suitable for session storage, caching, and pub/sub workloads. VPC placement keeps cache traffic within your private network. Redis 7..."
type: "preset"
rank: "03"
presetSlug: "03-redis"
componentSlug: "database-cluster"
componentTitle: "Database Cluster"
provider: "digitalocean"
icon: "package"
order: 3
---

# Redis Cache Cluster

This preset creates a managed Redis cache cluster with one node, suitable for session storage, caching, and pub/sub workloads. VPC placement keeps cache traffic within your private network. Redis 7 provides the latest features and performance improvements.

## When to Use

- Application caching (page cache, query cache, object cache)
- Session storage for web applications
- Rate limiting and temporary data
- Pub/sub messaging between services

## Key Configuration Choices

- **Redis 7** (`engine: redis`, `engineVersion: "7"`) -- latest stable Redis version with improved memory efficiency.
- **Single node** (`nodeCount: 1`) -- sufficient for most caching use cases. Redis clusters support up to 3 nodes for HA.
- **VPC placement** (`vpc`) -- recommended so app servers and cache communicate privately.
- **Smallest size** (`sizeSlug: db-s-1vcpu-1gb`) -- start small; scale up as cache size grows.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-id>` | UUID of the target VPC | DigitalOcean VPC console or `DigitalOceanVpc` status outputs |
| `nyc3` | Target DigitalOcean region slug | Must match the VPC's region |

## Related Presets

- **01-postgresql-ha** -- Use for relational data instead of caching
- **02-postgresql-dev** -- Use for dev/test relational databases
