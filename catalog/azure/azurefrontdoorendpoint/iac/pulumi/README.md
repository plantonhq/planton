# AzureFrontDoorEndpoint - Pulumi Module

Pulumi implementation for the AzureFrontDoorEndpoint deployment
component.

## Architecture

```
cdn.FrontdoorEndpoint (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`profile_id`) -- the provider
  takes the profile's full resource id directly and derives the resource
  group and profile name from it, so the spec carries one authoritative
  parent reference and nothing that could contradict it.
- **`enabled` is sent only when explicitly disabled** -- Azure's default
  is enabled, the platform materializes the documented default centrally,
  and IaC inputs never carry proto defaults; sending nothing and
  sending `true` are behaviorally identical.
- **`host_name` is the load-bearing output** -- Azure generates a
  globally unique `{name}-{hash}.z01.azurefd.net` hostname; custom-domain
  DNS records (e.g. an AzureDnsRecord CNAME) point at this value, and the
  hash suffix is why endpoint names only need per-profile uniqueness.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
