# SQL Server Private DNS Zone

This preset creates a Private DNS Zone for Azure SQL Server (MSSQL) Private Endpoint connectivity. The zone name `privatelink.database.windows.net` is required by Azure for DNS resolution of SQL Server private endpoints. Clients in the linked VNet resolve the server's FQDN to its private IP address.

## When to Use

- Azure SQL Server deployments using Private Endpoints for network isolation
- Enterprise environments requiring all database traffic to stay within private networks
- Multi-VNet architectures requiring DNS resolution for SQL Server across peered networks

## Key Configuration Choices

- **Zone name** (`name: privatelink.database.windows.net`) -- Azure-mandated zone name for SQL Server Private Link
- **VNet link** (`vnetId`) -- Links this DNS zone to the specified VNet for DNS resolution
- **Registration disabled** (`registrationEnabled: false`) -- Required for Private Link zones

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<vnet-resource-id>` | ARM resource ID of the VNet to link | `AzureVpc` status outputs |

## Related Presets

- **01-standard** -- PostgreSQL Private Link DNS zone
- **02-mysql** -- MySQL Private Link DNS zone
