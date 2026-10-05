# GCP Pub/Sub Topic

Deploys a Pub/Sub topic in a GCP project with configurable message retention, regional storage policies, schema validation, CMEK encryption, and data ingestion from external sources (AWS Kinesis, AWS MSK, Azure Event Hubs, Cloud Storage, Confluent Cloud). Topic-level message transforms (JavaScript UDFs or Vertex AI inference) redact, normalize, or filter every published message at the topic boundary — changing what all subscriptions see — before subscribers ever run.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Pub/Sub API enablement** -- the module enables `pubsub.googleapis.com` in the target project before creating the topic (never disabled on destroy)
- **Pub/Sub Topic** -- a named topic resource in the specified GCP project for publishing and distributing messages to subscribers
- **Message Storage Policy** -- created only when `messageStoragePolicy` is configured; restricts which GCP regions may persist messages, with optional in-transit enforcement that rejects publishes from non-allowed regions
- **Schema Validation** -- created only when `schemaSettings` is configured; validates all published messages against a Pub/Sub schema in JSON or BINARY encoding
- **Ingestion Pipeline** -- created only when `ingestionDataSourceSettings` is configured; ingests data from AWS Kinesis, AWS MSK, Azure Event Hubs, Cloud Storage, or Confluent Cloud into the topic
- **CMEK Encryption** -- created only when `kmsKeyName` is provided; encrypts messages at rest using a customer-managed Cloud KMS key instead of Google-managed keys
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the topic will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef. The module enables the Pub/Sub API itself — no manual API setup is needed.
- **Cloud KMS CryptoKey Encrypter/Decrypter role** granted to the Pub/Sub service account on the KMS key (only for CMEK encryption).
- **GCS bucket** accessible to the Pub/Sub service account (only for Cloud Storage ingestion).

## Deploy

### Console

Open the deployment store, find **GCP Pub/Sub Topic**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Topic** preset in the [Presets](#presets) tab to pre-populate a minimal configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubTopic
metadata:
  name: order-events
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  topicName: order-events
```

```shell
planton apply -f pubsub-topic.yaml
```

This creates a topic with Google-managed encryption, no message retention, no regional storage constraints, and no schema validation. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the topic to a GCP project and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: pubsub-encryption-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project and KMS key first, then provisions the topic with CMEK encryption.

## Key Configuration

These are the most important decisions when configuring a Pub/Sub topic. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Topic name** -- The `topicName` field is immutable after creation. Changing it requires destroying and recreating the topic along with all attached subscriptions. Choose a descriptive, stable name.

**Message retention** -- Set `messageRetentionDuration` (e.g., `"604800s"` for 7 days) to retain published messages at the topic level independently of subscription-level retention. Enables any subscription to seek to a timestamp within the retention window. Range: 600s to 2678400s.

**Regional storage policy** -- Configure `messageStoragePolicy.allowedPersistenceRegions` to restrict where messages are stored. Set `enforceInTransit: true` to additionally reject publish calls from non-allowed regions. Required for data residency compliance.

**Schema validation** -- Set `schemaSettings.schema` to a GcpPubSubSchema reference (its `schema_id` output) or a fully qualified literal path to validate all published messages. Invalid messages are rejected at publish time. Encoding must be `JSON` or `BINARY`. If the referenced schema is later deleted, the topic validates against the `_deleted-schema_` sentinel and publishes fail -- destroy topics before their schema.

**Schema revision pinning** -- `schemaSettings.firstRevisionId` and `lastRevisionId` bound which schema revisions the topic accepts; pinning BOTH to the same revision (reference the schema's `revision_id` output) freezes the contract exactly as this deploy committed it, instead of accepting whatever revision lands next.

**Data ingestion** -- Configure `ingestionDataSourceSettings` to stream data from external sources (AWS Kinesis, AWS MSK, Azure Event Hubs, Cloud Storage, or Confluent Cloud) directly into the topic without writing custom publisher code.

**Topic-level transforms** -- `messageTransforms` run in list order on every published message before it is stored; a step returning null drops the message. They change what EVERY subscription sees — redaction belongs here, per-consumer reshaping belongs on the subscription. The `disabled` flag stages a transform in or out without losing its pipeline position.

**Tags are create-time only** -- `resourceManagerTags` cannot change after creation: editing them REPLACES the topic and detaches every subscription with it. Decide the tag posture before the topic carries production traffic.

**Destroy semantics** -- Deleting a topic leaves its subscriptions alive but starving — they receive no new messages and drain to empty. For the topic an event pipeline publishes into, set `deletionPolicy: PREVENT` so destroy fails; `ABANDON` unmanages the topic while publishers and subscriptions keep working.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** (optional) | `kmsKeyName` | `status.outputs.key_id` |
| **GcpPubSubSchema** (optional) | `schemaSettings.schema` | `status.outputs.schema_id` |
| **GcpPubSubSchema** (optional) | `schemaSettings.firstRevisionId` / `lastRevisionId` | `status.outputs.revision_id` |
| **GcpGcsBucket** (optional) | `ingestionDataSourceSettings.cloudStorage.bucket` | `status.outputs.bucket_id` |
| **GcpServiceAccount** (optional) | ingestion `gcpServiceAccount` fields | `status.outputs.email` |
| **GcpVertexAiEndpoint** (optional, per transform) | `messageTransforms[].aiInference.endpoint` | `status.outputs.endpoint_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `topic_id` | Fully qualified topic ID (`projects/{project}/topics/{name}`) | GcpPubSubSubscription `topic` and `deadLetterPolicy.deadLetterTopic`, Cloud Function triggers, Cloud Scheduler targets |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic topic** -- Minimal topic with Google-managed encryption and no retention or storage policy. Suitable for development and simple event routing. Start from the **Basic Topic** preset.

**Regional encrypted** -- CMEK-encrypted topic with regional storage constraints and in-transit enforcement. Suitable for regulated workloads with data residency requirements. Start from the **Regional Encrypted Topic** preset.

**Message retention** -- Topic with 7-day message retention enabling subscription-level seek and replay. Suitable for event sourcing and audit trail patterns where consumers may need to reprocess historical messages. Start from the **Topic with Message Retention** preset.

**Cloud Storage ingestion** -- Topic configured to ingest data from a GCS bucket with glob-based filtering and text format parsing. Suitable for batch-to-stream bridge patterns where files landing in a bucket should trigger downstream event processing. Start from the **Cloud Storage Ingestion Topic** preset.

**Schema-validated topic** -- Topic attached to a GcpPubSubSchema so malformed messages are rejected at publish time instead of surfacing as consumer failures. The shape for contracts shared across producing teams. Start from the **Schema-Validated Topic** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the topic is created
- [**GCP Pub/Sub Schema**](/infra-catalog/gcp-pub-sub-schema) -- provides the message contract enforced at publish time
- [**GCP Pub/Sub Subscription**](/infra-catalog/gcp-pub-sub-subscription) -- consumes this topic's stream (pull, push, BigQuery, or Cloud Storage delivery)
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- provides the CMEK encryption key for messages at rest
- [**GCP GCS Bucket**](/infra-catalog/gcp-gcs-bucket) -- provides the source bucket for Cloud Storage ingestion
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the federated identity ingestion pipelines authenticate with, and the caller for AI-inference transforms
- [**GCP Vertex AI Endpoint**](/infra-catalog/gcp-vertex-ai-endpoint) -- the model endpoint behind AI-inference message transforms