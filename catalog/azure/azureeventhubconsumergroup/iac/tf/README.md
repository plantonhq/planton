# AzureEventHubConsumerGroup - Terraform Module

OpenTofu/Terraform implementation for the AzureEventHubConsumerGroup
kind, at 100% behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_eventhub_consumer_group` -- the group, on the hub referenced
  by the spec's ARM id

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- azurerm addresses consumer groups by discrete names (resource group,
  namespace, event hub) rather than the parent's ARM id; `locals.tf`
  parses all three out of the spec's single `event_hub_id` with an
  anchored regex that fails the plan loudly on a malformed id.
- No tags: ARM does not support tags on Event Hubs entities, so the
  platform's identity tags live on the parent namespace.
- `user_metadata` passes through as null when unset.

## Validate

```bash
tofu init -backend=false && tofu validate
```
