# AzureManagedRedisAccessPolicyAssignment - Terraform Module

Terraform implementation for the AzureManagedRedisAccessPolicyAssignment
kind.

## Resources Created

- `azurerm_managed_redis_access_policy_assignment.main` -- the
  data-plane grant binding the built-in "default" policy to an Entra
  principal on the instance's default database

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.managed_redis_id` | The instance's resolved ARM id; ForceNew |
| `spec.object_id` | The PRINCIPAL id (never the client id -- that fails at connect time); Azure also names the assignment after it |

## Provider Version

`azurerm ~> 5.0`. The `azurerm_managed_redis_access_policy_assignment`
resource predates the 5.0 line, so every resolvable 5.x provider
carries it.

## Usage

```hcl
module "redis_grant" {
  source = "./path/to/module"

  metadata = { name = "app-cache-grant" }
  spec = {
    managed_redis_id = "/subscriptions/.../redisEnterprise/app-cache"
    object_id        = "11111111-2222-3333-4444-555555555555"
  }
}
```
