---
title: "MySQL Production Instance"
description: "This preset creates a Multi-AZ RDS MySQL instance with encrypted storage and private network access. Same production-grade defaults as the PostgreSQL preset, but configured for MySQL 8.0. Suitable..."
type: "preset"
rank: "02"
presetSlug: "02-mysql-production"
componentSlug: "rds-instance"
componentTitle: "RDS Instance"
provider: "aws"
icon: "package"
order: 2
---

# MySQL Production Instance

This preset creates a Multi-AZ RDS MySQL instance with encrypted storage and private network access. Same production-grade defaults as the PostgreSQL preset, but configured for MySQL 8.0. Suitable for applications that require MySQL-compatible SQL.

## When to Use

- Production MySQL databases that need automatic failover (Multi-AZ)
- Applications migrating from on-premises MySQL to AWS
- Workloads that don't need Aurora's performance benefits and prefer standard RDS MySQL

## Key Configuration Choices

- **MySQL 8.0** (`engine: mysql`, `engineVersion: "8.0.35"`) -- Standard RDS MySQL; update to latest minor version
- **Multi-AZ** (`multiAz: true`) -- Synchronous standby for automatic failover
- **db.t3.medium** -- 2 vCPUs, 4 GiB RAM; increase for production workloads
- **50 GiB storage** -- GP2/GP3 SSD; increase based on data requirements
- **Encrypted and private** -- Same security defaults as the PostgreSQL preset

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing MySQL port (3306) from application tier | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<master-username>` | Master database username (e.g., `dbadmin`) | Your credential management system |
| `<master-password>` | Master database password | Your credential management system |

## Related Presets

- **01-postgresql-production** -- Use instead for PostgreSQL workloads
