# AzureServiceBusTopic - Terraform Module

OpenTofu/Terraform implementation for the AzureServiceBusTopic
kind, at 100% behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_servicebus_topic` -- the topic, addressed by the parent
  namespace's ARM id

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- The gate-state enum maps to ARM's wire values in `locals.tf`
  (Active/Disabled only -- direction gating is a subscription concern);
  unset deploys Active.
- Unset lifecycle dials pass through as null so Azure's own defaults
  apply.

## Validate

```bash
tofu init -backend=false && tofu validate
```
