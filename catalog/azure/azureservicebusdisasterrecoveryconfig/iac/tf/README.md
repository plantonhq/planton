# AzureServiceBusDisasterRecoveryConfig - Terraform Module

OpenTofu/Terraform implementation for the
AzureServiceBusDisasterRecoveryConfig kind, at 100%
behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_servicebus_namespace_disaster_recovery_config` -- the geo-DR
  pairing under its alias name

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- The provider choreographs the pairing lifecycle (create-wait,
  break-pair on partner change, name-release polling on destroy) -- the
  module adds no ordering of its own.
- The optional alias rule is sent only when non-empty so Azure defaults
  to the root rule.

## Validate

```bash
tofu init -backend=false && tofu validate
```
