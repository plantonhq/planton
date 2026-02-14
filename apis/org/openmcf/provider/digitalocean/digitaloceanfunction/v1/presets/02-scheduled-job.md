# Scheduled Background Job

This preset creates a serverless function on DigitalOcean that runs on a cron schedule. It is not exposed as an HTTP endpoint, making it suitable for ETL pipelines, data synchronization, report generation, and other periodic background tasks.

## When to Use

- Periodic data processing (ETL, aggregation, cleanup)
- Scheduled report generation or notifications
- Cron-triggered background tasks that don't serve HTTP traffic

## Key Configuration Choices

- **Python 3.11 runtime** (`runtime: python_311`) -- popular for data processing. Change to any supported runtime as needed.
- **Not a web endpoint** (`isWeb: false`) -- the function is triggered by the cron schedule only, not accessible via HTTP.
- **Hourly schedule** (`cronSchedule: "0 * * * *"`) -- runs at the top of every hour. Adjust the cron expression for your cadence.
- **512 MB memory** (`memoryMb: 512`) -- more headroom for data processing than the default 256 MB.
- **60-second timeout** (`timeoutMs: 60000`) -- allows longer execution for batch operations. Maximum is 300,000 ms (5 min).
- **Encrypted secrets** (`secretEnvironmentVariables`) -- database URLs and API keys are stored securely in App Platform's secret store.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<your-org>/<your-repo>` | GitHub repository in `owner/repo` format | Your GitHub repository |
| `/functions/etl` | Path within the repo containing function code | Your repository structure |
| `<your-database-connection-string>` | Database connection URL | `DigitalOceanDatabaseCluster` status outputs or database console |
| `nyc1` | Target DigitalOcean region slug | [App Platform regions](https://docs.digitalocean.com/products/app-platform/) |

## Related Presets

- **01-web-api** -- Use instead for functions that should be exposed as HTTP endpoints
