---
title: "PostgreSQL Production Instance"
description: "This preset creates a Multi-AZ RDS PostgreSQL instance with encrypted storage and private network access. Multi-AZ deploys a synchronous standby replica in a different Availability Zone for automatic..."
type: "preset"
rank: "01"
presetSlug: "01-postgresql-production"
componentSlug: "rds-instance"
componentTitle: "RDS Instance"
provider: "aws"
icon: "package"
order: 1
---

# PostgreSQL Production Instance

This preset creates a Multi-AZ RDS PostgreSQL instance with encrypted storage and private network access. Multi-AZ deploys a synchronous standby replica in a different Availability Zone for automatic failover. This is the standard production configuration for non-Aurora PostgreSQL workloads.

## When to Use

- Production PostgreSQL databases that need automatic failover (Multi-AZ)
- Workloads that don't need Aurora's performance benefits and prefer standard RDS pricing
- Applications requiring a specific PostgreSQL version or feature not available in Aurora

## Key Configuration Choices

- **PostgreSQL 15.4** (`engine: postgres`) -- Standard RDS PostgreSQL (not Aurora); update to latest minor version
- **Multi-AZ** (`multiAz: true`) -- Synchronous standby replica for automatic failover; ~2x cost of single-AZ
- **db.t3.medium** (`instanceClass`) -- 2 vCPUs, 4 GiB RAM; burstable; increase for production workloads
- **50 GiB storage** (`allocatedStorageGb: 50`) -- GP2/GP3 SSD storage; increase based on data requirements
- **Encrypted** (`storageEncrypted: true`) -- Data at rest encryption with AWS-managed key
- **Private** (`publiclyAccessible: false`) -- No public IP; accessible only from within the VPC

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing PostgreSQL port (5432) from application tier | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<master-username>` | Master database username (e.g., `dbadmin`) | Your credential management system |
| `<master-password>` | Master database password (store securely; consider using Secrets Manager) | Your credential management system |

## Related Presets

- **02-mysql-production** -- Use instead for MySQL workloads
