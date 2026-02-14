---
title: "PostgreSQL Flexible Server"
description: "PostgreSQL Flexible Server deployment documentation"
icon: "package"
order: 100
componentName: "azurepostgresqlflexibleserver"
---

# Azure PostgreSQL Flexible Server

Deploys an Azure Database for PostgreSQL Flexible Server with configurable compute tier, storage, high availability, backup retention, and network access mode. The component optionally creates named databases and firewall rules on the server as part of a single composite deployment.

## What Gets Created

When you deploy an AzurePostgresqlFlexibleServer resource, OpenMCF provisions:

- **PostgreSQL Flexible Server** -- a `postgresql.FlexibleServer` resource in the specified region and resource group, configured with the chosen SKU, PostgreSQL version, storage size, backup retention, and authentication settings (password authentication enabled by default)
- **Network Access** -- either private VNet access (when `delegatedSubnetId` is set, public access is automatically disabled) or public access with firewall rules controlling connectivity
- **Databases** -- a `postgresql.FlexibleServerDatabase` for each entry in `databases`, each with independent lifecycle, charset, and collation settings
- **Firewall Rules** -- a `postgresql.FlexibleServerFirewallRule` for each entry in `firewallRules`, allowing connections from specified IP address ranges in public access mode
- **High Availability** -- optional zone-redundant or same-zone standby when the `highAvailability` block is present (General Purpose and Memory Optimized SKUs only)
- **Azure Tags** -- resource metadata tags applied to the server for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the server will be created (can reference an AzureResourceGroup resource)
- **A globally unique server name** -- the name becomes the hostname `{name}.postgres.database.azure.com`
- **Network planning** -- decide between public access with firewall rules or private VNet access with a delegated subnet before deployment, as changing `delegatedSubnetId` after creation destroys and recreates the server

## Quick Start

Create a file `postgresql.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: my-pg-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzurePostgresqlFlexibleServer.my-pg-server
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-pg-server
  administratorLogin: pgadmin
  administratorPassword: "Ch@ngeMe1234!"
  skuName: B_Standard_B1ms
  storageMb: 32768
  databases:
    - name: myapp
```

Deploy:

```shell
openmcf apply -f postgresql.yaml
```

This creates a Burstable-tier PostgreSQL 16 Flexible Server with 32 GB storage, public network access, and a single application database named `myapp`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the server (e.g., `eastus`, `westeurope`). Must match the VNet/subnet region if VNet integration is used. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Globally unique server name. Becomes the hostname `{name}.postgres.database.azure.com`. Lowercase letters, numbers, and hyphens only; must start with a letter. **ForceNew**: changing this destroys and recreates the server. | Required, 3--63 characters, pattern `^[a-z][a-z0-9-]*$` |
| `administratorLogin` | `string` | Administrator login name. Cannot be `azure_superuser`, `admin`, `administrator`, `root`, `guest`, or `public`. Must start with a letter. **ForceNew**: changing this destroys and recreates the server. | Required, minimum length 1 |
| `administratorPassword` | `StringValueOrRef` | Administrator password. 8--128 characters with characters from at least three of: uppercase, lowercase, digits, special characters. Can reference another resource's output via `valueFrom`. | Required |
| `skuName` | `string` | Compute tier and size. Format: `{TIER}_Standard_{SIZE}`. See SKU naming below. | Required, minimum length 1 |
| `storageMb` | `int32` | Storage size in megabytes. Allowed values: 32768, 65536, 131072, 262144, 524288, 1048576, 2097152, 4194304, 8388608, 16777216, 33553408. Cannot be downgraded after creation. | Required, minimum 32768 |

**SKU naming convention** -- `{TIER}_Standard_{SIZE}` where TIER is:

- `B` (Burstable) -- dev/test workloads. Examples: `B_Standard_B1ms`, `B_Standard_B2s`, `B_Standard_B4ms`
- `GP` (General Purpose) -- production workloads. Examples: `GP_Standard_D2s_v3`, `GP_Standard_D4s_v3`, `GP_Standard_D8s_v3`
- `MO` (Memory Optimized) -- analytics and caching. Examples: `MO_Standard_E2s_v3`, `MO_Standard_E4s_v3`

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `version` | `string` | `"16"` | PostgreSQL major version. Valid values: `"12"`, `"13"`, `"14"`, `"15"`, `"16"`, `"17"`. |
| `autoGrowEnabled` | `bool` | `false` | Automatically increase storage when free storage falls below a threshold. |
| `backupRetentionDays` | `int32` | `7` | Number of days to retain automatic daily backups for point-in-time restore. Range: 7--35. |
| `geoRedundantBackupEnabled` | `bool` | `false` | Replicate backup data to a paired Azure region for cross-region disaster recovery. **ForceNew**: changing this destroys and recreates the server. |
| `delegatedSubnetId` | `StringValueOrRef` | -- | Subnet ID delegated to `Microsoft.DBforPostgreSQL/flexibleServers`. When set, the server uses private VNet access and public access is disabled. Can reference an AzureSubnet resource via `valueFrom`. **ForceNew**: changing this destroys and recreates the server. |
| `privateDnsZoneId` | `StringValueOrRef` | -- | Private DNS zone ID for server name resolution within the VNet. Typically used with `delegatedSubnetId`. Can reference an AzurePrivateDnsZone resource via `valueFrom`. |
| `zone` | `string` | -- | Availability zone for the primary server. Valid values: `"1"`, `"2"`, `"3"`. If omitted, Azure selects automatically. |
| `highAvailability.mode` | `string` | -- | High availability mode. `"ZoneRedundant"` places the standby in a different zone; `"SameZone"` places it in the same zone. Burstable SKUs do not support HA. |
| `highAvailability.standbyAvailabilityZone` | `string` | -- | Availability zone for the standby server. Valid values: `"1"`, `"2"`, `"3"`. Must differ from `zone` in ZoneRedundant mode. |
| `databases` | `list` | `[]` | Databases to create on the server. Each entry has `name` (required), `charset` (default `"UTF8"`), and `collation` (default `"en_US.utf8"`). |
| `firewallRules` | `list` | `[]` | Firewall rules for public access mode. Each entry has `name`, `startIpAddress`, and `endIpAddress` (all required). Use `0.0.0.0`/`0.0.0.0` to allow all Azure services. |

## Examples

### Development Server with Public Access

A minimal Burstable-tier server for development with a single database and a firewall rule allowing the developer machine:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: dev-pg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzurePostgresqlFlexibleServer.dev-pg
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-pg
  administratorLogin: devadmin
  administratorPassword: "DevP@ss2024!"
  skuName: B_Standard_B1ms
  storageMb: 32768
  version: "16"
  databases:
    - name: myapp
  firewallRules:
    - name: allow-dev-machine
      startIpAddress: "203.0.113.42"
      endIpAddress: "203.0.113.42"
    - name: allow-azure-services
      startIpAddress: "0.0.0.0"
      endIpAddress: "0.0.0.0"
```

### Production Server with High Availability

A General Purpose server with zone-redundant HA, geo-redundant backups, 35-day retention, storage auto-grow, and multiple databases:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: prod-pg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePostgresqlFlexibleServer.prod-pg
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: prod-pg
  administratorLogin: pgadmin
  administratorPassword: "Pr0dStr0ng!P@ss"
  skuName: GP_Standard_D4s_v3
  storageMb: 262144
  version: "16"
  autoGrowEnabled: true
  backupRetentionDays: 35
  geoRedundantBackupEnabled: true
  zone: "1"
  highAvailability:
    mode: ZoneRedundant
    standbyAvailabilityZone: "2"
  databases:
    - name: orders
    - name: inventory
    - name: analytics
      charset: UTF8
      collation: en_US.utf8
  firewallRules:
    - name: allow-office
      startIpAddress: "203.0.113.0"
      endIpAddress: "203.0.113.255"
    - name: allow-cicd
      startIpAddress: "198.51.100.10"
      endIpAddress: "198.51.100.10"
```

### Private VNet Access with Delegated Subnet

A server deployed into a VNet with private access, a private DNS zone for name resolution, and no public connectivity:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: private-pg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePostgresqlFlexibleServer.private-pg
spec:
  region: eastus
  resourceGroup: prod-rg
  name: private-pg
  administratorLogin: pgadmin
  administratorPassword: "Pr1v@teAccess!99"
  skuName: GP_Standard_D2s_v3
  storageMb: 131072
  delegatedSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/pg-subnet
  privateDnsZoneId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com
  zone: "1"
  highAvailability:
    mode: SameZone
  databases:
    - name: appdb
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding IDs. The resource group, subnet, and private DNS zone are resolved from their respective stack outputs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: ref-pg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePostgresqlFlexibleServer.ref-pg
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-pg
  administratorLogin: pgadmin
  administratorPassword: "R3f@Str0ng!Pass"
  skuName: GP_Standard_D2s_v3
  storageMb: 65536
  delegatedSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: pg-subnet
      field: status.outputs.subnet_id
  privateDnsZoneId:
    valueFrom:
      kind: AzurePrivateDnsZone
      name: pg-dns
      field: status.outputs.zone_id
  databases:
    - name: appdb
```

### Memory Optimized for Analytics

A Memory Optimized server with PostgreSQL 17 for analytics workloads, large storage, and extended backup retention:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzurePostgresqlFlexibleServer
metadata:
  name: analytics-pg
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzurePostgresqlFlexibleServer.analytics-pg
spec:
  region: southeastasia
  resourceGroup: analytics-rg
  name: analytics-pg
  administratorLogin: analyticsadmin
  administratorPassword: "An@lyt1cs!Str0ng"
  skuName: MO_Standard_E4s_v3
  storageMb: 4194304
  version: "17"
  autoGrowEnabled: true
  backupRetentionDays: 35
  geoRedundantBackupEnabled: true
  zone: "2"
  highAvailability:
    mode: ZoneRedundant
    standbyAvailabilityZone: "3"
  databases:
    - name: warehouse
    - name: reporting
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `server_id` | `string` | Azure Resource Manager ID of the PostgreSQL Flexible Server. Referenced by AzurePrivateEndpoint for establishing private connectivity. |
| `server_name` | `string` | Name of the PostgreSQL Flexible Server |
| `fqdn` | `string` | Fully qualified domain name (e.g., `{name}.postgres.database.azure.com`). Used to construct connection strings: `postgresql://{admin}:{password}@{fqdn}:5432/{database}?sslmode=require` |
| `administrator_login` | `string` | Administrator login name, exported so downstream resources can construct connection strings without duplicating this value |
| `database_ids` | `map<string, string>` | Map of database names to their Azure Resource Manager IDs. Only populated for databases defined in `databases`. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for server placement
- [AzureSubnet](/docs/catalog/azure/azuresubnet) -- provides a delegated subnet for private VNet access
- [AzurePrivateDnsZone](/docs/catalog/azure/azureprivatednszone) -- provides DNS resolution for VNet-connected clients
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) -- stores the administrator password or connection strings as secrets
- [AzureVpc](/docs/catalog/azure/azurevpc) -- provides the virtual network containing the delegated subnet
