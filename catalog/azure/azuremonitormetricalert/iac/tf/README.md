# AzureMonitorMetricAlert - Terraform Module

Terraform implementation for the AzureMonitorMetricAlert
kind.

## Resources Created

- `azurerm_monitor_metric_alert.main` -- the global alert rule with one
  dynamic block per condition family (exactly one renders, spec-enforced)

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.scopes` | Resolved ARM ids -- any metric-emitting resource; multi-scope needs `target_resource_type` + `target_resource_location` (Azure's apply-time contract) |
| criteria enums | Proto enum value names mapped verbatim in `locals.tf` (aggregation, the shared six-value operator vocabulary, dimension operators, sensitivities) |
| `spec.frequency` / `spec.window_size` | Spec-closed ISO 8601 vocabularies passed through |
| `spec.static_criteria[].threshold` | Plain number -- zero is meaningful ("any failed request") |

## Outputs

`metric_alert_id`, `metric_alert_name`.
