---
title: "Production PostgreSQL HA"
description: "This preset creates a production-grade PostgreSQL database cluster with three nodes for high availability, VPC isolation for secure private access, and PostgreSQL 16 on a 2 vCPU / 4 GB node size...."
type: "preset"
rank: "01"
presetSlug: "01-postgresql-ha"
componentSlug: "database-cluster"
componentTitle: "Database Cluster"
provider: "digitalocean"
icon: "package"
order: 1
---

# Production PostgreSQL HA

This preset creates a production-grade PostgreSQL database cluster with three nodes for high availability, VPC isolation for secure private access, and PostgreSQL 16 on a 2 vCPU / 4 GB node size. Suitable for mission-critical applications requiring automatic failover.

## When to Use

- Production applications requiring database high availability
- Workloads needing automatic failover when a primary node fails
- Environments where database traffic must stay within a private VPC

## Key Configuration Choices

- **Three nodes** (`nodeCount: 3`) -- primary plus two standby nodes for HA. DigitalOcean provides automatic failover within the cluster.
- **PostgreSQL 16** (`engine: pg`, `engineVersion: "16"`) -- latest stable major version with extended support.
- **VPC placement** (`vpc`) -- required for production; keeps database traffic off the public internet.
- **Node size** (`sizeSlug: db-s-2vcpu-4gb`) -- general-purpose sizing; scale up for heavier workloads.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-id>` | UUID of the target VPC | DigitalOcean VPC console or `DigitalOceanVpc` status outputs |
| `nyc3` | Target DigitalOcean region slug | Must match the VPC's region |

## Related Presets

- **02-postgresql-dev** -- Use instead for dev/test where HA and VPC are unnecessary
- **03-redis** -- Use for caching workloads instead of relational data
