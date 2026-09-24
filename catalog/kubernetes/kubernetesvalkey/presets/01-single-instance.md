# Single Instance Valkey

This preset deploys one Valkey instance (Redis-compatible) with append-only
persistence on a 1Gi volume, a memory ceiling with LRU eviction, and ACL
authentication. The most common shape for caching and session storage.

## When to Use

- Application caching, session storage, or rate limiting
- Development or staging environments
- Workloads where a single instance provides sufficient throughput

## Key Configuration Choices

- **Standalone topology** -- no `replication` block means one instance; add
  the block later to grow into primary/replica read scaling
- **Append-only persistence + 1Gi volume** -- a pod restart replays the AOF
  and recovers the dataset; without both, restarts start empty
- **`maxMemory: 256mb` with `allkeys-lru`** -- the cache posture: bounded
  memory with automatic eviction instead of write failures or OOM kills
- **ACL auth declared, password minted** -- the chart ships with auth OFF;
  declaring the `default` user is what actually requires credentials. No
  password is declared: the module generates one, lands it in the
  `<name>-auth` Secret (never in rendered chart values), and the
  `password_secret` output names where clients read it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **02-persistent-with-replicas** -- primary/replica Valkey with a read
  Service for production read scaling
