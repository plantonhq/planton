# AzureEventHubNamespaceCustomerManagedKey - Terraform Module

OpenTofu/Terraform implementation for the
AzureEventHubNamespaceCustomerManagedKey kind, at 100%
behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_eventhub_namespace_customer_managed_key` -- customer-managed-
  key (BYOK) encryption applied onto an existing Event Hubs namespace

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- ADD-ONLY lifecycle: once CMK is enabled Azure cannot remove it; the
  provider's Delete is deliberately a no-op, and returning to
  Microsoft-managed keys requires replacing the namespace itself.
- A user-assigned identity must already be attached to the parent
  namespace's identity block with wrap/unwrap vault access; unset falls
  back to the namespace's system-assigned identity (a null-if-empty
  local omits the argument).
- No tags: the configuration is a namespace property, not an ARM object
  of its own.

## Validate

```bash
tofu init -backend=false && tofu validate
```
