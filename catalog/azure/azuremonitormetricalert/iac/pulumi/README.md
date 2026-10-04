# AzureMonitorMetricAlert - Pulumi Module

Pulumi (Go) implementation for the AzureMonitorMetricAlert
kind, at 100% behavioral parity with the Terraform module.

## Resources Created

- `monitoring.MetricAlert` -- the global alert rule carrying the scopes,
  exactly one condition family (spec-enforced), the evaluation cadence,
  and the action-group wiring

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- Enum wire maps live in `locals.go` (aggregations, the shared six-value
  operator vocabulary, dimension operators, sensitivities), mechanically
  identical to the Terraform module's `locals.tf` maps.
- Presence-guarded defaults: severity 3, frequency PT1M, window PT5M,
  auto-mitigate true, and the dynamic criteria's 4/4 evaluation counts --
  sent explicitly so both engines carry the same values.
- Scope and action-group references read the resolved FK values.

## Build

```bash
make build
```
