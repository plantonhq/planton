# PostgreSQL Production Cluster

This preset creates a production PostgreSQL 14 PolarDB cluster with 3 nodes, proper collation settings, and deletion protection.

## When to Use

- PostgreSQL production workloads
- Applications requiring locale-aware collation
- Services needing read scaling with 2 read replicas
- Environments requiring reliable data retention on deletion

## Key Configuration Choices

- **PostgreSQL 14** -- latest stable PolarDB PostgreSQL version
- **3 nodes** -- 1 primary + 2 read replicas for read scaling and HA
- **polar.pg.x4.large** -- balanced production node class
- **UTF8 with en_US.UTF-8 collation** -- standard for internationalized applications
- **Deletion lock** -- prevents accidental cluster deletion
- **Backup retention** -- retains latest backup on cluster deletion

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<alibaba-cloud-region>` | Alibaba Cloud region code | Your deployment region |
| `<your-vswitch-resource>` | AliCloudVswitch resource name | Your VSwitch resource metadata.name |
| `<your-cluster-name>` | Cluster name (2-256 chars) | Choose a descriptive name |
| `<your-organization>` | Organization identifier | Your org name |
| `<your-vpc-cidr>` | VPC CIDR for security whitelist | Your VPC CIDR block |
| `<your-database-name>` | Database name | Choose a name |
| `<your-account-name>` | Account name | Choose a username |
| `<your-password>` | Account password | Use a secrets manager |
| `<your-team>` | Team tag value | Your team name |

## Related Presets

- **01-mysql-dev** -- Use for MySQL development with minimal resources
- **02-mysql-production** -- Use for MySQL production clusters
