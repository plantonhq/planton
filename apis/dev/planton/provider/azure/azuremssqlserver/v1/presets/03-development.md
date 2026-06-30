# Development MSSQL Server

This preset creates an Azure SQL logical server with a single Basic-tier database (~$5/month for 5 DTU, 2 GB max). Designed for development, testing, and staging environments where cost matters more than performance or resilience. Basic tier provides 5 DTU (a blended measure of CPU, memory, and I/O) — sufficient for schema migrations, integration tests, and light application development. Backup redundancy is set to Local (same datacenter) to minimize cost.

## When to Use

- Local development and feature branch testing against a real SQL Server instance
- CI/CD pipeline databases for integration tests and EF Core / Flyway migration validation
- .NET, Java, or Node.js application development against Azure SQL
- Staging environments that mirror production schema but don't need production performance
- Proof-of-concept deployments and demo environments

## Key Configuration Choices

- **Basic SKU** (`skuName: Basic`) -- 5 DTU at ~$5/month. The cheapest Azure SQL tier. Scale up to S0 (10 DTU, ~$15/month) if 5 DTU is insufficient for development workloads
- **2 GB max** (`maxSizeGb: 2`) -- Maximum database size. Basic tier supports up to 2 GB. Sufficient for schema development and test data
- **Local backup redundancy** (`storageAccountType: Local`) -- Backups stored in the same datacenter. Cheapest option (~$0 for dev). Use Geo for production
- **No zone redundancy** (`zoneRedundant: false`) -- Zone redundancy requires Business Critical tier and adds cost. Not needed for dev
- **Public access** (`publicNetworkAccessEnabled: true`) -- Server is accessible from the internet via firewall rules. The allow-azure-services rule enables connectivity from other Azure services
- **Default connection policy** (`connectionPolicy: Default`) -- Uses Redirect inside Azure (lower latency) and Proxy outside (through gateway). No need to change for dev

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (e.g., "eastus", "westeurope") | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-server-name>` | Globally unique server name (lowercase, numbers, hyphens; ForceNew) | Choose a name; becomes `{name}.database.windows.net` |
| `<admin-username>` | Administrator login name (cannot be `admin`, `sa`, `root`, `dbmanager`, etc.) | Choose a name |
| `<admin-password>` | Administrator password (8+ chars, mix of upper, lower, number, special) | Generate a strong password or reference a Key Vault secret |
| `<your-database-name>` | Name of the initial database | Choose a name for your application database |

## Related Presets

- **01-standard** -- Use instead for production workloads with Standard S0 tier (250 GB, Geo backup)
- **02-business-critical** -- Use instead for high-availability production with Business Critical tier and zone redundancy
