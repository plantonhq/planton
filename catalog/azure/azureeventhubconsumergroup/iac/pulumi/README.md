# AzureEventHubConsumerGroup - Pulumi Module

Pulumi (Go) implementation for the AzureEventHubConsumerGroup
kind, at 100% behavioral parity with the Terraform
module.

## Resources Created

- `eventhub.EventHubConsumerGroup` -- the group, on the hub referenced
  by the spec's ARM id

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- azurerm addresses consumer groups by discrete names (resource group,
  namespace, event hub) rather than the parent's ARM id; the module
  parses all three out of the spec's single `event_hub_id` with the
  same anchored contract as the Terraform module's regex, failing
  loudly on a malformed id.
- No tags: ARM does not support tags on Event Hubs entities, so the
  platform's identity tags live on the parent namespace.
- `user_metadata` is presence-guarded -- sent only when set.

## Build

```bash
make build
```
