# GcpPubSubSubscription -- Research & Design Documentation

## Deployment Landscape

Google Cloud Pub/Sub subscriptions are the consumer-side primitive of the Pub/Sub
messaging system. While topics handle message ingestion, subscriptions define how
messages are delivered to consuming applications. The subscription model has evolved
significantly -- from simple pull-only delivery to supporting four distinct delivery
methods (pull, push, BigQuery, Cloud Storage), each serving different consumption
patterns.

### Provisioning Methods

| Method | Tool | Maturity |
|--------|------|----------|
| Google Cloud Console | Web UI | GA |
| gcloud CLI | `gcloud pubsub subscriptions create` | GA |
| Terraform | `google_pubsub_subscription` | GA, widely adopted |
| Pulumi | `pubsub.Subscription` | GA |
| REST API | `projects.subscriptions.create` | GA |
| Client Libraries | Go, Java, Python, Node.js, etc. | GA |

### Key Design Decisions in GCP's API

1. **Four delivery methods, mutually exclusive**: Push, BigQuery, Cloud Storage, or
   pull (default). Only one can be active. This is enforced at the API level.

2. **Immutable fields**: `name`, `filter`, `enable_message_ordering`, and `tags` are
   ForceNew -- changing them requires destroying and recreating the subscription.

3. **Ack deadline drives redelivery**: Messages not acknowledged within the deadline
   are redelivered. Range is 10-600 seconds, with client libraries typically
   extending deadlines automatically.

4. **Dead-letter requires IAM setup**: The Pub/Sub service account needs both
   Subscriber permission on the subscription and Publisher permission on the
   dead-letter topic.

## What This Component Covers (80/20 Scoping)

### Included (High-Value, Widely Used)

| Feature | Rationale |
|---------|-----------|
| Pull delivery (default) | Most common pattern; no config needed |
| Push delivery with OIDC | Webhooks, Cloud Run/Functions triggers |
| Push with no_wrapper | Raw payload delivery for external webhooks |
| BigQuery delivery | Event streaming to analytics -- very popular |
| Cloud Storage delivery | Archival, data lake ingestion |
| Dead-letter policy | Production reliability pattern |
| Retry policy | Backoff configuration for failing consumers |
| Expiration policy | Auto-cleanup of abandoned subscriptions |
| Message ordering | FIFO delivery by ordering key |
| Exactly-once delivery | Deduplication guarantees |
| Filter | Attribute-based routing |

### Excluded (Niche or Emerging)

| Feature | Reason |
|---------|--------|
| `message_transforms` | JavaScript UDF transforms -- newer, niche feature. Users needing this can use raw Terraform/Pulumi directly. |
| `tags` | Resource manager tags are immutable and niche. Framework labels handle the common use case. |

## Delivery Method Comparison

| Aspect | Pull | Push | BigQuery | Cloud Storage |
|--------|------|------|----------|---------------|
| Consumer type | Application code | HTTPS endpoint | BigQuery table | GCS bucket |
| Latency | Low (long poll) | Low (HTTP POST) | Medium (batch) | Medium (batch) |
| Ordering | Supported | Supported | N/A | N/A |
| Exactly-once | Supported | Supported | N/A | N/A |
| Backpressure | Client-controlled | Rate-limited | Auto-scaled | Auto-scaled |
| Use case | General processing | Webhooks, serverless | Analytics | Archival |

## Cross-Resource Dependencies

### StringValueOrRef Fields (Infra Chart Composability)

| Field | Default Kind | Default Output Path |
|-------|-------------|-------------------|
| `project_id` | GcpProject | `status.outputs.project_id` |
| `topic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `dead_letter_policy.dead_letter_topic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `cloud_storage_config.bucket` | GcpGcsBucket | `status.outputs.bucket_id` |

### Infra Chart Composition Scenarios

1. **Event Pipeline**: Topic -> Subscription (pull) -> Cloud Function
2. **Analytics Pipeline**: Topic -> Subscription (BigQuery) -> BigQuery Dataset
3. **Archival Pipeline**: Topic -> Subscription (Cloud Storage) -> GCS Bucket
4. **Reliable Processing**: Topic -> Subscription (pull, dead-letter) -> DLQ Topic

## Validation Rules

| Rule | Type | Expression |
|------|------|-----------|
| Delivery mutual exclusion | Message CEL | At most one of push/bigquery/cloud_storage |
| BigQuery schema exclusion | Message CEL | Not both use_topic_schema and use_table_schema |
| subscription_name pattern | Field | `^[a-zA-Z][a-zA-Z0-9\-_\.~+%]*$`, 3-255 chars |
| ack_deadline_seconds range | Field CEL | 0 (default) or 10-600 |
| max_delivery_attempts range | Field CEL | 0 (default) or 5-100 |
| filter max length | Field | 256 bytes |

## Terraform Provider Notes

- All subscription features are available in Google provider `~> 5.0`.
  Unlike GcpPubSubTopic (which needs `~> 6.0` for ingestion features),
  subscriptions do not require the v6 provider.
- `bigquery_config`, `cloud_storage_config`, and `push_config` use `ConflictsWith`
  directives in the provider schema.
- Duration fields (`message_retention_duration`, `retry_policy.minimum_backoff`, etc.)
  use diff suppression functions to normalize string representations.

## Best Practices

1. **Always set a dead-letter policy for production subscriptions.** Without it,
   poison messages block the entire subscription indefinitely.

2. **Use message ordering only when needed.** Ordering reduces throughput because
   messages with the same key are serialized.

3. **Set expiration_policy.ttl to "" for critical subscriptions.** Prevents
   accidental auto-deletion of subscriptions that might be temporarily idle.

4. **Prefer BigQuery delivery over pull+insert for analytics.** BigQuery subscriptions
   handle batching, retries, and schema mapping automatically.

5. **Use filters judiciously.** Filtered-out messages are auto-acknowledged and
   cannot be recovered. Ensure the filter expression is correct before deploying.
