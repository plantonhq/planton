---
title: "Private Endpoint for Azure Blob Storage"
description: "This preset creates an Azure Private Endpoint that connects an Azure Storage Account's blob service to a VNet subnet via Private Link. It includes a DNS zone group registration so that the storage..."
type: "preset"
rank: "02"
presetSlug: "02-storage-account"
componentSlug: "private-endpoint"
componentTitle: "Private Endpoint"
provider: "azure"
icon: "package"
order: 2
---

# Private Endpoint for Azure Blob Storage

This preset creates an Azure Private Endpoint that connects an Azure Storage Account's blob service to a VNet subnet via Private Link. It includes a DNS zone group registration so that the storage account's blob FQDN resolves to a private IP within the VNet. This is the standard pattern for securing blob storage access without exposing it to the public internet.

## When to Use

- Azure Storage Accounts that must be accessible only via private IP within the VNet
- Applications that read/write blobs and need to avoid public internet traffic for security or compliance
- Data lake or data pipeline architectures where storage access must stay on the Microsoft backbone network
- Environments that disable public blob endpoints and rely exclusively on Private Link

## Key Configuration Choices

- **Sub-resource: blob** (`subresourceNames: [blob]`) -- Targets the blob sub-resource of the Storage Account. Use `table`, `queue`, or `file` for other storage services (each requires a separate private endpoint)
- **Auto-approved connection** -- The connection is auto-approved (not manual). The private endpoint owner must have appropriate permissions on the target storage account
- **DNS zone group** (`privateDnsZoneId`) -- Automatically registers an A-record in the specified `privatelink.blob.core.windows.net` zone so that `youraccount.blob.core.windows.net` resolves to the private IP
- **Dynamic IP allocation** -- The private endpoint receives a private IP dynamically from the specified subnet

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match the subnet region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-pe-name>` | Name for the private endpoint (unique within resource group) | Your naming convention |
| `<subnet-resource-id>` | Full ARM resource ID of the subnet for private IP allocation | Azure portal or `AzureSubnet` status outputs |
| `<storage-account-resource-id>` | Full ARM resource ID of the Azure Storage Account | Azure portal or `AzureStorageAccount` status outputs |
| `<private-dns-zone-id>` | Full ARM resource ID of the `privatelink.blob.core.windows.net` private DNS zone | Azure portal or `AzurePrivateDnsZone` status outputs |

## Related Presets

- **01-sql-server** -- Use instead for private connectivity to Azure SQL Database
