# AzureManagedRedis - Pulumi Module

Pulumi implementation for the AzureManagedRedis component.

## Architecture

```
managedredis.ManagedRedis (single resource: cluster + default database)
```

## Key Design Decisions

- **Cluster + database in one resource** -- Azure's own grain: the
  mapping is 1-to-1 and the database has no life without its cluster;
  the SDK provisions both and polls each to its Running state.
- **The SKU enum is mapped row-by-row** to ARM's `Balanced_B0`-style
  wire values -- a vocabulary drift fails loudly at preview instead of
  deploying a wrong SKU.
- **Database enum defaults are materialized explicitly**
  (Encrypted/OSSCluster/VolatileLRU -- Azure's own defaults) so both
  engines send identical request bodies; IaC inputs never carry proto
  defaults, so every optional-with-default field is presence-guarded.
- **Database-derived outputs ride the default_database block** (id,
  port, both access keys) through nil-guarded appliers -- the keys stay
  empty under the keyless default.
- **CMK takes the VERSIONED Key Vault key id**; the wrapping identity
  must also be attached through the identity block (an ARM pairing
  enforced at apply time, documented on both spec fields).

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the stack
input. Never construct the provider inline.
