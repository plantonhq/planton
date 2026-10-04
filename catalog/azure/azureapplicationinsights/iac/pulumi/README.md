# AzureApplicationInsights - Pulumi Module

Pulumi (Go) implementation for the AzureApplicationInsights
kind, at 100% behavioral parity with the Terraform module (one
documented bridge-lag PARITY-EXCEPTION, below).

## Resources Created

- `appinsights.Insights` -- the workspace-based component, carrying the
  application type, workspace binding, retention/sampling/cap dials,
  privacy/auth/network posture, and merged governance tags

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The application-type wire map in `locals.go` carries Azure's
  case-sensitive strings, mechanically identical to the Terraform
  module's map.
- **PARITY-EXCEPTION (bridge lag)**: pulumi-azure v6.38 bridges only the
  provider's DEPRECATED negative-form toggles (`disableIpMasking`,
  `localAuthenticationDisabled`, `dailyDataCapNotificationsDisabled`);
  this module inverts the spec's positive booleans. The wire property is
  identical for each pair -- behavior and outputs match the Terraform
  module exactly. Re-align when the bridge ships the positive forms.

## Build

```bash
make build
```
