---
title: "MySQL Production Cluster"
description: "This preset creates a production-grade MySQL 8.0 PolarDB cluster with 4 nodes, TDE encryption, audit logging, and deletion protection."
type: "preset"
rank: "02"
presetSlug: "02-mysql-production"
componentSlug: "polardb-cluster"
componentTitle: "PolarDB Cluster"
provider: "alicloud"
icon: "package"
order: 2
---

# MySQL Production Cluster

This preset creates a production-grade MySQL 8.0 PolarDB cluster with 4 nodes, TDE encryption, audit logging, and deletion protection.

## When to Use

- Production workloads requiring high availability
- Applications needing read scaling (3 read replicas)
- Environments with compliance requirements (encryption, audit logs)
- Long-running services with deletion protection

## Key Configuration Choices

- **Enterprise Edition (Normal)** -- shared distributed storage with auto-scaling
- **Exclusive sub-category** -- dedicated resources for consistent performance
- **4 nodes** -- 1 primary + 3 read replicas for read scaling
- **polar.mysql.x4.xlarge** -- production-class node; scale up as needed
- **TDE enabled** -- transparent data encryption at rest (irreversible)
- **Audit logging** -- SQL audit collector for compliance
- **Deletion lock** -- prevents accidental cluster deletion
- **Backup retention** -- retains latest backup on cluster deletion
- **utf8mb4 charset** -- full Unicode support including emojis

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code | Your deployment region |
| `<your-vswitch-resource>` | AliCloudVswitch resource name | Your VSwitch resource metadata.name |
| `<your-cluster-name>` | Cluster name (2-256 chars) | Choose a descriptive name |
| `<your-organization>` | Organization identifier | Your org name |
| `<your-vpc-cidr>` | VPC CIDR for security whitelist | Your VPC CIDR block |
| `<your-kms-key-id>` | KMS key ID for TDE | `AliCloudKmsKey` stack outputs |
| `<your-database-name>` | Database name | Choose a name |
| `<your-account-name>` | Account name | Choose a username |
| `<your-password>` | Account password | Use a secrets manager |
| `<your-team>` | Team tag value | Your team name |
| `<your-cost-center>` | Cost center tag value | Your cost center |

## Related Presets

- **01-mysql-dev** -- Use for development with minimal resources
- **03-postgresql-production** -- Use for PostgreSQL production clusters
