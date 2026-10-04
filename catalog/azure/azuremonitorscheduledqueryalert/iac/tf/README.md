# AzureMonitorScheduledQueryAlert - Terraform Module

Terraform implementation for the AzureMonitorScheduledQueryAlert
kind.

## Resources Created

- `azurerm_monitor_scheduled_query_rules_alert_v2.main` -- the regional
  log-search rule carrying the KQL criteria, cadence, noise controls,
  optional identity, and action wiring

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.region` | Must match the queried resource's region (Azure's contract) |
| `spec.scope` | The resolved workspace/App-Insights ARM id -- Azure caps scopes at one; the module wraps it in the provider's one-item list |
| criteria enums | Proto enum value names mapped verbatim in `locals.tf`; note this API's equality operator is "Equal", not "Equals" |
| `spec.auto_mitigation_enabled` / `spec.mute_actions_after_alert_duration` | Mutually exclusive (spec-enforced) |
| `spec.criteria[].failing_periods` | min <= total (spec-enforced -- a violating rule can never fire) |
| `spec.identity` | SystemAssigned / UserAssigned mapped in `locals.tf` |

## Outputs

`scheduled_query_alert_id`, `scheduled_query_alert_name`,
`identity_principal_id`.
