# AzureMssqlFailoverGroup - Terraform Module

Terraform implementation for the AzureMssqlFailoverGroup
kind.

## Resources Created

- `azurerm_mssql_failover_group.main` -- the cross-region failover
  group over a primary SQL server and its partner servers, with stable
  read-write and read-only listener endpoints

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.name` | Also the listener DNS label (`{name}.database.windows.net`) |
| `spec.server_id` | The primary server |
| `spec.partner_servers` | At least one, in a different region (dynamic blocks) |
| `spec.read_write_endpoint_failover_policy` | `AUTOMATIC` / `MANUAL` mapped to ARM wire values; `grace_minutes` is only legal with Automatic (CEL-enforced pairing, null for Manual) |
| `spec.database_ids` | Optional; an empty group is legal (empty list sent as null) |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- The listener FQDNs are composed in outputs
  (`{name}.database.windows.net` / `{name}.secondary.database.windows.net`)
  because Azure does not return them.
- `readonly_endpoint_failover_policy_enabled` unset deploys the
  provider's Disabled default -- identical on both engines.

## Usage

```hcl
module "failover_group" {
  source = "./path/to/module"

  metadata = { name = "orders-fog" }
  spec = {
    name            = "orders-fog"
    server_id       = "/subscriptions/.../servers/orders-primary"
    partner_servers = [{ id = "/subscriptions/.../servers/orders-dr" }]
    read_write_endpoint_failover_policy = {
      mode          = "AUTOMATIC"
      grace_minutes = 60
    }
    database_ids = ["/subscriptions/.../databases/orders"]
  }
}
```
