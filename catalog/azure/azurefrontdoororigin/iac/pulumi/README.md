# AzureFrontDoorOrigin - Pulumi Module

Pulumi implementation for the AzureFrontDoorOrigin component.

## Architecture

```
cdn.FrontdoorOrigin (single resource)
```

## Key Design Decisions

- **The parent is addressed by ARM id** (`origin_group_id`) -- the
  provider takes the origin group's full resource id directly, so the
  spec carries one authoritative parent reference.
- **`certificate_name_check_enabled` is always sent** -- the provider
  requires the value explicitly; the module materializes the documented
  `true` default (IaC inputs never carry proto defaults). Keeping it on
  is the secure posture, and Azure requires it with Private Link.
- **Private Link is PREMIUM-gated by Azure at apply time** -- the profile
  SKU lives on a different resource, so neither the spec nor the module
  can check it statically; Azure rejects the combination with a clear
  error. The spec's CELs enforce what IS statically checkable: Private
  Link requires certificate name checking, and `target_type` is required
  unless the target is a Private Link Service.
- **Target-type dialects differ per engine**: the pulumi bridge expects
  `blobSecondary`/`webSecondary` where the Terraform provider spells
  `blob_secondary`/`web_secondary`. Both land on the same ARM group id;
  each module maps the spec enum to its own provider's vocabulary.
- **No Azure tags** -- ARM does not support tags on Front Door origins;
  the platform's identity tags live on the profile.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
