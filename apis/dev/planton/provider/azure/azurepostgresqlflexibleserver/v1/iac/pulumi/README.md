# AzurePostgresqlFlexibleServer - Pulumi Module

Pulumi implementation for the AzurePostgresqlFlexibleServer deployment component.

## Architecture

The module creates three types of Azure resources:

1. **Flexible Server** (`postgresql.FlexibleServer`) -- The PostgreSQL server instance
2. **Databases** (`postgresql.FlexibleServerDatabase`) -- Application databases on the server
3. **Firewall Rules** (`postgresql.FlexibleServerFirewallRule`) -- IP-based access rules

## Resource Dependencies

```
FlexibleServer
├── FlexibleServerDatabase (per database, DependsOn: server)
└── FlexibleServerFirewallRule (per rule, DependsOn: server)
```

## Key Design Decisions

### Network Mode (C10)

Public vs private access is determined by `delegated_subnet_id`:
- Set --> `PublicNetworkAccessEnabled = false` (private VNet access)
- Not set --> `PublicNetworkAccessEnabled = true` (public, firewall rules apply)

### Authentication

Password auth is always enabled. AAD auth is omitted for v1 (80/20).

### Storage

Uses `StorageMb` (Azure's native unit). Cannot be downgraded.
`AutoGrowEnabled` available for unpredictable growth patterns.

## Running Locally

```bash
# Build
make build

# Run with Pulumi
make run

# Debug
./debug.sh
```

## Provider

Uses `pulumi-azure` v6 classic provider (`github.com/pulumi/pulumi-azure/sdk/v6/go/azure/postgresql`),
consistent with all other Azure modules in Planton.
