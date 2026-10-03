# AzureServiceBusNamespace - Pulumi Module

Pulumi (Go) implementation for the AzureServiceBusNamespace deployment
component, at 100% behavioral parity with the Terraform module.

## Resources Created

- `servicebus.Namespace` -- the namespace, with optional managed
  identity, customer-managed-key encryption, and the inline network
  rule set

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- Enum wire maps carry an unspecified row (Standard sku, Allow firewall
  action) so a zero enum never sends the empty string.
- `local_auth_enabled` and `public_network_access_enabled` are
  presence-guarded to Azure's defaults (true) -- direct iac-input
  paths do not materialize proto defaults.
- The Premium pairings (capacity, partitions) are sent only when
  present; the spec's CELs guarantee they exist exactly on PREMIUM.

## Build

```bash
make build
```
