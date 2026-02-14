# HTTP-Triggered Cloud Function

This preset creates a Gen 2 Cloud Function invoked via HTTP requests. It uses all recommended defaults: 256 MiB memory, 60-second timeout, scale-to-zero, and authenticated access. Source code is deployed from a GCS bucket archive.

## When to Use

- REST API endpoints, webhooks, or lightweight HTTP services
- Functions triggered by external HTTP clients or Cloud Scheduler
- Serverless backends where you want pay-per-invocation pricing

## Key Configuration Choices

- **HTTP trigger** (`triggerType: HTTP`) -- invoked via HTTPS endpoint
- **Node.js 20 runtime** (`runtime: nodejs20`) -- change to `python312`, `go122`, or `java21` as needed
- **Authenticated** (`allowUnauthenticated: false`) -- requires `cloudfunctions.invoker` IAM role; set to `true` for public webhooks
- **256 MiB memory** -- sufficient for most lightweight functions; higher memory also increases CPU allocation
- **60-second timeout** -- appropriate for API handlers; increase up to 3600s for long-running tasks
- **Scale-to-zero** (`minInstanceCount: 0`) -- no idle cost; cold starts apply
- **Max 100 instances** -- cost guardrail; adjust based on expected traffic

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |
| `<source-bucket-name>` | GCS bucket containing the function source zip | Your build pipeline or `GcpGcsBucket` outputs |
| `<source-archive-path>` | Path to the source zip in the bucket (e.g., `functions/my-func-v1.0.0.zip`) | Your build pipeline |

## Related Presets

- **02-pubsub-event** -- Use for event-driven functions triggered by Pub/Sub messages
