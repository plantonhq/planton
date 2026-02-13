# Azure MSSQL Server

Deploys an Azure SQL Database logical server with configurable databases, firewall rules, TLS policy, and connection policy. The component bundles the logical server with its databases and firewall rules because a server without at least one database and a connection path has no practical utility.

## What Gets Created

When you deploy an AzureMssqlServer resource, OpenMCF provisions:

- **SQL Server** -- a `mssql.Server` resource in the specified region and resource group, configured with administrator credentials, TLS version, public network access, and connection policy
- **Databases** -- a `mssql.Database` resource for each entry in the `databases` list, each with its own compute SKU, maximum storage size, collation, zone redundancy, license type, and backup storage type
- **Firewall Rules** -- a `mssql.FirewallRule` resource for each entry in the `firewallRules` list, defining IP address ranges allowed to connect when public access is enabled
- **Azure Tags** -- resource metadata tags applied to the server for tracking and governance

## Prerequisites

- **Azure credentials** configured via environment variables or OpenMCF provider config
- **An Azure Resource Group** where the server will be created (can reference an AzureResourceGroup resource)
- **Administrator password** meeting Azure SQL requirements: 8-128 characters with characters from at least three of uppercase, lowercase, digits, and special characters
- **Globally unique server name** -- the server name becomes the hostname `{name}.database.windows.net` and must be unique across all of Azure

## Quick Start

Create a file `mssqlserver.yaml`:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: my-sql-server
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureMssqlServer.my-sql-server
spec:
  region: eastus
  resourceGroup: my-rg
  name: my-sql-server
  administratorLogin: sqladmin
  administratorPassword:
    value: "P@ssw0rd1234!"
  databases:
    - name: myappdb
      skuName: S0
```

Deploy:

```shell
openmcf apply -f mssqlserver.yaml
```

This creates a SQL Server running version 12.0 with TLS 1.2, public network access enabled, the Default connection policy, and a single Standard-tier (S0) database.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Azure region for the SQL Server (e.g., `eastus`, `westeurope`). | Required, minimum length 1 |
| `resourceGroup` | `StringValueOrRef` | Azure Resource Group name. Can reference an AzureResourceGroup resource via `valueFrom`. | Required |
| `name` | `string` | Globally unique server name. Becomes the hostname `{name}.database.windows.net`. Lowercase letters, numbers, and hyphens only. Must start with a letter. | Required, 3-63 characters |
| `administratorLogin` | `string` | Administrator login name. Cannot be reserved names such as `admin`, `administrator`, `sa`, `root`, `dbo`, or `guest`. Must start with a letter. | Required, minimum length 1 |
| `administratorPassword` | `StringValueOrRef` | Administrator password. Can be a literal value or a reference to another resource's output (e.g., a generated random password or a Key Vault secret). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `version` | `string` | `"12.0"` | SQL Server version identifier. Values: `"12.0"` (current), `"2.0"` (legacy). |
| `minimumTlsVersion` | `string` | `"1.2"` | Minimum TLS version for client connections. Values: `"1.2"` (recommended), `"1.0"` (legacy). |
| `publicNetworkAccessEnabled` | `bool` | `true` | Whether the server is accessible over the public internet. When `false`, only private endpoints can reach the server. |
| `connectionPolicy` | `string` | `"Default"` | Connection routing policy. Values: `"Default"` (Redirect inside Azure, Proxy outside), `"Proxy"` (all connections proxied, higher latency), `"Redirect"` (direct connection, lower latency, requires ports 11000-11999). |
| `databases` | `AzureMssqlDatabase[]` | `[]` | Databases to create on the server. Each database has its own compute SKU and storage. See Database Fields below. |
| `firewallRules` | `AzureMssqlFirewallRule[]` | `[]` | Firewall rules for public access. Only effective when `publicNetworkAccessEnabled` is `true`. See Firewall Rule Fields below. |

### Database Fields

Each entry in the `databases` list supports:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | `string` | -- | Database name. Required. Must be unique within the server. Maximum 128 characters. |
| `skuName` | `string` | -- | Compute tier and performance level. Required. DTU-based: `"Basic"`, `"S0"`-`"S12"`, `"P1"`-`"P15"`. vCore-based: `"GP_Gen5_2"`, `"BC_Gen5_2"`, `"HS_Gen5_2"`. Serverless: `"GP_S_Gen5_1"`. |
| `maxSizeGb` | `int32` | SKU default | Maximum database size in gigabytes. Varies by tier (Basic: 2 GB, Standard: up to 250 GB, Premium: up to 4096 GB, Hyperscale: up to 100 TB). |
| `collation` | `string` | `"SQL_Latin1_General_CP1_CI_AS"` | Database collation for sort order and string comparison. Changing after creation destroys and recreates the database. |
| `zoneRedundant` | `bool` | `false` | Spread replicas across availability zones. Supported on Premium (DTU) and Business Critical (vCore) tiers only. |
| `licenseType` | `string` | `"LicenseIncluded"` | License model. Values: `"BasePrice"` (Azure Hybrid Benefit, bring your own license), `"LicenseIncluded"` (pay-as-you-go). Not applicable to serverless or Free SKUs. |
| `storageAccountType` | `string` | `"Geo"` | Backup storage redundancy. Values: `"Geo"` (paired region), `"GeoZone"` (zones and regions), `"Local"` (same region), `"Zone"` (availability zones, same region). |

### Firewall Rule Fields

Each entry in the `firewallRules` list supports:

| Field | Type | Description |
|-------|------|-------------|
| `name` | `string` | Firewall rule name. Required. Must be unique within the server. |
| `startIpAddress` | `string` | Start of the IP address range (inclusive). Required. Use `"0.0.0.0"` with end `"0.0.0.0"` to allow all Azure services. |
| `endIpAddress` | `string` | End of the IP address range (inclusive). Required. Set equal to `startIpAddress` for a single IP rule. |

## Examples

### Single Database for Development

A minimal server with one Standard-tier database and no firewall restrictions:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: dev-sql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureMssqlServer.dev-sql
spec:
  region: eastus
  resourceGroup: dev-rg
  name: dev-sql
  administratorLogin: devadmin
  administratorPassword:
    value: "D3v-P@ssw0rd!"
  databases:
    - name: appdb
      skuName: S0
```

### Production Server with Multiple Databases and Firewall Rules

A production server with a General Purpose vCore database, geo-redundant backups, and restricted access from office IP ranges and Azure services:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: prod-sql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureMssqlServer.prod-sql
spec:
  region: westeurope
  resourceGroup: prod-rg
  name: prod-sql
  administratorLogin: prodadmin
  administratorPassword:
    value: "Pr0d-$ecure-P@ss!"
  minimumTlsVersion: "1.2"
  connectionPolicy: Redirect
  databases:
    - name: appdb
      skuName: GP_Gen5_4
      maxSizeGb: 256
      storageAccountType: Geo
      licenseType: LicenseIncluded
    - name: analyticsdb
      skuName: GP_Gen5_2
      maxSizeGb: 128
      storageAccountType: Geo
  firewallRules:
    - name: allow-office
      startIpAddress: "203.0.113.0"
      endIpAddress: "203.0.113.255"
    - name: allow-azure-services
      startIpAddress: "0.0.0.0"
      endIpAddress: "0.0.0.0"
```

### Private-Only Server with Business Critical Database

A server with public access disabled, intended for use exclusively through an AzurePrivateEndpoint. Uses a Business Critical tier with zone redundancy and Azure Hybrid Benefit:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: private-sql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureMssqlServer.private-sql
spec:
  region: eastus
  resourceGroup: prod-rg
  name: private-sql
  administratorLogin: secureadmin
  administratorPassword:
    value: "Pr1v@te-$ql-P@ss!"
  publicNetworkAccessEnabled: false
  databases:
    - name: coredb
      skuName: BC_Gen5_4
      maxSizeGb: 512
      zoneRedundant: true
      licenseType: BasePrice
      storageAccountType: GeoZone
```

### Serverless Database for Cost Optimization

A server with a serverless database that auto-pauses during inactivity, suitable for intermittent workloads:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: serverless-sql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AzureMssqlServer.serverless-sql
spec:
  region: eastus
  resourceGroup: dev-rg
  name: serverless-sql
  administratorLogin: slsadmin
  administratorPassword:
    value: "S3rverl3ss-P@ss!"
  databases:
    - name: eventdb
      skuName: GP_S_Gen5_2
      maxSizeGb: 32
      storageAccountType: Local
  firewallRules:
    - name: allow-dev-machine
      startIpAddress: "198.51.100.42"
      endIpAddress: "198.51.100.42"
```

### Using Foreign Key References

Reference an OpenMCF-managed resource group and use a password from another resource output instead of hardcoding values:

```yaml
apiVersion: azure.openmcf.org/v1
kind: AzureMssqlServer
metadata:
  name: ref-sql
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AzureMssqlServer.ref-sql
spec:
  region: eastus
  resourceGroup:
    valueFrom:
      kind: AzureResourceGroup
      name: my-rg
      field: status.outputs.resource_group_name
  name: ref-sql
  administratorLogin: refsqladmin
  administratorPassword:
    valueFrom:
      kind: AzureKeyVault
      name: prod-vault
      field: status.outputs.secret_id_map.sql-admin-password
  databases:
    - name: appdb
      skuName: GP_Gen5_2
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `serverId` | `string` | Azure Resource Manager ID of the SQL Server. Referenced by AzurePrivateEndpoint for private connectivity (`privateConnectionResourceId` with `subresourceNames: ["sqlServer"]`). |
| `serverName` | `string` | Name of the SQL Server as created in Azure. |
| `fqdn` | `string` | Fully qualified domain name of the server (e.g., `{name}.database.windows.net`). Used to construct connection strings. |
| `administratorLogin` | `string` | Administrator login name. Exported so downstream resources and applications can build connection strings without duplicating this value. |
| `databaseIds` | `map<string, string>` | Map of database names to their Azure Resource Manager IDs. Only populated for databases defined in the spec. |

## Related Components

- [AzureResourceGroup](/docs/catalog/azure/azureresourcegroup) -- provides the resource group for server placement
- [AzurePrivateEndpoint](/docs/catalog/azure/azureprivateendpoint) -- establishes private connectivity to the SQL Server when public access is disabled (use `subresourceNames: ["sqlServer"]`)
- [AzureVpc](/docs/catalog/azure/azurevpc) -- provides the virtual network for private endpoint integration
- [AzureKeyVault](/docs/catalog/azure/azurekeyvault) -- stores the administrator password and other secrets; can be referenced via `valueFrom` for the `administratorPassword` field
