# PostgreSQL HA Production Instance

This preset creates a production-grade PostgreSQL 16 instance with high availability, SSL encryption, cross-AZ deployment, and fine-grained monitoring.

## When to Use

- Production PostgreSQL workloads requiring high availability
- Applications needing automatic failover with cross-AZ standby
- Environments where SSL-encrypted connections are mandatory
- Compliance scenarios requiring deletion protection and monitoring

## Key Configuration Choices

- **HighAvailability category** -- primary + standby with automatic failover
- **Cross-AZ deployment** -- primary and standby in different availability zones
- **rds.pg.s2.large** -- 2 vCPU, 4 GB RAM; scale instance_type for larger workloads
- **100 GB cloud_essd** -- balanced performance; increase for data-heavy workloads
- **SSL enabled** -- encrypted client connections
- **60-second monitoring** -- fine-grained metrics collection
- **Deletion protection** -- prevents accidental instance deletion
- **Maintenance window** -- `02:00Z-06:00Z` for off-peak maintenance

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code (e.g., `cn-shanghai`) | Your deployment region |
| `<your-vswitch-id>` | VSwitch ID for instance placement | `AlicloudVswitch` stack outputs |
| `<primary-zone-id>` | Primary AZ (e.g., `cn-shanghai-a`) | Region availability zones |
| `<standby-zone-id>` | Standby AZ (e.g., `cn-shanghai-b`) | Must differ from primary |
| `<your-vpc-cidr>` | VPC CIDR for IP whitelist (e.g., `10.0.0.0/8`) | `AlicloudVpc` stack outputs |
| `<your-instance-name>` | Instance name (2-256 chars) | Choose a descriptive name |
| `<your-org>` | Organization name | Your OpenMCF org |
| `<your-team>` | Team tag value | Your team name |
| `<your-database-name>` | Database name | Choose a name |
| `<your-account-name>` | Login account name | Choose a username |
| `<your-password>` | Account password (8+ chars) | Use a secrets manager |

## Related Presets

- **01-mysql-basic** -- Use for development MySQL instances
- **03-mysql-production** -- Use for production MySQL with encryption
