# AzureFrontDoorProfile - Pulumi Module

Pulumi implementation for the AzureFrontDoorProfile
kind.

## Architecture

```
cdn.FrontdoorProfile (single resource)
```

## Key Design Decisions

- **The profile is just the container** -- endpoints, origin groups,
  origins, and routes are their own components referencing this
  profile's `profile_id` output, mirroring Azure's ARM child-resource
  model. The module creates exactly one resource.
- **The sku default is materialized in the module** -- IaC inputs
  never carry proto defaults, so an unspecified sku deploys
  `Standard_AzureFrontDoor` exactly as the spec documents. The sku is
  ForceNew and Azure refuses a Premium -> Standard downgrade.
- **Log scrubbing is presence-enabled** -- Azure semantics: no rules
  means scrubbing disabled; each spec entry renders one rule with the
  service's only supported operator (match everything for the selected
  variable).
- **`identity_principal_id` exports empty rather than conditionally** --
  the output shape stays constant whether or not a system-assigned
  identity exists, so downstream references never break on a
  configuration change.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
