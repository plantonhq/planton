# AzureServiceBusAuthorizationRule - Terraform Module

OpenTofu/Terraform implementation for the
AzureServiceBusAuthorizationRule kind, at 100%
behavioral parity with the Pulumi module.

## Resources Created

Exactly ONE of (the spec's scope XOR picks it):

- `azurerm_servicebus_namespace_authorization_rule`
- `azurerm_servicebus_queue_authorization_rule`
- `azurerm_servicebus_topic_authorization_rule`

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- The three resources are count-gated on the scope discriminator;
  outputs coalesce PER ATTRIBUTE (never the whole resource object --
  that would taint non-sensitive outputs like the id with the key
  attributes' sensitivity).
- The alias connection-string outputs fall back to "" when the
  namespace has no geo-DR pairing.

## Validate

```bash
tofu init -backend=false && tofu validate
```
