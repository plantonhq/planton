# AzureMysqlFlexibleServer - Research & Design Documentation

## Deployment Landscape

### MySQL Managed Options on Azure

Azure has offered several managed MySQL deployment models over time. As of 2025, the landscape
has consolidated around a single recommended option:

| Service | Status | Notes |
|---------|--------|-------|
| Azure Database for MySQL - Single Server | **Retired** (Sept 2024) | Legacy, no new deployments, migrating to Flexible |
| Azure Database for MySQL - Flexible Server | **Active (recommended)** | Current-generation, full feature set |
| MySQL on Azure VMs | Active (IaaS) | Self-managed, full control, operational burden |
| Azure Database for MySQL - HeatWave | Preview | Analytics acceleration via Oracle HeatWave integration |

**Single Server** was Azure's first-generation managed MySQL offering. It reached end-of-life
in September 2024, and all existing Single Server instances are being migrated to Flexible Server.
No new Single Server deployments are possible.

**Flexible Server** is Microsoft's current-generation managed MySQL service and the only recommended
option for new deployments. It provides zone-redundant HA, burstable compute, VNet integration,
and finer control over maintenance windows and server parameters than its predecessor.

### Azure Database for MySQL Flexible Server Overview

Azure Database for MySQL Flexible Server is built on the MySQL Community Edition. It provides:

- **Managed infrastructure**: Automatic OS and MySQL patching, backups, monitoring
- **Compute flexibility**: Burstable, General Purpose, and Memory Optimized tiers
- **Storage flexibility**: 20 GB to 16 TB with auto-grow capability
- **High availability**: Zone-redundant and same-zone automatic failover
- **Networking**: Public access with firewall rules or private VNet integration
- **Security**: SSL/TLS enforcement, Azure AD authentication (optional), data encryption at rest
- **Monitoring**: Azure Monitor integration, slow query logs, audit logs
- **MySQL versions**: 5.7, 8.0.21, and 8.4

The service handles backup, restore, patching, scaling, and failover automatically. Users
configure compute tier, storage, networking, and HA mode. Runtime parameters (e.g., `max_connections`,
`innodb_buffer_pool_size`) can be tuned post-deployment through server configurations.

### Comparison: Single Server vs Flexible Server

| Feature | Single Server (retired) | Flexible Server |
|---------|------------------------|-----------------|
| High availability | No zone-redundant option | ZoneRedundant + SameZone |
| Compute | Fixed tiers | Burstable + GP + MO |
| VNet integration | Private Link only | Delegated subnets + Private Link |
| Stop/Start | Not supported | Supported (cost savings) |
| Maintenance window | Azure-controlled | User-configurable |
| Storage auto-grow | Available | Available (default: enabled) |
| Max storage | 16 TB | 16 TB |
| MySQL versions | 5.7, 8.0 | 5.7, 8.0.21, 8.4 |
| Encryption (CMK) | Available | Available |
| Read replicas | Up to 5 | Up to 10 |

Flexible Server is strictly superior. The only reason to reference Single Server knowledge
is for migration scenarios.

### Deployment Methods Compared

| Method | Strengths | Weaknesses |
|--------|-----------|------------|
| Azure Portal | Visual, guided, immediate | Manual, not repeatable |
| Azure CLI | Scriptable, CI/CD friendly | Imperative, state management burden |
| ARM Templates | Declarative, Azure-native | Verbose JSON, complex for multi-resource |
| Terraform (`azurerm`) | Declarative, state management, mature | HCL learning curve |
| Pulumi (azure classic) | Declarative, general-purpose languages | Smaller community than Terraform |
| Planton | Opinionated defaults, infra-chart composability | Opinionated (by design) |

### Why Planton

Planton's AzureMysqlFlexibleServer component provides:

1. **Opinionated defaults** -- Password auth enabled, `auto_grow_enabled: true`, `utf8mb4` charset, Standard create mode
2. **Infra-chart composability** -- StringValueOrRef fields enable wiring to subnets, DNS zones, resource groups
3. **Bundled sub-resources** -- Server + databases + firewall rules managed as a unit
4. **Dual IaC** -- Both Pulumi and Terraform modules with feature parity

## Compute Tiers Deep Dive

### Burstable (B_Standard_*)

Designed for workloads that don't need full CPU continuously. The server accumulates CPU credits
during idle periods and bursts above baseline when needed.

| SKU | vCPU | Memory (GiB) | Max IOPS | Use Case |
|-----|------|-------------|----------|----------|
| B_Standard_B1ms | 1 | 2 | 640 | Dev/test, micro-services |
| B_Standard_B2ms | 2 | 4 | 1280 | Small apps, CI databases |
| B_Standard_B4ms | 4 | 8 | 2560 | Light production |
| B_Standard_B8ms | 8 | 16 | 3200 | Medium dev workloads |
| B_Standard_B12ms | 12 | 24 | 3840 | Larger dev/staging |
| B_Standard_B16ms | 16 | 32 | 4096 | Heavy dev/test |
| B_Standard_B20ms | 20 | 40 | 5000 | Upper-tier burstable |

**Limitations**: No high availability support. Not recommended for sustained CPU workloads.

### General Purpose (GP_Standard_*)

Balanced compute-to-memory ratio for most production workloads. Provides consistent
performance without CPU credit mechanics.

| SKU | vCPU | Memory (GiB) | Max IOPS | Use Case |
|-----|------|-------------|----------|----------|
| GP_Standard_D2ds_v4 | 2 | 8 | 3200 | Small production |
| GP_Standard_D4ds_v4 | 4 | 16 | 6400 | Standard production |
| GP_Standard_D8ds_v4 | 8 | 32 | 12800 | Medium production |
| GP_Standard_D16ds_v4 | 16 | 64 | 20000 | Large production |
| GP_Standard_D32ds_v4 | 32 | 128 | 20000 | Very large production |
| GP_Standard_D48ds_v4 | 48 | 192 | 20000 | Enterprise |
| GP_Standard_D64ds_v4 | 64 | 256 | 20000 | Enterprise |

The v5 series (`GP_Standard_D2ads_v5` through `GP_Standard_D64ads_v5`) offers improved
price-performance with AMD processors and is available in select regions.

### Memory Optimized (MO_Standard_*)

High memory-to-compute ratio for workloads that benefit from large buffer pools,
caching, or in-memory analytics.

| SKU | vCPU | Memory (GiB) | Max IOPS | Use Case |
|-----|------|-------------|----------|----------|
| MO_Standard_E2ds_v4 | 2 | 16 | 3200 | Small analytics |
| MO_Standard_E4ds_v4 | 4 | 32 | 6400 | Standard analytics |
| MO_Standard_E8ds_v4 | 8 | 64 | 12800 | Large analytics |
| MO_Standard_E16ds_v4 | 16 | 128 | 20000 | Enterprise analytics |
| MO_Standard_E32ds_v4 | 32 | 256 | 20000 | Heavy analytics |
| MO_Standard_E48ds_v4 | 48 | 384 | 20000 | Extreme memory |
| MO_Standard_E64ds_v4 | 64 | 504 | 20000 | Maximum memory |

The v5 series (`MO_Standard_E2ads_v5` through `MO_Standard_E96ads_v5`) is available
in select regions with AMD EPYC processors.

## Storage Architecture

### Provisioned Storage

MySQL Flexible Server uses Azure Premium SSD-based storage measured in **gigabytes** (GB).
This differs from PostgreSQL Flexible Server which uses megabytes (MB).

- **Minimum**: 20 GB
- **Maximum**: 16,384 GB (16 TB)
- **Granularity**: 1 GB increments
- **Cannot be downgraded**: Decreasing `storage_size_gb` forces server recreation

### IOPS

IOPS scale with storage size. Azure provisions a baseline of 3 IOPS per GB of provisioned storage,
with a minimum of 360 IOPS and a maximum that depends on compute tier:

| Storage Size | Baseline IOPS | Notes |
|-------------|---------------|-------|
| 20 GB | 360 | Minimum baseline |
| 100 GB | 360 | Still at minimum |
| 200 GB | 600 | 3 IOPS/GB |
| 500 GB | 1500 | 3 IOPS/GB |
| 1 TB (1024 GB) | 3072 | 3 IOPS/GB |
| 4 TB (4096 GB) | 12288 | 3 IOPS/GB |
| 16 TB (16384 GB) | 20000 | Capped by tier maximum |

Additional IOPS can be pre-provisioned beyond the baseline for workloads that need
consistent high I/O performance, though this is not exposed in the Planton v1 spec (80/20).

### Auto-Grow

When `auto_grow_enabled` is `true` (the default for MySQL, unlike PostgreSQL where it defaults
to `false`), Azure automatically increases storage when free storage falls below a threshold:

- Triggers when free storage drops below 10% or 10 GB (whichever is smaller)
- Increases storage by 10% or 5 GB, whichever is greater
- New storage is provisioned without server restart
- Cannot be disabled once the server exceeds certain storage thresholds

This is a critical safety net for production databases with unpredictable growth patterns.

## High Availability Modes

### ZoneRedundant

- Primary and standby servers in **different** availability zones
- Protects against zone-level failures (power, cooling, network)
- Failover time: typically 60-120 seconds
- Data synchronization: synchronous replication
- Recommended for all production workloads

**How it works:**
1. Primary serves all read/write traffic
2. Standby maintains a synchronous copy via MySQL binary log replication
3. On primary failure, Azure detects via health monitoring (typically within 30s)
4. Automatic DNS failover points FQDN to the standby (now promoted to primary)
5. A new standby is provisioned in the original primary's zone

### SameZone

- Primary and standby servers in the **same** availability zone
- Protects against server-level failures (hardware, OS)
- Does NOT protect against zone-level failures
- Failover time: typically 30-60 seconds (faster than ZoneRedundant)
- Lower cost than ZoneRedundant (same-zone data transfer)

### HA Requirements and Limitations

- **Burstable SKUs do NOT support HA** -- minimum GP_Standard_D2ds_v4
- HA doubles the compute cost (standby uses identical SKU)
- HA can be enabled/disabled after server creation
- When using ZoneRedundant, the `zone` and `standby_availability_zone` should differ
- Geo-redundant backup is independent of HA (they protect against different failure modes)

## Networking

### Public Access (Default)

When `delegated_subnet_id` is NOT set, the server is accessible over the internet:

- Server has a public IP address
- FQDN resolves to the public IP: `{name}.mysql.database.azure.com`
- All connections blocked by default
- `firewall_rules` whitelist specific IP ranges
- Special rule `0.0.0.0`-`0.0.0.0` allows connections from all Azure services
- SSL/TLS is enforced by default (configurable post-deployment)

### VNet-Integrated (Private Access)

When `delegated_subnet_id` is set, the server is deployed into a VNet:

- Server gets a private IP in the delegated subnet
- Public access is automatically disabled (no public IP)
- The subnet MUST be delegated to `Microsoft.DBforMySQL/flexibleServers`
- Only one Flexible Server per delegated subnet (Azure limitation)
- `private_dns_zone_id` enables FQDN resolution to the private IP
- DNS zone is typically `privatelink.mysql.database.azure.com`

**Subnet delegation requirements:**
```
service_delegation_name: Microsoft.DBforMySQL/flexibleServers
actions:
  - Microsoft.Network/virtualNetworks/subnets/join/action
```

**Recommended subnet size**: `/28` (16 IPs) minimum, `/24` (256 IPs) recommended.
Azure reserves several IPs in each subnet for internal use.

### Private Endpoint (Alternative)

For servers in public access mode, a Private Endpoint can provide private connectivity
without VNet integration:

- Creates a private IP in a VNet that tunnels to the server's public endpoint
- Does NOT require subnet delegation
- Can coexist with firewall rules (both public and private access)
- Managed via the `AzurePrivateEndpoint` component referencing `server_id`

VNet integration is preferred for new deployments. Private Endpoint is useful for
connecting from VNets that can't use subnet delegation (e.g., shared VNets).

## 80/20 Scoping Rationale

### Included (covers 80%+ of production use cases)

- **Standard create mode** -- New server creation (most common)
- **Password authentication** -- Default and most widely used auth method
- **Storage provisioning** -- All Azure-supported sizes from 20 GB to 16 TB
- **High availability** -- Both ZoneRedundant and SameZone modes
- **VNet integration** -- Delegated subnet + private DNS zone
- **Multiple databases** -- With charset/collation customization
- **Firewall rules** -- IP-based access control for public mode
- **Backup configuration** -- Retention days (1-35) and geo-redundant backup
- **Auto-grow storage** -- Automatic storage scaling (default: enabled)

### Excluded (advanced/niche features deferred to v2)

- **Azure AD authentication** -- Requires tenant configuration, service principal setup. Can be enabled post-deployment via Azure portal.
- **Customer-managed encryption keys** -- Requires Key Vault with specific access policies. Enterprise-only feature.
- **Point-in-time restore / Replica creation** -- Uses different `create_mode` values. Restore operations are typically one-off, not declarative IaC.
- **Server configurations** (e.g., `max_connections`, `innodb_buffer_pool_size`) -- Runtime tuning done post-deployment. Azure provides defaults appropriate for the SKU.
- **Maintenance window** -- Terraform provider limitation (cannot set on create). Configure post-deployment.
- **Read replicas** -- Cross-region read scaling. Managed separately from primary.
- **Pre-provisioned IOPS** -- Beyond baseline IOPS scaling. For performance-critical workloads, tune post-deployment.
- **HeatWave integration** -- Oracle HeatWave for analytics acceleration. Preview feature, not yet GA.
- **Slow query log / Audit log configuration** -- Diagnostics settings managed post-deployment.
- **Data-in transit encryption configuration** -- SSL/TLS is enforced by default. Custom certificates are an advanced scenario.

## MySQL Version Strategy

### Version 5.7 (Legacy)

- **Status**: Approaching end-of-life (MySQL community EOL was October 2023; Azure extended support)
- **Use case**: Legacy migration only -- applications that cannot be upgraded to 8.0+
- **Recommendation**: Migrate to 8.0.21 or 8.4 as soon as possible
- **Collation note**: Use `utf8mb4_unicode_ci` instead of `utf8mb4_0900_ai_ci` (not available in 5.7)

### Version 8.0.21 (Default / Production Standard)

- **Status**: Active, widely deployed, well-tested
- **Use case**: Most production workloads
- **Why 8.0.21**: This is Azure's specific version string for the MySQL 8.0 line on Flexible Server. It includes all 8.0.x improvements through this patch level.
- **Key features**: CTEs, window functions, JSON enhancements, default `utf8mb4` charset, `caching_sha2_password` default auth plugin
- **Recommendation**: Default choice for new deployments

### Version 8.4 (Latest GA)

- **Status**: Latest GA release on Azure Flexible Server
- **Use case**: New deployments that want the latest features and performance improvements
- **Key features**: Performance schema improvements, information schema enhancements, improved optimizer, InnoDB improvements
- **Recommendation**: Preferred for greenfield projects; test thoroughly before upgrading existing workloads

### Version Selection Decision Tree

```
Is the application already running MySQL 5.7?
├── Yes → Use "5.7" initially, plan migration to "8.0.21"
└── No → Is this a greenfield project?
    ├── Yes → Use "8.4" (latest features)
    └── No → Use "8.0.21" (proven stability)
```

## Provider Research

### Terraform Provider (`azurerm`)

The `azurerm_mysql_flexible_server` resource supports approximately 25 top-level fields.
Key findings from provider source analysis:

**Storage model:**
- `storage` block with `size_gb` (integer, in GB) and `auto_grow_enabled` (bool, default true)
- `iops` and `io_scaling_enabled` are available but not exposed in Planton v1 (80/20)
- Storage tier is derived from size automatically

**Authentication model:**
- `authentication` is not a top-level block like PostgreSQL; password auth is implicit
- Azure AD authentication can be configured post-deployment

**Network model:**
- `delegated_subnet_id` and `private_dns_zone_id` are independent (both optional)
- `public_network_access_enabled` defaults based on subnet presence
- We derive public access from the presence of `delegated_subnet_id`

**ForceNew fields** (resource recreation on change):
- `name`, `administrator_login`, `delegated_subnet_id`, `private_dns_zone_id`, `geo_redundant_backup_enabled`
- Note: `private_dns_zone_id` is ForceNew for MySQL but NOT for PostgreSQL

### Pulumi Provider

Uses `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mysql` (classic provider),
consistent with all other Azure modules in Planton.

Key types:
- `mysql.FlexibleServer` -- Main server resource
- `mysql.FlexibleDatabase` -- Database resource
- `mysql.FlexibleServerFirewallRule` -- Firewall rule resource
- `mysql.FlexibleServerArgs` -- Server constructor arguments

## Design Decisions Applied

### C1-C2: Required resource_group and region (DD05 compliance)
Every Azure resource in Planton requires `resource_group` (StringValueOrRef) and `region` (string).

### C3: String+CEL for version (not proto enum)
Following the established pattern, version uses string with CEL `in` validation.
Valid values: `"5.7"`, `"8.0.21"`, `"8.4"`. Default: `"8.0.21"`.

### C4: Optional HA message (not bool+enum)
If the `high_availability` message is present, HA is enabled. No separate boolean needed.
Mode uses string+CEL with Azure's exact values ("ZoneRedundant", "SameZone").

### C5: Storage in GB (not MB)
Unlike PostgreSQL which uses `storage_mb`, MySQL uses `storage_size_gb` because:
- The underlying Azure resource API uses GB for MySQL
- The Terraform provider exposes `size_gb` in the storage block
- GB is more intuitive for users (no need to calculate MB equivalents)

### C6: auto_grow_enabled defaults to true
Azure's MySQL provider defaults `auto_grow_enabled` to `true`, unlike PostgreSQL which defaults
to `false`. We match the Azure default to avoid surprising behavior. This is a safer default
since running out of storage is a more catastrophic failure than unexpected storage growth.

### C7: Backup retention range 1-35 days
MySQL Flexible Server supports backup retention from 1 to 35 days, broader than PostgreSQL's
7-35 day range. The default is 7 days, matching the Azure default.

### C8: Repeated databases (not initial_database_name)
Changed from a single string to `repeated AzureMysqlDatabase` following the LB backend_pools
pattern. Supports multiple databases with custom charset/collation. Default charset is
`utf8mb4` with collation `utf8mb4_0900_ai_ci`.

### C9: Polymorphic StringValueOrRef for password
No `default_kind` annotation since the password source varies (literal, chart variable,
external secret).

### C10: Hardcoded public_network_access logic
Not exposed as a spec field. IaC modules derive from `delegated_subnet_id` presence:
- Subnet set -> public access disabled
- Subnet not set -> public access enabled

### C11: private_dns_zone_id is ForceNew
Unlike PostgreSQL where `private_dns_zone_id` can be changed, MySQL treats this as ForceNew.
This is documented prominently because it's a subtle difference that can cause data loss.

### C12: Server name can start with a number
Unlike PostgreSQL where server names must start with a letter, MySQL Flexible Server allows
names starting with a number. The CEL validation pattern is `'^[a-z0-9][a-z0-9-]*[a-z0-9]$'`.

## Best Practices for Production

### Compute Sizing

1. **Start with GP_Standard_D2ds_v4** for most workloads and scale up based on metrics
2. **Monitor CPU utilization** -- if consistently >70%, scale up
3. **Monitor memory utilization** -- if buffer pool hit ratio drops below 99%, consider Memory Optimized
4. **Use Burstable ONLY for dev/test** -- not suitable for production workloads
5. **Size for peak load** -- MySQL Flexible Server supports online scaling but it causes a brief connection drop

### Storage

1. **Provision 2-3x current data size** for growth headroom
2. **Keep auto_grow_enabled = true** as a safety net
3. **Monitor storage usage** -- Azure alerts when approaching limits
4. **Remember: storage cannot shrink** -- over-provisioning has a cost but under-provisioning risks downtime

### High Availability

1. **Always enable ZoneRedundant HA for production** -- the cost of standby is insurance
2. **Use SameZone only when** zone-level outages are not a concern (e.g., non-critical internal apps)
3. **Set explicit zone and standby_availability_zone** for predictable placement
4. **Test failover** before production launch -- Azure supports planned failover for testing

### Networking

1. **Use VNet integration for production** -- never expose production databases to the internet
2. **Use a dedicated /24 subnet** for each MySQL server
3. **Set up private DNS zone** for reliable name resolution
4. **If public access needed (dev/test)**, use specific IP allowlists, never `0.0.0.0/0`

### Backup and DR

1. **Set backup_retention_days to 35** for production (maximum retention)
2. **Enable geo_redundant_backup** if RPO requirements span regions
3. **Remember geo-redundant backup is ForceNew** -- enable at creation time
4. **Test restore regularly** -- backups are only useful if restore works

### Security

1. **Use a RandomPassword resource** for the admin password (never hardcode in specs)
2. **Enable SSL/TLS** (enforced by default, don't disable it)
3. **Restrict firewall rules** to the minimum necessary IP ranges
4. **Use Azure AD auth post-deployment** for application-level access control
5. **Rotate admin credentials** periodically

## Connection String Format

Applications connect to MySQL Flexible Server using standard MySQL connection strings:

```
mysql://{administrator_login}:{password}@{fqdn}:3306/{database}?ssl-mode=REQUIRED
```

**Components:**

| Part | Source | Example |
|------|--------|---------|
| `administrator_login` | Spec field / Stack output | `mysqladmin` |
| `password` | Spec field (administrator_password) | `Pr0dS3cur3P@ss!` |
| `fqdn` | Stack output | `myapp-prod-mysql.mysql.database.azure.com` |
| `port` | Always `3306` | `3306` |
| `database` | Database name from spec | `myapp` |
| `ssl-mode` | Always `REQUIRED` (default) | `REQUIRED` |

**Language-specific connection examples:**

```
# Python (mysql-connector-python)
mysql.connector.connect(
    host="myapp-prod-mysql.mysql.database.azure.com",
    user="mysqladmin",
    password="...",
    database="myapp",
    ssl_verify_cert=True,
    port=3306
)

# Node.js (mysql2)
mysql.createConnection({
    host: "myapp-prod-mysql.mysql.database.azure.com",
    user: "mysqladmin",
    password: "...",
    database: "myapp",
    ssl: { rejectUnauthorized: true },
    port: 3306
})

# JDBC (Java)
jdbc:mysql://myapp-prod-mysql.mysql.database.azure.com:3306/myapp?useSSL=true&requireSSL=true

# Go (go-sql-driver/mysql)
mysqladmin:...@tcp(myapp-prod-mysql.mysql.database.azure.com:3306)/myapp?tls=true
```

## Differences from PostgreSQL Flexible Server in Azure

Understanding the differences between MySQL and PostgreSQL Flexible Server in Azure
is important for teams managing both, and for ensuring the Planton components
accurately reflect each service's behavior.

### API and Resource Differences

| Aspect | MySQL Flexible Server | PostgreSQL Flexible Server |
|--------|----------------------|---------------------------|
| Resource provider | `Microsoft.DBforMySQL` | `Microsoft.DBforPostgreSQL` |
| Terraform resource | `azurerm_mysql_flexible_server` | `azurerm_postgresql_flexible_server` |
| Pulumi module | `mysql.FlexibleServer` | `postgresql.FlexibleServer` |
| Storage field | `storage.size_gb` (GB) | `storage_mb` (MB) |
| Storage minimum | 20 GB | 32 GB (32768 MB) |
| Storage maximum | 16 TB | 32 TB |
| Default port | 3306 | 5432 |
| FQDN suffix | `.mysql.database.azure.com` | `.postgres.database.azure.com` |

### Behavioral Differences

| Behavior | MySQL | PostgreSQL |
|----------|-------|------------|
| auto_grow_enabled default | `true` | `false` |
| Backup retention range | 1-35 days | 7-35 days |
| private_dns_zone_id ForceNew | Yes | No |
| Server name pattern | Can start with number | Must start with letter |
| Default charset | `utf8mb4` | `UTF8` |
| Default collation | `utf8mb4_0900_ai_ci` | `en_US.utf8` |
| Authentication block | Implicit (password) | Explicit block |
| SKU families | `_v4`, `_v5` series | `_v3` series |

### Version Differences

| MySQL Versions | PostgreSQL Versions |
|---------------|-------------------|
| 5.7 (approaching EOL) | 12 (approaching EOL) |
| 8.0.21 (default) | 13, 14, 15 |
| 8.4 (latest) | 16 (default) |
| -- | 17 (latest, with cluster support) |

### Networking Differences

Both services support the same networking models (public with firewall, VNet-integrated,
Private Endpoint), but the subnet delegation services differ:

- MySQL: `Microsoft.DBforMySQL/flexibleServers`
- PostgreSQL: `Microsoft.DBforPostgreSQL/flexibleServers`

Private DNS zone names also differ:
- MySQL: `privatelink.mysql.database.azure.com`
- PostgreSQL: `privatelink.postgres.database.azure.com`

### Planton Spec Differences

| Spec Field | MySQL | PostgreSQL |
|-----------|-------|------------|
| Storage field name | `storage_size_gb` | `storage_mb` |
| Storage type | `int32` (GB) | `int32` (MB) |
| auto_grow default | `"true"` | `"false"` |
| Version values | `"5.7"`, `"8.0.21"`, `"8.4"` | `"12"`-`"17"` |
| Version default | `"8.0.21"` | `"16"` |
| Backup retention min | 1 | 7 |
| Database charset default | `utf8mb4` | `UTF8` |
| Database collation default | `utf8mb4_0900_ai_ci` | `en_US.utf8` |

## Infra Chart Integration

### database-stack chart pattern

```
AzureResourceGroup
└── AzureSubnet (delegated to MySQL)
    └── AzurePrivateDnsZone (privatelink.mysql.database.azure.com)
        └── AzureMysqlFlexibleServer
            ├── delegated_subnet_id: valueFrom AzureSubnet
            ├── private_dns_zone_id: valueFrom AzurePrivateDnsZone
            └── resource_group: valueFrom AzureResourceGroup
```

### Referenced by
- **AzurePrivateEndpoint** -- `private_connection_resource_id` references `server_id` (for non-VNet-integrated servers)

### References
- **AzureResourceGroup** -- `resource_group` (required)
- **AzureSubnet** -- `delegated_subnet_id` (optional, for VNet integration)
- **AzurePrivateDnsZone** -- `private_dns_zone_id` (optional, for private DNS)
