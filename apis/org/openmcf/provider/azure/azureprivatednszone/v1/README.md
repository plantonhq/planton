# AzurePrivateDnsZone

## Overview

**AzurePrivateDnsZone** creates an Azure Private DNS Zone with a Virtual Network link, enabling private name resolution within Azure Virtual Networks.

Azure Private DNS zones serve two primary purposes:

1. **Private Link DNS resolution** -- When using Azure Private Endpoints to access PaaS services (PostgreSQL, MySQL, Key Vault, Storage, etc.) over a private IP, the corresponding privatelink DNS zone ensures that the service's FQDN resolves to its private IP address instead of the public one. Without a properly configured private DNS zone, clients in the VNet would resolve the public IP and bypass the private endpoint entirely.

2. **Custom internal DNS** -- For internal name resolution (e.g., `contoso.internal`), private DNS zones provide hostname resolution for VMs and other resources without requiring external DNS infrastructure. With `registration_enabled` set to `true`, Azure automatically registers and deregisters VM A-records as VMs are created and deleted.

## When to Use

- You need private connectivity to Azure PaaS services (PostgreSQL Flexible Server, MySQL Flexible Server, Redis, Key Vault, etc.) via Private Endpoints
- You want internal DNS resolution for VMs or containers within a VNet
- You're building the **database-stack** infra chart and need privatelink zones for each database type
- You're setting up the networking foundation for a Private Link-based architecture

## Key Configuration

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `resource_group` | StringValueOrRef | Yes | Azure Resource Group for the zone |
| `name` | string | Yes | DNS zone name (e.g., `privatelink.postgres.database.azure.com`) |
| `vnet_id` | StringValueOrRef | Yes | VNet to link for DNS resolution |
| `registration_enabled` | bool | No | Auto-register VM A-records (default: `false`) |

## Outputs

| Output | Description |
|--------|-------------|
| `zone_id` | Azure Resource Manager ID of the zone |
| `zone_name` | Name of the private DNS zone |

## Azure Resources Created

- `azurerm_private_dns_zone` -- The private DNS zone itself
- `azurerm_private_dns_zone_virtual_network_link` -- Links the zone to a VNet for name resolution

## Key Behaviors

- **Global resource** -- Private DNS zones have no region; they are globally available within the subscription
- **Bundled VNet link** -- A zone without a VNet link is unreachable; this component always creates one link per DD03 (Composite Bundling Rules)
- **Private Link naming** -- For Private Endpoint scenarios, the zone name must match Azure's defined privatelink zone name for the target service (see examples below)
- **Registration** -- When `registration_enabled` is `true`, A-records are auto-managed for VMs in the linked VNet; use this only for custom internal zones, not privatelink zones

## Common Private Link Zone Names

| Azure Service | Private DNS Zone Name |
|--------------|----------------------|
| PostgreSQL Flexible Server | `privatelink.postgres.database.azure.com` |
| MySQL Flexible Server | `privatelink.mysql.database.azure.com` |
| Azure SQL Database | `privatelink.database.windows.net` |
| Cosmos DB | `privatelink.documents.azure.com` |
| Azure Cache for Redis | `privatelink.redis.cache.windows.net` |
| Azure Blob Storage | `privatelink.blob.core.windows.net` |
| Azure Key Vault | `privatelink.vaultcore.azure.net` |
| Azure Container Registry | `privatelink.azurecr.io` |

## Dependencies

- **AzureResourceGroup** -- The resource group where the zone is created
- **AzureVpc** -- The VNet linked to this zone for DNS resolution

## Referenced By

- **AzurePrivateEndpoint** -- `private_dns_zone_id` for DNS zone group registration
- **AzurePostgresqlFlexibleServer** -- `private_dns_zone_id` for VNet-integrated deployment
- **AzureMysqlFlexibleServer** -- `private_dns_zone_id` for VNet-integrated deployment

## Infra Chart Usage

- **database-stack** -- Creates privatelink zones for each database type (PostgreSQL, MySQL, MSSQL, Redis) and wires them to private endpoints
- **enterprise-network-foundation** -- Optional component for private DNS infrastructure

## Known Limitations

- **Single VNet link per instance** -- Each deployment creates one zone with one VNet link. For hub-spoke topologies requiring links to multiple VNets, deploy separate instances or manage additional links outside this component.
