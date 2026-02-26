# Development MySQL Flexible Server

This preset creates a minimal Azure Database for MySQL Flexible Server using the Burstable B1ms SKU (~$12/month) with 20 GB storage, no high availability, no geo-redundant backup, and public access via firewall rules. Designed for development, testing, and staging environments where cost matters more than resilience. The Burstable tier provides 1 vCPU and 2 GB RAM with credit-based burst capability — sufficient for typical dev workloads.

## When to Use

- Local development and feature branch testing against a real MySQL server
- CI/CD pipeline databases for integration tests and migration validation
- Staging environments that mirror production schema but don't need production resilience
- WordPress, Laravel, or other PHP/MySQL application development
- Proof-of-concept deployments and demo environments

## Key Configuration Choices

- **Burstable B1ms** (`skuName: B_Standard_B1ms`) -- 1 vCPU, 2 GB RAM at ~$12/month. Accumulates CPU credits during idle periods. Scale up to B2ms (2 vCPU, 4 GB) if dev workloads consistently exceed burst credits
- **20 GB storage** (`storageSizeGb: 20`) -- The minimum provisioned storage. Storage cannot be reduced after creation, so start small for dev
- **Auto-grow disabled** (`autoGrowEnabled: false`) -- Keeps costs predictable. The default is enabled, but dev environments rarely need automatic growth
- **No HA** -- High availability is omitted entirely. Dev servers don't need zone-redundant failover
- **No geo-redundant backup** (`geoRedundantBackupEnabled: false`) -- Backups stay in the same region. ForceNew field — cannot be enabled later without recreating the server
- **7-day backup retention** (`backupRetentionDays: 7`) -- Minimum retention. Sufficient for dev where point-in-time recovery is rarely needed
- **Public access** with allow-azure-services firewall rule -- Simple connectivity for dev. Add your developer IP ranges or use VNet integration for stricter access

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-server-name>` | Globally unique server name (lowercase, numbers, hyphens; ForceNew) | Choose a name; becomes `{name}.mysql.database.azure.com` |
| `<admin-username>` | Administrator login name (ForceNew — cannot change after creation) | Choose a name; cannot be `admin`, `azure_superuser`, or common reserved names |
| `<admin-password>` | Administrator password (8-128 chars, 3 of: upper, lower, number, special) | Generate a strong password or reference a Key Vault secret |
| `<your-database-name>` | Name of the initial database | Choose a name for your application database |

## Related Presets

- **01-production-public** -- Use instead for production workloads with General Purpose compute (GP_Standard_D2ds_v4)
- **02-production-vnet** -- Use instead for production workloads requiring VNet isolation with private DNS
