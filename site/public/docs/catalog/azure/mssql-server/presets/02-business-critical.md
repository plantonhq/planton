---
title: "Business Critical Azure SQL Database"
description: "This preset creates an Azure SQL logical server with a Business Critical (BC_Gen5_2) vCore-based database. Business Critical tier provides local SSD storage for low-latency IO, zone-redundant..."
type: "preset"
rank: "02"
presetSlug: "02-business-critical"
componentSlug: "mssql-server"
componentTitle: "MSSQL Server"
provider: "azure"
icon: "package"
order: 2
---

# Business Critical Azure SQL Database

This preset creates an Azure SQL logical server with a Business Critical (BC_Gen5_2) vCore-based database. Business Critical tier provides local SSD storage for low-latency IO, zone-redundant replicas for high availability, and a built-in read-only secondary endpoint. This is the recommended configuration for mission-critical workloads that require the highest performance and availability guarantees.

## When to Use

- Mission-critical applications requiring 99.995% SLA with zone redundancy
- Latency-sensitive workloads benefiting from local SSD (NVMe) storage
- Applications that can offload read queries to the built-in read-only replica
- Financial, healthcare, or e-commerce systems where data availability is paramount

## Key Configuration Choices

- **Business Critical BC_Gen5_2** (`databases[0].skuName: BC_Gen5_2`) -- 2 vCores with local SSD storage. Scale up to BC_Gen5_4, BC_Gen5_8, or higher for heavier workloads
- **Zone redundancy** (`databases[0].zoneRedundant: true`) -- Replicas spread across availability zones for the highest availability SLA (99.995%)
- **100 GB max storage** (`databases[0].maxSizeGb: 100`) -- Business Critical supports up to 4096 GB. Adjust based on your data volume
- **Geo-redundant backups** (`databases[0].storageAccountType: Geo`) -- Backups replicated to the paired region for disaster recovery
- **Default connection policy** (`connectionPolicy: Default`) -- Uses Redirect for Azure-to-Azure and Proxy for external. Consider explicit "Redirect" for lowest latency between Azure services
- **Public network access** (`publicNetworkAccessEnabled: true`) -- Set to false and use `AzurePrivateEndpoint` for private-only access in production
- **Read-only replica** -- Business Critical tier automatically provides a free read-only endpoint at `{name}.database.windows.net` with `ApplicationIntent=ReadOnly` in the connection string

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region with availability zone support | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-server-name>` | Globally unique server name (3-63 chars, lowercase/numbers/hyphens) | Choose a name; becomes `{name}.database.windows.net` |
| `<admin-username>` | Administrator login (cannot be "admin", "sa", "root", etc.) | Your credentials policy |
| `<admin-password>` | Administrator password (8-128 chars, 3 of 4 character types) | Generate a strong password or reference a Key Vault secret |
| `<your-database-name>` | Name of the application database (max 128 chars) | Your application configuration |

## Related Presets

- **01-standard** -- Use instead for cost-effective workloads that don't require zone redundancy or local SSD
