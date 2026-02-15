---
title: "Production Redis with Replicas"
description: "This preset deploys a 3-replica Redis deployment with persistence and production-grade resources. Provides read scaling and data durability through replication and persistent storage."
type: "preset"
rank: "02"
presetSlug: "02-persistent-with-replicas"
componentSlug: "redis"
componentTitle: "Redis"
provider: "kubernetes"
icon: "package"
order: 2
---

# Production Redis with Replicas

This preset deploys a 3-replica Redis deployment with persistence and production-grade resources. Provides read scaling and data durability through replication and persistent storage.

## When to Use

- Production caching, session storage, or pub/sub workloads
- Environments requiring read replicas for higher throughput
- Workloads where data persistence across pod restarts is critical

## Key Configuration Choices

- **3 replicas** -- one primary with two replicas for read scaling and failover
- **Persistence enabled** with 5Gi disk -- durable across restarts; increase for larger datasets
- **Higher resources** (`100m`/`256Mi` requests, `1000m`/`2Gi` limits) -- production-appropriate for moderate to high throughput

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **01-single-instance** -- Minimal single-replica Redis for development
