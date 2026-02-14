# Single Instance PostgreSQL

This preset deploys a single-replica PostgreSQL instance with 10Gi of persistent storage and no backup configuration. Suitable for development, testing, or low-criticality workloads.

## When to Use

- Development or staging PostgreSQL databases
- Low-traffic applications where a single instance is sufficient
- Workloads where backups are not yet needed or are managed externally

## Key Configuration Choices

- **Single replica** -- no streaming replication or automatic failover
- **10Gi disk** -- persistent storage for the PostgreSQL data directory; increase for larger datasets
- **No backups** -- backup configuration is omitted; add `backupConfig` for production use
- **No databases/users** -- managed separately or created by the application at startup

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |

## Related Presets

- **02-production-with-backup** -- Multi-replica PostgreSQL with backup configuration
