# Scheduled Job Function

This preset creates a private Scaleway Serverless Function with a CRON trigger that runs daily at 2:00 AM UTC. The function uses the Python 3.11 runtime and scales to a single instance. This is the standard configuration for background tasks like data cleanup, report generation, and automated maintenance.

## When to Use

- Scheduled background tasks (data cleanup, report generation, cache warming)
- Periodic data synchronization between systems
- Any recurring job that does not need to respond to HTTP requests

## Key Configuration Choices

- **Python 3.11 runtime** (`runtime: python311`) -- well-suited for data processing and scripting; change to `node20` or `go122` as needed
- **Private privacy** (`privacy: private`) -- the function is not HTTP-accessible from the internet; only invoked by the CRON trigger
- **Single instance** (`maxScale: 1`) -- prevents concurrent executions of the same scheduled job
- **Daily at 2 AM UTC** (`schedule: "0 2 * * *"`) -- standard low-traffic maintenance window; adjust the CRON expression for your needs
- **5-minute timeout** (`timeoutSeconds: 300`) -- maximum execution time; increase for long-running batch jobs

## Placeholders to Replace

No placeholders -- this preset is ready to deploy. Update the CRON schedule and deploy your function code via Scaleway CLI.

## Related Presets

- **01-http-api** -- Use instead for functions that respond to HTTP requests
