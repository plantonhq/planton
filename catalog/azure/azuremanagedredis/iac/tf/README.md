# AzureManagedRedis - Terraform Module

Terraform implementation for the AzureManagedRedis
kind.

## Resources Created

- `azurerm_managed_redis.main` -- the Managed Redis cluster plus its
  default database (Azure maps them 1-to-1; the provider provisions
  both from one resource)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.sku_name` | Spec enum name (`BALANCED_B0` style) mapped row-by-row to ARM's `Balanced_B0` wire values in `locals.tf` |
| `spec.high_availability_enabled` | Default true; ForceNew |
| `spec.customer_managed_key` | The VERSIONED Key Vault key id + the wrapping identity (which must also be attached via `spec.identity` -- ARM enforces the pairing at apply) |
| `spec.default_database` | Required; the database enums deploy Azure's defaults explicitly so both engines send identical bodies |
| `spec.public_network_access_enabled` | Bool mapped to the provider's Enabled/Disabled string |

## Provider Version

`azurerm ~> 5.0`. The `azurerm_managed_redis` resource family predates
the 5.0 line, so every resolvable 5.x provider carries it.

## Behavior Notes

- Provisioning and deletion run tens of minutes each (the provider
  budgets 45/30 minutes).
- Changing `clustering_policy`, `geo_replication_group_name`, or the
  module set recreates the DATABASE in place (data loss, brief
  unavailability) without replacing the cluster; the provider handles
  the delete/create sequence.
- In-place SKU changes are validated against the live instance by the
  provider; disallowed changes replace the instance.

## Usage

```hcl
module "managed_redis" {
  source = "./path/to/module"

  metadata = { name = "app-cache" }
  spec = {
    region         = "eastus"
    resource_group = "app-rg"
    cluster_name   = "app-cache"
    sku_name       = "BALANCED_B1"
    default_database = {
      access_keys_authentication_enabled = false
    }
  }
}
```
