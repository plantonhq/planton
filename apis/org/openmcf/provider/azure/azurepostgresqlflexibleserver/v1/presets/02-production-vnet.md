# Production PostgreSQL with VNet Integration

This preset creates an Azure Database for PostgreSQL Flexible Server injected into a virtual network subnet. Public network access is automatically disabled when `delegatedSubnetId` is set, ensuring the database is only reachable from within the VNet. A private DNS zone enables FQDN resolution to the server's private IP address. No firewall rules are needed in VNet mode.

## When to Use

- Production databases requiring private network isolation with no public internet exposure
- Compliance-driven environments (PCI-DSS, HIPAA, SOC 2) mandating private-only database access
- Applications running on VMs, AKS clusters, or other VNet-connected compute within the same network
- Zero-trust architectures where all database traffic must stay within the private network

## Key Configuration Choices

- **VNet injection** (`delegatedSubnetId`) -- Server is deployed into a subnet delegated to `Microsoft.DBforPostgreSQL/flexibleServers`. Public access is automatically disabled (ForceNew field)
- **Private DNS zone** (`privateDnsZoneId`) -- Enables VNet clients to resolve `{name}.postgres.database.azure.com` to the server's private IP
- **No firewall rules** -- Firewall rules are only effective in public access mode. Network access is controlled by VNet/subnet security instead
- **General Purpose SKU** (`skuName: GP_Standard_D2ds_v4`) -- 2 vCPU, 8 GiB RAM. Scale up for heavier workloads
- **32 GB storage** (`storageMb: 32768`) -- Minimum allowed. Cannot be downgraded after creation
- **No high availability** -- Add `highAvailability.mode: ZoneRedundant` for production SLA requirements

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | Azure region (must match VNet/subnet region) | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-server-name>` | Globally unique server name (3-63 chars, lowercase/numbers/hyphens) | Choose a name; becomes `{name}.postgres.database.azure.com` |
| `<admin-username>` | Administrator login (cannot be "admin", "root", etc.) | Your credentials policy |
| `<admin-password>` | Administrator password (8-128 chars, 3 of 4 character types) | Generate a strong password or reference a Key Vault secret |
| `<delegated-subnet-resource-id>` | ARM resource ID of a subnet delegated to PostgreSQL | `AzureSubnet` status outputs (use the `02-delegated-postgresql` subnet preset) |
| `<private-dns-zone-resource-id>` | ARM resource ID of the private DNS zone (e.g., privatelink.postgres.database.azure.com) | `AzurePrivateDnsZone` status outputs |
| `<your-database-name>` | Name of the application database | Your application configuration |

## Related Presets

- **01-production-public** -- Use instead when public access with firewall rules is acceptable
- **AzureSubnet / 02-delegated-postgresql** -- Creates the required delegated subnet for VNet injection
- **AzurePrivateDnsZone / 01-standard** -- Creates the private DNS zone for FQDN resolution
