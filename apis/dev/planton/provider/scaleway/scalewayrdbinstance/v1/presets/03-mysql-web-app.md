# MySQL Web Application Database

This preset creates a Scaleway RDB instance running MySQL 8 on a general-purpose node with Private Network connectivity. It is sized for typical web application databases (CMS, e-commerce, SaaS) where MySQL is the preferred engine.

## When to Use

- Web applications built on PHP/Laravel, Ruby on Rails, or WordPress that require MySQL
- SaaS applications and CMSes with existing MySQL schemas
- Any workload where MySQL compatibility is a requirement

## Key Configuration Choices

- **MySQL 8** (`engine: MySQL-8`) -- latest stable MySQL version with JSON support, window functions, and CTEs
- **DB-GP-XS node** (`nodeType: DB-GP-XS`) -- general-purpose production node; sufficient for most web application databases
- **Private Network** (`privateNetworkId`) -- database is reachable only via private IPs
- **20 GB storage** (`volumeSizeInGb: 20`) -- starting size for a typical web app; increase as data grows
- **No HA** -- single-node for cost efficiency; add `isHaCluster: true` for production workloads requiring failover

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-private-network-id>` | UUID of the Private Network for database connectivity | Scaleway console or `ScalewayPrivateNetwork` status outputs |
| `<your-admin-user>` | Database admin username (max 63 characters) | Choose a username for your database |
| `<your-admin-password>` | Database admin password (min 8 characters) | Generate a strong password |

## Related Presets

- **01-dev-postgres** -- Use instead for PostgreSQL development databases
- **02-production-postgres-ha** -- Use instead for production PostgreSQL with HA
