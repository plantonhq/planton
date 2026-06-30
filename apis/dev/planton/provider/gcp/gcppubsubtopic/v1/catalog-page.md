# GCP Pub/Sub Topic

Deploys a Google Cloud Pub/Sub topic with optional CMEK encryption, message retention, regional storage policies, schema validation, and cross-cloud ingestion from AWS Kinesis, AWS MSK, Azure Event Hubs, Cloud Storage, or Confluent Cloud. The topic is labeled automatically from resource metadata.

## What Gets Created

When you deploy a GcpPubSubTopic resource, Planton provisions:

- **Pub/Sub Topic** — a `google_pubsub_topic` resource in the specified GCP project, with GCP labels derived from `metadata.org`, `metadata.env`, and `metadata.id`
- **CMEK Encryption** — configured only when `kmsKeyName` is provided, encrypts messages at rest using a customer-managed Cloud KMS key
- **Message Storage Policy** — applied only when `messageStoragePolicy` is provided, restricts message persistence to the listed GCP regions and optionally enforces in-transit guarantees
- **Schema Validation** — configured only when `schemaSettings` is provided, validates every published message against the referenced Pub/Sub schema
- **Ingestion Pipeline** — configured only when `ingestionDataSourceSettings` is provided, streams data from an external source (AWS Kinesis, AWS MSK, Azure Event Hubs, Cloud Storage, or Confluent Cloud) into the topic

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** with the Pub/Sub API enabled
- **A Cloud KMS key** if enabling CMEK encryption (the Pub/Sub service account needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key)
- **A Pub/Sub schema** if enabling schema validation
- **Cross-cloud IAM roles** if configuring ingestion from AWS or Azure sources

## Quick Start

Create a file `pubsub-topic.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubTopic
metadata:
  name: my-topic
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpPubSubTopic.my-topic
spec:
  projectId:
    value: my-gcp-project
  topicName: my-topic
```

Deploy:

```shell
planton apply -f pubsub-topic.yaml
```

This creates a Pub/Sub topic named `my-topic` in the specified GCP project with Google-managed encryption and no retention policy.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `StringValueOrRef` | GCP project where the topic will be created. Can reference a GcpProject resource via `valueFrom`. | Required |
| `topicName` | `string` | Name of the Pub/Sub topic. Immutable after creation. | 3–255 characters, must start with a letter, allows letters, numbers, hyphens, underscores, periods, tildes, `+`, `%` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `kmsKeyName` | `StringValueOrRef` | — | Cloud KMS key for encrypting messages at rest (CMEK). Format: `projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}`. Can reference a GcpKmsKey resource via `valueFrom`. |
| `messageRetentionDuration` | `string` | — | Duration to retain published messages on the topic. Format: duration string (e.g., `"604800s"` for 7 days). Range: `600s` to `2678400s`. When unset, retention is controlled by individual subscriptions. |
| `messageStoragePolicy.allowedPersistenceRegions` | `string[]` | — | GCP region IDs where messages may be stored. Messages from non-allowed regions are routed to an allowed region. Minimum 1 item when the policy is set. |
| `messageStoragePolicy.enforceInTransit` | `bool` | `false` | When `true`, publish calls from non-allowed regions are rejected and subscriptions in non-allowed regions fail. |
| `schemaSettings.schema` | `string` | — | Fully qualified Pub/Sub schema name. Format: `projects/{project}/schemas/{schema}`. Required when `schemaSettings` is set. |
| `schemaSettings.encoding` | `string` | — | Message encoding validated against the schema. Valid values: `"JSON"` or `"BINARY"`. |
| `ingestionDataSourceSettings.awsKinesis` | `object` | — | Ingest from Amazon Kinesis Data Streams. Requires `streamArn`, `consumerArn`, `awsRoleArn`, and `gcpServiceAccount`. |
| `ingestionDataSourceSettings.awsMsk` | `object` | — | Ingest from Amazon MSK. Requires `clusterArn`, `topic`, `awsRoleArn`, and `gcpServiceAccount`. |
| `ingestionDataSourceSettings.azureEventHubs` | `object` | — | Ingest from Azure Event Hubs. Fields: `resourceGroup`, `namespace`, `eventHub`, `clientId`, `tenantId`, `subscriptionId`, `gcpServiceAccount`. |
| `ingestionDataSourceSettings.cloudStorage` | `object` | — | Ingest from a GCS bucket. Requires `bucket` (StringValueOrRef, can reference GcpGcsBucket). Optional: `matchGlob`, `minimumObjectCreateTime`, and one of `textFormat`, `avroFormat`, or `pubsubAvroFormat`. |
| `ingestionDataSourceSettings.confluentCloud` | `object` | — | Ingest from Confluent Cloud. Requires `bootstrapServer`, `topic`, `identityPoolId`, and `gcpServiceAccount`. Optional: `clusterId`. |
| `ingestionDataSourceSettings.platformLogsSettings.severity` | `string` | — | Minimum severity for ingestion platform logs. Valid values: `"DISABLED"`, `"DEBUG"`, `"INFO"`, `"WARNING"`, `"ERROR"`. |

## Examples

### Topic with Message Retention

Retain messages for 7 days so any subscription can seek back within the retention window:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubTopic
metadata:
  name: orders-topic
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpPubSubTopic.orders-topic
spec:
  projectId:
    value: my-gcp-project
  topicName: orders
  messageRetentionDuration: "604800s"
  messageStoragePolicy:
    allowedPersistenceRegions:
      - us-central1
      - us-east1
```

### CMEK-Encrypted Topic with Schema Validation

Encrypt messages with a customer-managed key and enforce JSON schema validation:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubTopic
metadata:
  name: events-topic
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpPubSubTopic.events-topic
spec:
  projectId:
    value: my-gcp-project
  topicName: events
  kmsKeyName:
    value: projects/my-gcp-project/locations/us-central1/keyRings/my-ring/cryptoKeys/my-key
  schemaSettings:
    schema: projects/my-gcp-project/schemas/event-schema
    encoding: JSON
```

### Topic with Cloud Storage Ingestion

Ingest objects from a GCS bucket in text format:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubTopic
metadata:
  name: logs-ingest-topic
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpPubSubTopic.logs-ingest-topic
spec:
  projectId:
    value: my-gcp-project
  topicName: logs-ingest
  ingestionDataSourceSettings:
    cloudStorage:
      bucket:
        value: my-logs-bucket
      matchGlob: "**/*.log"
      textFormat:
        delimiter: "\n"
    platformLogsSettings:
      severity: WARNING
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding IDs:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubTopic
metadata:
  name: ref-topic
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpPubSubTopic.ref-topic
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  topicName: ref-topic
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: my-key
      field: status.outputs.key_id
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `topic_id` | `string` | Fully qualified topic ID. Format: `projects/{project}/topics/{name}` |
| `topic_name` | `string` | Short topic name (same as the `topicName` input) |

## Related Components

- [GcpKmsKey](/docs/catalog/gcp/gcpkmskey) — provides a customer-managed encryption key for CMEK
- [GcpGcsBucket](/docs/catalog/gcp/gcpgcsbucket) — source bucket for Cloud Storage ingestion
- [GcpPubSubSubscription](/docs/catalog/gcp/gcppubsubsubscription) — creates a subscription attached to this topic
