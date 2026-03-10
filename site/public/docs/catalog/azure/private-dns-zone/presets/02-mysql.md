---
title: "MySQL Private DNS Zone"
description: "This preset creates a Private DNS Zone for Azure Database for MySQL Flexible Server Private Link connectivity. The zone name `privatelink.mysql.database.azure.com` is required by Azure for DNS..."
type: "preset"
rank: "02"
presetSlug: "02-mysql"
componentSlug: "private-dns-zone"
componentTitle: "Private DNS Zone"
provider: "azure"
icon: "package"
order: 2
---

# MySQL Private DNS Zone

This preset creates a Private DNS Zone for Azure Database for MySQL Flexible Server Private Link connectivity. The zone name `privatelink.mysql.database.azure.com` is required by Azure for DNS resolution of VNet-integrated MySQL servers. Clients in the linked VNet resolve the server's FQDN to its private IP address instead of the public endpoint.

## When to Use

- VNet-integrated MySQL Flexible Server deployments (using the `02-production-vnet` MySQL preset)
- Private networking architectures where MySQL must not have a public endpoint
- Multi-VNet environments requiring DNS resolution across peered networks

## Key Configuration Choices

- **Zone name** (`name: privatelink.mysql.database.azure.com`) -- Azure-mandated zone name for MySQL Private Link. Must be exactly this value
- **VNet link** (`vnetId`) -- Links this DNS zone to the specified VNet for DNS resolution. Add additional VNet links for peered networks
- **Registration disabled** (`registrationEnabled: false`) -- Private Link zones must not auto-register VM A records. Auto-registration is only for custom internal DNS zones

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<vnet-resource-id>` | ARM resource ID of the VNet to link | `AzureVpc` status outputs |

## Related Presets

- **01-standard** -- PostgreSQL Private Link DNS zone
