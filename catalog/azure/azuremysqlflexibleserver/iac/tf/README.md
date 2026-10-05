# AzureMysqlFlexibleServer - Terraform Module

Terraform implementation for the AzureMysqlFlexibleServer
kind.

## Resources Created

- `azurerm_mysql_flexible_server.main` -- the server: compute, storage,
  networking, identities, encryption, and lifecycle (replica/restore)
- `azurerm_mysql_flexible_database.main` -- one per `databases` entry
- `azurerm_mysql_flexible_server_firewall_rule.main` -- one per
  `firewall_rules` entry
- `azurerm_mysql_flexible_server_configuration.main` -- one per
  `server_parameters` entry (destroy resets the parameter to its default)
- `azurerm_mysql_flexible_server_active_directory_administrator.main` --
  the single Entra administrator (MySQL supports exactly one)
- `data.azurerm_client_config.current` -- the tenant fallback for the
  Entra administrator grant

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.create_mode` | Spec enum name string (`DEFAULT`, `REPLICA`, `POINT_IN_TIME_RESTORE`, `GEO_RESTORE`); unset is not sent, matching azurerm's omitted default |
| `spec.version` | Only forwarded for a fresh (DEFAULT) server -- replicas and restores inherit the source's version |
| `spec.administrator_login` / `_password` | Sent only when non-empty (replicas and restores omit them; MySQL never disables password auth) |
| `spec.replication_role` | Day-2 only: Azure rejects it at create; setting `NONE` on an existing replica promotes it |
| `spec.storage` | MySQL's storage block: `size_gb` + provisioned `iops` XOR `io_scaling_enabled`, `auto_grow_enabled` (Azure default true), `log_on_disk_enabled` |
| `spec.public_network_access` | Spec enum name string (`ENABLED`/`DISABLED`); unset maps to null so Azure derives the value |
| `spec.high_availability.mode` | Spec enum name string mapped to ARM values in `locals.tf` |
| `spec.user_assigned_identity_ids` | MySQL supports user-assigned identities only; a non-empty list renders the `UserAssigned` identity block |
| `spec.aad_administrator.tenant_id` | Falls back to the deploying credential's tenant when unset |

## Usage

```hcl
module "mysql" {
  source = "./path/to/module"

  metadata = {
    name = "my-mysql"
    org  = "mycompany"
    env  = "production"
  }

  spec = {
    region                 = "eastus"
    resource_group         = "my-rg"
    server_name            = "myorg-prod-mysql"
    administrator_login    = "mysqladmin"
    administrator_password = "P@ssw0rd1234!"
    sku_name               = "GP_Standard_D4ds_v4"

    storage = {
      size_gb            = 256
      io_scaling_enabled = true
    }

    high_availability = {
      mode                      = "ZONE_REDUNDANT"
      standby_availability_zone = "2"
    }

    databases = [
      { name = "myapp" }
    ]

    server_parameters = {
      "require_secure_transport" = "ON"
    }
  }
}
```

## Feature Parity

This Terraform module has feature parity with the Pulumi implementation:
lifecycle modes (replica / point-in-time restore / geo-restore), the
storage profile with elastic IOPS scaling, user-assigned identities,
customer-managed-key encryption with the geo-backup key pair, the Entra
administrator, maintenance windows, VNet injection, high availability,
databases, firewall rules, server parameters, and user tags.
