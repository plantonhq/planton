# AzureMonitorDiagnosticSetting - Terraform Module

Terraform implementation for the AzureMonitorDiagnosticSetting
kind.

## Resources Created

- `azurerm_monitor_diagnostic_setting.main` -- the extension resource on
  the target, carrying the log/metric selection and destination wiring

## Variable Highlights

| Variable | Notes |
| --- | --- |
| `spec.target_resource_id` | The resolved target ARM id -- any resource. ForceNew |
| `spec.enabled_logs` | category XOR category_group per entry (spec-enforced); at least one log or metric overall |
| `spec.log_analytics_destination_type` | Proto enum value name mapped in `locals.tf` (DEDICATED -> Dedicated, ...); absent lets Azure pick |
| destinations | At least one of workspace/storage/eventhub-rule/partner (spec-enforced); `eventhub_name` rides with its rule |

## Outputs

`diagnostic_setting_id` -- the CONSTRUCTED ARM extension-resource id (the
provider's own state id is a `{target}|{name}` composite no API consumes),
`diagnostic_setting_name`, `target_resource_id`.
