# Production MySQL with Public Access

This preset creates an Azure Database for MySQL Flexible Server with General Purpose compute, 32 GB storage with auto-grow enabled, and public network access controlled by firewall rules. A starter firewall rule allows connections from Azure services, and one application database is pre-configured with `utf8mb4` encoding for full Unicode support. This is the recommended starting point for production MySQL workloads that do not require VNet isolation.

## When to Use

- Production applications needing a managed MySQL database accessible over the internet
- Workloads where IP-based firewall rules provide sufficient network security
- WordPress, Laravel, Django, or other frameworks with MySQL as the primary datastore
- Services running in Azure App Service, Azure Functions, or other PaaS offerings that connect via Azure backbone

## Key Configuration Choices

- **General Purpose SKU** (`skuName: GP_Standard_D2ds_v4`) -- 2 vCPU, 8 GiB RAM. Scale up to D4ds/D8ds for heavier workloads
- **32 GB storage** (`storageSizeGb: 32`) -- Minimum practical size. Cannot be downgraded after creation
- **Auto-grow enabled** (`autoGrowEnabled: true`) -- Default for MySQL. Azure automatically increases storage when free space is low
- **MySQL 8.0.21** (`version: "8.0.21"`) -- Default. Use "8.4" for the latest GA features or "5.7" only for legacy migration
- **utf8mb4 charset** (`databases[0].charset: utf8mb4`) -- Full Unicode including emojis and supplementary characters
- **7-day backup retention** (`backupRetentionDays: 7`) -- Default. Increase to 35 for compliance requirements
- **Azure services firewall rule** (`firewallRules[0]`) -- Allows connections from any Azure service. Add additional rules for your application's IP ranges
- **No high availability** -- Single server instance. Add `highAvailability.mode: ZoneRedundant` for production SLA requirements

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-server-name>` | Globally unique server name (3-63 chars, lowercase/numbers/hyphens) | Choose a name; becomes `{name}.mysql.database.azure.com` |
| `<admin-username>` | Administrator login (1-32 chars, cannot be "admin", "root", etc.) | Your credentials policy |
| `<admin-password>` | Administrator password (8-128 chars, 3 of 4 character types) | Generate a strong password or reference a Key Vault secret |
| `<your-database-name>` | Name of the application database | Your application configuration |

## Related Presets

- **02-production-vnet** -- Use instead for private VNet-injected MySQL without public internet exposure
