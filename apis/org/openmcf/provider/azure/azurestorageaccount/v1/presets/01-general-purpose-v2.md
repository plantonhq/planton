# General-Purpose v2 Storage Account

This preset creates an Azure Storage Account with General-Purpose v2 (StorageV2), Standard tier, and locally redundant storage. It includes a default-deny network posture, blob versioning, 7-day soft delete, and two starter containers. This is the standard configuration for most application storage needs.

## When to Use

- Application data storage (blobs, files, queues, tables)
- Storing backups, logs, or artifacts
- Any workload needing a general-purpose storage account with sensible security defaults

## Key Configuration Choices

- **StorageV2 / Standard / LRS** -- General-purpose with HDD-backed locally redundant storage. The most cost-effective tier for most workloads
- **Network rules** (`networkRules.defaultAction: DENY`) -- Denies all public access by default. Azure trusted services can still access the account
- **Blob versioning** (`blobProperties.enableVersioning: true`) -- Maintains previous versions of blobs for data protection
- **7-day soft delete** -- Deleted blobs and containers are recoverable for 7 days
- **HTTPS only + TLS 1.2** -- Enforces encrypted connections with modern TLS
- **Two containers** -- `data` for application data, `backups` for backup storage. Adjust or add containers as needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |

## Related Presets

- **02-production-geo-redundant** -- Use instead for production workloads requiring geo-replication and longer retention
