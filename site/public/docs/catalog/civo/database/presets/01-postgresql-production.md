---
title: "Production PostgreSQL Database"
description: "This preset creates a production-grade managed PostgreSQL 16 database with 2 read replicas (3 nodes total), VPC networking, and firewall protection. This is the most common configuration for..."
type: "preset"
rank: "01"
presetSlug: "01-postgresql-production"
componentSlug: "database"
componentTitle: "Database"
provider: "civo"
icon: "package"
order: 1
---

# Production PostgreSQL Database

This preset creates a production-grade managed PostgreSQL 16 database with 2 read replicas (3 nodes total), VPC networking, and firewall protection. This is the most common configuration for mission-critical applications requiring high availability and read scaling.

## When to Use

- Production applications backed by PostgreSQL
- Workloads requiring read replicas for scaling read-heavy queries
- Environments where the database must be isolated within a private network with firewall rules

## Key Configuration Choices

- **PostgreSQL 16** (`engine: postgres`, `engineVersion: "16"`) -- latest stable major version with proven reliability
- **2 replicas** (`replicas: 2`) -- primary + 2 read replicas for a 3-node cluster; provides automatic failover and read scaling
- **Medium instance** (`sizeSlug: g3.db.medium`) -- balanced CPU/RAM for most production workloads; scale up for heavier queries
- **VPC networking** (`networkId`) -- database traffic stays within the private network
- **Firewall protection** (`firewallIds`) -- restricts access to the application tier only

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-network-id>` | Network ID of the target CivoVpc | `CivoVpc` status outputs |
| `<database-firewall-id>` | Firewall ID for database-tier rules | `CivoFirewall` status outputs |

## Related Presets

- **02-mysql-development** -- Use instead for MySQL workloads or development environments without replicas
