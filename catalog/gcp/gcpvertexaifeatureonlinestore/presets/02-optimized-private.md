# Optimized Private

## Use Case

Low-latency feature serving for a fraud or risk model, reachable only over Private Service Connect from the serving projects, encrypted under your key, with a view that streams changes from BigQuery continuously.

## When to Use

- Real-time decisions where every millisecond counts
- Serving traffic that must never cross the public internet
- A store production depends on (`PREVENT`)

## What This Creates

- An Optimized online store `realtime_store` in `us-central1` whose dedicated endpoint is served only over Private Service Connect to the two allowlisted projects
- Encryption under the `GcpKmsKey` named `vertex-ai-key`
- A feature view `merchant_risk` materializing `my-gcp-project.features.merchant_risk_latest` (the modules add `bq://`), keyed by `merchant_id`, synced continuously
- `deletionPolicy: PREVENT`, fanned to the view

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `dedicatedServingEndpoint.privateServiceConnectConfig.projectAllowlist` | two projects | The projects whose forwarding rules may target the store. |
| `featureViews[].bigQuerySource.uri` | a literal table | A `GcpBigQueryTable` reference when the table is in the same chart. |
| `featureViews[].syncConfig` | `continuous: true` | A `cron` when periodic freshness is enough (cheaper). |
| `kmsKeyName` | `vertex-ai-key` reference | Your key, in the store's region. |
