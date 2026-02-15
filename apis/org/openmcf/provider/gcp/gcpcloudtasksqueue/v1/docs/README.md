# GcpCloudTasksQueue - Research Documentation

## GCP Cloud Tasks Overview

Cloud Tasks is a fully managed service that manages the execution, dispatch, and delivery of a large number of distributed tasks. It provides at-least-once delivery with configurable rate limiting and retry policies.

### Core Concepts

- **Queue**: A named entity that manages the dispatch of tasks. Controls rate, retries, and routing.
- **Task**: A unit of work defined as an HTTP request (or App Engine request). Tasks are added to queues and dispatched according to queue configuration.
- **Dispatch**: The act of sending the task's HTTP request to its target handler.

### Task Target Types

1. **HTTP tasks** (modern, recommended): Tasks dispatched to any HTTP endpoint. The target URL, method, headers, and body are defined per-task or overridden at the queue level.
2. **App Engine tasks** (legacy): Tasks dispatched to App Engine request handlers. Routing is based on App Engine service/version/instance.

This component focuses on HTTP tasks, which is the modern and recommended approach.

## Terraform Resource: google_cloud_tasks_queue

**Provider**: `hashicorp/google ~> 6.0`

### Key Schema Fields

| Field | Type | Required | ForceNew | Description |
|-------|------|----------|----------|-------------|
| `name` | string | yes | yes | Queue name |
| `location` | string | yes | yes | GCP region |
| `project` | string | no | yes | GCP project |
| `desired_state` | string | no | no | RUNNING or PAUSED |

### Nested Blocks

- `rate_limits` -- max_dispatches_per_second, max_concurrent_dispatches, max_burst_size (computed)
- `retry_config` -- max_attempts, max_retry_duration, min_backoff, max_backoff, max_doublings
- `stackdriver_logging_config` -- sampling_ratio
- `http_target` -- http_method, header_overrides, oauth_token, oidc_token, uri_override
- `app_engine_routing_override` -- service, version, instance (NOT included in this component)

### Notable Behaviors

- `max_burst_size` in rate_limits is **computed-only** -- GCP calculates it from `max_dispatches_per_second`
- `desired_state` is a **virtual field** that triggers separate pause/resume API calls, not a PATCH update
- `name`, `location`, and `project` are all `ForceNew` -- changing them destroys and recreates the queue
- Duration fields (`min_backoff`, `max_backoff`, `max_retry_duration`) use GCP's duration format (e.g., "300s", "0.100s")

## Pulumi Resource: cloudtasks.Queue

**SDK**: `pulumi-gcp/sdk/v9/go/gcp/cloudtasks`

The Pulumi SDK mirrors the Terraform schema closely. Notable differences:
- `State` is available as an output property (not available in Terraform outputs)
- Duration diff suppression is handled automatically
- `DesiredState` maps directly to the queue's pause/resume state

## Design Decisions

### Included Features

1. **http_target** -- Queue-level HTTP configuration is the modern Cloud Tasks pattern. It enables:
   - Central authentication config (OIDC/OAuth) for all tasks in a queue
   - URI overrides for consistent routing
   - Header defaults
   This is critical for the Cloud Run + Cloud Tasks integration pattern.

2. **desired_state** -- RUNNING/PAUSED control is valuable for:
   - Maintenance windows in infra-chart deployments
   - Gradual rollout of task processing
   - Emergency queue freeze without destroying infrastructure

3. **rate_limits** and **retry_config** -- Core queue behavior that most users will configure.

4. **stackdriver_logging_config** -- Operational observability for task dispatch.

### Excluded Features

1. **app_engine_routing_override** -- App Engine is in decline. Most modern Cloud Tasks usage is HTTP-based. App Engine routing is typically configured per-task in application code, not per-queue in infrastructure. Adds significant spec complexity for a niche, legacy use case.

2. **max_burst_size** (as input) -- This field is computed-only in both Terraform and Pulumi. GCP calculates it automatically from `max_dispatches_per_second`. Attempting to set it produces errors.

3. **GCP labels** -- Cloud Tasks queues do NOT support GCP labels (confirmed in both Terraform schema and Pulumi SDK). This is a GCP API limitation, not a design choice.

### Flattening Decisions

The Terraform/Pulumi resource has several unnecessarily nested structures:

- `http_target.uri_override.path_override.path` -- Flattened to `http_target.uri_override.path`
- `http_target.uri_override.query_override.query_params` -- Flattened to `http_target.uri_override.query_params`
- `http_target.header_overrides[].header.{key,value}` -- Flattened to `http_target.header_overrides[].{key,value}`

These wrapper messages add zero semantic value. Our spec provides a cleaner, more intuitive API while the IaC modules handle the mapping back to the provider's nested structure.

### StringValueOrRef Fields (Infra-Chart Composability)

Three StringValueOrRef fields enable infra-chart composition:

1. **`project_id`** -> GcpProject (`status.outputs.project_id`)
2. **`oauth_token.service_account_email`** -> GcpServiceAccount (`status.outputs.email`)
3. **`oidc_token.service_account_email`** -> GcpServiceAccount (`status.outputs.email`)

This enables a common infra-chart pattern: create a dedicated service account for task dispatch, then wire its email into the queue's authentication config.

## Cloud Tasks Pricing (as of 2024)

- **First 1 million tasks/month**: Free
- **Next 1-10 billion tasks/month**: $0.40 per million
- **Over 10 billion tasks/month**: Custom pricing

No charges for queue creation or management. Costs are driven entirely by task creation volume.

## Common Integration Patterns

### Cloud Run + Cloud Tasks

The most common modern pattern:

```
Producer Service -> Cloud Tasks Queue -> Cloud Run Handler Service
                    (rate limited)      (OIDC authenticated)
```

The queue manages dispatch rate and retries. OIDC tokens authenticate requests to the Cloud Run service. This is configured entirely at the queue level via `http_target`.

### Cloud Functions + Cloud Tasks

Similar to Cloud Run, but targeting Cloud Functions:

```
Event Source -> Cloud Tasks Queue -> Cloud Function
               (retry config)     (OIDC authenticated)
```

### Cloud Scheduler + Cloud Tasks

For scheduled task batches:

```
Cloud Scheduler Job -> Creates Tasks -> Cloud Tasks Queue -> HTTP Handler
(cron trigger)        (via API)        (rate limited)       (authenticated)
```

## Comparison: Cloud Tasks vs Pub/Sub

| Feature | Cloud Tasks | Pub/Sub |
|---------|------------|---------|
| Pattern | Task dispatch (1:1) | Message fan-out (1:N) |
| Delivery | At-least-once | At-least-once |
| Rate limiting | Built-in | Not built-in |
| Retry | Configurable per-queue | Per-subscription |
| Target | HTTP endpoint | Push/Pull/BigQuery/GCS |
| Task dedup | Task naming | Message ID |
| Use case | Background jobs, API calls | Event streaming, pub/sub |

**Rule of thumb**: Use Cloud Tasks when you want to control *when* and *how fast* work is done. Use Pub/Sub when you want to broadcast events to multiple consumers.
