# AzureMssqlServer

**Azure SQL Database Logical Server** with bundled databases and firewall rules.

## Overview

`AzureMssqlServer` provisions an Azure SQL Database logical server -- the managed
endpoint for Azure SQL databases. Unlike PostgreSQL and MySQL Flexible Servers
(which are physical servers carrying their own compute and storage), Azure SQL uses
a **logical server + database** architecture:

- The **server** is an administrative container that defines the login endpoint,
  firewall rules, TLS policy, and connection policy. It has no compute or storage.
- Each **database** carries its own compute SKU (DTU-based or vCore-based),
  maximum storage size, zone redundancy, and license type. Databases are billed
  independently.

This component bundles the server with its databases and firewall rules per DD03
(Composite Bundling Rules), because a server without at least one database and a
connection path has no practical utility.

## When to Use

- **Relational SQL workloads** that require T-SQL, stored procedures, or SQL Server
  compatibility features not available in PostgreSQL/MySQL
- **Enterprise applications** migrating from on-premises SQL Server
- **Cost-optimized deployments** using Azure Hybrid Benefit (bring your own
  SQL Server license for up to 55% savings)
- **Multi-database architectures** where each database needs independent compute
  and storage sizing

## Key Configuration

### Server-Level Settings

| Field | Description | Default |
|-------|-------------|---------|
| `version` | SQL Server version (`"12.0"` or `"2.0"`) | `"12.0"` |
| `minimum_tls_version` | Minimum TLS for connections | `"1.2"` |
| `public_network_access_enabled` | Allow public internet access | `true` |
| `connection_policy` | How clients connect (`Default`, `Proxy`, `Redirect`) | `"Default"` |

### Database-Level Settings

Each database in the `databases` list has its own compute and storage:

| Field | Description | Default |
|-------|-------------|---------|
| `sku_name` | Compute tier (e.g., `"S0"`, `"GP_Gen5_2"`, `"BC_Gen5_4"`) | Required |
| `max_size_gb` | Maximum storage in GB | SKU default |
| `collation` | Sort order and string comparison | `SQL_Latin1_General_CP1_CI_AS` |
| `zone_redundant` | Spread replicas across availability zones | `false` |
| `license_type` | `BasePrice` (Hybrid Benefit) or `LicenseIncluded` | `LicenseIncluded` |
| `storage_account_type` | Backup redundancy (`Geo`, `Local`, `Zone`, `GeoZone`) | `"Geo"` |

### Network Access

Azure SQL does **not** support VNet delegation. Private connectivity is exclusively
through `AzurePrivateEndpoint` (subresource: `sqlServer`). The
`public_network_access_enabled` boolean controls public endpoint visibility.

### Connection Policy

- **Default**: Redirect for Azure-internal connections, Proxy for external
- **Redirect**: Lower latency, requires ports 11000-11999 in addition to 1433
- **Proxy**: All traffic through the gateway, simpler firewall rules

For Azure-to-Azure workloads, `Redirect` gives best performance.

## Outputs

| Output | Description |
|--------|-------------|
| `server_id` | Azure Resource Manager ID (used by AzurePrivateEndpoint) |
| `server_name` | Server name |
| `fqdn` | `{name}.database.windows.net` -- for connection strings |
| `administrator_login` | Admin login for connection strings |
| `database_ids` | Map of database name to ARM resource ID |

## Infra Chart Usage

This resource participates in the **database-stack** infra chart as the SQL Server
option alongside PostgreSQL and MySQL. The `server_id` output is referenced by
`AzurePrivateEndpoint` for private connectivity within the stack.

## Related Resources

- **AzurePrivateEndpoint** -- Private connectivity (subresource_names: `["sqlServer"]`)
- **AzureResourceGroup** -- Container for the server and its resources
- **AzurePostgresqlFlexibleServer** -- PostgreSQL alternative (different architecture)
- **AzureMysqlFlexibleServer** -- MySQL alternative (different architecture)
