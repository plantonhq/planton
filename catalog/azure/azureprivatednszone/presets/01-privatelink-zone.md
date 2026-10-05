# Private Link Zone

This preset creates the Azure-defined privatelink zone for PostgreSQL
Flexible Server -- the zone private endpoints register their records in
so VNet-connected clients resolve the server's FQDN to its private IP.
Swap the `name` for any other service's privatelink zone:

| Service | Zone name |
|---------|-----------|
| PostgreSQL Flexible Server | `privatelink.postgres.database.azure.com` |
| MySQL Flexible Server | `privatelink.mysql.database.azure.com` |
| Azure SQL Database | `privatelink.database.windows.net` |
| Cosmos DB (SQL) | `privatelink.documents.azure.com` |
| Azure Cache for Redis | `privatelink.redis.cache.windows.net` |
| Blob Storage | `privatelink.blob.core.windows.net` |
| Key Vault | `privatelink.vaultcore.azure.net` |

The zone answers nobody until it is linked: pair it with an
`AzurePrivateDnsZoneVirtualNetworkLink` per network whose workloads call
the service (see that kind's presets).

## When to Use

- Before (or alongside) creating a private endpoint for an Azure PaaS
  service
- One zone per service type per DNS scope -- typically shared by every
  environment in the same network topology

## Key Configuration Choices

- **The name is fixed by Azure** -- privatelink zone names must match the
  service exactly, or endpoint DNS registration misses the zone
- **No registration, no SOA tuning** -- privatelink zones are populated
  by private endpoints; defaults are correct

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<resource-group-name>` | The resource group to create the zone in (or use `valueFrom` against an `AzureResourceGroup`) | The resource group's `status.outputs.resource_group_name` |
| `<cost-center>` | Your org's cost-attribution tag value | Your tagging convention |
