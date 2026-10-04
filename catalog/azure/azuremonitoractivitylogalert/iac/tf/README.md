# AzureMonitorActivityLogAlert - Terraform Module

Terraform implementation for the AzureMonitorActivityLogAlert
kind.

## Resources Created

- `azurerm_monitor_activity_log_alert.main` -- the subscription-plane
  alert that fires action groups on Activity Log events (administrative
  operations, service health, resource health, autoscale,
  recommendations)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.scopes` | At least one ARM ID the alert watches |
| `spec.criteria` | Required; `category` selects the Activity Log slice, and the recommendation / resource-health / service-health sub-criteria are mutually exclusive (CEL-enforced) |
| `spec.actions` | Action groups to fire, with optional webhook properties |
| `spec.location` | Unset deploys `global` -- the right value for almost every alert |
| `spec.enabled` | Presence-guarded; unset leaves Azure's default (true) |

## Provider Version

`azurerm ~> 5.0`.

## Behavior Notes

- The spec's enum vocabularies (categories, levels, statuses, health
  reasons, recommendation types) map row-by-row to ARM wire strings in
  `locals.tf` -- a vocabulary drift fails loudly at plan instead of
  deploying a wrong filter.
- Criteria fields are sent only when set, and empty lists become null,
  so unspecified filters match everything in the chosen category.

## Usage

```hcl
module "sh_alert" {
  source = "./path/to/module"

  metadata = { name = "service-health" }
  spec = {
    resource_group = "observability-rg"
    name           = "service-health-alert"
    scopes         = ["/subscriptions/00000000-0000-0000-0000-000000000000"]
    criteria = {
      category       = "SERVICE_HEALTH"
      service_health = { locations = ["East US"] }
    }
    actions = [{ action_group_id = "/subscriptions/.../actionGroups/oncall" }]
  }
}
```
