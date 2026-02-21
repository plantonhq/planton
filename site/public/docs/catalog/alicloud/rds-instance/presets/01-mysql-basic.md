---
title: "MySQL Basic Development Instance"
description: "This preset creates a minimal MySQL 8.0 instance with a single database and account, suitable for development and testing."
type: "preset"
rank: "01"
presetSlug: "01-mysql-basic"
componentSlug: "rds-instance"
componentTitle: "RDS Instance"
provider: "alicloud"
icon: "package"
order: 1
---

# MySQL Basic Development Instance

This preset creates a minimal MySQL 8.0 instance with a single database and account, suitable for development and testing.

## When to Use

- Development and testing environments
- Proof-of-concept deployments
- Learning and experimentation with Alibaba Cloud RDS
- Environments where high availability is not required

## Key Configuration Choices

- **Basic category** -- single-node deployment (no standby) for lowest cost
- **rds.mysql.t1.small** -- smallest instance class; upgrade for production use
- **20 GB cloud_essd** -- minimum storage with modern SSD performance
- **Single database and account** -- one ReadWrite account for the database
- **Postpaid billing** (default) -- pay-as-you-go, no commitment

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-hangzhou`) | Your deployment region |
| `<your-vswitch-id>` | VSwitch ID to place the instance in | `AliCloudVswitch` stack outputs |
| `<your-instance-name>` | Instance name (2-256 chars) | Choose a descriptive name |
| `<your-database-name>` | Database name (e.g., `appdb`) | Choose a name for your database |
| `<your-account-name>` | Login account name | Choose a username |
| `<your-password>` | Account password (8+ chars, mixed complexity) | Use a secrets manager |

## Related Presets

- **02-postgresql-ha** -- Use for production PostgreSQL with high availability
- **03-mysql-production** -- Use for production MySQL with encryption and monitoring
