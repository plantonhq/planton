# AzurePostgresqlFlexibleServer - Terraform Module

Terraform implementation for the AzurePostgresqlFlexibleServer
kind.

## Resources Created

- `azurerm_postgresql_flexible_server.main` -- the server: compute, storage,
  networking, authentication, encryption, and lifecycle (replica/restore)
- `azurerm_postgresql_flexible_server_database.main` -- one per `databases`
  entry
- `azurerm_postgresql_flexible_server_firewall_rule.main` -- one per
  `firewall_rules` entry
- `azurerm_postgresql_flexible_server_configuration.main` -- one per
  `server_parameters` entry (destroy resets the parameter to its default)
- `azurerm_postgresql_flexible_server_active_directory_administrator.main`
  -- one per `aad_administrators` entry
- `data.azurerm_client_config.current` -- the tenant fallback for Entra
  auth and administrator grants

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.create_mode` | Spec enum name string (`DEFAULT`, `REPLICA`, `POINT_IN_TIME_RESTORE`, `GEO_RESTORE`, `REVIVE_DROPPED`); unset is not sent, matching azurerm's omitted default |
| `spec.version` | Only forwarded for a fresh (DEFAULT) server -- replicas and restores inherit the source's version |
| `spec.administrator_login` / `_password` | Sent only when non-empty (Entra-only servers and replicas omit them) |
| `spec.replication_role` | Day-2 only: Azure rejects it at create; setting `NONE` on an existing replica promotes it |
| `spec.storage_tier` | Spec enum name string (`P4`-`P80`), mapped through an exhaustive lookup so unknown values fail the plan loudly |
| `spec.identity.type` / `spec.aad_administrators[].principal_type` / `spec.high_availability.mode` | Spec enum name strings mapped to ARM values in `locals.tf` |
| `spec.authentication.tenant_id` | Falls back to the deploying credential's tenant when unset |

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
    region                 = "eastus"
    resource_group         = "my-rg"
    server_name            = "myorg-prod-pg"
    administrator_login    = "pgadmin"
    administrator_password = "P@ssw0rd1234!"
    sku_name               = "GP_Standard_D4ds_v5"
    storage_mb             = 131072

    high_availability = {
      mode                      = "ZONE_REDUNDANT"
      standby_availability_zone = "2"
    }

    databases = [
      { name = "myapp" }
    ]

    server_parameters = {
      "azure.extensions" = "PGCRYPTO"
    }
  }
}
```

## Feature Parity

This Terraform module has feature parity with the Pulumi implementation:
lifecycle modes (replica / point-in-time restore / geo-restore / revive),
Entra authentication + administrators, managed identity, customer-managed-key
encryption, elastic clusters, storage tiers, maintenance windows, VNet
injection, high availability, databases, firewall rules, server parameters,
and user tags. The `identity_principal_id` output is conditionally populated
on both engines only when the identity type includes SYSTEM_ASSIGNED.
