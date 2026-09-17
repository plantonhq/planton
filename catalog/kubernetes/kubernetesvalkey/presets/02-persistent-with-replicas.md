# Production Valkey with Replicas

This preset deploys a primary plus two replicas (Redis-compatible) with
per-pod append-only persistence, write safety, a dedicated read Service, a
PodDisruptionBudget, and ACL authentication. Read scaling and restart
durability for production workloads.

## When to Use

- Production caching, session storage, or pub/sub workloads
- Environments that need read replicas for higher read throughput
- Workloads where data must survive pod restarts

## Key Configuration Choices

- **1 primary + 2 replicas** -- the chart's replication topology: writes go
  to the primary Service (`<name>`), reads load-balance across every pod
  through `<name>-read`. Note this is read scaling, NOT automated failover:
  the chart ships no Sentinel — durability through a primary restart comes
  from persistence, and Kubernetes restarts the primary in place
- **`minReplicasToWrite: 1`** -- the primary refuses writes unless at least
  one replica is connected and in sync, so a partitioned primary cannot
  accept writes the replicas would never see
- **Append-only persistence on every pod (5Gi)** -- required in replication
  mode; an ephemeral primary would replicate an empty dataset after every
  restart
- **PDB with `maxUnavailable: 1`** -- node drains take at most one pod down
  at a time
- **ACL auth declared, password minted** -- the `default` user is declared
  without a password, so the module generates one and lands it in the
  `<name>-auth` Secret; replicas authenticate to the primary with the same
  ACL machinery

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **01-single-instance** -- one instance for development and simple caching
