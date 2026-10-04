# AzureRedisCache - Terraform Module

Terraform implementation for the AzureRedisCache kind.

## Resources Created

- `azurerm_redis_cache.main` -- the cache at the chosen tier and size
- `azurerm_redis_firewall_rule.rules` -- one per `firewall_rules` entry
  (public-endpoint IP allow-list)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.cache_name` | Globally unique DNS label; becomes `{name}.redis.cache.windows.net` |
| `spec.sku_name` | Spec enum name strings (BASIC/STANDARD/PREMIUM); absent coalesces to STANDARD (tfvars drops zero-valued proto fields); the size-family letter is derived from it |
| `spec.capacity` | `optional(number, 0)` -- zero is meaningful (C0), so the attribute defaults to 0 rather than null |
| `spec.access_keys_authentication_enabled` | The keyless posture: only false once Entra auth is on (spec-enforced) |
| `spec.redis_configuration` | Emitted as a block only when present, so an omitted block deploys Azure's engine defaults; unset memory dials are never sent |
| `spec.identity` | Type enum name strings mapped to ARM's values; pairs with MANAGED_IDENTITY persistence auth |
| `spec.patch_schedules[].day_of_week` | Spec enum name strings (MONDAY..SUNDAY) mapped to ARM's capitalized day names |

## Usage

```hcl
module "redis_cache" {
  source = "./path/to/module"

  metadata = {
    name = "app-cache"
    org  = "mycompany"
  }

  spec = {
    region         = "eastus"
    resource_group = "app-rg"
    cache_name     = "my-app-cache"
    sku_name       = "STANDARD"
    capacity       = 1
  }
}
```

Provisioning runs 15-40 minutes (azurerm's own timeouts are 3 hours).
The keys and connection strings are secret-bearing outputs; both
primary and secondary faces are exported so clients rotate with zero
downtime.
