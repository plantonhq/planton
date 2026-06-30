---
title: "MySQL Production with Encryption"
description: "This preset creates a production MySQL 8.0 instance with high availability, TDE encryption, KMS disk encryption, SSL, monitoring, and performance-tuned parameters."
type: "preset"
rank: "03"
presetSlug: "03-mysql-production"
componentSlug: "rds-instance"
componentTitle: "RDS Instance"
provider: "alicloud"
icon: "package"
order: 3
---

# MySQL Production with Encryption

This preset creates a production MySQL 8.0 instance with high availability, TDE encryption, KMS disk encryption, SSL, monitoring, and performance-tuned parameters.

## When to Use

- Production MySQL workloads with strict security requirements
- Financial or compliance-sensitive data (PCI-DSS, SOC2)
- Applications requiring transparent data encryption at rest
- Environments needing a DBA superuser alongside application accounts

## Key Configuration Choices

- **HighAvailability category** -- primary + standby with automatic failover
- **Cross-AZ deployment** -- primary and standby in different zones
- **rds.mysql.s2.xlarge** -- 4 vCPU, 8 GB RAM; adjust for workload size
- **200 GB cloud_essd** -- ample storage with high IOPS
- **TDE enabled** -- transparent data encryption at rest (irreversible once enabled)
- **KMS encryption** -- customer-managed key for disk encryption
- **SSL enabled** -- encrypted client connections
- **Performance tuning** -- optimized InnoDB buffer pool and connection limits
- **Dual databases** -- primary and secondary with differentiated access
- **Super account** -- DBA administrative access alongside application accounts

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code | Your deployment region |
| `<your-vswitch-id>` | VSwitch ID for instance placement | `AliCloudVswitch` stack outputs |
| `<primary-zone-id>` | Primary AZ | Region availability zones |
| `<standby-zone-id>` | Standby AZ | Must differ from primary |
| `<your-vpc-cidr>` | VPC CIDR for IP whitelist | `AliCloudVpc` stack outputs |
| `<your-kms-key-id>` | KMS key ID for encryption | `AliCloudKmsKey` stack outputs |
| `<your-resource-group-id>` | Resource group ID | Alibaba Cloud console |
| `<your-instance-name>` | Instance name | Choose a descriptive name |
| `<your-org>` | Organization name | Your Planton org |
| `<your-team>` | Team tag value | Your team name |
| `<compliance-standard>` | Compliance standard tag | e.g., `pci-dss`, `soc2` |
| `<primary-database-name>` | Primary database name | e.g., `transactions` |
| `<secondary-database-name>` | Secondary database name | e.g., `audit_log` |
| `<app-account-name>` | Application account name | e.g., `app_svc` |
| `<app-password>` | Application account password | Use a secrets manager |
| `<admin-account-name>` | DBA account name | e.g., `dba_admin` |
| `<admin-password>` | DBA account password | Use a secrets manager |

## Related Presets

- **01-mysql-basic** -- Use for development instances
- **02-postgresql-ha** -- Use for PostgreSQL production instances
