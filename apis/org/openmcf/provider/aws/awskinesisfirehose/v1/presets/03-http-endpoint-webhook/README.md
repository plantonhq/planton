# HTTP Endpoint Webhook Preset

HTTP endpoint destination for third-party log and metrics integrations. This preset targets Datadog's HTTP intake API with GZIP compression and custom attributes.

## When to Use

- **Third-party integrations** — Send logs and events to external observability platforms
- **Datadog** — Centralized log aggregation and APM
- **New Relic** — Application performance monitoring and log management
- **Sumo Logic** — Log analytics and security
- **Custom webhooks** — Any HTTPS endpoint that accepts JSON payloads

## Key Configuration

- **Datadog HTTP intake URL** — `https://http-intake.logs.datadoghq.com/v1/input`
- **Access key** — Replace with your Datadog API key (sent in `X-Amz-Firehose-Access-Key` header)
- **GZIP content encoding** — Reduces payload size for cost and latency
- **Custom attribute** — `env=production` for environment tagging
- **FailedDataOnly backup** — Failed HTTP deliveries are backed up to S3 for retry or audit

## Common Endpoints

| Provider | URL | Notes |
|----------|-----|-------|
| **Datadog** | `https://http-intake.logs.datadoghq.com/v1/input` | Use API key as access_key |
| **New Relic** | `https://log-api.newrelic.com/log/v1` | Use License key; may require Lambda for header format |
| **Sumo Logic** | `https://endpoint1.collection.region.sumologic.com/receiver/v1/http/...` | Use HTTP Source URL |
| **Honeycomb** | `https://api.honeycomb.io/1/kinesis_events/<dataset>` | Use API key as access_key |

## Prerequisites

| Resource | Description |
|----------|-------------|
| **HTTP endpoint** | HTTPS URL that accepts POST requests with JSON body |
| **API/access key** | Authentication credential from the target service |
| **S3 backup bucket** | Required for FailedDataOnly backup of failed deliveries |
| **IAM role** | Permissions for Firehose to call the endpoint and write to S3 |

## Placeholders to Replace

| Placeholder | Description |
|-------------|-------------|
| `<your-datadog-api-key>` | Datadog API key from Integrations → APIs |
| `my-firehose-backup-bucket` | S3 bucket for failed delivery backup |
| `123456789012` | Your AWS account ID |
| `firehose-http-delivery-role` | IAM role for HTTP and S3 access |
