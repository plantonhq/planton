# AzureEventHubNamespace - Pulumi Module

Pulumi (Go) implementation for the AzureEventHubNamespace deployment
component, at 100% behavioral parity with the Terraform module.

## Resources Created

- `eventhub.EventHubNamespace` -- the namespace, with optional managed
  identity, auto-inflate, dedicated-cluster placement, and the inline
  network rule set

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- Enum wire maps carry an unspecified row (Standard sku) so a zero enum
  never sends the empty string; the firewall's default action has no
  fallback row because the spec requires an explicit choice.
- `local_authentication_enabled` and `public_network_access_enabled` are
  presence-guarded to Azure's defaults (true) -- direct iac-input
  paths do not materialize proto defaults.
- `capacity`, `auto_inflate_enabled`, and `maximum_throughput_units` are
  sent only when present so Azure's defaults apply otherwise.
- Each `ip_rules` entry is emitted as an allow rule -- Azure's per-rule
  action accepts exactly one value (Allow).
- `identity_principal_id` exports "" unless the identity block includes
  SYSTEM_ASSIGNED, mirroring the Terraform module's
  `try(identity[0].principal_id, "")`.

## Build

```bash
make build
```
