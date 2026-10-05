# AzureEventHubSchemaGroup - Terraform Module

OpenTofu/Terraform implementation for the AzureEventHubSchemaGroup
kind, at 100% behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_eventhub_namespace_schema_group` -- the schema group,
  addressed by the parent namespace's ARM id

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- The compatibility and format enums arrive as the full proto value
  names and are mapped to ARM's wire values in `locals.tf`; both are
  required in the spec, so no unspecified fallback row exists -- an
  unmapped value fails the plan loudly, which is the right outcome.
- Every field is ForceNew (the resource has no update surface); any
  spec change plans as a replacement that drops the group's registered
  schemas.
- No tags: ARM does not support tags on Event Hubs entities, so the
  platform's identity tags live on the parent namespace.

## Validate

```bash
tofu init -backend=false && tofu validate
```
