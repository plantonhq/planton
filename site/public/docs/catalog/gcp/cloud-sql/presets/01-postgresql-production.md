---
title: "PostgreSQL Production Instance"
description: "This preset creates a production-grade Cloud SQL PostgreSQL instance with high availability, private IP networking, automated backups with point-in-time recovery, query insights, and deletion..."
type: "preset"
rank: "01"
presetSlug: "01-postgresql-production"
componentSlug: "cloud-sql"
componentTitle: "Cloud SQL"
provider: "gcp"
icon: "package"
order: 1
---

# PostgreSQL Production Instance

This preset creates a production-grade Cloud SQL PostgreSQL instance with high availability, private IP networking, automated backups with point-in-time recovery, query insights, and deletion protection. It represents the standard configuration for production relational databases on GCP.

## When to Use

- Production applications needing a managed PostgreSQL database
- Workloads requiring automatic regional failover (HA)
- Environments where data protection and recovery are critical

## Key Configuration Choices

- **PostgreSQL 15** (`databaseVersion: POSTGRES_15`) -- latest stable major version
- **Custom 2-8192 tier** -- 2 vCPU, 8 GB RAM; adjust based on workload
- **Private IP only** (`privateIpEnabled: true`, no public IP) -- accessible only within the VPC
- **High availability** -- regional HA with automatic failover to a standby in the specified zone
- **Automated backups** -- daily at 2 AM UTC, 7-day retention
- **Point-in-time recovery** -- enables recovery to any specific second within the retention window
- **Query Insights** -- performance monitoring and query analysis without overhead
- **Deletion protection** -- prevents accidental deletion of the production database
- **Stable maintenance track** -- maintenance updates on Sundays at 4 AM UTC

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |
| `<vpc-network-id>` | VPC network ID for private IP | `GcpVpc` status outputs (requires Private Services Access) |
| `<failover-zone>` | Zone for HA failover (e.g., `us-central1-b`) | Different zone in the same region |

## Related Presets

- **02-mysql-production** -- Use for MySQL workloads with the same production posture
- **03-postgresql-development** -- Use for dev/test environments without HA or deletion protection
