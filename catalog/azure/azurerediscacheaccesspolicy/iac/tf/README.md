# AzureRedisCacheAccessPolicy - Terraform Module

Terraform implementation for the AzureRedisCacheAccessPolicy
kind.

## Resources Created

- `azurerm_redis_cache_access_policy.main` -- the custom data-plane
  permission set on the referenced cache

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.redis_cache_id` | The parent cache's resolved ARM id; ForceNew |
| `spec.policy_name` | What assignments reference; the spec rejects the built-in policy names |
| `spec.permissions` | Raw Redis ACL syntax; updatable in place |

## Usage

```hcl
module "redis_access_policy" {
  source = "./path/to/module"

  metadata = {
    name = "orders-read-only"
    org  = "mycompany"
  }

  spec = {
    redis_cache_id = "/subscriptions/.../providers/Microsoft.Cache/redis/app-cache"
    policy_name    = "orders-read-only"
    permissions    = "+@read +@connection ~orders:*"
  }
}
```

No tags: ARM does not support tags on access policies (cache children).
