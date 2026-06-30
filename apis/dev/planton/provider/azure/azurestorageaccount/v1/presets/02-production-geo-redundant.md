# Production Geo-Redundant Storage Account

This preset creates an Azure Storage Account with geo-redundant storage (GRS), 30-day soft delete retention, IP-restricted network access, and blob versioning. GRS replicates data to Azure's paired region, providing 6 copies across two datacenters for disaster recovery. This is the standard configuration for production data that requires durability beyond a single region.

## When to Use

- Production application data that must survive a regional Azure outage
- Compliance workloads requiring geo-redundant data storage
- Critical backups and archives with stricter retention requirements

## Key Configuration Choices

- **Geo-redundant storage** (`replicationType: GRS`) -- 6 copies total: 3 in the primary region + 3 in the Azure-paired region. Use `RA_GRS` if read access to the secondary region is needed
- **30-day soft delete** (`softDeleteRetentionDays: 30`) -- 4x longer than default for production data protection
- **IP-restricted access** (`networkRules.ipRules`) -- Only specified IP ranges can access the account over the public endpoint. Combine with VNet service endpoints or private endpoints for full isolation
- **Three containers** -- `data`, `backups`, and `archives` for tiered storage organization

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (primary region; Azure selects the paired region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-trusted-cidr>` | Trusted IP range for access (e.g., `203.0.113.0/24`) | Your network team |

## Related Presets

- **01-general-purpose-v2** -- Use instead for standard workloads where geo-redundancy is not required
