# AzureServiceBusSubscription - Pulumi Module

Pulumi (Go) implementation for the AzureServiceBusSubscription
kind, at 100% behavioral parity with the Terraform
module.

## Resources Created

- `servicebus.Subscription` -- the consuming view, addressed by the
  parent topic's ARM id
- `servicebus.SubscriptionRule` -- one per declared rule, parented to
  the subscription (folded: rules have no life outside their
  subscription)

## Implementation Notes

- The Azure provider is built through the shared
  `pulumiazureprovider.Get` builder -- never inline `NewProvider`.
- The gate-state and filter-type wire maps carry the provider's
  case-sensitive values; the unspecified status row deploys Active.
- `dead_lettering_on_filter_evaluation_error` is presence-guarded to
  Azure's default (true).
- The topic/namespace names are parsed from the resolved parent id with
  the same anchored contract as the Terraform module's regexes.

## Build

```bash
make build
```
