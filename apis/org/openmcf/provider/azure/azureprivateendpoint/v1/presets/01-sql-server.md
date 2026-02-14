# Private Endpoint for Azure SQL Database

This preset creates an Azure Private Endpoint that connects an Azure SQL Database server to a VNet subnet via Private Link. It includes a DNS zone group registration so that the SQL server's FQDN resolves to a private IP within the VNet instead of the public endpoint. This is the standard pattern for securing database connectivity in Azure.

## When to Use

- Azure SQL Database instances that must be accessible only via private IP within the VNet
- Environments that require data exfiltration protection by eliminating public database endpoints
- Multi-tier architectures where application subnets connect to databases over the Microsoft backbone network
- Compliance requirements that mandate private-only database connectivity

## Key Configuration Choices

- **Sub-resource: sqlServer** (`subresourceNames: [sqlServer]`) -- Targets the SQL Server sub-resource of the Azure SQL Database service
- **Auto-approved connection** -- The connection is auto-approved (not manual). The private endpoint owner must have appropriate permissions on the target SQL server
- **DNS zone group** (`privateDnsZoneId`) -- Automatically registers an A-record in the specified `privatelink.database.windows.net` zone so that `yourserver.database.windows.net` resolves to the private IP
- **Dynamic IP allocation** -- The private endpoint receives a private IP dynamically from the specified subnet

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the subnet region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-pe-name>` | Name for the private endpoint (unique within resource group) | Your naming convention |
| `<subnet-resource-id>` | Full ARM resource ID of the subnet for private IP allocation | Azure portal or `AzureSubnet` status outputs |
| `<sql-server-resource-id>` | Full ARM resource ID of the Azure SQL Server | Azure portal or `AzureMssqlServer` status outputs |
| `<private-dns-zone-id>` | Full ARM resource ID of the `privatelink.database.windows.net` private DNS zone | Azure portal or `AzurePrivateDnsZone` status outputs |

## Related Presets

- **02-storage-account** -- Use instead for private connectivity to Azure Blob Storage
