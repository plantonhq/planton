# GcpPubSubTopic

A GcpPubSubTopic provisions a Pub/Sub topic -- the named channel to which
publishers send messages and from which subscribers consume via subscriptions.
Topics are the foundation of event-driven and streaming architectures in GCP,
decoupling producers from consumers and supporting one-to-many message delivery.

## When to Use

Use GcpPubSubTopic when you need:

- **An event bus** for asynchronous communication between microservices
- **A streaming ingestion endpoint** for real-time data pipelines (logs, metrics, events)
- **Cross-cloud data ingestion** from AWS Kinesis, AWS MSK, Azure Event Hubs, or Confluent Cloud
- **Cloud Storage ingestion** to stream object data into Pub/Sub messages
- **Customer-managed encryption** (CMEK) for topics handling sensitive or regulated data
- **Regional message storage guarantees** with configurable in-transit enforcement
- **Schema validation** to enforce message contracts at publish time

## Prerequisites

- A GCP project with the Pub/Sub API enabled
- Appropriate IAM permissions (`roles/pubsub.admin` or `roles/pubsub.editor`)
- For CMEK: an existing KMS key with `roles/cloudkms.cryptoKeyEncrypterDecrypter` granted to the Pub/Sub service account
- For schema validation: a Pub/Sub schema resource already created in the project

## Quick Start

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubTopic
metadata:
  name: my-topic
spec:
  projectId:
    value: "my-gcp-project"
  topicName: order-events
```

This creates a Pub/Sub topic with Google-managed encryption and no message
retention -- subscribers control their own retention via subscriptions.

## Configuration Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | StringValueOrRef | Yes | GCP project ID |
| `topicName` | string | Yes | Topic name (3-255 chars, starts with letter, immutable) |
| `kmsKeyName` | StringValueOrRef | No | CMEK encryption key (full resource name) |
| `messageRetentionDuration` | string | No | Retain messages for this duration (600s-2678400s) |
| `messageStoragePolicy` | object | No | Regional storage constraints |
| `messageStoragePolicy.allowedPersistenceRegions` | list | Yes* | GCP regions for message storage (* required when policy is set) |
| `messageStoragePolicy.enforceInTransit` | bool | No | Reject publishes from non-allowed regions |
| `schemaSettings` | object | No | Schema validation settings |
| `schemaSettings.schema` | string | Yes* | Schema resource name (* required when schemaSettings is set) |
| `schemaSettings.encoding` | string | No | Message encoding: JSON or BINARY |
| `ingestionDataSourceSettings` | object | No | External data source ingestion |
| `ingestionDataSourceSettings.awsKinesis` | object | No | AWS Kinesis Data Streams ingestion |
| `ingestionDataSourceSettings.awsMsk` | object | No | AWS MSK (Kafka) ingestion |
| `ingestionDataSourceSettings.azureEventHubs` | object | No | Azure Event Hubs ingestion |
| `ingestionDataSourceSettings.cloudStorage` | object | No | GCS bucket ingestion |
| `ingestionDataSourceSettings.confluentCloud` | object | No | Confluent Cloud ingestion |
| `ingestionDataSourceSettings.platformLogsSettings` | object | No | Ingestion pipeline logging |

## Important Notes

**Topic name is immutable.** Changing `topicName` destroys and recreates the
topic along with all its subscriptions. Choose names carefully.

**CMEK requires IAM setup.** The Pub/Sub service account
(`service-{PROJECT_NUMBER}@gcp-sa-pubsub.iam.gserviceaccount.com`) must have
`roles/cloudkms.cryptoKeyEncrypterDecrypter` on the KMS key before the topic
is created.

**Message retention is topic-level.** When `messageRetentionDuration` is set,
messages are retained regardless of subscriber acknowledgement. This enables
replay via subscription seek operations.

**Ingestion sources are mutually exclusive per topic.** Configure at most one
ingestion data source (Kinesis, MSK, Event Hubs, Cloud Storage, or Confluent
Cloud) per topic.

## Related Components

- [GcpPubSubSubscription](../gcppubsubsubscription/v1/) -- Subscriptions that consume from this topic
- [GcpKmsKey](../gcpkmskey/v1/) -- CMEK encryption key for topic encryption
- [GcpProject](../gcpproject/v1/) -- Parent GCP project
- [GcpGcsBucket](../gcpgcsbucket/v1/) -- Source bucket for Cloud Storage ingestion
