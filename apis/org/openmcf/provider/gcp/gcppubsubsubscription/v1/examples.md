# GcpPubSubSubscription Examples

## Basic Pull Subscription

The simplest subscription: pull delivery with all defaults.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubSubscription
metadata:
  name: my-pull-sub
spec:
  projectId:
    value: "my-gcp-project"
  subscriptionName: order-events-sub
  topic:
    value: "projects/my-gcp-project/topics/order-events"
```

## Push Subscription with OIDC Authentication

Push messages to an HTTPS endpoint with OIDC-based authentication.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubSubscription
metadata:
  name: my-push-sub
spec:
  projectId:
    value: "my-gcp-project"
  subscriptionName: webhook-delivery-sub
  topic:
    value: "projects/my-gcp-project/topics/user-events"
  ackDeadlineSeconds: 30
  pushConfig:
    pushEndpoint: "https://my-app.run.app/webhook"
    oidcToken:
      serviceAccountEmail: "push-invoker@my-gcp-project.iam.gserviceaccount.com"
      audience: "https://my-app.run.app"
```

## Push Subscription with Unwrapped Payload

Send raw message data to a webhook without the Pub/Sub envelope.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubSubscription
metadata:
  name: my-raw-push-sub
spec:
  projectId:
    value: "my-gcp-project"
  subscriptionName: raw-webhook-sub
  topic:
    value: "projects/my-gcp-project/topics/notifications"
  pushConfig:
    pushEndpoint: "https://third-party.example.com/ingest"
    noWrapper:
      writeMetadata: true
```

## BigQuery Delivery with Topic Schema

Stream messages directly into a BigQuery table using the topic's schema.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubSubscription
metadata:
  name: my-bq-sub
spec:
  projectId:
    value: "my-gcp-project"
  subscriptionName: analytics-bq-sub
  topic:
    value: "projects/my-gcp-project/topics/click-events"
  bigqueryConfig:
    table: "my-gcp-project.analytics_dataset.click_events"
    useTopicSchema: true
    dropUnknownFields: true
    writeMetadata: true
```

## Cloud Storage Archival

Archive messages to Cloud Storage with Avro format for data lake ingestion.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubSubscription
metadata:
  name: my-gcs-sub
spec:
  projectId:
    value: "my-gcp-project"
  subscriptionName: archive-gcs-sub
  topic:
    value: "projects/my-gcp-project/topics/audit-logs"
  cloudStorageConfig:
    bucket:
      value: "my-audit-archive-bucket"
    filenamePrefix: "pubsub/audit-logs/"
    filenameSuffix: ".avro"
    maxBytes: 104857600  # 100 MB
    maxDuration: "600s"  # 10 minutes
    avroConfig:
      useTopicSchema: true
      writeMetadata: true
```

## Dead Letter and Retry Configuration

Pull subscription with dead-letter handling and custom retry backoff.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubSubscription
metadata:
  name: my-reliable-sub
spec:
  projectId:
    value: "my-gcp-project"
  subscriptionName: payment-processing-sub
  topic:
    value: "projects/my-gcp-project/topics/payment-events"
  ackDeadlineSeconds: 60
  messageRetentionDuration: "1209600s"  # 14 days
  retainAckedMessages: true
  enableExactlyOnceDelivery: true
  deadLetterPolicy:
    deadLetterTopic:
      value: "projects/my-gcp-project/topics/payment-events-dlq"
    maxDeliveryAttempts: 10
  retryPolicy:
    minimumBackoff: "30s"
    maximumBackoff: "300s"
```

## Filtered Subscription with Message Ordering

Only receive high-priority messages, delivered in publish order.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubSubscription
metadata:
  name: my-filtered-sub
spec:
  projectId:
    value: "my-gcp-project"
  subscriptionName: high-priority-sub
  topic:
    value: "projects/my-gcp-project/topics/task-events"
  filter: 'attributes.priority = "high"'
  enableMessageOrdering: true
  expirationPolicy:
    ttl: ""  # Never expires
```

## Infra Chart Composition (valueFrom References)

Wire a subscription to a topic and dead-letter topic from the same infra chart.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpPubSubSubscription
metadata:
  name: events-sub
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: "{{ values.env }}-project"
      fieldPath: status.outputs.project_id
  subscriptionName: "{{ values.env }}-events-sub"
  topic:
    valueFrom:
      kind: GcpPubSubTopic
      name: "{{ values.env }}-events-topic"
      fieldPath: status.outputs.topic_id
  deadLetterPolicy:
    deadLetterTopic:
      valueFrom:
        kind: GcpPubSubTopic
        name: "{{ values.env }}-events-dlq"
        fieldPath: status.outputs.topic_id
    maxDeliveryAttempts: 15
```
