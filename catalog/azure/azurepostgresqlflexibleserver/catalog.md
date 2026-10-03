# Azure PostgreSQL Flexible Server

Deploys an Azure Database for PostgreSQL Flexible Server with configurable compute tiers, storage, high availability, VNet integration, and bundled databases with firewall rules. Network access mode is determined by whether a delegated subnet is provided -- public access with firewall rules or private VNet-only access. Lifecycle modes (`createMode`) cover the full server story: a fresh server, a read replica, a point-in-time restore, and a geo-restore from geo-redundant backups.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **PostgreSQL Flexible Server** -- a managed PostgreSQL server in the specified Azure region and resource group, with configurable compute SKU (Burstable, General Purpose, or Memory Optimized), storage with optional auto-grow, PostgreSQL version, password authentication, backup retention, and optional geo-redundant backups
- **High Availability Standby** -- created only when `highAvailability` is configured; a standby server in `ZONE_REDUNDANT` mode (different availability zone) or `SAME_ZONE` mode (same zone, faster failover)
- **PostgreSQL Databases** -- created only when `databases` entries are provided; each database has its own charset and collation settings
- **Firewall Rules** -- created only when `firewallRules` entries are provided; each rule allows connections from a range of IPv4 addresses (only effective in public access mode)
- **Microsoft Entra Administrators** -- created only when `aadAdministrators` entries are provided (requires `authentication.activeDirectoryAuthEnabled`); each grant is its own Azure sub-resource
- **Managed Identity & Customer-Managed Key** -- created only when `identity` / `customerManagedKey` are configured; CMK encrypts the server's data with your Key Vault key through a user-assigned identity
- **Elastic Cluster** -- created only when `cluster` is configured on a fresh PostgreSQL 17+ server; provisions a sharded, citus-based cluster instead of a single node
- **Server Parameter Overrides** -- applied only when `serverParameters` entries are provided; user overrides on Azure's per-SKU engine defaults
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the server for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the PostgreSQL server will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A delegated subnet** (optional, for VNet integration) -- a subnet delegated to `Microsoft.DBforPostgreSQL/flexibleServers`. When provided, public access is automatically disabled. Use the AzureSubnet kind with a delegation block. Provide the subnet ID directly or reference via ValueFromRef. This is a ForceNew field.
- **A Private DNS Zone** (optional, for VNet integration) -- enables FQDN resolution to the server's private IP within the VNet.

## Deploy

### Console

Open the deployment store, find **Azure PostgreSQL Flexible Server**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Production Private Server with Zone-Redundant HA** preset in the [Presets](#presets) tab for the production baseline, or **Development Burstable Server** for the smallest practical server.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: app-postgres
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  serverName: acme-app-pg-prod
  administratorLogin: pgadmin
  administratorPassword:
    value: "$secret/postgres-admin-password"
  skuName: GP_Standard_D2ds_v5
  storageMb: 32768
  databases:
    - name: appdb
```

```shell
planton apply -f postgresql-server.yaml
```

This creates a PostgreSQL Flexible Server with public access, PostgreSQL 16, General Purpose compute (2 vCPU, 8 GiB), 32 GB storage, password authentication, and one database. No firewall rules or high availability are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the PostgreSQL server to a resource group, delegated subnet, and private DNS zone deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  # VNet injection requires public access explicitly OFF -- Azure never
  # derives it for PostgreSQL Flexible Server.
  publicNetworkAccessEnabled: false
  delegatedSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: pg-subnet
      fieldPath: status.outputs.subnet_id
  privateDnsZoneId:
    valueFrom:
      kind: AzurePrivateDnsZone
      name: pg-dns
      fieldPath: status.outputs.zone_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, subnet, and DNS zone first, then provisions the PostgreSQL server with the resolved values.

## Key Configuration

These are the most important decisions when configuring a PostgreSQL Flexible Server. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Compute tier** -- `skuName` determines the compute capacity. Burstable (`B_Standard_B1ms`) is for development. General Purpose (`GP_Standard_D2ds_v5`) is for production workloads. Memory Optimized (`MO_Standard_E2s_v3`) is for analytics and caching. Burstable SKUs do not support high availability or read replicas.

**Network access mode** -- Setting `delegatedSubnetId` activates private VNet access and requires `publicNetworkAccessEnabled: false` set explicitly (validation rejects an injected server without it -- unlike MySQL Flexible Server, Azure never derives it). Without a delegated subnet, the server uses public access controlled by `firewallRules`. The delegated subnet is a ForceNew field -- changing the network mode destroys and recreates the server.

**High availability** -- Add a `highAvailability` block with `mode: ZONE_REDUNDANT` for zone-level failure protection or `mode: SAME_ZONE` for faster failover within the same zone. Only supported on General Purpose and Memory Optimized SKUs.

**Storage and backups** -- `storageMb` uses Azure-allowed values (32768 for 32 GB up to 33553408 for 32 TB) and cannot be downgraded. `autoGrowEnabled` defaults to `false` for PostgreSQL. `backupRetentionDays` ranges from 7-35 (default 7). `geoRedundantBackupEnabled` is a ForceNew field.

**PostgreSQL version** -- `version` supports `"11"` through `"18"`. Unset applies `"16"`, the current production standard. Version `"17"` or `"18"` (set explicitly) is required for elastic clusters; in-place upgrades go upward only.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** (optional) | `delegatedSubnetId` | `status.outputs.subnet_id` |
| **AzurePrivateDnsZone** (optional) | `privateDnsZoneId` | `status.outputs.zone_id` |
| **AzurePostgresqlFlexibleServer** (optional) | `sourceServerId` (replica/restore modes) | `status.outputs.server_id` |
| **AzureUserAssignedIdentity** (optional) | `identity.identityIds`, `customerManagedKey.primaryUserAssignedIdentityId`, `aadAdministrators[].objectId` | `status.outputs.identity_id` / `status.outputs.principal_id` |
| **AzureKeyVaultKey** (optional) | `customerManagedKey.keyVaultKeyId` | `status.outputs.versionless_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `server_id` | Azure resource ID of the PostgreSQL Flexible Server | Private endpoint connections, replica/restore `source_server_id`, diagnostic settings |
| `fqdn` | Fully qualified domain name (`{name}.postgres.database.azure.com`) | Application connection strings |
| `administrator_login` | Administrator login name | Application connection strings |
| `database_ids` | Map of database names to Azure resource IDs (e.g. `status.outputs.database_ids.appdb`) | Any resource that targets one database by ARM ID |
| `identity_principal_id` | Principal ID of the system-assigned managed identity (empty unless the identity type includes system-assigned) | AzureRoleAssignment grants |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development** -- The smallest practical server: a single Burstable instance on the public endpoint, one application database, and the Azure-services firewall rule; everything else rides Azure's defaults. Start from the **Development Burstable Server** preset.

**Production, private and highly available** -- A General Purpose server injected into a delegated subnet (no public endpoint), a zone-redundant standby with automatic failover, geo-redundant 35-day backups, storage auto-grow, and a pinned maintenance window. Start from the **Production Private Server with Zone-Redundant HA** preset.

**Hardened compliance posture** -- Password authentication disabled entirely (Entra-only logins), administration through a Microsoft Entra group, and customer-managed-key encryption through a user-assigned identity. Start from the **Hardened Server: Entra-Only Auth + Customer-Managed Key** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the PostgreSQL server is created
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- provides the delegated subnet for VNet-integrated private access
- [**Azure Private DNS Zone**](/infra-catalog/azure-private-dns-zone) -- provides the private DNS zone for FQDN resolution in VNet mode
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed encryption key (`customerManagedKey.keyVaultKeyId`)
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- unwraps the CMK and serves as an Entra administrator principal