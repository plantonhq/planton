# AzureMonitorDiagnosticSetting - Pulumi Module

Pulumi (Go) implementation for the AzureMonitorDiagnosticSetting
kind, at 100% behavioral parity with the Terraform module.

## Resources Created

- `monitoring.DiagnosticSetting` -- the extension resource on the target,
  carrying the log/metric selection and destination wiring

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The destination-type wire map lives in `locals.go`, mechanically
  identical to the Terraform module's map.
- `diagnostic_setting_id` exports the CONSTRUCTED ARM extension-resource
  id (`{target}/providers/Microsoft.Insights/diagnosticSettings/{name}`)
  -- the provider's own state id is a `{target}|{name}` composite no API
  consumes.
- The setting carries no tags (ARM does not support them on diagnostic
  settings).

## Build

```bash
make build
```
