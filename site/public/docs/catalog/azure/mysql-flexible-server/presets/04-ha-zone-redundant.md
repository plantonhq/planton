---
title: "HA Zone-Redundant MySQL Flexible Server"
description: "This preset creates an Azure Database for MySQL Flexible Server with Zone-Redundant high availability, geo-redundant backup, and a General Purpose D4ds_v4 SKU (4 vCPU, 16 GiB RAM). The primary runs..."
type: "preset"
rank: "04"
presetSlug: "04-ha-zone-redundant"
componentSlug: "mysql-flexible-server"
componentTitle: "MySQL Flexible Server"
provider: "azure"
icon: "package"
order: 4
---

# HA Zone-Redundant MySQL Flexible Server

This preset creates an Azure Database for MySQL Flexible Server with Zone-Redundant high availability, geo-redundant backup, and a General Purpose D4ds_v4 SKU (4 vCPU, 16 GiB RAM). The primary runs in zone 1 with a warm standby in zone 2 — automatic failover takes 60–120 seconds during zone-level failures. Geo-redundant backup replicates to the paired Azure region for disaster recovery.

## When to Use

- Production MySQL databases where downtime must be minimized (99.99% SLA with HA)
- Multi-zone applications requiring resilience against datacenter-level failures
- WordPress, Laravel, or Java applications with MySQL backends requiring enterprise availability
- Compliance-driven environments mandating zone-redundant data services

## Key Configuration Choices

- **General Purpose D4ds_v4** (`skuName: GP_Standard_D4ds_v4`) -- 4 vCPU, 16 GiB at ~$250/mo. HA doubles compute to ~$500/mo
- **Zone-Redundant HA** (`highAvailability.mode: ZoneRedundant`) -- Standby in zone 2, primary in zone 1. Failover 60–120 seconds
- **64 GB storage** (`storageSizeGb: 64`) -- Provides 360 baseline IOPS with auto-grow enabled as safety net
- **14-day backup retention** (`backupRetentionDays: 14`) -- Two weeks of point-in-time restore
- **Geo-redundant backup** (`geoRedundantBackupEnabled: true`) -- ForceNew. Backups replicated to paired region
- **Auto-grow enabled** (`autoGrowEnabled: true`) -- MySQL default; prevents disk-full outages

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-server-name>` | Globally unique server name (ForceNew) | Choose a name; becomes `{name}.mysql.database.azure.com` |
| `<admin-username>` | Administrator login name (ForceNew) | Choose a name |
| `<admin-password>` | Administrator password | Generate a strong password or reference a Key Vault secret |
| `<your-database-name>` | Name of the initial database | Choose a name for your application database |

## Related Presets

- **01-production-public** -- Use instead for production without HA at lower cost
- **02-production-vnet** -- Combine VNet networking approach with this HA configuration
- **03-development** -- Use instead for cost-effective dev/test without HA
