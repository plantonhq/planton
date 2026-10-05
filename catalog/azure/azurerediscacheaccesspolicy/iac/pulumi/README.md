# AzureRedisCacheAccessPolicy - Pulumi Module

Pulumi implementation for the AzureRedisCacheAccessPolicy
kind.

## Architecture

```
redis.CacheAccessPolicy (single resource)
```

## Key Design Decisions

- **The parent cache is addressed by ARM id** (`redis_cache_id`) --
  azurerm's modern addressing; the policy is a pure child resource.
- **Permissions are raw Redis ACL syntax** -- the spec documents the
  building blocks (`+@category`, `+command`, `-@carveout`, `~pattern`)
  rather than inventing an abstraction over Redis's own vocabulary.
- **Built-in policies are never modeled as resources** -- "Data Owner",
  "Data Contributor", and "Data Reader" exist on every cache;
  assignments reference them by literal name, and the spec rejects
  creating a custom policy under a built-in name.
- **Permissions update in place**; the name and cache are ForceNew.
- **No Azure tags**: ARM does not support tags on access policies
  (cache children).

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
