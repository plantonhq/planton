# Single Instance Redis

This preset deploys a single-replica Redis instance with persistence enabled and 1Gi of disk storage. The most common Redis configuration for caching and session storage.

## When to Use

- Application caching, session storage, or rate limiting
- Development or staging environments
- Workloads where a single Redis instance provides sufficient throughput

## Key Configuration Choices

- **Single replica** -- no Redis Sentinel or cluster mode
- **Persistence enabled** (`true`) -- data survives pod restarts via RDB snapshots and AOF
- **1Gi disk** -- sufficient for most caching use cases; increase for larger datasets

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **02-persistent-with-replicas** -- Multi-replica Redis with persistence for production
