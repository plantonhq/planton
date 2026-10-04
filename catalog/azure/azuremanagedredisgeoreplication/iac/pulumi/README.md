# AzureManagedRedisGeoReplication - Pulumi Module

Pulumi implementation for the AzureManagedRedisGeoReplication
kind.

## Architecture

```
managedredis.GeoReplication (single resource)
```

## Key Design Decisions

- **Group membership is its own resource** -- linking mutates every
  member's replication state out of band, so it cannot honestly be a
  property of any single instance; ONE resource manages the whole group
  (linking is reciprocal).
- **Members are addressed by cluster ARM id** -- the group operates on
  each member's default database, whose path the provider derives from
  the cluster id.
- **The self-link and distinct-members contracts are documented, not
  validated statically** -- the members are references whose values
  resolve before the module runs, and Azure rejects violations at apply
  with its own diagnostics.
- **No tags** -- the group has no ARM object of its own; its resource
  ID is the managing cluster's ARM ID.

## Provider

Built via the shared `pulumiazureprovider.Get` builder -- static client
secret, keyless web identity, or ambient chain, resolved from the IaC
input. Never construct the provider inline.
