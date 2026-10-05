# AzurePostgresqlFlexibleServer

Azure Database for PostgreSQL Flexible Server is Azure's managed PostgreSQL:
per-server compute/storage sizing, zone-redundant high availability with
automatic failover, Microsoft Entra (Azure AD) authentication,
customer-managed-key encryption, read replicas, and point-in-time restore.
Databases, firewall rules, server parameters, and Entra administrators are
configured on the server and managed with it.

## When to Use

Use AzurePostgresqlFlexibleServer when you need:

- **Managed PostgreSQL** with automatic patching, backups, and monitoring
- **Flexible compute tiers** from burstable dev/test (`B_Standard_B1ms`) to
  memory-optimized production (`MO_Standard_E64ds_v5`)
- **High availability** with a synchronously replicated standby
  (zone-redundant or same-zone) and automatic failover
- **Private networking** via VNet injection (delegated subnet + private DNS
  zone) or private endpoints
- **Entra-only authentication** to eliminate static database passwords
- **Customer-managed-key encryption** composed against `AzureKeyVaultKey`
- **Read replicas and restores** -- a replica or restored server is another
  `AzurePostgresqlFlexibleServer` referencing the source's `server_id`

## Key Configuration

### Compute Tiers (`sku_name`)

| Tier | Prefix | Use Case | Example |
|------|--------|----------|---------|
| Burstable | `B_Standard_` | Dev/test, low-traffic | `B_Standard_B1ms` (1 vCPU, 2 GiB) |
| General Purpose | `GP_Standard_` | Production workloads | `GP_Standard_D4ds_v5` (4 vCPU, 16 GiB) |
| Memory Optimized | `MO_Standard_` | Analytics, caching | `MO_Standard_E4s_v3` (4 vCPU, 32 GiB) |

Burstable SKUs support neither high availability nor read replicas.

### Storage (`storage_mb` + `storage_tier`)

Storage comes from a fixed ladder (32768 = 32 GiB up to 33553408 = 32 TiB)
and only grows -- shrinking replaces the server. Each size has a default
performance tier (IOPS class) and a bounded valid range; set `storage_tier`
to buy more IOPS without growing capacity. Enable `auto_grow_enabled` for
databases with unpredictable growth.

### Networking

Two independent dials, matching Azure's contract:

- **`public_network_access_enabled`** (default true) -- the public endpoint,
  allowlisted by `firewall_rules`
- **VNet injection** -- `delegated_subnet_id` (a subnet delegated to
  `Microsoft.DBforPostgreSQL/flexibleServers`) + `private_dns_zone_id`;
  requires public access explicitly off

### Authentication

- **Password auth** (default): `administrator_login` + `administrator_password`
- **Microsoft Entra**: enable `authentication.active_directory_auth_enabled`
  and declare `aad_administrators`; disable password auth entirely for an
  Entra-only posture (credentials must then be omitted)

### Lifecycle (`create_mode`)

| Mode | What it creates | Requires |
|------|-----------------|----------|
| DEFAULT (unset) | A fresh, empty server | `sku_name` (+ credentials with password auth) |
| REPLICA | An asynchronous read replica | `source_server_id` |
| POINT_IN_TIME_RESTORE | Same-region restore | `source_server_id` + restore timestamp |
| GEO_RESTORE | Paired-region restore | source with geo-redundant backups + timestamp |
| REVIVE_DROPPED | Revive a soft-deleted server | `source_server_id` |

Promote a replica by setting `replication_role: NONE` on it (irreversible).

### Versions

Supported: 11-18. Default 16. In-place upgrades go up only; elastic
clusters (`cluster` block, sharded/citus-based) require 17+.

## Fields That Replace the Server When Changed

`server_name`, `region`, `resource_group`, `administrator_login` (once
set), `delegated_subnet_id`, `geo_redundant_backup_enabled`,
`customer_managed_key`, `cluster`, the restore/replica trio -- plus any
`version` downgrade or `storage_mb` shrink.

## Outputs

| Output | Description |
|--------|-------------|
| `server_id` | ARM ID -- the seam for private endpoints and replica/restore sources |
| `server_name` | Server name |
| `fqdn` | `{name}.postgres.database.azure.com` -- resolves privately when VNet-injected |
| `administrator_login` | Admin login (empty on Entra-only servers) |
| `database_ids` | Name-keyed map of database ARM IDs |
| `identity_principal_id` | System-assigned identity's principal -- the role-assignment seam |

## Related Resources

- **AzureResourceGroup** -- the server's container
- **AzureSubnet** -- the delegated subnet for VNet injection
- **AzurePrivateDnsZone** -- private name resolution for the fqdn
- **AzurePrivateEndpoint** -- alternative private connectivity (Private Link)
- **AzureUserAssignedIdentity** -- CMK unwrap identity / Entra administrators
- **AzureKeyVaultKey** -- the customer-managed encryption key

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
