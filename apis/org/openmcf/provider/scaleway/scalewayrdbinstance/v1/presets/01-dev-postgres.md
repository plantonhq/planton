# Development PostgreSQL Instance

This preset creates a minimal Scaleway RDB instance running PostgreSQL 16 on the smallest available node type. It uses local SSD storage and has no Private Network attachment or HA -- the simplest path to a working PostgreSQL database for development and testing.

## When to Use

- Development and testing databases
- Small applications with light database loads
- Learning and prototyping with managed PostgreSQL

## Key Configuration Choices

- **PostgreSQL 16** (`engine: PostgreSQL-16`) -- the latest stable major version; supports all modern PostgreSQL features
- **DB-DEV-S node** (`nodeType: DB-DEV-S`) -- the most affordable managed database node; sufficient for development workloads
- **Local SSD storage** (`volumeType: lssd`) -- high I/O performance for development; data is stored on the compute node's local disk
- **No Private Network** -- accessible via public endpoint with TLS; add `privateNetworkId` for production
- **No HA** (`isHaCluster` not set) -- single-node instance; acceptable for non-critical environments
- **Backups enabled** (default) -- automatic daily backups retained for 7 days by default

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-admin-user>` | Database admin username (max 63 characters) | Choose a username for your database |
| `<your-admin-password>` | Database admin password (min 8 characters) | Generate a strong password |

## Related Presets

- **02-production-postgres-ha** -- Use instead for production with HA, Private Network, and configured backups
- **03-mysql-web-app** -- Use instead for MySQL workloads
