---
title: "Development Minimal Storage Account"
description: "This preset creates a minimal Azure Storage Account with LRS replication (single datacenter, 11 nines durability), no blob versioning, 7-day soft delete, and open network access. Designed for..."
type: "preset"
rank: "03"
presetSlug: "03-development-minimal"
componentSlug: "storage-account"
componentTitle: "Storage Account"
provider: "azure"
icon: "package"
order: 3
---

# Development Minimal Storage Account

This preset creates a minimal Azure Storage Account with LRS replication (single datacenter, 11 nines durability), no blob versioning, 7-day soft delete, and open network access. Designed for development, testing, and staging where cost and setup speed matter more than geo-redundancy or strict network controls. LRS storage costs ~$0.02/GB/month for Hot tier — the cheapest option.

## When to Use

- Local development needing a real storage account for blob uploads, queue messaging, or table storage
- CI/CD pipeline storage for test artifacts, build caches, or temporary data
- Staging environments that don't require production-level data protection
- Proof-of-concept deployments and demo environments

## Key Configuration Choices

- **LRS replication** (`replicationType: 1`) -- 3 copies in 1 datacenter at ~$0.02/GB/month (Hot tier). No cross-zone or cross-region redundancy. Upgrade to ZRS or GRS for production
- **No blob versioning** (`enableVersioning: false`) -- Saves storage cost by not maintaining previous blob versions. Enable for production data protection
- **7-day soft delete** -- Minimal safety net for accidental deletion. Production presets use 30 days
- **Open network access** (`defaultAction: ALLOW`) -- All IPs can access the account. Production presets use DENY with specific IP/VNet rules
- **Standard tier** (`accountTier: 1`) -- HDD-backed storage. Cheapest option, adequate for dev workloads
- **Single container** -- Pre-provisions an `uploads` container with private access

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |

## Related Presets

- **01-general-purpose-v2** -- Use instead for production with versioning, deny-by-default networking, and LRS
- **02-production-geo-redundant** -- Use instead for production with GRS, 30-day soft delete, and IP rules
