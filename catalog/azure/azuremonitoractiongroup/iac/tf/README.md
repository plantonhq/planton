# AzureMonitorActionGroup - Terraform Module

Terraform implementation for the AzureMonitorActionGroup
kind.

## Resources Created

- `azurerm_monitor_action_group.main` -- the global notification hub,
  with one dynamic block per receiver family (all eleven)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.short_name` | 1-12 chars (spec-enforced); the SMS/push sender identity |
| `spec.enabled` | The group-level kill switch; a disabled group swallows alerts |
| receiver lists | All optional; a receiver-less group is a legal "null" routing target. `use_common_alert_schema` exists only on the seven payload-aware families |
| `spec.webhook_receivers[].aad_auth` | Entra-authenticated webhooks -- the keyless posture |
| `spec.itsm_receivers[].ticket_configuration` | JSON that must carry PayloadRevision + WorkItemType (spec-enforced) |

Location is left to the provider's "global" default -- the only value that
makes sense for alerting infrastructure (recorded skip in the research doc).

## Outputs

`action_group_id` (the FK seam alert rules reference), `action_group_name`.
