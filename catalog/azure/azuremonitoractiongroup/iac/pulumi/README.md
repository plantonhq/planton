# AzureMonitorActionGroup - Pulumi Module

Pulumi (Go) implementation for the AzureMonitorActionGroup
kind, at 100% behavioral parity with the Terraform module.

## Resources Created

- `monitoring.ActionGroup` -- the global notification hub, with one
  receiver array per family (all eleven)

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- Receiver arrays are built only when non-empty; optional per-receiver
  UUIDs (webhook tenant, event hub tenant/subscription) are sent only
  when set so Azure's home-tenant defaults apply.
- Location is left to the provider's "global" default, matching the
  Terraform module (recorded skip in the research doc).
- FK-backed receiver fields (`function_app_resource_id`, `role_id`,
  `event_hub_namespace`) read the resolved reference values.

## Build

```bash
make build
```
