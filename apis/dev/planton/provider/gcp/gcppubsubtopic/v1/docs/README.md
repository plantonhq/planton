# GcpPubSubTopic -- Research & Design Documentation

## Pub/Sub in the GCP Ecosystem

Google Cloud Pub/Sub is a fully managed, real-time messaging service that allows
you to send and receive messages between independent applications. It follows the
publish-subscribe pattern: **publishers** send messages to a **topic**, and
**subscribers** receive messages from **subscriptions** attached to that topic.

A **topic** is the foundational resource in Pub/Sub. It is a named channel that:

- Receives messages from one or more publishers
- Fans out each message to all attached subscriptions
- Optionally validates messages against a schema
- Optionally encrypts messages with a customer-managed key (CMEK)
- Optionally enforces regional storage constraints
- Optionally retains messages for replay

Topics are global resources -- they have no inherent region. However, the
**message storage policy** can constrain where message data is persisted. This
distinction is important: the topic exists globally, but the data it holds can
be regionally bounded.

### Architecture

```
Publishers ──► Topic ──► Subscription A ──► Subscriber A (push/pull)
                   ├──► Subscription B ──► Subscriber B (push/pull)
                   └──► Subscription C ──► BigQuery / GCS (export)
```

Each subscription receives a copy of every message published to the topic after
the subscription was created. This one-to-many delivery model is what makes
Pub/Sub suitable for event-driven architectures, streaming pipelines, and
fan-out patterns.

### Topics vs. Subscriptions

| Concern | Topic | Subscription |
|---------|-------|-------------|
| Created by | Infrastructure (Planton) | Infrastructure (Planton) |
| Encryption | CMEK or Google-managed | Inherits from topic |
| Message retention | Optional (600s-31 days) | Required (10 min-31 days) |
| Schema validation | Configured here | Enforced at publish time |
| Delivery mode | N/A | Pull, push, or export |
| Ordering | Configured here (ordering key) | Enforced per-subscription |
| Dead-letter | N/A | Configured per-subscription |

Planton models topics and subscriptions as separate resources because they have
independent lifecycles -- a topic may outlive many subscription changes.

## Deployment Landscape

### Method Comparison

| Method | Topic | Subscriptions | CMEK | Schema | Ingestion | Retention |
|--------|-------|---------------|------|--------|-----------|-----------|
| GCP Console | Yes | Yes | Yes | Yes | Yes | Yes |
| `gcloud pubsub topics create` | Yes | Separate cmd | Yes | Yes | Limited | Yes |
| Terraform (`google_pubsub_topic`) | Yes | Separate resource | Yes | Yes | Yes | Yes |
| Pulumi (`gcp.pubsub.Topic`) | Yes | Separate resource | Yes | Yes | Yes | Yes |
| Planton (this component) | Yes | Separate resource | Yes | Yes | Yes | Yes |

Planton adds cross-resource composability that Terraform and Pulumi lack
natively -- topics can reference projects, KMS keys, and GCS buckets from
other Planton resources using `valueFrom`.

### Terraform Mapping

| Planton Field | Terraform Attribute |
|---------------|-------------------|
| `projectId` | `project` |
| `topicName` | `name` |
| `kmsKeyName` | `kms_key_name` |
| `messageRetentionDuration` | `message_retention_duration` |
| `messageStoragePolicy` | `message_storage_policy` |
| `schemaSettings` | `schema_settings` |
| `ingestionDataSourceSettings` | `ingestion_data_source_settings` |

### Pulumi Mapping

| Planton Field | Pulumi Property |
|---------------|----------------|
| `projectId` | `project` |
| `topicName` | `name` |
| `kmsKeyName` | `kmsKeyName` |
| `messageRetentionDuration` | `messageRetentionDuration` |
| `messageStoragePolicy` | `messageStoragePolicy` |
| `schemaSettings` | `schemaSettings` |
| `ingestionDataSourceSettings` | `ingestionDataSourceSettings` |

## Field Analysis

### Immutable Fields (ForceNew)

These fields cannot be changed after creation. Any change destroys and recreates
the topic (and all its subscriptions):

- `topic_name` -- the Pub/Sub topic name

### Mutable Fields

- `kms_key_name` -- CMEK encryption key (rotation does not affect existing messages)
- `message_retention_duration` -- can be added, changed, or removed
- `message_storage_policy` -- regional constraints
- `schema_settings` -- schema validation (changing may break existing publishers)
- `ingestion_data_source_settings` -- ingestion configuration
- `labels` -- managed by Planton framework

### Labels Support

Pub/Sub topics support GCP labels. The Planton framework applies standard labels:

- `planton-resource: true`
- `planton-resource-name: <topic_name>`
- `planton-resource-kind: gcppubsubtopic`
- `planton-organization: <metadata.org>` (if set)
- `planton-environment: <metadata.env>` (if set)
- `planton-resource-id: <metadata.id>` (if set)

### Topic Name Constraints

The topic name must satisfy:
- 3-255 characters
- Start with a letter
- Contain only: `[a-zA-Z0-9\-_\.~+%]`
- Unique within the project

The name becomes part of the fully qualified topic ID:
`projects/{project}/topics/{name}`.

## Message Retention

Pub/Sub supports two independent retention mechanisms:

### Topic-Level Retention

Configured via `message_retention_duration`. When set:

- Messages are retained for the specified duration after publication
- Any subscription can seek to a timestamp within the retention window
- Retained messages are available even if no subscription existed at publish time
- Valid range: 600s (10 minutes) to 2678400s (31 days)
- Storage costs apply proportional to message volume x retention duration

### Subscription-Level Retention

Configured on individual subscriptions (not in this component). Each subscription
has its own `message_retention_duration` (default: 7 days, max: 31 days) that
controls how long unacknowledged messages are retained.

### Interaction Between the Two

| Topic Retention | Subscription Retention | Behavior |
|----------------|----------------------|----------|
| Not set | 7 days (default) | Messages retained per-subscription only |
| 7 days | 7 days | Topic stores for 7d; subscription retains for 7d |
| 31 days | 7 days | Topic stores for 31d; subscription can seek to 31d |
| Not set | 31 days | Subscription retains for 31d; no topic-level replay |

**Key insight:** Topic-level retention enables cross-subscription replay. Without
it, each subscription only sees messages published after it was created.

## Message Storage Policy

The message storage policy constrains where message data is persisted at rest.

### `allowed_persistence_regions`

A list of GCP region IDs. When specified:

- Messages are only stored in the listed regions
- Publishers in non-allowed regions have their messages routed to an allowed region
- At least one region must be specified when the policy is present

### `enforce_in_transit`

When `true` (in addition to `allowed_persistence_regions`):

- Publish calls from non-allowed regions are **rejected** (not routed)
- Subscribe calls from non-allowed regions are **rejected**
- Provides end-to-end data residency guarantees

This is critical for compliance scenarios (GDPR, data sovereignty) where data
must never transit through unauthorized regions.

### Common Patterns

| Scenario | Regions | enforce_in_transit |
|----------|---------|-------------------|
| EU data residency | `europe-west1`, `europe-west4` | `true` |
| US data residency | `us-central1`, `us-east1` | `true` |
| Low-latency US | `us-central1` | `false` |
| Global (no constraint) | (omit policy) | N/A |

## Schema Validation

Pub/Sub schemas enforce message structure at publish time. When `schema_settings`
is configured on a topic:

1. Every published message is validated against the schema
2. Messages that fail validation are rejected with an error
3. The schema resource must already exist in the project

### Schema Resource (External to This Component)

Schemas are separate Pub/Sub resources (not managed by GcpPubSubTopic):

```
projects/{project}/schemas/{schema_name}
```

They define message structure using either:
- **Protocol Buffers** (`.proto` definitions)
- **Apache Avro** (`.avsc` JSON schemas)

### Encoding Options

| Encoding | Description | Use Case |
|----------|-------------|----------|
| `JSON` | Messages are JSON strings validated against the schema | Human-readable, debugging |
| `BINARY` | Messages are binary-encoded (Avro or Proto) | High throughput, compact |

### Schema Evolution

- Schemas support revisions -- new versions can be created without breaking existing publishers
- The topic references a specific schema (not a revision), so it uses the latest revision
- Breaking schema changes require creating a new schema and updating the topic

## CMEK (Customer-Managed Encryption Keys)

All Pub/Sub messages are encrypted at rest by default using Google-managed keys.
CMEK provides additional control over the encryption key lifecycle.

### Setup Requirements

1. **Create a KMS key** in the same region(s) where messages will be stored
2. **Grant IAM access** to the Pub/Sub service account:
   ```
   service-{PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com
   ```
   needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key
3. **Configure the topic** with `kms_key_name` pointing to the key

### Key Location Requirements

| Message Storage | KMS Key Location |
|----------------|-----------------|
| No storage policy | Key in any region |
| Single region (us-central1) | Key in us-central1 or global |
| Multi-region | Key in matching multi-region or global |

### CMEK Key Rotation

- Automatic key rotation is supported (configurable on the KMS key)
- New messages use the new key version
- Existing messages remain encrypted with the key version used at publish time
- Key destruction renders encrypted messages unreadable

### Composability with Planton

The `kms_key_name` field uses `StringValueOrRef` with `default_kind = GcpKmsKey`.
In infra charts, this creates a dependency edge:

```yaml
kmsKeyName:
  valueFrom:
    kind: GcpKmsKey
    name: pubsub-cmek
    fieldPath: status.outputs.key_id
```

The KMS key is provisioned and IAM is configured before the topic is created.

## Ingestion Data Sources

Pub/Sub supports ingesting data from five external sources directly into a topic.
This eliminates the need for custom ingestion pipelines for common cross-platform
data flows.

### Supported Sources

| Source | Use Case | Auth Model |
|--------|----------|-----------|
| **AWS Kinesis** | Migrate Kinesis streams to GCP | Federated Identity (AWS IAM role) |
| **AWS MSK** | Ingest Kafka topics from MSK | Federated Identity (AWS IAM role) |
| **Azure Event Hubs** | Ingest from Azure messaging | Azure AD + Federated Identity |
| **Cloud Storage** | Stream GCS objects as messages | GCP IAM (service account) |
| **Confluent Cloud** | Ingest from managed Kafka | Workload Identity Pool |

### When to Use Ingestion Sources

**AWS Kinesis / AWS MSK ingestion:**
- Migrating workloads from AWS to GCP
- Hybrid architectures with real-time data flow between clouds
- Consolidating streaming data from multiple clouds into GCP

**Azure Event Hubs ingestion:**
- Hybrid Azure-GCP architectures
- Migrating event-driven workloads from Azure to GCP

**Cloud Storage ingestion:**
- Processing log files, data exports, or batch data as streaming messages
- Re-importing previously exported Pub/Sub messages
- Converting batch data pipelines to streaming

**Confluent Cloud ingestion:**
- Integrating managed Kafka with GCP-native consumers
- Migrating from Confluent Cloud to GCP Pub/Sub

### Architecture: Cloud Storage Ingestion

Cloud Storage ingestion is the most common source for GCP-native workloads:

```
GCS Bucket ──► [Object Created] ──► Pub/Sub Ingestion Pipeline ──► Topic ──► Subscriptions
                                          │
                                          ├── text_format (line-delimited)
                                          ├── avro_format (binary Avro)
                                          └── pubsub_avro_format (re-import)
```

Configuration:
- `bucket` -- source bucket name (supports `valueFrom` for GcpGcsBucket reference)
- `match_glob` -- glob pattern to filter objects (e.g., `**/*.json`)
- `minimum_object_create_time` -- only ingest objects created after this timestamp
- Format: exactly one of `text_format`, `avro_format`, or `pubsub_avro_format`

### Platform Logging

All ingestion sources support `platform_logs_settings.severity` to control
pipeline observability:

| Severity | Description |
|----------|-------------|
| `DISABLED` | No platform logs |
| `DEBUG` | Verbose debugging information |
| `INFO` | Normal operational events |
| `WARNING` | Potential issues |
| `ERROR` | Failed ingestion attempts |

## Infra-Chart Composability

GcpPubSubTopic is a **Layer 1** resource in infra chart topology:

```
Layer 0: GcpProject
Layer 0-1: GcpKmsKeyRing -> GcpKmsKey
Layer 1: GcpPubSubTopic (references Project, optionally KmsKey)
Layer 1: GcpGcsBucket (source for Cloud Storage ingestion)
Layer 2: GcpPubSubSubscription (references Topic)
Layer 3+: Cloud Functions, Dataflow, BigQuery (consumers)
```

### Key Outputs for Composition

| Output | Type | Used By |
|--------|------|---------|
| `topic_id` | string | Subscriptions, Cloud Functions, Dataflow jobs, Cloud Scheduler |
| `topic_name` | string | Application code, gcloud commands, monitoring |

### What Depends on GcpPubSubTopic

| Downstream Resource | References |
|--------------------|-----------|
| GcpPubSubSubscription | `topic_id` for subscription creation |
| Cloud Functions | `topic_id` as event trigger |
| Cloud Scheduler | `topic_id` as target |
| Dataflow | `topic_id` as streaming source |
| BigQuery subscriptions | `topic_id` for direct export |
| Cloud Storage subscriptions | `topic_id` for message archival |

### What GcpPubSubTopic Depends On

| Upstream Resource | Field | Purpose |
|------------------|-------|---------|
| GcpProject | `project_id` | Parent project |
| GcpKmsKey | `kms_key_name` | CMEK encryption |
| GcpGcsBucket | `ingestion_data_source_settings.cloud_storage.bucket` | Ingestion source |

### Common Infra Chart Patterns

**Event Pipeline:**
```
GcpProject -> GcpPubSubTopic -> GcpPubSubSubscription -> Cloud Function
```

**Streaming Analytics:**
```
GcpProject -> GcpPubSubTopic -> GcpPubSubSubscription -> Dataflow -> BigQuery
```

**Data Lake Ingestion:**
```
GcpProject -> GcpGcsBucket -> GcpPubSubTopic (Cloud Storage ingestion) -> Subscriptions
```

**Secure Event Bus:**
```
GcpProject -> GcpKmsKeyRing -> GcpKmsKey -> GcpPubSubTopic (CMEK) -> Subscriptions
```

## Deliberate Exclusions

| Feature | Reason |
|---------|--------|
| `message_transforms` | Message transformation pipelines. Advanced feature, evolving API. |
| Subscriptions | Separate lifecycle; modeled as GcpPubSubSubscription. |
| Dead-letter topics | Configured on subscriptions, not topics. |
| Ordering keys | Application-level concern set per-message, not per-topic. |
| Resource tags | GCP resource tags (different from labels). Not widely adopted. |
| IAM policy bindings | Topic-level IAM is typically project-scoped. Add if demand materializes. |
| Snapshots | Operational concern, not infrastructure provisioning. |
| Seek operations | Operational concern, not infrastructure provisioning. |

These can be added in future versions if demand materializes.

## Best Practices

1. **Choose topic names carefully.** Topic names are immutable. Use descriptive,
   stable names like `order-events` or `audit-logs` rather than names tied to
   implementation details.

2. **Enable message retention for critical topics.** Without topic-level
   retention, messages are only available to existing subscriptions. Set
   `messageRetentionDuration` for topics where replay or late-binding
   subscriptions are needed.

3. **Use CMEK for regulated data.** Any topic carrying PII, financial data, or
   health information should use customer-managed encryption. Set up IAM before
   creating the topic.

4. **Constrain storage regions for compliance.** Use `messageStoragePolicy` with
   `enforceInTransit: true` for data sovereignty requirements. This prevents
   data from transiting through unauthorized regions.

5. **Use schema validation for shared topics.** When multiple teams publish to
   the same topic, schema validation prevents malformed messages from reaching
   subscribers.

6. **One ingestion source per topic.** While the API allows multiple sources,
   best practice is to dedicate a topic per ingestion source for clear
   lineage and troubleshooting.

7. **Use `valueFrom` for cross-resource references.** Rather than hardcoding
   project IDs and KMS key names, use `valueFrom` to create explicit dependency
   edges in infra charts.

8. **Monitor ingestion pipelines.** Set `platformLogsSettings.severity` to
   at least `INFO` for ingestion topics to get visibility into pipeline health.

9. **Plan for subscription topology.** Topics are cheap -- create separate topics
   for distinct event types rather than multiplexing. This simplifies filtering
   and reduces unnecessary message delivery.

10. **Consider message size.** Pub/Sub messages have a 10 MB limit. For larger
    payloads, publish a reference (GCS URI) rather than the payload itself.
