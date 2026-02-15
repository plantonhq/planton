# GcpCloudTasksQueue

Provision and manage GCP Cloud Tasks queues with configurable dispatch rates, retry policies, HTTP target routing, and authentication.

## What is Cloud Tasks?

[Cloud Tasks](https://cloud.google.com/tasks) is a fully managed service for executing, dispatching, and delivering tasks asynchronously. A **queue** manages the rate at which tasks are dispatched, the retry behavior on failure, and the routing of tasks to their HTTP handlers.

Cloud Tasks is ideal for:

- **Background processing** -- offloading work from request-response cycles
- **Rate-limited API calls** -- throttling dispatch to respect downstream limits
- **Reliable task delivery** -- automatic retries with exponential backoff
- **Microservice choreography** -- decoupling producers from consumers via HTTP dispatch

## When to Use This Component

Use `GcpCloudTasksQueue` when you need:

- A managed task queue with configurable dispatch rate limits
- Queue-level HTTP target authentication (OIDC/OAuth) for Cloud Run or Cloud Functions
- Retry policies with exponential backoff for unreliable downstream services
- Operational control (pause/resume) over task dispatch

**Not suitable for:**

- Pub/Sub-style fan-out messaging (use `GcpPubSubTopic` instead)
- Cron-scheduled jobs (use `GcpCloudSchedulerJob` instead)
- Long-running batch processing (use `GcpDataprocCluster` instead)

## Key Configuration Options

### Queue Identity

- **`queue_name`** -- Immutable name for the queue (RFC 1035 compliant, 1-63 chars)
- **`location`** -- GCP region (e.g., `us-central1`). Immutable after creation.
- **`project_id`** -- GCP project (supports `valueFrom` for infra-chart composition)

### HTTP Target (Queue-Level)

Configure default HTTP settings that apply to all tasks in the queue:

- **`http_method`** -- Default HTTP method (POST, GET, PUT, etc.)
- **`uri_override`** -- Default scheme, host, port, path, and query parameters
- **`oidc_token`** -- OIDC authentication for Cloud Run / Cloud Functions
- **`oauth_token`** -- OAuth2 authentication for Google APIs
- **`header_overrides`** -- Default HTTP headers for all tasks

This is the recommended pattern for microservices: configure auth and routing at the queue level, then enqueue tasks with just a request body.

### Rate Limits

- **`max_dispatches_per_second`** -- Maximum dispatch rate (tasks/second)
- **`max_concurrent_dispatches`** -- Maximum concurrent task executions

Note: `max_burst_size` is computed by GCP from `max_dispatches_per_second` and cannot be set directly.

### Retry Config

- **`max_attempts`** -- Maximum retry attempts (-1 for unlimited)
- **`min_backoff`** / **`max_backoff`** -- Backoff duration bounds
- **`max_doublings`** -- Exponential backoff doublings before linear increase
- **`max_retry_duration`** -- Maximum total retry time

### Operational Control

- **`desired_state`** -- `RUNNING` (default) or `PAUSED`. Paused queues accept new tasks but do not dispatch them.

## Important Notes

- Cloud Tasks queues do **NOT** support GCP labels.
- Queue name and location are **immutable** after creation.
- This component manages the queue only. Individual tasks are created by your application code using the Cloud Tasks API.
- App Engine routing override is not supported by this component (legacy feature).

## Outputs

| Output | Description |
|--------|-------------|
| `queue_id` | Fully qualified queue path: `projects/{project}/locations/{location}/queues/{name}` |
| `queue_name` | Short queue name |
| `state` | Current queue state: RUNNING, PAUSED, or DISABLED (Pulumi only) |
