# GcpCloudSchedulerJob

A deployment component for provisioning [Google Cloud Scheduler](https://cloud.google.com/scheduler) jobs through Planton.

## Overview

Cloud Scheduler is a fully managed enterprise-grade cron job scheduler. It allows you to schedule virtually any job -- including batch processing, big data operations, cloud infrastructure tasks, and other recurring workloads -- with guaranteed execution and configurable retry behavior.

GcpCloudSchedulerJob provisions a single Cloud Scheduler job that executes on a unix-cron schedule and dispatches to one of three target types:

- **HTTP targets** -- Trigger any HTTP endpoint (Cloud Run, Cloud Functions, webhooks, APIs)
- **Pub/Sub targets** -- Publish a message to a Pub/Sub topic for event-driven processing
- **App Engine targets** -- Dispatch to an App Engine handler within the same project

## When to Use

Use GcpCloudSchedulerJob when you need:

- Recurring execution of HTTP endpoints (daily reports, periodic syncs, health checks)
- Scheduled Pub/Sub message publishing (triggering batch pipelines, ETL jobs)
- Cron-based invocation of App Engine handlers
- Reliable scheduling with configurable retry and exponential backoff
- OIDC or OAuth authentication for secure endpoint invocation

## Key Configuration

| Field | Description |
|-------|-------------|
| `schedule` | Unix-cron expression (e.g., `"0 9 * * 1-5"` for weekdays at 9am) |
| `time_zone` | Timezone for schedule interpretation (default: `Etc/UTC`) |
| `http_target` | HTTP endpoint with optional OIDC/OAuth authentication |
| `pubsub_target` | Pub/Sub topic with message data and attributes |
| `app_engine_http_target` | App Engine handler with routing configuration |
| `retry_config` | Retry count, backoff durations, and doublings |
| `paused` | Create job in paused state (won't execute until resumed) |

## Important Notes

- **Exactly one target** must be specified per job (HTTP, Pub/Sub, or App Engine)
- **No GCP labels** -- Cloud Scheduler jobs do not support labels
- **`job_name` and `location` are immutable** -- changing them destroys and recreates the job
- **Body fields are base64-encoded** for HTTP and App Engine targets
- **`attempt_deadline`** varies by target type: 15s-30min for HTTP, 15s-24h for App Engine, ignored for Pub/Sub

## Presets

| Preset | Description |
|--------|-------------|
| `basic-http-job` | Simple HTTP GET on a cron schedule |
| `pubsub-publisher` | Publishes a Pub/Sub message on a schedule |
| `secure-cloud-run-trigger` | OIDC-authenticated HTTP POST to Cloud Run |

## Related Components

- [GcpPubSubTopic](/docs/catalog/gcp/pubsub-topic) -- Target for Pub/Sub scheduled publishing
- [GcpCloudTasksQueue](/docs/catalog/gcp/cloud-tasks-queue) -- Asynchronous task dispatch (not cron-based)
- [GcpCloudRun](/docs/catalog/gcp/cloud-run) -- Common HTTP target for scheduled jobs
- [GcpServiceAccount](/docs/catalog/gcp/service-account) -- Identity for OIDC/OAuth authentication
