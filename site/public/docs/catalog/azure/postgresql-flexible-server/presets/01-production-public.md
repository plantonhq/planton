---
title: "Production PostgreSQL with Public Access"
description: "This preset creates an Azure Database for PostgreSQL Flexible Server with General Purpose compute, 32 GB storage, auto-grow disabled, and public network access controlled by firewall rules. A starter..."
type: "preset"
rank: "01"
presetSlug: "01-production-public"
componentSlug: "postgresql-flexible-server"
componentTitle: "PostgreSQL Flexible Server"
provider: "azure"
icon: "package"
order: 1
---

# Production PostgreSQL with Public Access

This preset creates an Azure Database for PostgreSQL Flexible Server with General Purpose compute, 32 GB storage, auto-grow disabled, and public network access controlled by firewall rules. A starter firewall rule allows connections from Azure services, and one application database is pre-configured with UTF-8 encoding. This is the recommended starting point for production PostgreSQL workloads that do not require VNet isolation.

## When to Use

- Production applications needing a managed PostgreSQL database accessible over the internet
- Workloads where IP-based firewall rules provide sufficient network security
- Teams that prefer simple connectivity without VNet infrastructure overhead
- Services running in Azure App Service, Azure Functions, or other PaaS offerings that connect via Azure backbone

## Key Configuration Choices

- **General Purpose SKU** (`skuName: GP_Standard_D2ds_v4`) -- 2 vCPU, 8 GiB RAM. Scale up to D4ds/D8ds for heavier workloads
- **32 GB storage** (`storageMb: 32768`) -- Minimum allowed. Cannot be downgraded after creation; choose conservatively and rely on auto-grow if needed
- **Auto-grow disabled** (`autoGrowEnabled: false`) -- Default. Enable for databases with unpredictable growth to prevent out-of-storage failures
- **7-day backup retention** (`backupRetentionDays: 7`) -- Default. Increase to 35 for compliance or disaster recovery requirements
- **No geo-redundant backup** (`geoRedundantBackupEnabled: false`) -- Backups are locally redundant. Enable for cross-region disaster recovery (ForceNew)
- **Azure services firewall rule** (`firewallRules[0]`) -- Allows connections from any Azure service. Add additional rules for your application's IP ranges
- **No high availability** -- Single server instance. Add `highAvailability.mode: ZoneRedundant` for production SLA requirements

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-server-name>` | Globally unique server name (3-63 chars, lowercase/numbers/hyphens) | Choose a name; becomes `{name}.postgres.database.azure.com` |
| `<admin-username>` | Administrator login (cannot be "admin", "root", etc.) | Your credentials policy |
| `<admin-password>` | Administrator password (8-128 chars, 3 of 4 character types) | Generate a strong password or reference a Key Vault secret |
| `<your-database-name>` | Name of the application database | Your application configuration |

## Related Presets

- **02-production-vnet** -- Use instead for private VNet-injected PostgreSQL without public internet exposure
