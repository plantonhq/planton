---
title: "Pub/Sub Event-Driven Cloud Function"
description: "This preset creates a Gen 2 Cloud Function triggered by Pub/Sub messages. It processes one event at a time per instance, uses internal-only ingress (no public HTTP endpoint), and does not retry..."
type: "preset"
rank: "02"
presetSlug: "02-pubsub-event"
componentSlug: "cloud-function"
componentTitle: "Cloud Function"
provider: "gcp"
icon: "package"
order: 2
---

# Pub/Sub Event-Driven Cloud Function

This preset creates a Gen 2 Cloud Function triggered by Pub/Sub messages. It processes one event at a time per instance, uses internal-only ingress (no public HTTP endpoint), and does not retry failed invocations by default.

## When to Use

- Asynchronous event processing triggered by Pub/Sub messages
- Background tasks like data transformation, notification dispatch, or audit logging
- Decoupled architectures where producers and consumers communicate via message queues

## Key Configuration Choices

- **Pub/Sub event trigger** (`triggerType: EVENT_TRIGGER`) -- invoked when a message is published to the topic
- **Python 3.12 runtime** (`runtime: python312`) -- change to `nodejs20`, `go122`, or `java21` as needed
- **Single concurrency** (`maxInstanceRequestConcurrency: 1`) -- processes one event per instance; simplifies event handling logic
- **Internal-only ingress** (`ingressSettings: ALLOW_INTERNAL_ONLY`) -- event-driven functions don't need public HTTP access
- **No retry** (`retryPolicy: RETRY_POLICY_DO_NOT_RETRY`) -- at-most-once delivery; change to `RETRY_POLICY_RETRY` for at-least-once (requires idempotent code)
- **Scale-to-zero** -- no cost when no messages are being processed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcp-region>` | GCP region (e.g., `us-central1`) | Your deployment region |
| `<source-bucket-name>` | GCS bucket containing the function source zip | Your build pipeline or `GcpGcsBucket` outputs |
| `<source-archive-path>` | Path to the source zip in the bucket | Your build pipeline |
| `<pubsub-topic-resource-name>` | Full Pub/Sub topic resource name (`projects/{project}/topics/{topic}`) | GCP Pub/Sub console |

## Related Presets

- **01-http-trigger** -- Use for functions invoked via HTTP requests
