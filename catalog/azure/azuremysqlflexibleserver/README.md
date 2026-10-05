# AzureMysqlFlexibleServer

Azure Database for MySQL Flexible Server is Azure's managed MySQL:
per-server compute/storage sizing, zone-redundant high availability with
automatic failover, a Microsoft Entra (Azure AD) administrator,
customer-managed-key encryption, read replicas, and point-in-time restore.
Databases, firewall rules, server parameters, and the Entra administrator
are configured on the server and managed with it.

## When to Use

Use AzureMysqlFlexibleServer when you need:

- **Managed MySQL** with automatic patching, backups, and monitoring
- **Flexible compute tiers** from burstable dev/test (`B_Standard_B1ms`) to
  memory-optimized production (`MO_Standard_E4ds_v4`)
- **High availability** with a synchronously replicated standby
  (zone-redundant or same-zone) and automatic failover
- **Private networking** via VNet injection (delegated subnet + private DNS
  zone) or private endpoints
- **An Entra administrator** for token-based administration alongside
  password auth (MySQL cannot disable password auth)
- **Customer-managed-key encryption** composed against `AzureKeyVaultKey`
- **Read replicas and restores** -- a replica or restored server is another
  `AzureMysqlFlexibleServer` referencing the source's `server_id`

## Key Configuration

### Compute Tiers (`sku_name`)

| Tier | Prefix | Use Case | Example |
|------|--------|----------|---------|
| Burstable | `B_Standard_` | Dev/test, low-traffic | `B_Standard_B1ms` (1 vCPU, 2 GiB) |
| General Purpose | `GP_Standard_` | Production workloads | `GP_Standard_D4ds_v4` (4 vCPU, 16 GiB) |
| Memory Optimized | `MO_Standard_` | Analytics, caching | `MO_Standard_E4ds_v4` (4 vCPU, 32 GiB) |

Burstable SKUs support neither high availability nor read replicas.

### Storage (the `storage` block)

Capacity is `size_gb` (20-16384; only grows -- shrinking replaces the
server). IOPS comes in two mutually exclusive shapes: a provisioned `iops`
value (360-48000, bounded by SKU and size) or elastic scaling
(`io_scaling_enabled`) that rides workload demand. `auto_grow_enabled`
defaults to true (Azure's MySQL default); `log_on_disk_enabled` places the
slow query log on the server's own storage for compliance regimes that
require it.

### Networking

Two postures, matching Azure's contract:

- **Public endpoint** (Azure's derived default) -- allowlisted by
  `firewall_rules`
- **VNet injection** -- `delegated_subnet_id` (a subnet delegated to
  `Microsoft.DBforMySQL/flexibleServers`) + `private_dns_zone_id`; a
  VNet-injected server cannot have a public endpoint (leave
  `public_network_access` unset and Azure derives DISABLED)

### Authentication

- **Password auth is always on** -- MySQL Flexible Server cannot disable it
  (unlike PostgreSQL Flexible Server)
- **Microsoft Entra is additive**: declare the single `aad_administrator`
  (a group admits a team), backed by a user-assigned identity attached via
  `user_assigned_identity_ids`

### Lifecycle (`create_mode`)

| Mode | What it creates | Requires |
|------|-----------------|----------|
| DEFAULT (unset) | A fresh, empty server | `sku_name` + credentials |
| REPLICA | An asynchronous read replica | `source_server_id` |
| POINT_IN_TIME_RESTORE | Same-region restore | `source_server_id` + restore timestamp |
| GEO_RESTORE | Paired-region restore of the latest geo-backup | source with geo-redundant backups (no timestamp) |

Promote a replica by setting `replication_role: NONE` on it (irreversible).

### Versions

Supported: `5.7` (legacy migrations), `8.0.21` (the 8.0 series -- the
default and production standard), `8.4` (newest LTS). In-place upgrade goes
5.7 to 8.0.21 only; downgrades replace the server.

## Fields That Replace the Server When Changed

`server_name`, `region`, `resource_group`, `administrator_login` (once
set), `delegated_subnet_id`, `private_dns_zone_id`,
`geo_redundant_backup_enabled`, the restore/replica trio, a `version`
downgrade, and a `storage.size_gb` shrink.

## Outputs

| Output | Description |
|--------|-------------|
| `server_id` | ARM ID -- the seam for private endpoints and replica/restore sources |
| `server_name` | Server name |
| `fqdn` | `{name}.mysql.database.azure.com` -- resolves privately when VNet-injected |
| `administrator_login` | Admin login, echoed for connection strings |
| `database_ids` | Name-keyed map of database ARM IDs |
| `replica_capacity` | How many read replicas the server can still accept |

**Connection string format:**

```text
mysql://{administrator_login}:{password}@{fqdn}:3306/{database}?ssl-mode=REQUIRED
```

## Related Resources

- **AzureResourceGroup** -- the server's container
- **AzureSubnet** -- the delegated subnet for VNet injection
- **AzurePrivateDnsZone** -- private name resolution for the fqdn
- **AzurePrivateEndpoint** -- alternative private connectivity (Private Link)
- **AzureUserAssignedIdentity** -- CMK unwrap identity / the Entra administrator's backing identity
- **AzureKeyVaultKey** -- the customer-managed encryption key

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
