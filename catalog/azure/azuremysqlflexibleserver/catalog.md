# Azure MySQL Flexible Server

Deploys an Azure Database for MySQL Flexible Server with configurable compute tiers, storage, high availability, VNet integration, and bundled databases with firewall rules. Network access mode is determined by whether a delegated subnet is provided -- public access with firewall rules or private VNet-only access. Lifecycle modes (`createMode`) cover the full server story: a fresh server, a read replica, a point-in-time restore, and a geo-restore from geo-redundant backups.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **MySQL Flexible Server** -- a managed MySQL server in the specified Azure region and resource group, with configurable compute SKU (Burstable, General Purpose, or Memory Optimized), storage with optional auto-grow / provisioned IOPS / elastic IO scaling, MySQL version, password authentication, backup retention, and optional geo-redundant backups
- **High Availability Standby** -- created only when `highAvailability` is configured; a standby server in `ZONE_REDUNDANT` mode (different availability zone) or `SAME_ZONE` mode (same zone, faster failover)
- **MySQL Databases** -- created only when `databases` entries are provided; each database has its own charset and collation settings
- **Firewall Rules** -- created only when `firewallRules` entries are provided; each rule allows connections from a range of IPv4 addresses (only effective in public access mode)
- **Microsoft Entra Administrator** -- created only when `aadAdministrator` is configured; MySQL admits exactly one Entra admin whose `objectId` resolves the identity's **client_id** (not principal_id)
- **User-Assigned Identities & Customer-Managed Key** -- created only when `userAssignedIdentityIds` / `customerManagedKey` are configured; MySQL has no system-assigned identity flavor
- **Server Parameter Overrides** -- applied only when `serverParameters` entries are provided; user overrides on Azure's per-SKU engine defaults
- **Azure Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the server for tracking and governance

## Before You Deploy

### Planton Setup

- **Azure Provider Connection** -- an active connection in the Connect module with credentials for the target Azure subscription. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Azure Subscription

- **An Azure Resource Group** where the MySQL server will be created. Provide the name directly or reference an AzureResourceGroup Infra Component via ValueFromRef.
- **A delegated subnet** (optional, for VNet integration) -- a subnet delegated to `Microsoft.DBforMySQL/flexibleServers`. When provided, public access must not be ENABLED. Use the AzureSubnet kind with a delegation block. Provide the subnet ID directly or reference via ValueFromRef. This is a ForceNew field.
- **A Private DNS Zone** (optional, for VNet integration) -- enables FQDN resolution to the server's private IP within the VNet. This is a ForceNew field.

## Deploy

### Console

Open the deployment store, find **Azure MySQL Flexible Server**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Development Burstable Server** preset in the [Presets](#presets) tab for the smallest practical server, or **Production Private Server with Zone-Redundant HA** for the production baseline.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureMysqlFlexibleServer
metadata:
  name: app-mysql
  org: acme-corp
  env: prod
spec:
  region: eastus
  resourceGroup:
    value: "acme-prod-rg"
  serverName: acme-app-mysql-prod
  administratorLogin: mysqladmin
  administratorPassword:
    value: "$secret/mysql-admin-password"
  skuName: GP_Standard_D2ds_v4
  storage:
    sizeGb: 32
  databases:
    - name: appdb
```

```shell
planton apply -f mysql-server.yaml
```

This creates a MySQL Flexible Server with public access, MySQL 8.0.21, General Purpose compute (2 vCPU, 8 GiB), 32 GB storage with auto-grow enabled (Azure's MySQL default), and one database. No firewall rules or high availability are configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the MySQL server to a resource group, delegated subnet, and private DNS zone deployed in the same InfraPipeline:

```yaml
spec:
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: production-rg
      fieldPath: status.outputs.resource_group_name
  delegatedSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: mysql-subnet
      fieldPath: status.outputs.subnet_id
  privateDnsZoneId:
    valueFrom:
      kind: AzurePrivateDnsZone
      name: mysql-dns
      fieldPath: status.outputs.zone_id
```

The InfraPipeline resolves the dependency graph, deploys the resource group, subnet, and DNS zone first, then provisions the MySQL server with the resolved values.

## Key Configuration

These are the most important decisions when configuring a MySQL Flexible Server. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Compute tier** -- `skuName` determines the compute capacity. Burstable (`B_Standard_B1ms`) is for development. General Purpose (`GP_Standard_D2ds_v4`) is for production workloads. Memory Optimized (`MO_Standard_E2ds_v4`) is for analytics and caching. Burstable SKUs do not support high availability or read replicas.

**Network access mode** -- Setting `delegatedSubnetId` activates private VNet access. `publicNetworkAccess` is a meaningful enum (`UNSPECIFIED` lets Azure derive; `ENABLED` / `DISABLED` pin the posture). Without a delegated subnet, the server uses public access controlled by `firewallRules`. Network mode is ForceNew -- changing it destroys and recreates the server.

**High availability** -- Add a `highAvailability` block with `mode: ZONE_REDUNDANT` for zone-level failure protection or `mode: SAME_ZONE` for faster failover within the same zone. Only supported on General Purpose and Memory Optimized SKUs.

**Storage and backups** -- Storage lives under the nested `storage` message (`storage.sizeGb`, `storage.autoGrowEnabled`, `storage.iops`, `storage.ioScalingEnabled`, `storage.logOnDiskEnabled`). Size cannot be downgraded after creation. Provisioned IOPS and elastic IO scaling are mutually exclusive. `backupRetentionDays` ranges from 1-35 (default 7). `geoRedundantBackupEnabled` is a ForceNew field.

**MySQL version** -- `version` accepts Azure's exact strings: `"5.7"`, `"8.0.21"` (the whole 8.0 series identifier), or `"8.4"`. Unset applies `"8.0.21"`.

**Entra administrator** -- Unlike PostgreSQL, MySQL has no authentication posture block (password auth is always on) and admits exactly ONE `aadAdministrator`. Its `objectId` must resolve the identity's **client_id** -- MySQL validates tokens against the client id, not the principal id.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AzureResourceGroup** | `resourceGroup` | `status.outputs.resource_group_name` |
| **AzureSubnet** (optional) | `delegatedSubnetId` | `status.outputs.subnet_id` |
| **AzurePrivateDnsZone** (optional) | `privateDnsZoneId` | `status.outputs.zone_id` |
| **AzureMysqlFlexibleServer** (optional) | `sourceServerId` (replica/restore modes) | `status.outputs.server_id` |
| **AzureUserAssignedIdentity** (optional) | `userAssignedIdentityIds`, `aadAdministrator.identityId` / `objectId`, `customerManagedKey.primaryUserAssignedIdentityId` | `status.outputs.identity_id` / `status.outputs.client_id` |
| **AzureKeyVaultKey** (optional) | `customerManagedKey.keyVaultKeyId` | `status.outputs.versionless_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `server_id` | Azure resource ID of the MySQL Flexible Server | Private endpoint connections, replica/restore `source_server_id`, diagnostic settings |
| `fqdn` | Fully qualified domain name (`{name}.mysql.database.azure.com`) | Application connection strings |
| `administrator_login` | Administrator login name | Application connection strings |
| `database_ids` | Map of database names to Azure resource IDs (e.g. `status.outputs.database_ids.appdb`) | Any resource that targets one database by ARM ID |
| `replica_capacity` | How many more read replicas this server can take (Burstable reports 0) | Capacity planning before creating replicas |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Development** -- The smallest practical server: a single Burstable instance on the public endpoint, one application database, and the Azure-services firewall rule; everything else rides Azure's defaults. Start from the **Development Burstable Server** preset.

**Production, private and highly available** -- A General Purpose server injected into a delegated subnet (no public endpoint), a zone-redundant standby with automatic failover, geo-redundant backups, and storage auto-grow. Start from the **Production Private Server with Zone-Redundant HA** preset.

**Hardened compliance posture** -- Customer-managed-key encryption through a user-assigned identity plus a Microsoft Entra administrator (password auth stays on — MySQL never turns it off). Start from the **Hardened Server with CMK Encryption and an Entra Administrator** preset.

## Works With

- [**Azure Resource Group**](/infra-catalog/azure-resource-group) -- provides the resource group where the MySQL server is created
- [**Azure Subnet**](/infra-catalog/azure-subnet) -- provides the delegated subnet for VNet-integrated private access
- [**Azure Private DNS Zone**](/infra-catalog/azure-private-dns-zone) -- provides the private DNS zone for FQDN resolution in VNet mode
- [**Azure Key Vault Key**](/infra-catalog/azure-key-vault-key) -- the customer-managed encryption key (`customerManagedKey.keyVaultKeyId`)
- [**Azure User Assigned Identity**](/infra-catalog/azure-user-assigned-identity) -- unwraps the CMK and serves as the Entra administrator principal
