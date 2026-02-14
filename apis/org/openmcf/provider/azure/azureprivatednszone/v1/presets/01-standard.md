# Standard Private DNS Zone

This preset creates an Azure Private DNS Zone for PostgreSQL Flexible Server Private Link resolution, linked to a Virtual Network. When a private endpoint is created for a PostgreSQL server, its private IP is automatically registered in this zone, allowing VNet-connected clients to resolve the server's FQDN to its private IP instead of its public endpoint. This is the most common Private DNS Zone use case.

## When to Use

- Enabling Private Link DNS resolution for PostgreSQL Flexible Server
- Any VNet-injected database deployment that requires internal DNS resolution
- Replace `name` with the appropriate privatelink zone for other services (see Key Configuration Choices)

## Key Configuration Choices

- **Zone name** (`name: privatelink.postgres.database.azure.com`) -- Azure-defined zone name for PostgreSQL Flexible Server Private Link. Change to the appropriate zone for other services:
  - `privatelink.mysql.database.azure.com` -- MySQL Flexible Server
  - `privatelink.database.windows.net` -- Azure SQL Database
  - `privatelink.documents.azure.com` -- Cosmos DB
  - `privatelink.redis.cache.windows.net` -- Azure Cache for Redis
  - `privatelink.blob.core.windows.net` -- Blob Storage
  - `privatelink.vaultcore.azure.net` -- Key Vault
- **Auto-registration disabled** (`registrationEnabled: false`) -- Private Link zones should not auto-register VM records. DNS records are managed by the private endpoint resource itself
- **Single VNet link** -- One link per zone instance. For hub-spoke topologies with multiple VNets, deploy separate instances

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<vnet-resource-id>` | Full ARM resource ID of the VNet to link | Azure portal or `AzureVpc` status outputs |
