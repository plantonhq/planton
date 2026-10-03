# AzureRedisLinkedServer - Pulumi Module

Pulumi implementation for the AzureRedisLinkedServer component.

## Architecture

```
redis.LinkedServer (single resource)
```

## Key Design Decisions

- **The primary cache is addressed by ARM id** (`target_redis_cache_id`)
  -- its name and resource group are parsed from the id with a loud
  error on a malformed id, so the parent is never spelled twice. The
  type segments match case-insensitively (ARM has emitted both
  `.../Redis/{name}` and `.../redis/{name}` casings).
- **The linked cache's location is a reference, not a repeated string**
  -- it defaults to the same referenced cache's `region` output, so the
  location can never disagree with the cache it describes.
- **Deleting the link IS the failover operation** -- unlinking makes the
  secondary writable. The module never wraps this in extra ceremony;
  the resource's lifecycle is the DR workflow.
- **Azure names the link after the linked (secondary) cache** -- there is
  no name argument; the metadata name identifies the Planton resource
  only.
- **No Azure tags**: ARM does not support tags on linked servers; the
  platform's identity tags live on the caches.

## Provider

The Azure provider is built by the shared
`pulumiazureprovider.Get(ctx, iacInput.ProviderConfig)` builder, which
dispatches static client-secret, keyless web-identity (OIDC), and
ambient credential chains. Never construct a provider inline.
