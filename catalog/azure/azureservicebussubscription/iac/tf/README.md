# AzureServiceBusSubscription - Terraform Module

OpenTofu/Terraform implementation for the AzureServiceBusSubscription
kind, at 100% behavioral parity with the Pulumi module.

## Resources Created

- `azurerm_servicebus_subscription` -- the consuming view, addressed by
  the parent topic's ARM id
- `azurerm_servicebus_subscription_rule` -- one per declared rule,
  keyed by rule name (folded: rules have no life outside their
  subscription)

## Implementation Notes

- The provider block is EMPTY -- credentials arrive as ARM_* environment
  variables (service principal or keyless OIDC).
- Rules are keyed by `rule_name`, so a duplicate name fails the plan
  instead of silently upserting one rule at ARM.
- `$Default` cannot be declared: the provider's create refuses to adopt
  the service-created catch-all (import-existence check). The spec
  reserves the name; restrictive delivery removes the catch-all once,
  out-of-band.
- The client-scoped block's presence drives
  `client_scoped_subscription_enabled`; the provider round-trips the
  internal `{name}${client_id}$D` entity naming.

## Validate

```bash
tofu init -backend=false && tofu validate
```
