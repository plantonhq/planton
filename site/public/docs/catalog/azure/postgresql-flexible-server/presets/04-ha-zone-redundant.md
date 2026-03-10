---
title: "HA Zone-Redundant PostgreSQL Flexible Server"
description: "This preset creates an Azure Database for PostgreSQL Flexible Server with Zone-Redundant high availability, geo-redundant backup, and a General Purpose D4ds_v4 SKU (4 vCPU, 16 GiB RAM). The primary..."
type: "preset"
rank: "04"
presetSlug: "04-ha-zone-redundant"
componentSlug: "postgresql-flexible-server"
componentTitle: "PostgreSQL Flexible Server"
provider: "azure"
icon: "package"
order: 4
---

# HA Zone-Redundant PostgreSQL Flexible Server

This preset creates an Azure Database for PostgreSQL Flexible Server with Zone-Redundant high availability, geo-redundant backup, and a General Purpose D4ds_v4 SKU (4 vCPU, 16 GiB RAM). The primary runs in zone 1 with a warm standby in zone 2 — automatic failover takes 60–120 seconds during zone-level failures. Geo-redundant backup replicates to the paired Azure region for disaster recovery. This is the recommended configuration for production databases requiring the highest availability.

## When to Use

- Production databases where downtime must be minimized (99.99% SLA with HA)
- Multi-zone applications requiring resilience against datacenter-level failures
- Compliance-driven environments (SOC 2, ISO 27001) mandating zone-redundant data services
- Databases backing business-critical APIs or transaction processing systems

## Key Configuration Choices

- **General Purpose D4ds_v4** (`skuName: GP_Standard_D4ds_v4`) -- 4 vCPU, 16 GiB at ~$250/mo. HA doubles compute to ~$500/mo. Scale up to D8ds_v4 or D16ds_v4 if 4 vCPU is insufficient
- **Zone-Redundant HA** (`highAvailability.mode: ZoneRedundant`) -- Standby in zone 2, primary in zone 1. Failover takes 60–120 seconds. Requires General Purpose or Memory Optimized tier
- **64 GB storage** (`storageMb: 65536`) -- Provides 360 baseline IOPS. Scale up for higher throughput
- **14-day backup retention** (`backupRetentionDays: 14`) -- Two weeks of point-in-time restore capability
- **Geo-redundant backup** (`geoRedundantBackupEnabled: true`) -- ForceNew field. Backups replicated to paired region for disaster recovery
- **Public access with Azure services** -- For VNet-integrated HA, combine with the `02-production-vnet` preset's networking approach

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-server-name>` | Globally unique server name (lowercase, numbers, hyphens; ForceNew) | Choose a name; becomes `{name}.postgres.database.azure.com` |
| `<admin-username>` | Administrator login name (ForceNew) | Choose a name |
| `<admin-password>` | Administrator password (8-128 chars) | Generate a strong password or reference a Key Vault secret |
| `<your-database-name>` | Name of the initial database | Choose a name for your application database |

## Related Presets

- **01-production-public** -- Use instead for production without HA at lower cost
- **02-production-vnet** -- Combine VNet networking approach with this HA configuration
- **03-development** -- Use instead for cost-effective dev/test without HA
