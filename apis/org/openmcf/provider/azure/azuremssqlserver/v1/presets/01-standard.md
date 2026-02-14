# Standard Azure SQL Database

This preset creates an Azure SQL logical server with a Standard-tier (S0) database. The S0 SKU provides 10 DTUs of compute, 250 GB max storage, and geo-redundant backups -- a cost-effective entry point for production SQL Server workloads. A firewall rule allows connections from Azure services, and the server enforces TLS 1.2 with the default connection policy.

## When to Use

- Small to medium production applications needing SQL Server compatibility
- Line-of-business applications, internal tools, and APIs with moderate query loads
- Workloads where DTU-based pricing provides predictable cost (bundled compute, IO, storage)
- Teams migrating from on-premises SQL Server that need a managed PaaS equivalent

## Key Configuration Choices

- **Standard S0 SKU** (`databases[0].skuName: S0`) -- 10 DTUs, suitable for light production workloads. Scale up to S1 (20 DTU), S2 (50 DTU), or switch to vCore-based GP_Gen5_2 for more control
- **250 GB max storage** (`databases[0].maxSizeGb: 250`) -- Standard tier maximum. Sufficient for most small-to-medium databases
- **Geo-redundant backups** (`databases[0].storageAccountType: Geo`) -- Default. Backups replicated to the paired region for disaster recovery
- **Default connection policy** (`connectionPolicy: Default`) -- Uses Redirect for Azure-to-Azure connections and Proxy for external clients
- **Public network access** (`publicNetworkAccessEnabled: true`) -- Server has a public endpoint. Disable and use `AzurePrivateEndpoint` for private-only access
- **TLS 1.2** (`minimumTlsVersion: "1.2"`) -- Industry standard. Lowers to "1.0" only for legacy client compatibility
- **No zone redundancy** (`databases[0].zoneRedundant: false`) -- Not supported on Standard tier. Use Business Critical for zone-redundant deployments

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-server-name>` | Globally unique server name (3-63 chars, lowercase/numbers/hyphens) | Choose a name; becomes `{name}.database.windows.net` |
| `<admin-username>` | Administrator login (cannot be "admin", "sa", "root", etc.) | Your credentials policy |
| `<admin-password>` | Administrator password (8-128 chars, 3 of 4 character types) | Generate a strong password or reference a Key Vault secret |
| `<your-database-name>` | Name of the application database (max 128 chars) | Your application configuration |

## Related Presets

- **02-business-critical** -- Use instead for mission-critical workloads requiring zone redundancy, local SSD, and higher IOPS
