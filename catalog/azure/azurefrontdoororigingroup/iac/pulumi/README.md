# AzureFrontDoorOriginGroup - Pulumi Module

Pulumi implementation for the AzureFrontDoorOriginGroup deployment
component.

## Architecture

```
cdn.FrontdoorOriginGroup (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`profile_id`) -- the provider
  takes the profile's full resource id directly, so the spec carries one
  authoritative parent reference.
- **Load balancing is always sent** -- Azure requires load-balancing
  settings on every origin group. When the spec omits the block (or a
  field), the module materializes Azure's own defaults (sample size 4,
  3 successful samples, 50 ms additional latency), which is exactly what
  the spec documents an unset block to mean.
- **The health probe is sent only when configured** -- Front Door treats
  ABSENT probe settings as probing disabled, so an omitted block is a
  real behavior (all origins assumed healthy), not a defaults shortcut.
  Note for maintainers: Azure's PATCH would silently null probe settings
  on unrelated updates; azurerm ships a workaround client for this, and
  both engines inherit that behavior from the provider layer.
- **No Azure tags** -- ARM does not support tags on Front Door origin
  groups; the platform's identity tags live on the profile.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
