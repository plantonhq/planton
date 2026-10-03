# AzureFrontDoorRoute - Pulumi Module

Pulumi implementation for the AzureFrontDoorRoute component.

## Architecture

```
cdn.FrontdoorRoute (single resource)
```

## Key Design Decisions

- **Two references, one parent** -- the route's ARM parent is the
  ENDPOINT (`endpoint_id`, ForceNew); the origin group (`origin_group_id`)
  is its updatable destination. Repointing the origin group is how
  traffic moves between backend pools without recreating the route.
- **`origin_ids` never reaches Azure** -- the provider uses the list
  purely to order provisioning and teardown, because ARM rejects a route
  whose origin group has no origins yet. Composed manifests list the
  group's origins here; standalone routes over pre-existing origins may
  omit it.
- **The cache block is sent only when configured** -- Front Door treats
  absent cache settings as caching disabled (the provider transmits an
  explicit null), so omitting the block is a real behavior switch, not a
  defaults bundle.
- **Rule sets and custom domains are sent only when populated** --
  Front Door distinguishes an EMPTY collection ("disassociate") from an
  absent one, so empty `rule_set_ids`/`custom_domain_ids` lists are
  omitted and absence and emptiness agree.
- **Enum defaults are materialized in the module** -- IaC inputs never
  carry proto defaults, so unspecified `forwarding_protocol` deploys
  `MatchRequest` and unspecified `query_string_caching_behavior` deploys
  `IgnoreQueryString`, exactly as the spec documents.
- **No Azure tags** -- ARM does not support tags on Front Door routes;
  the platform's identity tags live on the profile.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
