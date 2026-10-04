# AzureServiceBusQueue - Terraform Module

OpenTofu/Terraform implementation for the AzureServiceBusQueue
kind, at 100% behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_servicebus_queue` -- the queue, addressed by the parent
  namespace's ARM id

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- The gate-state enum arrives as the full proto value name and is mapped
  to ARM's wire values in `locals.tf`; unset deploys Active.
- Unset lifecycle dials pass through as null so Azure's own defaults
  apply (TTL unbounded, lock PT1M, 10 deliveries).
- `namespace_name` is parsed from the resolved parent id with an
  anchored regex that fails the plan loudly on a malformed id.

## Validate

```bash
tofu init -backend=false && tofu validate
```
