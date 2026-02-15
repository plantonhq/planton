# GcpPubSubTopic -- Examples

## Example 1: Minimal Topic

The simplest configuration: a topic with default settings. Suitable for
development, prototyping, or workloads without special requirements.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubTopic
metadata:
  name: dev-events
spec:
  projectId:
    value: "my-gcp-project"
  topicName: dev-events
```

**Notes:**
- Google-managed encryption (default)
- No message retention at the topic level -- subscribers control retention
- No regional constraints on message storage
- Messages delivered to all subscriptions attached to this topic

## Example 2: Topic with CMEK Encryption

For regulated workloads requiring customer-managed encryption keys. All messages
published to this topic are encrypted with the specified KMS key.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubTopic
metadata:
  name: secure-events
spec:
  projectId:
    value: "my-prod-project"
  topicName: secure-audit-events
  kmsKeyName:
    value: "projects/my-prod-project/locations/us-central1/keyRings/prod-ring/cryptoKeys/pubsub-cmek"
```

**Notes:**
- The KMS key must be in a region where messages are stored
- The Pub/Sub service account needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key
- Suitable for PCI, HIPAA, and SOX compliance scenarios

## Example 3: Topic with Message Retention

Retain messages at the topic level for 7 days, enabling replay and seek
operations across all subscriptions.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubTopic
metadata:
  name: event-stream
spec:
  projectId:
    value: "my-gcp-project"
  topicName: transaction-events
  messageRetentionDuration: "604800s"
```

**Notes:**
- `604800s` = 7 days (valid range: 600s to 2678400s)
- Enables `seek` operations on any subscription attached to this topic
- Topic-level retention is independent of subscription-level retention
- Increases storage costs proportional to message volume and retention window

## Example 4: Topic with Regional Storage Policy

Constrain message storage to specific regions with in-transit enforcement.
Required for data residency compliance (GDPR, data sovereignty).

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubTopic
metadata:
  name: eu-events
spec:
  projectId:
    value: "my-gcp-project"
  topicName: eu-customer-events
  messageStoragePolicy:
    allowedPersistenceRegions:
      - us-central1
    enforceInTransit: true
```

**Notes:**
- Messages are only stored in `us-central1`
- `enforceInTransit: true` rejects publish calls from clients outside `us-central1`
- Without `enforceInTransit`, messages from other regions are routed to an allowed region
- Combine with CMEK for full data sovereignty

## Example 5: Topic with Schema Validation

Enforce message contracts at publish time using a Pub/Sub schema. Messages
that do not conform to the schema are rejected.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubTopic
metadata:
  name: validated-events
spec:
  projectId:
    value: "my-gcp-project"
  topicName: order-events
  schemaSettings:
    schema: "projects/my-gcp-project/schemas/order-event-schema"
    encoding: JSON
```

**Notes:**
- The schema must be created separately (Pub/Sub Schema resource)
- `encoding: JSON` validates and expects JSON-encoded messages
- `encoding: BINARY` expects messages encoded in the schema's binary format (e.g., Avro)
- Schema validation adds publish-time latency but prevents malformed messages

## Example 6: Cross-Resource Reference (Infra Chart Pattern)

When composing resources in an infra chart, use `valueFrom` to reference
the project and KMS key from other OpenMCF resources.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubTopic
metadata:
  name: pipeline-events
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: data-project
      fieldPath: status.outputs.project_id
  topicName: pipeline-events
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: pubsub-cmek-key
      fieldPath: status.outputs.key_id
  messageRetentionDuration: "604800s"
```

**Notes:**
- `valueFrom` creates dependency edges -- project and KMS key are provisioned first
- This pattern enables fully declarative infrastructure composition
- The topic waits for both the project and KMS key to be ready before creation

## Example 7: Topic with Cloud Storage Ingestion

Ingest data from a GCS bucket into Pub/Sub messages. Each matching object's
content is published as one or more Pub/Sub messages based on the selected format.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubTopic
metadata:
  name: gcs-ingest-topic
spec:
  projectId:
    value: "my-gcp-project"
  topicName: gcs-data-ingest
  ingestionDataSourceSettings:
    cloudStorage:
      bucket:
        value: "my-data-landing-bucket"
      matchGlob: "**/*.json"
      textFormat:
        delimiter: "\n"
    platformLogsSettings:
      severity: INFO
```

**Notes:**
- `bucket` uses `StringValueOrRef` and supports `valueFrom` to reference a GcpGcsBucket
- `matchGlob: "**/*.json"` only ingests JSON files; omit to ingest all objects
- `textFormat` splits object content by newlines into individual Pub/Sub messages
- Alternative formats: `avroFormat` (binary Avro) or `pubsubAvroFormat` (re-import exported messages)
- `platformLogsSettings.severity: INFO` enables pipeline observability
