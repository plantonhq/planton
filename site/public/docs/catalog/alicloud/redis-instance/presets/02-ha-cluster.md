---
title: "HA Cluster Redis"
description: "This preset creates a production-grade Redis cluster with multi-zone high availability, horizontal sharding for throughput, and read replicas for read scaling."
type: "preset"
rank: "02"
presetSlug: "02-ha-cluster"
componentSlug: "redis-instance"
componentTitle: "Redis Instance"
provider: "alicloud"
icon: "package"
order: 2
---

# HA Cluster Redis

This preset creates a production-grade Redis cluster with multi-zone high availability, horizontal sharding for throughput, and read replicas for read scaling.

## When to Use

- Production workloads requiring high availability
- Applications with high cache throughput needs
- Read-heavy workloads that benefit from read replicas
- Services that need cross-AZ failover protection

## Key Configuration Choices

- **redis.sharding.mid.default** -- mid-tier sharding instance class; adjust for your throughput needs
- **4 shards** -- cluster mode for horizontal scaling; increase for higher throughput
- **2 read replicas** -- distributes read traffic; adjust based on read/write ratio
- **Cross-AZ deployment** -- primary and standby in different availability zones
- **Backup schedule** -- Monday/Wednesday/Friday at 03:00-04:00 UTC
- **Deletion protection** -- prevents accidental deletion
- **maxmemory-policy: allkeys-lru** -- evicts least recently used keys when memory is full

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Region code (e.g., `cn-shanghai`) | Your deployment region |
| `<your-vswitch-resource-name>` | VSwitch resource name for `valueFrom` reference | Your AliCloudVswitch resource |
| `<your-instance-name>` | Instance name | Choose a descriptive name |
| `<your-organization>` | Organization identifier | Your org slug |
| `<primary-zone-id>` | Primary AZ (e.g., `cn-shanghai-a`) | Available AZs in your region |
| `<standby-zone-id>` | Standby AZ (e.g., `cn-shanghai-b`) | A different AZ from primary |
| `<your-password>` | Instance password (8-32 chars) | Use a secrets manager |
| `<your-application-cidr>` | Application CIDR (e.g., `10.0.0.0/8`) | Your VPC CIDR range |
| `<your-team>` | Team tag value | Your team name |

## Related Presets

- **01-standard-single** -- Use for development and testing
- **03-production-encrypted** -- Use when encryption at rest and in transit is required
