# AzureMysqlFlexibleServer

Azure Database for MySQL Flexible Server is a fully managed relational database service
that provides granular control over database management and configuration. It supports
burstable, general-purpose, and memory-optimized compute tiers, zone-redundant and same-zone
high availability, automatic backups with point-in-time restore, and private VNet integration.

## When to Use

Use AzureMysqlFlexibleServer when you need:

- **Managed MySQL** with automatic patching, backups, and monitoring
- **Flexible compute tiers** from burstable dev/test (B_Standard_B1ms) to production (GP_Standard_D4ds_v4)
- **High availability** with automatic failover (ZoneRedundant or SameZone)
- **Private VNet access** using delegated subnets and private DNS zones
- **Multiple databases** on a single server instance

## Key Configuration

### Compute Tiers (sku_name)

| Tier | Prefix | Use Case | Example |
|------|--------|----------|---------|
| Burstable | `B_Standard_` | Dev/test, low-traffic | `B_Standard_B1ms` (1 vCPU, 2 GiB) |
| General Purpose | `GP_Standard_` | Production workloads | `GP_Standard_D2ds_v4` (2 vCPU, 8 GiB) |
| Memory Optimized | `MO_Standard_` | Analytics, caching | `MO_Standard_E2ds_v4` (2 vCPU, 16 GiB) |

MySQL Flexible Server uses the `_v4` and `_v5` series SKU families (e.g., `D2ds_v4`, `D4ads_v5`),
which differ from the PostgreSQL Flexible Server `_v3` series. Common production SKUs:

- `GP_Standard_D2ds_v4` -- 2 vCPU, 8 GiB RAM
- `GP_Standard_D4ds_v4` -- 4 vCPU, 16 GiB RAM
- `GP_Standard_D8ds_v4` -- 8 vCPU, 32 GiB RAM
- `MO_Standard_E4ds_v4` -- 4 vCPU, 32 GiB RAM

### Storage (storage_size_gb)

Storage is provisioned in **GB** (not MB like PostgreSQL). Range: 20 GB to 16,384 GB (16 TB).

| Size | Value |
|------|-------|
| 20 GB | `20` |
| 64 GB | `64` |
| 128 GB | `128` |
| 256 GB | `256` |
| 512 GB | `512` |
| 1 TB | `1024` |
| 16 TB | `16384` |

Storage cannot be downgraded once provisioned. `auto_grow_enabled` defaults to **true** (unlike
PostgreSQL where it defaults to false), so storage automatically expands when free space is low.

### Network Access

**Public access** (default): The server is accessible over the internet. Use `firewall_rules` to restrict which IP addresses can connect.

**Private VNet access**: Set `delegated_subnet_id` to a subnet delegated to `Microsoft.DBforMySQL/flexibleServers`. Public access is automatically disabled. Optionally set `private_dns_zone_id` for DNS resolution within the VNet (typically `privatelink.mysql.database.azure.com`).

### High Availability

Set the `high_availability` block to enable automatic failover:

- **ZoneRedundant**: Standby in a different availability zone (recommended for production)
- **SameZone**: Standby in the same zone (faster failover, no zone-level protection)

Burstable SKUs do NOT support high availability.

### MySQL Versions

Supported: **5.7**, **8.0.21** (default), **8.4**

- `"5.7"` -- Approaching EOL, use only for legacy migration
- `"8.0.21"` -- Current production standard, recommended for most workloads
- `"8.4"` -- Latest GA release with performance improvements and new features

### Database Defaults

- **Charset**: `utf8mb4` (full Unicode support including emojis)
- **Collation**: `utf8mb4_0900_ai_ci` (accent-insensitive, case-insensitive; for MySQL 5.7 use `utf8mb4_unicode_ci`)

## ForceNew Fields

Changing these fields destroys and recreates the server (data loss risk):

- `name` -- server hostname
- `administrator_login` -- admin username
- `delegated_subnet_id` -- VNet integration
- `private_dns_zone_id` -- private DNS resolution
- `geo_redundant_backup_enabled` -- backup configuration

## Stack Outputs

| Output | Description |
|--------|-------------|
| `server_id` | Azure Resource Manager ID |
| `server_name` | Server name |
| `fqdn` | Fully qualified domain name (`{name}.mysql.database.azure.com`) |
| `administrator_login` | Admin login name |
| `database_ids` | Map of database name to resource ID |

**Connection string format:**

```
mysql://{administrator_login}:{password}@{fqdn}:3306/{database}?ssl-mode=REQUIRED
```

## Related Resources

- **AzureSubnet** -- Delegated subnet for VNet integration (delegation: `Microsoft.DBforMySQL/flexibleServers`)
- **AzurePrivateDnsZone** -- DNS resolution for private access (`privatelink.mysql.database.azure.com`)
- **AzurePrivateEndpoint** -- Alternative private connectivity (via Private Link)
- **AzureResourceGroup** -- Container for the server

## Infra Chart Usage

This resource is a key component of:

- **database-stack** -- MySQL + optional PostgreSQL + MSSQL + Redis + PrivateEndpoint
- **container-apps-environment** -- Optional database backend
- **web-app-environment** -- Optional database backend

## Differences from AzurePostgresqlFlexibleServer

| Aspect | MySQL | PostgreSQL |
|--------|-------|------------|
| Storage unit | GB (`storage_size_gb`) | MB (`storage_mb`) |
| Storage minimum | 20 GB | 32 GB (32768 MB) |
| auto_grow_enabled default | `true` | `false` |
| Versions | 5.7, 8.0.21, 8.4 | 12-17 |
| FQDN suffix | `.mysql.database.azure.com` | `.postgres.database.azure.com` |
| Default charset | `utf8mb4` | `UTF8` |
| Default collation | `utf8mb4_0900_ai_ci` | `en_US.utf8` |
| Backup retention range | 1-35 days | 7-35 days |
| Server name start char | Letter or number | Letter only |
| Subnet delegation | `Microsoft.DBforMySQL/flexibleServers` | `Microsoft.DBforPostgreSQL/flexibleServers` |
| SKU families | `_v4`, `_v5` series | `_v3` series |
| ForceNew: private_dns_zone_id | Yes | No |
| Default port | 3306 | 5432 |
