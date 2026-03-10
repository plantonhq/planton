---
title: "Storage Account Private DNS Zone (Blob)"
description: "This preset creates a Private DNS Zone for Azure Storage Account Blob Private Endpoint connectivity. The zone name `privatelink.blob.core.windows.net` is required by Azure for DNS resolution of..."
type: "preset"
rank: "04"
presetSlug: "04-storage"
componentSlug: "private-dns-zone"
componentTitle: "Private DNS Zone"
provider: "azure"
icon: "package"
order: 4
---

# Storage Account Private DNS Zone (Blob)

This preset creates a Private DNS Zone for Azure Storage Account Blob Private Endpoint connectivity. The zone name `privatelink.blob.core.windows.net` is required by Azure for DNS resolution of storage blob private endpoints. Clients in the linked VNet resolve the storage account's blob FQDN to its private IP address.

## When to Use

- Storage accounts using Private Endpoints to disable public blob access
- Enterprise environments with default-deny network rules on storage accounts
- Data processing pipelines (ADF, Databricks, AKS) accessing storage over private networks

## Key Configuration Choices

- **Zone name** (`name: privatelink.blob.core.windows.net`) -- Azure-mandated zone name for Blob Storage Private Link. For other storage services, use: `privatelink.file.core.windows.net` (Files), `privatelink.table.core.windows.net` (Tables), `privatelink.queue.core.windows.net` (Queues), `privatelink.dfs.core.windows.net` (Data Lake)
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
- **03-sql-server** -- SQL Server Private Link DNS zone
