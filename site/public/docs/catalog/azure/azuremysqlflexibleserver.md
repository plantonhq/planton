---
title: "My S Q Lflexibleserver"
description: "My S Q Lflexibleserver deployment documentation"
icon: "package"
order: 100
componentName: "azuremysqlflexibleserver"
---

# Azure MySQL Flexible Server

Deploys an Azure Database for MySQL Flexible Server with configurable compute tier, storage, high availability, backup retention, and network access mode. The component optionally creates named databases and firewall rules on the server.

## What Gets Created

When you deploy an AzureMysqlFlexibleServer resource, OpenMCF provisions:

- **MySQL Flexible Server** — a `mysql.FlexibleServer` resource in the specified region and resource group, configured with the chosen SKU, MySQL version, storage size, backup retention, and high availability settings
- **Network Access** — public access with firewall rules when no delegated subnet is provided, or private VNet access when `delegatedSubnetId` is set (public access is automatically disabled)
- **Databases** — a `mysql.FlexibleDatabase` resource for each entry in `databases`, each with its own charset and collation
- **Firewall Rules** — a `mysql.FlexibleServerFirewallRule` resource for each entry in `firewallRules`, controlling IP-based access in public access mode
- **Azure Tags** — resource metadata tags applied to the server for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the server will be created (can reference an AzureResourceGroup resource)
- **Network planning** — if using private VNet access, a subnet delegated to `Microsoft.DBforMySQL/flexibleServers` and optionally a private DNS zone for name resolution

## Quick Start

Create a file `mysql.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: my-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureMysqlFlexibleServer.my-mysql
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-mysql
  administratorLogin: myadmin
  administratorPassword: "Ch@ngeMe123!"
  skuName: B_Standard_B1ms
  storageSizeGb: 20
```

Deploy:

```shell
openmcf apply -f mysql.yaml
```

This creates a Burstable-tier MySQL 8.0.21 server with 20 GB storage, auto-grow enabled, 7-day backup retention, and public access with no firewall rules (all connections blocked until rules are added).

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the server (e.g., `eastus`, `westeurope`). Must match the VNet region if using VNet integration. | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Globally unique server name. Forms the hostname: `{name}.mysql.database.azure.com`. Lowercase letters, numbers, and hyphens only. **ForceNew**: changing this destroys and recreates the server. | Required, 3–63 characters, must start and end with a letter or number |
| `administratorLogin` | `string` | Administrator login name. Cannot be reserved names such as `admin`, `root`, or `azure_superuser`. **ForceNew**: changing this destroys and recreates the server. | Required, 1–32 characters |
| `administratorPassword` | `StringValueOrRef` | Administrator password. Must contain characters from at least three of: uppercase, lowercase, digits, special characters. Can reference another resource's output via `valueFrom`. | Required, 8–128 characters |
| `skuName` | `string` | Compute tier and size. Format: `{TIER}_Standard_{SIZE}`. Tiers: `B` (Burstable), `GP` (General Purpose), `MO` (Memory Optimized). Examples: `B_Standard_B1ms`, `GP_Standard_D2ds_v4`, `MO_Standard_E2ds_v4`. | Required, minimum length 1 |
| `storageSizeGb` | `int32` | Storage size in gigabytes. Cannot be downgraded after creation. | Required, minimum 20 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `version` | `string` | `8.0.21` | MySQL version. Valid values: `5.7` (legacy, approaching EOL), `8.0.21` (recommended), `8.4` (latest GA). |
| `autoGrowEnabled` | `bool` | `true` | Automatically increase storage when free storage falls below a threshold. |
| `delegatedSubnetId` | `StringValueOrRef` | — | Subnet ID delegated to `Microsoft.DBforMySQL/flexibleServers`. When set, enables private VNet access and disables public access. Can reference an AzureSubnet resource via `valueFrom`. **ForceNew**: changing this destroys and recreates the server. |
| `privateDnsZoneId` | `StringValueOrRef` | — | Private DNS zone ID for server name resolution within the VNet. Typically `privatelink.mysql.database.azure.com`. Can reference an AzurePrivateDnsZone resource via `valueFrom`. **ForceNew**: changing this destroys and recreates the server. |
| `zone` | `string` | — | Availability zone for the primary server. Valid values: `1`, `2`, `3`. If omitted, Azure selects automatically. |
| `highAvailability.mode` | `string` | — | HA mode. `ZoneRedundant` places the standby in a different zone (recommended for production). `SameZone` places the standby in the same zone. Burstable SKUs do not support HA. |
| `highAvailability.standbyAvailabilityZone` | `string` | — | Availability zone for the standby. Must differ from `zone` when using `ZoneRedundant`. |
| `backupRetentionDays` | `int32` | `7` | Number of days to retain automatic backups for point-in-time restore. Range: 1–35. |
| `geoRedundantBackupEnabled` | `bool` | `false` | Replicate backups to a paired Azure region for cross-region disaster recovery. **ForceNew**: changing this destroys and recreates the server. |
| `databases` | `list` | `[]` | Databases to create on the server. Each entry has: `name` (required), `charset` (default `utf8mb4`), `collation` (default `utf8mb4_0900_ai_ci`). |
| `firewallRules` | `list` | `[]` | Firewall rules for public access mode. Each entry has: `name` (required), `startIpAddress` (required), `endIpAddress` (required). Use `0.0.0.0`/`0.0.0.0` to allow all Azure services. |

## Examples

### Development Server with Burstable SKU

A minimal server for development and testing with the smallest compute tier:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: dev-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureMysqlFlexibleServer.dev-mysql
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-mysql
  administratorLogin: devadmin
  administratorPassword: "D3v$ecure!Pass"
  skuName: B_Standard_B1ms
  storageSizeGb: 20
  backupRetentionDays: 1
  databases:
    - name: appdb
  firewallRules:
    - name: allow-all-azure
      startIpAddress: "0.0.0.0"
      endIpAddress: "0.0.0.0"
```

### Production Server with HA and Firewall Rules

A General Purpose server with zone-redundant high availability, multiple databases, and restricted network access:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: prod-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureMysqlFlexibleServer.prod-mysql
spec:
  region: eastus
  resourceGroup: prod-rg
  name: prod-mysql
  administratorLogin: prodadmin
  administratorPassword: "Pr0d$ecure!Passw0rd"
  version: "8.0.21"
  skuName: GP_Standard_D4ds_v4
  storageSizeGb: 256
  autoGrowEnabled: true
  zone: "1"
  highAvailability:
    mode: ZoneRedundant
    standbyAvailabilityZone: "2"
  backupRetentionDays: 35
  geoRedundantBackupEnabled: true
  databases:
    - name: appdb
    - name: analytics
      charset: utf8mb4
      collation: utf8mb4_0900_ai_ci
  firewallRules:
    - name: allow-office
      startIpAddress: "203.0.113.0"
      endIpAddress: "203.0.113.255"
    - name: allow-ci
      startIpAddress: "198.51.100.42"
      endIpAddress: "198.51.100.42"
```

### Private VNet Access with Delegated Subnet

A server deployed into a private VNet with no public endpoint, using a delegated subnet and private DNS zone:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: private-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureMysqlFlexibleServer.private-mysql
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: private-mysql
  administratorLogin: dbadmin
  administratorPassword: "Pr!vat3Acc3ss#99"
  skuName: GP_Standard_D2ds_v4
  storageSizeGb: 128
  delegatedSubnetId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/mysql-subnet
  privateDnsZoneId: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/privateDnsZones/privatelink.mysql.database.azure.com
  highAvailability:
    mode: SameZone
  databases:
    - name: appdb
```

### Using Foreign Key References

Reference OpenMCF-managed resources instead of hardcoding Azure resource IDs:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: ref-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureMysqlFlexibleServer.ref-mysql
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-mysql
  administratorLogin: myadmin
  administratorPassword: "R3f$ecure!Pass"
  skuName: GP_Standard_D2ds_v4
  storageSizeGb: 64
  delegatedSubnetId:
    valueFrom:
      kind: AzureSubnet
      name: mysql-subnet
      field: status.outputs.subnet_id
  privateDnsZoneId:
    valueFrom:
      kind: AzurePrivateDnsZone
      name: mysql-dns
      field: status.outputs.zone_id
  databases:
    - name: appdb
    - name: jobs
```

### MySQL 8.4 with Memory Optimized SKU

A high-performance server running the latest MySQL version on a Memory Optimized tier for analytics workloads:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMysqlFlexibleServer
metadata:
  name: analytics-mysql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureMysqlFlexibleServer.analytics-mysql
spec:
  region: eastus
  resourceGroup: analytics-rg
  name: analytics-mysql
  administratorLogin: analyticsadmin
  administratorPassword: "An@lytics!P4ss"
  version: "8.4"
  skuName: MO_Standard_E2ds_v4
  storageSizeGb: 512
  autoGrowEnabled: true
  backupRetentionDays: 14
  databases:
    - name: warehouse
    - name: reporting
  firewallRules:
    - name: allow-office
      startIpAddress: "203.0.113.0"
      endIpAddress: "203.0.113.255"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `server_id` | `string` | Azure Resource Manager ID of the MySQL Flexible Server. Referenced by AzurePrivateEndpoint for private connectivity. |
| `server_name` | `string` | Name of the MySQL Flexible Server |
| `fqdn` | `string` | Fully qualified domain name (e.g., `{name}.mysql.database.azure.com`). Used to construct connection strings. |
| `administrator_login` | `string` | Administrator login name for constructing connection strings |
| `database_ids` | `map<string, string>` | Map of database names to their Azure Resource Manager IDs. Only populated for databases defined in `databases`. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) — provides the resource group for server placement
- [AzureSubnet](/docs/catalog/azure/azuresubnet) — provides a delegated subnet for private VNet access
- [AzurePrivateDnsZone](/docs/catalog/azure/azureprivatednszone) — provides a private DNS zone for VNet name resolution
- [AzureVpc](/docs/catalog/azure/azurevpc) — provides the virtual network containing delegated subnets
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) — can store the administrator password as a secret
