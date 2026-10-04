# AzureRedisCacheAccessPolicyAssignment - Terraform Module

Terraform implementation for the AzureRedisCacheAccessPolicyAssignment
kind.

## Resources Created

- `azurerm_redis_cache_access_policy_assignment.main` -- the data-plane
  grant binding a policy to an Entra principal on the cache

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.redis_cache_id` | The parent cache's resolved ARM id; ForceNew |
| `spec.access_policy_name` | Built-in names pass as literals; custom policies resolve from an AzureRedisCacheAccessPolicy reference |
| `spec.object_id` | The PRINCIPAL id (never the client id -- that fails at connect time) |
| `spec.object_id_alias` | Readable label; doubles as an alternative Redis username |

## Usage

```hcl
module "redis_grant" {
  source = "./path/to/module"

  metadata = {
    name = "orders-app-data-reader"
    org  = "mycompany"
  }

  spec = {
    redis_cache_id     = "/subscriptions/.../providers/Microsoft.Cache/redis/app-cache"
    assignment_name    = "orders-app-data-reader"
    access_policy_name = "Data Reader"
    object_id          = "11111111-2222-3333-4444-555555555555"
    object_id_alias    = "orders-app"
  }
}
```

Every argument is ForceNew: replacing a grant momentarily revokes and
re-grants. No tags: ARM does not support tags on access policy
assignments (cache children).
