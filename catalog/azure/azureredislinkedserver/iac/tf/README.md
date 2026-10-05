# AzureRedisLinkedServer - Terraform Module

Terraform implementation for the AzureRedisLinkedServer
kind.

## Resources Created

- `azurerm_redis_linked_server.main` -- the geo-replication link on the
  primary cache

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.target_redis_cache_id` | The primary cache's resolved ARM id; its NAME and RESOURCE GROUP are parsed from it in `locals.tf` (case-insensitive on the type segment -- ARM has emitted both `Redis` and `redis` casings) |
| `spec.linked_redis_cache_id` | The secondary cache's resolved ARM id |
| `spec.linked_redis_cache_location` | The secondary's region -- normally a reference to the same cache's `region` output so it can never disagree |
| `spec.server_role` | Spec enum name strings (PRIMARY/SECONDARY) mapped to ARM's capitalized values |

## Usage

```hcl
module "redis_linked_server" {
  source = "./path/to/module"

  metadata = {
    name = "app-cache-geo-link"
    org  = "mycompany"
  }

  spec = {
    target_redis_cache_id       = "/subscriptions/.../providers/Microsoft.Cache/redis/app-cache-east"
    linked_redis_cache_id       = "/subscriptions/.../providers/Microsoft.Cache/redis/app-cache-west"
    linked_redis_cache_location = "westus2"
    server_role                 = "SECONDARY"
  }
}
```

Every argument is ForceNew. Deleting the resource IS the failover
operation: unlinking makes the secondary writable. No tags: ARM does
not support tags on linked servers.
