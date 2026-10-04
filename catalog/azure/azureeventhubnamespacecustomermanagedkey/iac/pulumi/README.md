# AzureEventHubNamespaceCustomerManagedKey - Pulumi Module

Pulumi (Go) implementation for the
AzureEventHubNamespaceCustomerManagedKey kind, at 100%
behavioral parity with the Terraform module.

## Resources Created

- `eventhub.NamespaceCustomerManagedKey` -- customer-managed-key (BYOK)
  encryption applied onto an existing Event Hubs namespace

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- ADD-ONLY lifecycle: once CMK is enabled Azure cannot remove it; the
  provider's Delete is deliberately a no-op, and returning to
  Microsoft-managed keys requires replacing the namespace itself.
- A user-assigned identity must already be attached to the parent
  namespace's identity block with wrap/unwrap vault access; the module
  sends it only when non-empty, and sends
  `infrastructure_encryption_enabled` only when set.
- No tags: the configuration is a namespace property, not an ARM object
  of its own.

## Build

```bash
make build
```
