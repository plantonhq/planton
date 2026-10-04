# AzureRedisCacheAccessPolicyAssignment - Pulumi Module

Pulumi implementation for the AzureRedisCacheAccessPolicyAssignment
kind.

## Architecture

```
redis.CacheAccessPolicyAssignment (single resource)
```

## Key Design Decisions

- **The parent cache is addressed by ARM id** (`redis_cache_id`) --
  azurerm's modern addressing; the assignment is a pure child resource.
- **The policy is a reference with built-in escape hatch** -- built-in
  policies ("Data Owner", "Data Contributor", "Data Reader") are passed
  as literal values; custom policies reference an
  AzureRedisCacheAccessPolicy's name output.
- **The object id defaults to an identity's PRINCIPAL id** -- the
  workload-identity grant is the common case, and granting the client
  id instead is the classic mistake that fails at connect time, not
  deploy time (called out in the spec).
- **Everything is ForceNew** -- replacing a grant momentarily revokes
  and re-grants, which is safe for the grant class.
- **No Azure tags**: ARM does not support tags on access policy
  assignments (cache children).

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
