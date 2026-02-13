# AzurePostgresqlFlexibleServer - Terraform Module

Terraform implementation for the AzurePostgresqlFlexibleServer deployment component.

## Resources Created

- `azurerm_postgresql_flexible_server` -- The PostgreSQL Flexible Server
- `azurerm_postgresql_flexible_server_database` -- Databases (one per entry in `databases`)
- `azurerm_postgresql_flexible_server_firewall_rule` -- Firewall rules (one per entry)

## Usage

```hcl
module "postgresql" {
  source = "./path/to/module"

  metadata = {
    name = "my-pg"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region                = "eastus"
    resource_group        = "my-rg"
    name                  = "my-pg-server"
    administrator_login   = "pgadmin"
    administrator_password = "P@ssw0rd1234!"
    sku_name              = "GP_Standard_D2s_v3"
    storage_mb            = 131072

    databases = [
      { name = "myapp" }
    ]

    high_availability = {
      mode = "ZoneRedundant"
    }
  }
}
```

## Feature Parity

This Terraform module has feature parity with the Pulumi implementation:

- VNet integration (delegated subnet + private DNS zone)
- High availability (ZoneRedundant / SameZone)
- Multiple databases with custom charset/collation
- Firewall rules for public access mode
- Auto-grow storage
- Geo-redundant backup
- Standard Azure tagging
