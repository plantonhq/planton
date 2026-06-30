# Aurora MySQL Cluster

This preset creates a production-ready Aurora MySQL cluster with the same security and resilience defaults as the PostgreSQL preset: managed password, encrypted storage, deletion protection, and 7-day backups. Error and slow query logs are exported to CloudWatch for operational visibility.

## When to Use

- Production relational databases using MySQL-compatible SQL
- Applications migrating from MySQL or MariaDB to Aurora
- Workloads benefiting from Aurora's MySQL-compatible engine (up to 5x throughput over standard MySQL)

## Key Configuration Choices

- **Aurora MySQL 8.0** (`engine: aurora-mysql`, `engineVersion: "8.0.mysql_aurora.3.05.2"`) -- MySQL 8.0 compatible; update to latest minor version
- **Managed password** (`manageMasterUserPassword: true`) -- Password stored and rotated in Secrets Manager
- **CloudWatch logs** (`enabledCloudwatchLogsExports: [error, slowquery]`) -- Error logs for debugging, slow query logs for performance optimization
- **Same production defaults** as PostgreSQL preset: encrypted storage, deletion protection, 7-day backups, final snapshot

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing database port (3306) from application tier | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<database-name>` | Name of the initial database to create | Your application configuration |
| `<final-snapshot-name>` | Identifier for the final snapshot | Your naming convention |

## Related Presets

- **01-aurora-postgresql** -- Use instead for PostgreSQL-compatible workloads
- **03-aurora-serverless-v2** -- Use instead for variable traffic patterns with auto-scaling compute
