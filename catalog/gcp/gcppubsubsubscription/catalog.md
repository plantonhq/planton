# GCP Pub/Sub Subscription

Deploys a Pub/Sub subscription attached to a topic with configurable delivery method (pull, push, BigQuery, or Cloud Storage), acknowledgement deadlines, message ordering, exactly-once delivery, dead-letter handling, and retry policies. Per-consumer message transforms (JavaScript UDFs or Vertex AI inference) reshape payloads for this subscription without changing what other subscriptions on the topic see, and attribute filters auto-acknowledge non-matching messages before they reach the consumer.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Pub/Sub API enablement** -- the module enables `pubsub.googleapis.com` in the target project before creating the subscription (never disabled on destroy)
- **Pub/Sub Subscription** -- a named subscription resource in the specified GCP project, attached to the referenced topic, with the configured delivery method and message handling policies
- **Push Configuration** -- created only when `pushConfig` is set; configures HTTP POST delivery to an HTTPS endpoint with optional OIDC authentication and unwrapped payload mode
- **BigQuery Configuration** -- created only when `bigqueryConfig` is set; streams messages directly into a BigQuery table with optional schema mapping and metadata columns
- **Cloud Storage Configuration** -- created only when `cloudStorageConfig` is set; batches messages into Cloud Storage objects based on size, duration, or message count thresholds
- **Dead Letter Policy** -- created only when `deadLetterPolicy` is set; forwards messages that exceed the maximum delivery attempts to a dead-letter topic
- **Retry Policy** -- created only when `retryPolicy` is set; configures exponential backoff between delivery attempts after a NACK or ack deadline exceeded event
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the subscription will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **An existing Pub/Sub topic** to subscribe to. Provide the fully qualified topic ID or reference a GcpPubSubTopic Infra Component via ValueFromRef. The module enables the Pub/Sub API itself — no manual API setup is needed.
- **BigQuery Data Editor role** on the target table for the writing identity (only for BigQuery delivery) — the Pub/Sub service agent by default, or the `serviceAccountEmail` you set.
- **Storage Object Creator role** on the target bucket for the writing identity (only for Cloud Storage delivery).
- **Dead-letter grants** (only for `deadLetterPolicy`) — the Pub/Sub service agent needs Subscriber on this subscription and Publisher on the dead-letter topic.

## Deploy

### Console

Open the deployment store, find **GCP Pub/Sub Subscription**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Pull Subscription** preset in the [Presets](#presets) tab to pre-populate a minimal pull subscription.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubSubscription
metadata:
  name: order-processor
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  subscriptionName: order-processor
  topic:
    value: "projects/acme-prod-12345/topics/order-events"
```

```shell
planton apply -f pubsub-subscription.yaml
```

This creates a pull subscription with a 10-second ack deadline, 7-day message retention, no dead-letter handling, and no message ordering. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the subscription to a project and topic deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  topic:
    valueFrom:
      kind: GcpPubSubTopic
      name: order-events
      fieldPath: status.outputs.topic_id
```

The InfraPipeline resolves the dependency graph, deploys the project and topic first, then provisions the subscription with the resolved topic reference.

## Key Configuration

These are the most important decisions when configuring a Pub/Sub subscription. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Delivery method** -- Choose between pull (default, no config needed), push (`pushConfig`), BigQuery (`bigqueryConfig`), or Cloud Storage (`cloudStorageConfig`). Only one delivery method can be active. Pull requires application-side polling; push delivers to an HTTPS endpoint; BigQuery and Cloud Storage write directly without consumer code.

**Acknowledgement deadline** -- Set `ackDeadlineSeconds` (10-600, default 10) to control how long Pub/Sub waits for an acknowledgement before redelivering a message. Set higher for long-running processing tasks. Too low causes unnecessary redeliveries; too high delays retry on actual failures.

**Dead-letter handling** -- Configure `deadLetterPolicy` with a dead-letter topic and `maxDeliveryAttempts` (5-100) to route unprocessable messages instead of retrying indefinitely. The Pub/Sub service account needs Subscriber permissions on this subscription and Publisher permissions on the dead-letter topic.

**Exactly-once delivery** -- Set `enableExactlyOnceDelivery: true` to guarantee each message is delivered exactly once within the ack deadline window. Adds latency to acknowledgement processing. Does not prevent duplicates from publishers sending the same message multiple times.

**Message ordering** -- Set `enableMessageOrdering: true` to deliver messages with the same ordering key in publish order. Immutable after creation. Requires publishers to set an ordering key on messages.

**Idle expiry is on by default** -- With no `expirationPolicy`, GCP deletes the subscription after 31 days without an active consumer — a staging queue nobody polls quietly vanishes along with its backlog. Set `expirationPolicy.ttl: ""` for a subscription that never expires.

**Per-consumer transforms** -- `messageTransforms` run in list order on every message before delivery: a JavaScript UDF or a Vertex AI inference call per step, and a step returning null drops the message. Use them to reshape or redact payloads for THIS consumer; a transform the topic applies changes what every subscription sees. The `disabled` flag stages a transform in or out without losing its pipeline position.

**Destroy semantics** -- Deleting a subscription drops its unacknowledged backlog immediately. For consumers whose backlog must never silently disappear, set `deletionPolicy: PREVENT` so destroy fails instead; `ABANDON` unmanages the subscription but leaves it accumulating messages in GCP.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpPubSubTopic** | `topic` | `status.outputs.topic_id` |
| **GcpPubSubTopic** (optional) | `deadLetterPolicy.deadLetterTopic` | `status.outputs.topic_id` |
| **GcpCloudRun** (optional) | `pushConfig.pushEndpoint` | `status.outputs.url` |
| **GcpServiceAccount** (optional) | `pushConfig.oidcToken.serviceAccountEmail` | `status.outputs.email` |
| **GcpBigQueryTable** (optional) | `bigqueryConfig.table` | `status.outputs.qualified_name` |
| **GcpServiceAccount** (optional) | `bigqueryConfig.serviceAccountEmail` / `cloudStorageConfig.serviceAccountEmail` | `status.outputs.email` |
| **GcpGcsBucket** (optional) | `cloudStorageConfig.bucket` | `status.outputs.bucket_id` |
| **GcpVertexAiEndpoint** (optional, per transform) | `messageTransforms[].aiInference.endpoint` | `status.outputs.endpoint_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `subscription_id` | Fully qualified subscription ID (`projects/{project}/subscriptions/{name}`) | The handle pull consumers and monitoring filters address the subscription by |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic pull** -- Minimal pull subscription with default ack deadline and retention. Suitable for application-driven message consumption where the subscriber controls the pull rate. Start from the **Basic Pull Subscription** preset.

**Push with OIDC** -- Push delivery to an HTTPS endpoint with OIDC-based authentication. Suitable for webhook integrations, Cloud Run services, and serverless backends that receive messages as HTTP POST requests. Start from the **Push with OIDC Authentication** preset.

**BigQuery delivery** -- Direct streaming into a BigQuery table with topic schema mapping and metadata columns. Suitable for analytics pipelines, event archival, and real-time dashboards without writing consumer code. Start from the **BigQuery Delivery** preset.

**Dead letter** -- Subscription with exactly-once delivery, 14-day retention, retained ack'd messages, dead-letter routing after 10 attempts, and exponential retry backoff. Suitable for payment processing, order fulfillment, and other workflows that require reliable delivery with poison message isolation. Start from the **Reliable Processing with Dead-Lettering** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the subscription is created
- [**GCP Pub/Sub Topic**](/infra-catalog/gcp-pub-sub-topic) -- provides the topic that the subscription receives messages from, and optionally the dead-letter topic
- [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) -- the canonical push target; its `url` output feeds `pushConfig.pushEndpoint`
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the OIDC identity for authenticated push, or the writing identity for BigQuery/Cloud Storage delivery
- [**GCP BigQuery Table**](/infra-catalog/gcp-big-query-table) -- the destination table for BigQuery delivery
- [**GCP GCS Bucket**](/infra-catalog/gcp-gcs-bucket) -- provides the destination bucket for Cloud Storage delivery
- [**GCP Vertex AI Endpoint**](/infra-catalog/gcp-vertex-ai-endpoint) -- the model endpoint behind AI-inference message transforms