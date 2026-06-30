# Production PostgreSQL HA Instance

This preset creates a high-availability Scaleway RDB instance running PostgreSQL 16 with a standby replica, Private Network connectivity, encryption at rest, and frequent backups. This is the standard production database configuration for applications requiring durability and low-latency failover.

## When to Use

- Production applications that cannot tolerate database downtime
- Data-sensitive workloads requiring encryption at rest and Private Network isolation
- Applications requiring point-in-time recovery with frequent backup snapshots

## Key Configuration Choices

- **PostgreSQL 16** (`engine: PostgreSQL-16`) -- latest stable version with all modern features
- **DB-GP-XS node** (`nodeType: DB-GP-XS`) -- general-purpose production node; upgrade to `DB-GP-S` or larger for higher throughput
- **High availability** (`isHaCluster: true`) -- a synchronous standby replica provides automatic failover; RPO is near-zero
- **Private Network** (`privateNetworkId`) -- database is reachable only via private IPs; not exposed to the internet
- **50 GB storage** (`volumeSizeInGb: 50`) -- starting size; adjust based on data volume
- **6-hour backup frequency** (`backupScheduleFrequencyHours: 6`) -- 4 snapshots per day for fine-grained point-in-time recovery
- **30-day backup retention** (`backupScheduleRetentionDays: 30`) -- a full month of backup history
- **Encryption at rest** (`encryptionAtRest: true`) -- data on disk is encrypted

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-private-network-id>` | UUID of the Private Network for database connectivity | Scaleway console or `ScalewayPrivateNetwork` status outputs |
| `<your-admin-user>` | Database admin username (max 63 characters) | Choose a username for your database |
| `<your-admin-password>` | Database admin password (min 8 characters) | Generate a strong password |

## Related Presets

- **01-dev-postgres** -- Use instead for development with minimal cost and no HA
- **03-mysql-web-app** -- Use instead for MySQL workloads
