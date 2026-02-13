# AzureMssqlServer Pulumi Module Architecture

## Resource Graph

```
mssql.Server (logical server)
├── mssql.Database (per database in spec.databases)
│   └── Uses ServerId from server
└── mssql.FirewallRule (per rule in spec.firewall_rules)
    └── Uses ServerId from server
```

## Key Differences from PostgreSQL/MySQL Modules

| Aspect | PostgreSQL/MySQL | MSSQL |
|--------|-----------------|-------|
| Server has SKU | Yes | No (logical container) |
| Server has storage | Yes | No (per-database) |
| Database has SKU | No | Yes (sku_name) |
| Database has storage | No | Yes (max_size_gb) |
| VNet delegation | delegated_subnet_id | Not supported |
| Public access | Derived from subnet | Explicit boolean |
| HA config | Server-level message | Per-database zone_redundant |
| DB reference | ServerId (PG) / ServerName+RG (MySQL) | ServerId |
| Connection policy | N/A | Default/Proxy/Redirect |
| License type | N/A | BasePrice/LicenseIncluded |

## Module Structure

```
module/
├── main.go      # Resources() function - creates server, DBs, firewall rules
├── locals.go    # Locals struct - resource group, tags
└── outputs.go   # Output constant names
```

## Data Flow

1. `stackInput.Target.Spec` provides the server configuration
2. `locals.ResourceGroupName` extracted from `spec.ResourceGroup.GetValue()`
3. Server created with version, TLS, public_network_access, connection_policy
4. Each database created with its own SKU, max_size_gb, zone_redundant, license_type
5. Firewall rules created using server ID
6. Outputs exported: server_id, server_name, fqdn, administrator_login, database_ids
