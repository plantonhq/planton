# AzureMssqlServer - Terraform Module

## Overview

This Terraform module provisions an Azure SQL Database logical server with
databases and firewall rules using the `azurerm` provider (~> 4.0).

## Resources Created

| Resource | Type | For Each |
|----------|------|----------|
| SQL Server | `azurerm_mssql_server` | 1 |
| Database | `azurerm_mssql_database` | Per `spec.databases` |
| Firewall Rule | `azurerm_mssql_firewall_rule` | Per `spec.firewall_rules` |

## Key Differences from PostgreSQL/MySQL Modules

- No `authentication` block on the server
- No `delegated_subnet_id` or `private_dns_zone_id`
- No `high_availability` dynamic block
- No `storage_mb` or `auto_grow_enabled` on server
- Each database has its own `sku_name`, `max_size_gb`, `zone_redundant`
- Uses `connection_policy` on the server (unique to MSSQL)
- Uses `license_type` on each database (Azure Hybrid Benefit)
- Database outputs reference `azurerm_mssql_database` (not flexible server DB)

## Inputs

See `variables.tf` for the full input specification.

## Outputs

| Output | Description |
|--------|-------------|
| `server_id` | ARM resource ID of the SQL Server |
| `server_name` | Name of the SQL Server |
| `fqdn` | Fully qualified domain name |
| `administrator_login` | Admin login name |
| `database_ids` | Map of database names to ARM IDs |

## Usage

```hcl
module "mssql" {
  source = "./path/to/module"

  metadata = {
    name = "my-sql"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region                        = "eastus"
    resource_group                = "my-rg"
    name                          = "my-sql-server"
    administrator_login           = "sqladmin"
    administrator_password        = "P@ssw0rd!"
    connection_policy             = "Redirect"
    databases = [{
      name               = "myapp"
      sku_name           = "GP_Gen5_2"
      max_size_gb        = 100
      license_type       = "BasePrice"
      zone_redundant     = true
    }]
    firewall_rules = [{
      name             = "allow-azure"
      start_ip_address = "0.0.0.0"
      end_ip_address   = "0.0.0.0"
    }]
  }
}
```
