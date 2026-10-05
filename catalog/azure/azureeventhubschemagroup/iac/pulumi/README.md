# AzureEventHubSchemaGroup - Pulumi Module

Pulumi (Go) implementation for the AzureEventHubSchemaGroup
kind, at 100% behavioral parity with the Terraform module.

## Resources Created

- `eventhub.NamespaceSchemaGroup` -- the schema group, addressed by the
  parent namespace's ARM id

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The compatibility and format enums map to ARM's wire values through
  maps with no unspecified fallback row -- both enums are required in
  the spec, so an unmapped value fails loudly at the provider, which is
  the right outcome.
- Every field is ForceNew (the resource has no update surface); any
  spec change plans as a replacement that drops the group's registered
  schemas.
- No tags: ARM does not support tags on Event Hubs entities, so the
  platform's identity tags live on the parent namespace.

## Build

```bash
make build
```
