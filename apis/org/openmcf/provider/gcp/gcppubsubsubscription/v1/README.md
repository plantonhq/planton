# GcpPubSubSubscription

Provision and manage Pub/Sub subscriptions -- the named resources that receive messages
from a topic and deliver them to subscribing applications via pull, push, BigQuery, or
Cloud Storage.

## Overview

GcpPubSubSubscription creates a subscription attached to a Pub/Sub topic within a GCP
project. Subscriptions define how messages are delivered to consumers. This resource
supports all four delivery methods:

- **Pull** (default): Subscribers pull messages via the API or client libraries.
- **Push**: Pub/Sub sends messages as HTTP POST requests to an HTTPS endpoint.
- **BigQuery**: Pub/Sub writes messages directly to a BigQuery table for analytics.
- **Cloud Storage**: Pub/Sub writes messages as objects to a Cloud Storage bucket.

Only one delivery method can be active per subscription. If none is configured, the
subscription defaults to pull delivery.

## When to Use

Use GcpPubSubSubscription when you need to:

- Consume messages from a Pub/Sub topic
- Route events to an HTTPS endpoint for webhook-driven integrations
- Stream messages into BigQuery for real-time analytics
- Archive messages to Cloud Storage for data lake ingestion
- Configure dead-letter handling for unprocessable messages
- Enable exactly-once delivery guarantees
- Filter messages by attributes before delivery

## Key Configuration Options

### Delivery Methods

| Method | Config Field | Use Case |
|--------|-------------|----------|
| Pull | *(default)* | Client library consumers, batch processing |
| Push | `push_config` | Webhooks, Cloud Run, Cloud Functions |
| BigQuery | `bigquery_config` | Real-time analytics, event logging |
| Cloud Storage | `cloud_storage_config` | Archival, data lake, compliance |

### Message Handling

| Field | Description |
|-------|-------------|
| `ack_deadline_seconds` | Time to acknowledge before redelivery (10-600s) |
| `message_retention_duration` | Backlog retention (600s-2678400s, default 7d) |
| `retain_acked_messages` | Keep acknowledged messages for replay |
| `enable_message_ordering` | Deliver in publish order by ordering key |
| `enable_exactly_once_delivery` | Guarantee exactly-once within ack deadline |
| `filter` | Attribute-based message filtering (immutable) |

### Reliability

| Field | Description |
|-------|-------------|
| `dead_letter_policy` | Route unprocessable messages to a dead-letter topic |
| `retry_policy` | Configure backoff between delivery retries |
| `expiration_policy` | Auto-delete inactive subscriptions |

## Important Behavioral Notes

- **Immutable fields**: `subscription_name`, `filter`, and `enable_message_ordering`
  cannot be changed after creation. Modifying them requires destroying and recreating
  the subscription.

- **Delivery method exclusivity**: Only one of `push_config`, `bigquery_config`, or
  `cloud_storage_config` can be set. Setting more than one is a validation error.

- **Dead-letter IAM**: The Pub/Sub service account
  (`service-{PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com`) must have
  Subscriber permission on this subscription and Publisher permission on the
  dead-letter topic.

- **Exactly-once delivery** guarantees a message is not resent before its ack deadline
  expires. It does not prevent duplicates from the publisher side.

## Dependencies

| Dependency | Field | Required |
|-----------|-------|----------|
| GcpPubSubTopic | `topic` | Yes |
| GcpProject | `project_id` | Yes |
| GcpPubSubTopic | `dead_letter_policy.dead_letter_topic` | No |
| GcpGcsBucket | `cloud_storage_config.bucket` | No |

## Related Resources

- [GcpPubSubTopic](/docs/catalog/gcp/pubsub-topic) -- Source topic for this subscription
- [GcpGcsBucket](/docs/catalog/gcp/gcs-bucket) -- Target bucket for Cloud Storage delivery
- [GcpProject](/docs/catalog/gcp/project) -- Parent GCP project
