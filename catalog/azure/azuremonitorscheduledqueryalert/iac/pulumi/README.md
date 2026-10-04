# AzureMonitorScheduledQueryAlert - Pulumi Module

Pulumi (Go) implementation for the AzureMonitorScheduledQueryAlert
kind, at 100% behavioral parity with the Terraform module.

## Resources Created

- `monitoring.ScheduledQueryRulesAlertV2` -- the regional log-search rule
  carrying the KQL criteria, cadence, noise controls, optional identity,
  and action wiring

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The provider caps scopes at exactly one, so the pulumi bridge flattens
  the one-item list to a singular string input; the Terraform module
  wraps the same value in a one-item list. Wire-identical.
- Enum wire maps live in `locals.go` (aggregations, operators -- this
  API's equality operator is "Equal" -- dimension operators, identity
  types), mechanically identical to the Terraform module's maps.
- Presence-guarded defaults: severity 3, evaluation frequency and window
  PT5M -- sent explicitly so both engines carry the same values.

## Build

```bash
make build
```
