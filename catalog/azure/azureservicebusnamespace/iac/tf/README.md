# AzureServiceBusNamespace - Terraform Module

OpenTofu/Terraform implementation for the AzureServiceBusNamespace
kind, at 100% behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_servicebus_namespace` -- the namespace, with optional managed
  identity, customer-managed-key encryption, and the inline network
  rule set

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- The SKU/identity/firewall enums arrive as full proto value names and
  are mapped to ARM's wire values in `locals.tf`; unset sku deploys
  Standard.
- The Premium pairings (capacity, partitions) and the CMK/firewall
  Premium gates are enforced by the spec's CELs, so the module passes
  them through unchanged.
- CMK removal forces namespace replacement (Azure cannot un-set BYOK) --
  the provider encodes this as a ForceNew diff.

## Validate

```bash
tofu init -backend=false && tofu validate
```
