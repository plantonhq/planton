# AzureEventHubAuthorizationRule - Terraform Module

OpenTofu/Terraform implementation for the
AzureEventHubAuthorizationRule kind, at 100% behavioral
parity with the Pulumi module.

## Resources Created

Exactly ONE of (the spec's scope XOR picks it):

- `azurerm_eventhub_namespace_authorization_rule`
- `azurerm_eventhub_authorization_rule`

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- Both azurerm resources still use legacy name addressing
  (`namespace_name` + `resource_group_name` [+ `eventhub_name`]), so
  `locals.tf` parses the parent names from the resolved ARM id with
  anchored regexes that fail the plan loudly on a malformed id.
- The two resources are count-gated on the scope discriminator; outputs
  coalesce PER ATTRIBUTE (never the whole resource object -- that would
  taint non-sensitive outputs like the id with the key attributes'
  sensitivity and fail the plan).
- The alias connection-string outputs fall back to "" when the
  namespace has no geo-DR pairing (a try-chain rather than coalesce,
  which rejects all-null).

## Validate

```bash
tofu init -backend=false && tofu validate
```
