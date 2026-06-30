# AzurePostgresqlFlexibleServer - Research & Design Documentation

## Deployment Landscape

### Azure Database for PostgreSQL: Flexible Server

Azure Database for PostgreSQL Flexible Server is Microsoft's current-generation managed PostgreSQL offering. It replaced the previous "Single Server" deployment option, which reached end-of-life and is being retired. Flexible Server provides more granular control over compute, storage, networking, and high availability than its predecessor.

**Key differentiators from Single Server:**
- Zone-redundant and same-zone high availability
- Burstable compute tier for dev/test
- VNet integration via delegated subnets (not just Private Link)
- Maintenance window control
- Storage auto-grow
- Customer-managed encryption keys
- PostgreSQL 12 through 17 support

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

Planton's AzurePostgresqlFlexibleServer component provides:

1. **Opinionated defaults** -- Password auth enabled, sensible storage defaults, Standard create mode
2. **Infra-chart composability** -- StringValueOrRef fields enable wiring to subnets, DNS zones, resource groups
3. **Bundled sub-resources** -- Server + databases + firewall rules managed as a unit
4. **Dual IaC** -- Both Pulumi and Terraform modules with feature parity

## 80/20 Scoping Rationale

### Included (covers 80%+ of production use cases)

- **Standard create mode** -- New server creation (most common)
- **Password authentication** -- Default and most widely used auth method
- **Storage provisioning** -- All Azure-supported sizes from 32 GB to 32 TB
- **High availability** -- Both ZoneRedundant and SameZone modes
- **VNet integration** -- Delegated subnet + private DNS zone
- **Multiple databases** -- With charset/collation customization
- **Firewall rules** -- IP-based access control for public mode
- **Backup configuration** -- Retention days (7-35) and geo-redundant backup
- **Auto-grow storage** -- Automatic storage scaling

### Excluded (advanced/niche features deferred to v2)

- **Azure AD authentication** -- Requires tenant configuration, service principal setup. Can be enabled post-deployment via Azure portal.
- **Customer-managed encryption keys** -- Requires Key Vault with specific access policies. Enterprise-only feature.
- **Point-in-time restore / Replica creation** -- Uses different `create_mode` values. Restore operations are typically one-off, not declarative IaC.
- **Server configurations** (e.g., `max_connections`, `work_mem`) -- Runtime tuning done post-deployment. Azure provides defaults appropriate for the SKU.
- **Maintenance window** -- Terraform provider limitation (cannot set on create). Configure post-deployment.
- **Cluster feature** (PostgreSQL 17+) -- New feature for horizontal read scaling. Requires cluster-aware application design.
- **Read replicas** -- Cross-region read scaling. Managed separately from primary.

## Provider Research

### Terraform Provider (`azurerm`)

The `azurerm_postgresql_flexible_server` resource (API version 2025-08-01) supports approximately 25 top-level fields. Key findings from provider source analysis:

**Storage model:**
- `storage_mb` is in MB (not GB) with specific allowed values
- `storage_tier` is separate but defaults based on `storage_mb`
- We expose `storage_mb` only (80/20) and let Azure choose the tier

**Authentication model:**
- `authentication` block with `password_auth_enabled` and `active_directory_auth_enabled`
- We hardcode password auth = true (80/20)

**Network model:**
- `delegated_subnet_id` and `private_dns_zone_id` are independent (both optional)
- `public_network_access_enabled` defaults to true
- We derive public access from the presence of `delegated_subnet_id`

**ForceNew fields** (resource recreation on change):
- `name`, `administrator_login`, `delegated_subnet_id`, `geo_redundant_backup_enabled`
- `customer_managed_key` (not exposed), `cluster` (not exposed)

### Pulumi Provider

Uses `github.com/pulumi/pulumi-azure/sdk/v6/go/azure/postgresql` (classic provider), consistent with all other Azure modules in Planton. The classic provider mirrors Terraform's schema closely.

Key types:
- `postgresql.FlexibleServer` -- Main server resource
- `postgresql.FlexibleServerDatabase` -- Database resource
- `postgresql.FlexibleServerFirewallRule` -- Firewall rule resource
- `postgresql.FlexibleServerArgs` -- Server constructor arguments
- `postgresql.FlexibleServerAuthenticationArgs` -- Auth configuration

## Design Decisions Applied

### C1-C2: Required resource_group and region (DD05 compliance)
Every Azure resource in Planton requires `resource_group` (StringValueOrRef) and `region` (string). These were missing from the original T02 spec.

### C3: String+CEL for version (not proto enum)
Following the established pattern from R02, R06, R09, version uses string with CEL `in` validation. This preserves Azure's exact API values and avoids proto enum maintenance burden.

### C4: Optional HA message (not bool+enum)
If the `high_availability` message is present, HA is enabled. No separate boolean needed. Mode uses string+CEL with Azure's exact values ("ZoneRedundant", "SameZone").

### C7: auto_grow_enabled
Added as a production safety net. Default false matches Azure's behavior. Critical for databases with unpredictable growth patterns.

### C8: Repeated databases (not initial_database_name)
Changed from a single string to `repeated AzurePostgresqlDatabase` following the LB backend_pools pattern. Supports multiple databases with custom charset/collation.

### C9: Polymorphic StringValueOrRef for password
No `default_kind` annotation since the password source varies (literal, chart variable, external secret). Matches the UserAssignedIdentity `scope` pattern.

### C10: Hardcoded public_network_access logic
Not exposed as a spec field. IaC modules derive from `delegated_subnet_id` presence:
- Subnet set -> public access disabled
- Subnet not set -> public access enabled

### C11: Maintenance window omitted
Cannot be set on create in the Terraform provider. Deferred to v2.

## Infra Chart Integration

### database-stack chart pattern

```
AzureResourceGroup
└── AzureSubnet (delegated to PostgreSQL)
    └── AzurePrivateDnsZone (privatelink.postgres.database.azure.com)
        └── AzurePostgresqlFlexibleServer
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
