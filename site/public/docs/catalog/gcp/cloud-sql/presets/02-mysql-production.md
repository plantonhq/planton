---
title: "MySQL Production Instance"
description: "This preset creates a production-grade Cloud SQL MySQL 8.0 instance with the same security and reliability posture as the PostgreSQL production preset: high availability, private IP, automated..."
type: "preset"
rank: "02"
presetSlug: "02-mysql-production"
componentSlug: "cloud-sql"
componentTitle: "Cloud SQL"
provider: "gcp"
icon: "package"
order: 2
---

# MySQL Production Instance

This preset creates a production-grade Cloud SQL MySQL 8.0 instance with the same security and reliability posture as the PostgreSQL production preset: high availability, private IP, automated backups with PITR, and deletion protection.

## When to Use

- Production applications that require MySQL (WordPress, legacy apps, MySQL-specific features)
- Workloads requiring automatic regional failover
- Environments migrating from self-managed MySQL to Cloud SQL

## Key Configuration Choices

- **MySQL 8.0** (`databaseVersion: MYSQL_8_0`) -- latest stable MySQL version with JSON support, CTEs, and window functions
- **Custom 2-8192 tier** -- 2 vCPU, 8 GB RAM; adjust based on workload
- **Private IP only** -- accessible only within the VPC
- **High availability** -- regional HA with automatic failover
- **Automated backups with PITR** -- daily backups with point-in-time recovery
- **Query Insights** -- performance monitoring
- **Deletion protection** -- prevents accidental deletion

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |
| `<vpc-network-id>` | VPC network ID for private IP | `GcpVpc` status outputs |
| `<failover-zone>` | Zone for HA failover | Different zone in the same region |

## Related Presets

- **01-postgresql-production** -- Use for PostgreSQL workloads
- **03-postgresql-development** -- Use as a template for a MySQL development instance (change engine to MYSQL)
