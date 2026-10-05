# AzureServiceBusDisasterRecoveryConfig - Pulumi Module

Pulumi (Go) implementation for the
AzureServiceBusDisasterRecoveryConfig kind, at 100%
behavioral parity with the Terraform module.

## Resources Created

- `servicebus.NamespaceDisasterRecoveryConfig` -- the geo-DR pairing
  under its alias name

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The provider choreographs the pairing lifecycle (create-wait,
  break-pair on partner change, name-release polling on destroy) -- the
  module adds no ordering of its own.
- The optional alias rule is sent only when non-empty so Azure defaults
  to the root rule.

## Build

```bash
make build
```
