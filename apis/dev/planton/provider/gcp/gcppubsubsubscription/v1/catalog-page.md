# GCP Pub/Sub Subscription

Deploys a GCP Pub/Sub subscription attached to a topic, with support for four delivery methods: pull (default), push to an HTTPS endpoint, streaming to a BigQuery table, or batched writes to Cloud Storage. Only one delivery method can be active per subscription.

## What Gets Created

When you deploy a GcpPubSubSubscription resource, Planton provisions:

- **Pub/Sub Subscription** — a `google_pubsub_subscription` resource bound to the specified topic, configured with the chosen delivery method, acknowledgement deadline, retention policy, and retry/dead-letter settings
- **GCP Labels** — automatically applied labels containing the organization, environment, resource ID, resource name, and resource kind for consistent resource tracking

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **A GCP project** where the subscription will be created
- **An existing Pub/Sub topic** that the subscription will receive messages from
- **An HTTPS endpoint** if using push delivery
- **A BigQuery table** if using BigQuery delivery (the Pub/Sub service account must have `bigquery.dataEditor` on the dataset)
- **A Cloud Storage bucket** if using Cloud Storage delivery (the Pub/Sub service account must have `storage.objectCreator` on the bucket)
- **A dead-letter topic** if using dead-letter policy (the Pub/Sub service account must have Publisher permissions on the dead-letter topic and Subscriber permissions on this subscription)

## Quick Start

Create a file `pubsub-subscription.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubSubscription
metadata:
  name: my-subscription
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpPubSubSubscription.my-subscription
spec:
  projectId:
    value: my-gcp-project
  subscriptionName: my-subscription
  topic:
    value: projects/my-gcp-project/topics/my-topic
```

Deploy:

```shell
planton apply -f pubsub-subscription.yaml
```

This creates a pull subscription with default settings: 10-second ack deadline, 7-day message retention, and 31-day expiration for inactivity.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `projectId` | `string` | GCP project where the subscription will be created. Can reference GcpProject via `valueFrom`. | Required |
| `subscriptionName` | `string` | Name of the Pub/Sub subscription. Immutable after creation. | 3–255 characters, starts with a letter, alphanumeric plus `-_. ~+%` |
| `topic` | `string` | Topic from which this subscription receives messages. Immutable after creation. Can reference GcpPubSubTopic via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ackDeadlineSeconds` | `int` | `10` | Maximum time in seconds a subscriber has to acknowledge a message before redelivery. Range: 10–600. |
| `messageRetentionDuration` | `string` | `"604800s"` | How long unacknowledged messages are retained. Also controls seek range when `retainAckedMessages` is true. Range: `"600s"` to `"2678400s"`. |
| `retainAckedMessages` | `bool` | `false` | When true, acknowledged messages are retained in the backlog for the `messageRetentionDuration` window, enabling replay via seek. |
| `expirationPolicy.ttl` | `string` | `"2678400s"` | Duration of inactivity before the subscription is automatically deleted. Minimum: `"86400s"`. Set to `""` for a subscription that never expires. |
| `filter` | `string` | — | Attribute filter expression. Non-matching messages are auto-acknowledged. Max 256 bytes. Immutable after creation. |
| `enableMessageOrdering` | `bool` | `false` | Delivers messages with the same ordering key in publish order. Immutable after creation. |
| `enableExactlyOnceDelivery` | `bool` | `false` | Guarantees each message is not resent before its ack deadline expires. Does not prevent duplicates from the publisher. |
| `deadLetterPolicy.deadLetterTopic` | `string` | — | Topic to forward undeliverable messages to. Can reference GcpPubSubTopic via `valueFrom`. |
| `deadLetterPolicy.maxDeliveryAttempts` | `int` | `5` | Number of delivery attempts before dead-lettering. Range: 5–100. |
| `retryPolicy.minimumBackoff` | `string` | `"10s"` | Minimum delay between delivery retries after a NACK. Range: `"0s"` to `"600s"`. |
| `retryPolicy.maximumBackoff` | `string` | `"600s"` | Maximum delay between delivery retries after a NACK. Range: `"0s"` to `"600s"`. |
| `pushConfig.pushEndpoint` | `string` | — | HTTPS URL to which Pub/Sub pushes messages. Required when using push delivery. |
| `pushConfig.attributes` | `map<string,string>` | — | Endpoint configuration attributes. Supports `x-goog-version` (`"v1beta1"` or `"v1"`). |
| `pushConfig.oidcToken.serviceAccountEmail` | `string` | — | Service account used to generate OIDC tokens for authenticated push requests. |
| `pushConfig.oidcToken.audience` | `string` | push endpoint URL | Audience claim for the OIDC token. |
| `pushConfig.noWrapper.writeMetadata` | `bool` | `false` | When true, sends the raw message body without the Pub/Sub envelope and writes metadata as HTTP headers. |
| `bigqueryConfig.table` | `string` | — | BigQuery table to write messages to. Format: `{project}.{dataset}.{table}`. Required when using BigQuery delivery. |
| `bigqueryConfig.useTopicSchema` | `bool` | `false` | Maps message fields to BigQuery columns using the topic schema. Mutually exclusive with `useTableSchema`. |
| `bigqueryConfig.useTableSchema` | `bool` | `false` | Maps message fields to BigQuery columns using the table schema. Mutually exclusive with `useTopicSchema`. |
| `bigqueryConfig.dropUnknownFields` | `bool` | `false` | Silently drops message fields not present in the BigQuery table schema. Requires `useTopicSchema` or `useTableSchema`. |
| `bigqueryConfig.writeMetadata` | `bool` | `false` | Writes subscription name, messageId, publishTime, attributes, and orderingKey to additional BigQuery columns. |
| `bigqueryConfig.serviceAccountEmail` | `string` | Pub/Sub service agent | Service account for BigQuery writes. |
| `cloudStorageConfig.bucket` | `string` | — | Cloud Storage bucket name (without `gs://`). Required when using Cloud Storage delivery. Can reference GcpGcsBucket via `valueFrom`. |
| `cloudStorageConfig.filenamePrefix` | `string` | — | Prefix for Cloud Storage object names. |
| `cloudStorageConfig.filenameSuffix` | `string` | — | Suffix for Cloud Storage object names. Must not end in `/`. |
| `cloudStorageConfig.filenameDatetimeFormat` | `string` | — | Datetime format string for Cloud Storage object names. |
| `cloudStorageConfig.maxBytes` | `int` | — | Maximum bytes per object before a new object is created. Range: 1024–10737418240. |
| `cloudStorageConfig.maxDuration` | `string` | `"300s"` | Maximum duration before a new object is created. Range: `"60s"` to `"600s"`. Must not exceed `ackDeadlineSeconds`. |
| `cloudStorageConfig.maxMessages` | `int` | — | Maximum messages per object. Minimum: 1000. |
| `cloudStorageConfig.avroConfig.useTopicSchema` | `bool` | `false` | Serializes output in Avro format using the topic schema. |
| `cloudStorageConfig.avroConfig.writeMetadata` | `bool` | `false` | Includes subscription metadata as additional Avro fields. |
| `cloudStorageConfig.serviceAccountEmail` | `string` | Pub/Sub service agent | Service account for Cloud Storage writes. |

## Examples

### Push Subscription with OIDC Authentication

Delivers messages to an HTTPS endpoint with OIDC-based authentication:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubSubscription
metadata:
  name: push-subscription
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.GcpPubSubSubscription.push-subscription
spec:
  projectId:
    value: my-gcp-project
  subscriptionName: push-subscription
  topic:
    value: projects/my-gcp-project/topics/events-topic
  ackDeadlineSeconds: 30
  pushConfig:
    pushEndpoint: https://my-service.example.com/pubsub/push
    oidcToken:
      serviceAccountEmail: push-invoker@my-gcp-project.iam.gserviceaccount.com
      audience: https://my-service.example.com
```

### BigQuery Delivery with Dead-Letter Policy

Streams messages directly to a BigQuery table, forwarding failures to a dead-letter topic after 10 attempts:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubSubscription
metadata:
  name: bq-subscription
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpPubSubSubscription.bq-subscription
spec:
  projectId:
    value: my-gcp-project
  subscriptionName: bq-subscription
  topic:
    value: projects/my-gcp-project/topics/analytics-topic
  enableExactlyOnceDelivery: true
  bigqueryConfig:
    table: my-gcp-project.analytics_dataset.events
    useTopicSchema: true
    dropUnknownFields: true
    writeMetadata: true
  deadLetterPolicy:
    deadLetterTopic:
      value: projects/my-gcp-project/topics/analytics-dlq
    maxDeliveryAttempts: 10
  retryPolicy:
    minimumBackoff: "15s"
    maximumBackoff: "300s"
```

### Cloud Storage Delivery with Avro Format

Batches messages into Avro-formatted objects in Cloud Storage:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubSubscription
metadata:
  name: gcs-subscription
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpPubSubSubscription.gcs-subscription
spec:
  projectId:
    value: my-gcp-project
  subscriptionName: gcs-subscription
  topic:
    value: projects/my-gcp-project/topics/logs-topic
  ackDeadlineSeconds: 600
  messageRetentionDuration: "2678400s"
  cloudStorageConfig:
    bucket:
      value: my-logs-archive-bucket
    filenamePrefix: pubsub/logs/
    filenameSuffix: .avro
    maxBytes: 1073741824
    maxDuration: "600s"
    maxMessages: 10000
    avroConfig:
      useTopicSchema: true
      writeMetadata: true
```

### Using Foreign Key References

Reference other Planton-managed resources instead of hardcoding values:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpPubSubSubscription
metadata:
  name: ref-subscription
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.GcpPubSubSubscription.ref-subscription
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      field: status.outputs.project_id
  subscriptionName: ref-subscription
  topic:
    valueFrom:
      kind: GcpPubSubTopic
      name: my-topic
      field: status.outputs.topic_id
  retainAckedMessages: true
  enableMessageOrdering: true
  expirationPolicy:
    ttl: ""
  deadLetterPolicy:
    deadLetterTopic:
      valueFrom:
        kind: GcpPubSubTopic
        name: my-dlq-topic
        field: status.outputs.topic_id
    maxDeliveryAttempts: 20
  cloudStorageConfig:
    bucket:
      valueFrom:
        kind: GcpGcsBucket
        name: my-archive-bucket
        field: status.outputs.bucket_id
    filenamePrefix: events/
    maxDuration: "300s"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `subscription_id` | `string` | Fully qualified subscription ID (e.g., `projects/my-project/subscriptions/my-subscription`) |
| `subscription_name` | `string` | Short subscription name, same as the `subscriptionName` input |

## Related Components

- [GcpPubSubTopic](/docs/catalog/gcp/gcppubsubtopic) — provides the topic that the subscription receives messages from
- [GcpGcsBucket](/docs/catalog/gcp/gcpgcsbucket) — provides the Cloud Storage bucket for Cloud Storage delivery
- [GcpProject](/docs/catalog/gcp/gcpproject) — provides the GCP project where the subscription is created
