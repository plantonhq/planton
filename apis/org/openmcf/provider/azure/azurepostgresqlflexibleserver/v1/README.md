# AzurePostgresqlFlexibleServer

Azure Database for PostgreSQL Flexible Server is a fully managed relational database service
that provides granular control over database management and configuration. It supports
burstable, general-purpose, and memory-optimized compute tiers, zone-redundant and same-zone
high availability, automatic backups with point-in-time restore, and private VNet integration.

## When to Use

Use AzurePostgresqlFlexibleServer when you need:

- **Managed PostgreSQL** with automatic patching, backups, and monitoring
- **Flexible compute tiers** from burstable dev/test (B_Standard_B1ms) to production (GP_Standard_D16s_v3)
- **High availability** with automatic failover (ZoneRedundant or SameZone)
- **Private VNet access** using delegated subnets and private DNS zones
- **Multiple databases** on a single server instance

## Key Configuration

### Compute Tiers (sku_name)

| Tier | Prefix | Use Case | Example |
|------|--------|----------|---------|
| Burstable | `B_Standard_` | Dev/test, low-traffic | `B_Standard_B1ms` (1 vCPU, 2 GiB) |
| General Purpose | `GP_Standard_` | Production workloads | `GP_Standard_D2s_v3` (2 vCPU, 8 GiB) |
| Memory Optimized | `MO_Standard_` | Analytics, caching | `MO_Standard_E2s_v3` (2 vCPU, 16 GiB) |

### Storage (storage_mb)

Storage is provisioned in MB. Common sizes:

| Size | Value |
|------|-------|
| 32 GB | `32768` |
| 64 GB | `65536` |
| 128 GB | `131072` |
| 256 GB | `262144` |
| 512 GB | `524288` |
| 1 TB | `1048576` |

Storage cannot be downgraded once provisioned. Enable `auto_grow_enabled` for databases with unpredictable growth.

### Network Access

**Public access** (default): The server is accessible over the internet. Use `firewall_rules` to restrict which IP addresses can connect.

**Private VNet access**: Set `delegated_subnet_id` to a subnet delegated to `Microsoft.DBforPostgreSQL/flexibleServers`. Public access is automatically disabled. Optionally set `private_dns_zone_id` for DNS resolution within the VNet.

### High Availability

Set the `high_availability` block to enable automatic failover:

- **ZoneRedundant**: Standby in a different availability zone (recommended for production)
- **SameZone**: Standby in the same zone (faster failover, no zone-level protection)

Burstable SKUs do NOT support high availability.

### PostgreSQL Versions

Supported: 12, 13, 14, 15, 16 (default), 17

Version 16 is recommended for new deployments. Version 17 adds new features including cluster support.

## ForceNew Fields

Changing these fields destroys and recreates the server (data loss risk):

- `name` -- server hostname
- `administrator_login` -- admin username
- `delegated_subnet_id` -- VNet integration
- `geo_redundant_backup_enabled` -- backup configuration

## Stack Outputs

| Output | Description |
|--------|-------------|
| `server_id` | Azure Resource Manager ID |
| `server_name` | Server name |
| `fqdn` | Fully qualified domain name |
| `administrator_login` | Admin login name |
| `database_ids` | Map of database name to resource ID |

## Related Resources

- **AzureSubnet** -- Delegated subnet for VNet integration
- **AzurePrivateDnsZone** -- DNS resolution for private access
- **AzurePrivateEndpoint** -- Alternative private connectivity (via Private Link)
- **AzureResourceGroup** -- Container for the server

## Infra Chart Usage

This resource is a key component of:

- **database-stack** -- PostgreSQL + optional MSSQL + Redis + PrivateEndpoint
- **container-apps-environment** -- Optional database backend
- **web-app-environment** -- Optional database backend
