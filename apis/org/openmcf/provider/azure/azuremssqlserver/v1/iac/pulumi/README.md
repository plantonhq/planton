# AzureMssqlServer - Pulumi Module

## Overview

This Pulumi module provisions an Azure SQL Database logical server with databases
and firewall rules using the Azure Classic provider (v6).

## Package

```go
import "github.com/pulumi/pulumi-azure/sdk/v6/go/azure/mssql"
```

## Resources Created

| Resource | Pulumi Constructor | Identifier |
|----------|-------------------|------------|
| SQL Server | `mssql.NewServer` | Server name |
| Database | `mssql.NewDatabase` | `{server}-{db}` |
| Firewall Rule | `mssql.NewFirewallRule` | `{server}-{rule}` |

## Architecture Note

Unlike PostgreSQL (`postgresql.FlexibleServer`) and MySQL (`mysql.FlexibleServer`),
the Azure SQL logical server has **no compute or storage**. Each database
(`mssql.Database`) carries its own SKU and max storage size. The server is purely
an administrative endpoint.

## Entry Point

```go
// module/main.go
func Resources(ctx *pulumi.Context, stackInput *AzureMssqlServerStackInput) error
```

## Outputs

- `server_id` - Azure Resource Manager ID
- `server_name` - Server name
- `fqdn` - `{name}.database.windows.net`
- `administrator_login` - Admin login
- `database_ids` - Map of database name to ARM ID

## Local Development

```bash
make deps    # go mod tidy
make build   # compile module and entrypoint
make test    # run tests
make run     # run the Pulumi program
```

## Debugging

```bash
./debug.sh   # starts delve debugger on port 2345
```
