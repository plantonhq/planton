# Database Backup CronJob

This preset creates a daily database backup CronJob that runs at 2:00 AM UTC. Designed for PostgreSQL backup to S3-compatible storage; adapt the command for your database and storage backend.

## When to Use

- Daily automated database backups
- Any backup job that should run on a schedule with retry-on-failure semantics

## Key Configuration Choices

- **Daily at 2:00 AM UTC** (`0 2 * * *`) -- runs during typical low-traffic window; adjust for your timezone and load patterns
- **Starting deadline** (`600` seconds) -- if the job misses its schedule by more than 10 minutes, it is skipped
- **Restart on failure** (`OnFailure`) -- the pod restarts in place on failure, preserving any partial state
- **7-day history** -- retains 7 successful job records for audit trail; adjust based on your retention policy
- **Example command** -- `pg_dump` piped to gzip and uploaded to S3; replace with your backup tooling

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-backup-image>` | Image with backup tools (pg_dump, aws CLI, etc.) | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |
| `<your-database-connection-string>` | Database connection URL | Your database configuration or secrets |
| `<your-backup-bucket-name>` | S3 bucket name for storing backups | Your cloud storage console |

## Related Presets

- **01-scheduled-task** -- Generic CronJob for periodic tasks
