# Scheduled Task CronJob

This preset creates a CronJob that runs every 6 hours with a Forbid concurrency policy. The most common CronJob pattern: a periodic task that should not overlap with previous runs.

## When to Use

- Periodic data processing, report generation, or cleanup tasks
- Scheduled health checks or monitoring jobs
- Any recurring task that should not run concurrently with itself

## Key Configuration Choices

- **Schedule** (`0 */6 * * *`) -- runs at the top of every 6th hour; adjust the cron expression for your schedule
- **Concurrency policy** (`Forbid`) -- if a previous run is still active when the next schedule fires, the new run is skipped
- **Restart policy** (`Never`) -- failed pods are not restarted; a new pod is created up to `backoffLimit` times
- **Backoff limit** (`3`) -- maximum 3 retries before the job is marked as failed
- **History** -- keeps 3 successful and 1 failed job for debugging

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-image>` | Container image for the task | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline output |

## Related Presets

- **02-database-backup** -- CronJob specifically configured for database backup tasks
