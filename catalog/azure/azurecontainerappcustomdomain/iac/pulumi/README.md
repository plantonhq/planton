# AzureContainerAppCustomDomain - Pulumi Module

Pulumi implementation for the AzureContainerAppCustomDomain
kind.

## Architecture

```
containerapp.CustomDomain (one binding; managed vs BYO decided by spec)
```

## Key Design Decisions

- **One resource, conditional drift handling** -- where the Terraform
  module dispatches two `count` variants, this module creates a single
  `containerapp.CustomDomain` and applies `pulumi.IgnoreChanges` on the
  certificate fields ONLY when the spec leaves them unset (the managed
  flow, where Azure attaches its certificate asynchronously); BYO drift
  is never swallowed.
- **Create blocks on DNS ownership proof** -- the `asuid` TXT record and
  the CNAME/A record must exist before deploy; the app must have ingress
  enabled.
- **Enum wire map** matches the Terraform module exactly:
  `SNI_ENABLED` -> `SniEnabled`, `DISABLED` -> `Disabled`,
  `AUTO` -> `Auto`.
- **No tags** -- the binding is ingress configuration, not a taggable
  ARM resource.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the IaC
input. Never construct the provider inline.
