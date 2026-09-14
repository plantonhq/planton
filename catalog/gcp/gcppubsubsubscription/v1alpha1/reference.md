# GcpPubSubSubscription

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpPubSubSubscriptionSpec defines the configuration for a GCP Pub/Sub subscription.

A subscription is a named resource representing the stream of messages from a
single, specific topic, to be delivered to the subscribing application. Pub/Sub
supports four delivery methods:

  - **Pull** (default): The subscriber pulls messages using the API.
  - **Push**: Pub/Sub sends messages as HTTP POST requests to an endpoint.
  - **BigQuery**: Pub/Sub writes messages directly to a BigQuery table.
  - **Cloud Storage**: Pub/Sub writes messages to Cloud Storage objects.

Only one delivery method can be active. If none of push_config, bigquery_config,
or cloud_storage_config is set, the subscription defaults to pull delivery.

Important behavioral notes:

  - The subscription_name and filter fields are immutable after creation.
    Changing either requires destroying and recreating the subscription.

  - enable_message_ordering is immutable after creation. Once enabled,
    messages with the same ordering key are delivered in publish order.

  - enable_exactly_once_delivery guarantees each message is delivered exactly
    once within the acknowledgement deadline window. Does not prevent duplicate
    delivery if the publisher sends the same message multiple times.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubSubscription
metadata:
  name: test-subscription
spec:
  projectId:
    value: "test-project"
  subscriptionName: test-subscription
  topic:
    value: "projects/test-project/topics/test-topic"
  # Explicit destroy behavior: DELETE removes the subscription (and its
  # unacknowledged backlog) with the stack.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.subscriptionName` | `string` | yes |  |  |
| `spec.topic` | `string \| valueFrom` | yes |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.ackDeadlineSeconds` | `int32` |  |  |  |
| `spec.messageRetentionDuration` | `string` |  |  |  |
| `spec.retainAckedMessages` | `bool` |  |  |  |
| `spec.expirationPolicy` | `GcpPubSubSubscriptionExpirationPolicy` |  |  |  |
| `spec.expirationPolicy.ttl` | `string` |  |  |  |
| `spec.filter` | `string` |  |  |  |
| `spec.enableMessageOrdering` | `bool` |  |  |  |
| `spec.enableExactlyOnceDelivery` | `bool` |  |  |  |
| `spec.deadLetterPolicy` | `GcpPubSubSubscriptionDeadLetterPolicy` |  |  |  |
| `spec.deadLetterPolicy.deadLetterTopic` | `string \| valueFrom` |  |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.deadLetterPolicy.maxDeliveryAttempts` | `int32` |  |  |  |
| `spec.retryPolicy` | `GcpPubSubSubscriptionRetryPolicy` |  |  |  |
| `spec.retryPolicy.minimumBackoff` | `string` |  |  |  |
| `spec.retryPolicy.maximumBackoff` | `string` |  |  |  |
| `spec.pushConfig` | `GcpPubSubSubscriptionPushConfig` |  |  |  |
| `spec.pushConfig.pushEndpoint` | `string \| valueFrom` | yes |  | GcpCloudRun (`status.outputs.url`) |
| `spec.pushConfig.attributes` | `map<string, string>` |  |  |  |
| `spec.pushConfig.oidcToken` | `GcpPubSubSubscriptionPushConfigOidcToken` |  |  |  |
| `spec.pushConfig.oidcToken.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.pushConfig.oidcToken.audience` | `string` |  |  |  |
| `spec.pushConfig.noWrapper` | `GcpPubSubSubscriptionPushConfigNoWrapper` |  |  |  |
| `spec.pushConfig.noWrapper.writeMetadata` | `bool` |  |  |  |
| `spec.pushConfig.pubsubWrapper` | `GcpPubSubSubscriptionPushConfigPubsubWrapper` |  |  |  |
| `spec.bigqueryConfig` | `GcpPubSubSubscriptionBigQueryConfig` |  |  |  |
| `spec.bigqueryConfig.table` | `string \| valueFrom` | yes |  | GcpBigQueryTable (`status.outputs.qualified_name`) |
| `spec.bigqueryConfig.useTopicSchema` | `bool` |  |  |  |
| `spec.bigqueryConfig.useTableSchema` | `bool` |  |  |  |
| `spec.bigqueryConfig.dropUnknownFields` | `bool` |  |  |  |
| `spec.bigqueryConfig.writeMetadata` | `bool` |  |  |  |
| `spec.bigqueryConfig.serviceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.cloudStorageConfig` | `GcpPubSubSubscriptionCloudStorageConfig` |  |  |  |
| `spec.cloudStorageConfig.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.cloudStorageConfig.filenamePrefix` | `string` |  |  |  |
| `spec.cloudStorageConfig.filenameSuffix` | `string` |  |  |  |
| `spec.cloudStorageConfig.filenameDatetimeFormat` | `string` |  |  |  |
| `spec.cloudStorageConfig.maxBytes` | `int64` |  |  |  |
| `spec.cloudStorageConfig.maxDuration` | `string` |  |  |  |
| `spec.cloudStorageConfig.maxMessages` | `int64` |  |  |  |
| `spec.cloudStorageConfig.avroConfig` | `GcpPubSubSubscriptionCloudStorageConfigAvroConfig` |  |  |  |
| `spec.cloudStorageConfig.avroConfig.useTopicSchema` | `bool` |  |  |  |
| `spec.cloudStorageConfig.avroConfig.writeMetadata` | `bool` |  |  |  |
| `spec.cloudStorageConfig.textConfig` | `GcpPubSubSubscriptionCloudStorageConfigTextConfig` |  |  |  |
| `spec.cloudStorageConfig.serviceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.messageTransforms` | `[]GcpPubSubSubscriptionMessageTransform` |  |  |  |
| `spec.messageTransforms[].javascriptUdf` | `GcpPubSubSubscriptionMessageTransformJavascriptUdf` |  |  |  |
| `spec.messageTransforms[].javascriptUdf.functionName` | `string` | yes |  |  |
| `spec.messageTransforms[].javascriptUdf.code` | `string` | yes |  |  |
| `spec.messageTransforms[].disabled` | `bool` |  |  |  |
| `spec.messageTransforms[].aiInference` | `GcpPubSubSubscriptionMessageTransformAiInference` |  |  |  |
| `spec.messageTransforms[].aiInference.endpoint` | `string \| valueFrom` | yes |  | GcpVertexAiEndpoint (`status.outputs.endpoint_id`) |
| `spec.messageTransforms[].aiInference.serviceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.messageTransforms[].aiInference.unstructuredInference` | `GcpPubSubSubscriptionMessageTransformAiInferenceUnstructuredInference` |  |  |  |
| `spec.messageTransforms[].aiInference.unstructuredInference.parameters` | `map<string, string>` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the subscription will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.subscriptionName

`string` · required

Name of the Pub/Sub subscription.
Must be 3-255 characters, start with a letter, and contain only letters,
numbers, hyphens, underscores, periods, tildes, plus signs, and percent
signs. Names beginning with "goog" are reserved by Google and rejected
at create time. Immutable after creation.

- rule: subscription names beginning with 'goog' are reserved by Google — choose a different name
- rule: {"required":true,"string":{"minLen":"3","maxLen":"255","pattern":"^[a-zA-Z][a-zA-Z0-9\\-_\\.~+%]*$"}}

### spec.topic

`string | valueFrom` · required

The topic from which this subscription receives messages.
Format: projects/{project}/topics/{name} or just the topic name if
the topic is in the same project as the subscription.
Immutable after creation.

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.ackDeadlineSeconds

`int32`

Maximum time (in seconds) after a subscriber receives a message before the
subscriber should acknowledge the message. After the deadline expires,
the message is redelivered.
Range: 10 to 600 seconds. Defaults to 10 seconds.

- rule: ack_deadline_seconds must be between 10 and 600, or 0 for default (10)

### spec.messageRetentionDuration

`string`

How long to retain unacknowledged messages in the subscription's backlog.
If retain_acked_messages is true, this also controls retention of acknowledged
messages and determines how far back a seek operation can go.
Format: duration string (e.g., "604800s" for 7 days).
Range: 600s (10 minutes) to 2678400s (31 days). Default: 604800s (7 days).

### spec.retainAckedMessages

`bool`

When true, acknowledged messages are retained in the backlog until they
fall out of the message_retention_duration window. Enables replay via seek.

### spec.expirationPolicy

`GcpPubSubSubscriptionExpirationPolicy`

Expiration policy for the subscription. Controls automatic deletion of
inactive subscriptions. If not set, GCP defaults to 31 days TTL.

### spec.expirationPolicy.ttl

`string`

Duration after which the subscription expires if inactive.
Format: duration string (e.g., "2592000s" for 30 days).
Minimum: 86400s (1 day). Set to "" for a subscription that never expires.

### spec.filter

`string`

Message attribute filter expression. Only messages matching the filter are
delivered; non-matching messages are automatically acknowledged.
Maximum length: 256 bytes. Immutable after creation.

- rule: {"string":{"maxLen":"256"}}

### spec.enableMessageOrdering

`bool`

When true, messages with the same ordering key are delivered to subscribers
in the order they were published. Immutable after creation.

### spec.enableExactlyOnceDelivery

`bool`

When true, Pub/Sub guarantees that a message is not resent before its
acknowledgement deadline expires. An acknowledged message will not be
resent. Note: subscribers may still receive duplicates if the publisher
sends the same message multiple times.

### spec.deadLetterPolicy

`GcpPubSubSubscriptionDeadLetterPolicy`

Dead-letter policy. Messages that cannot be processed after repeated
delivery attempts are forwarded to the configured dead-letter topic.

### spec.deadLetterPolicy.deadLetterTopic

`string | valueFrom`

The topic to which dead-letter messages are published.
Format: projects/{project}/topics/{topic}

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.deadLetterPolicy.maxDeliveryAttempts

`int32`

Maximum number of delivery attempts before a message is dead-lettered.
A delivery attempt is counted as 1 + (NACKs + ack deadline exceeded events).
Range: 5 to 100. Defaults to 5 when set to 0.

- rule: max_delivery_attempts must be between 5 and 100, or 0 for default (5)

### spec.retryPolicy

`GcpPubSubSubscriptionRetryPolicy`

Retry policy. Controls the backoff between consecutive delivery attempts
after a NACK or ack deadline exceeded event.

### spec.retryPolicy.minimumBackoff

`string`

Minimum delay between consecutive delivery attempts of a given message.
Format: duration string (e.g., "10s").
Range: 0s to 600s. Defaults to 10s.

### spec.retryPolicy.maximumBackoff

`string`

Maximum delay between consecutive delivery attempts of a given message.
Format: duration string (e.g., "600s").
Range: 0s to 600s. Defaults to 600s.

### spec.pushConfig

`GcpPubSubSubscriptionPushConfig`

Push delivery configuration. When set, Pub/Sub sends messages as HTTP POST
requests to the configured endpoint. Mutually exclusive with bigquery_config
and cloud_storage_config.

### spec.pushConfig.pushEndpoint

`string | valueFrom` · required

URL to which Pub/Sub pushes messages. Must use HTTPS.
Accepts a literal URL or a reference to a GcpCloudRun service — pushing to
a Cloud Run service in the same environment is the canonical serverless
consumer pattern, and the service URL contains a generated suffix that can
only be known by reading the deployed service's output. Pair a Cloud Run
push endpoint with oidc_token (the service account must hold run.invoker)
unless the service allows unauthenticated invocations.

- references: GcpCloudRun (`status.outputs.url`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudRun, name: <that resource's name>, fieldPath: status.outputs.url}} -- a bare string does not parse

### spec.pushConfig.attributes

`map<string, string>`

Endpoint configuration attributes. The supported attribute is "x-goog-version"
which controls the push message format ("v1beta1" or "v1").

### spec.pushConfig.oidcToken

`GcpPubSubSubscriptionPushConfigOidcToken`

OIDC token configuration for authenticating push requests.

### spec.pushConfig.oidcToken.serviceAccountEmail

`string | valueFrom` · required

Service account used to generate the OIDC token. Accepts a literal
email or a reference to a GcpServiceAccount resource.
The caller must have iam.serviceAccounts.actAs permission on this account.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.pushConfig.oidcToken.audience

`string`

Audience claim for the OIDC token. Identifies the intended recipient.
Defaults to the push endpoint URL if not specified.

### spec.pushConfig.noWrapper

`GcpPubSubSubscriptionPushConfigNoWrapper`

The message payload is sent unwrapped (no Pub/Sub envelope).

### spec.pushConfig.noWrapper.writeMetadata

`bool`

When true, Pub/Sub message metadata is written as HTTP headers
(x-goog-pubsub-<key>:<value>) and message attributes as plain headers.

### spec.pushConfig.pubsubWrapper

`GcpPubSubSubscriptionPushConfigPubsubWrapper`

The message is wrapped in the standard Pub/Sub envelope.

### spec.bigqueryConfig

`GcpPubSubSubscriptionBigQueryConfig`

BigQuery delivery configuration. When set, Pub/Sub writes messages directly
to a BigQuery table. Mutually exclusive with push_config and
cloud_storage_config.

### spec.bigqueryConfig.table

`string | valueFrom` · required

The BigQuery table to write messages to.
Format: {project_id}.{dataset_id}.{table_id}
Accepts a literal or a reference to a GcpBigQueryTable resource (its
qualified_name output is exactly this dotted form).

- references: GcpBigQueryTable (`status.outputs.qualified_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBigQueryTable, name: <that resource's name>, fieldPath: status.outputs.qualified_name}} -- a bare string does not parse

### spec.bigqueryConfig.useTopicSchema

`bool`

When true, use the Pub/Sub topic's schema to map message fields to BigQuery columns.
Only one of use_topic_schema and use_table_schema can be true.

### spec.bigqueryConfig.useTableSchema

`bool`

When true, use the BigQuery table's schema to determine which message fields
to write. Only one of use_topic_schema and use_table_schema can be true.

### spec.bigqueryConfig.dropUnknownFields

`bool`

When true (and use_topic_schema or use_table_schema is true), message fields
not present in the BigQuery table schema are silently dropped. When false,
messages with extra fields are not written and remain in the backlog.

### spec.bigqueryConfig.writeMetadata

`bool`

When true, the subscription name, messageId, publishTime, attributes, and
orderingKey are written to additional columns in the BigQuery table.

### spec.bigqueryConfig.serviceAccountEmail

`string | valueFrom`

Service account to use for writing to BigQuery. Accepts a literal
email or a reference to a GcpServiceAccount resource. Defaults to the
Pub/Sub service agent if not specified.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.cloudStorageConfig

`GcpPubSubSubscriptionCloudStorageConfig`

Cloud Storage delivery configuration. When set, Pub/Sub writes messages to
Cloud Storage objects in batches. Mutually exclusive with push_config and
bigquery_config.

### spec.cloudStorageConfig.bucket

`string | valueFrom` · required

The Cloud Storage bucket to write messages to (without "gs://" prefix).
The bucket must already exist.

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.cloudStorageConfig.filenamePrefix

`string`

Prefix for Cloud Storage filenames.

### spec.cloudStorageConfig.filenameSuffix

`string`

Suffix for Cloud Storage filenames. Must not end in "/".

### spec.cloudStorageConfig.filenameDatetimeFormat

`string`

Format string for datetime in Cloud Storage filenames.

### spec.cloudStorageConfig.maxBytes

`int64`

Maximum bytes per Cloud Storage file before a new file is created.
Range: 1024 (1 KB) to 10737418240 (10 GiB).

### spec.cloudStorageConfig.maxDuration

`string`

Maximum duration before a new Cloud Storage file is created.
Format: duration string (e.g., "300s").
Range: 60s (1 minute) to 600s (10 minutes). Default: 300s (5 minutes).
Must not exceed the subscription's ack_deadline_seconds.

### spec.cloudStorageConfig.maxMessages

`int64`

Maximum number of messages per Cloud Storage file. Minimum: 1000.

### spec.cloudStorageConfig.avroConfig

`GcpPubSubSubscriptionCloudStorageConfigAvroConfig`

Messages are written as Avro records, optionally using the topic schema
and carrying message metadata.

### spec.cloudStorageConfig.avroConfig.useTopicSchema

`bool`

When true, serialize output using the topic schema.

### spec.cloudStorageConfig.avroConfig.writeMetadata

`bool`

When true, include subscription name, messageId, publishTime, attributes,
and orderingKey as additional fields in the Avro output.

### spec.cloudStorageConfig.textConfig

`GcpPubSubSubscriptionCloudStorageConfigTextConfig`

Messages are written as raw text.

### spec.cloudStorageConfig.serviceAccountEmail

`string | valueFrom`

Service account to use for writing to Cloud Storage. Accepts a literal
email or a reference to a GcpServiceAccount resource. Defaults to the
Pub/Sub service agent if not specified.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.labels

`map<string, string>`

User-defined labels attached to the subscription, for cost attribution
and fleet queries. Merged with Planton's platform labels (which win on
key conflicts).

### spec.messageTransforms

`[]GcpPubSubSubscriptionMessageTransform`

Ordered pipeline of transforms applied to every message before
delivery to this subscription — reshape payloads for this consumer
without changing what other subscriptions on the topic see.
Transforms run in list order; a transform returning null drops the
message.

- rule: each transform step is exactly one of: javascript_udf or ai_inference

### spec.messageTransforms[].javascriptUdf

`GcpPubSubSubscriptionMessageTransformJavascriptUdf`

A JavaScript user-defined function transform.

### spec.messageTransforms[].javascriptUdf.functionName

`string` · required

Name of the JavaScript function to invoke from the code below.
Must be unique across all transforms on the resource.

- rule: {"required":true}

### spec.messageTransforms[].javascriptUdf.code

`string` · required

The JavaScript source code defining the function. The function
signature is (message, metadata) => message; return null or undefined
to drop the message.

- rule: {"required":true}

### spec.messageTransforms[].disabled

`bool`

When true, this transform is kept in the pipeline definition but not
applied — the staging lever for rolling a transform in or out without
losing its position in the ordered list.

### spec.messageTransforms[].aiInference

`GcpPubSubSubscriptionMessageTransformAiInference`

An AI inference transform backed by a Vertex AI model endpoint.

### spec.messageTransforms[].aiInference.endpoint

`string | valueFrom` · required

The Vertex AI model endpoint inference requests are sent to. Accepts a
literal path — projects/{project}/locations/{location}/endpoints/{endpoint}
for a dedicated endpoint, or
projects/{project}/locations/{location}/publishers/{publisher}/models/{model}
for a publisher model — or a reference to a GcpVertexAiEndpoint resource
(its endpoint_id output is exactly the dedicated-endpoint form).

- references: GcpVertexAiEndpoint (`status.outputs.endpoint_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVertexAiEndpoint, name: <that resource's name>, fieldPath: status.outputs.endpoint_id}} -- a bare string does not parse

### spec.messageTransforms[].aiInference.serviceAccountEmail

`string | valueFrom`

The service account used to make prediction requests against the
endpoint (it needs Vertex AI invocation permission on the model).
Accepts a literal email or a reference to a GcpServiceAccount resource.
Defaults to the Pub/Sub service agent if not specified.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.messageTransforms[].aiInference.unstructuredInference

`GcpPubSubSubscriptionMessageTransformAiInferenceUnstructuredInference`

Configuration for making inferences using arbitrary JSON payloads
(rather than a model-specific request schema).

### spec.messageTransforms[].aiInference.unstructuredInference.parameters

`map<string, string>`

A parameters object included in each inference request (e.g. model
temperature or system-prompt knobs the endpoint understands). Combined
with the message data to form the request body.

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the subscription for org-policy and
IAM conditions. Keys in the form "tagKeys/{id}", values
"tagValues/{id}". Create-time only: changing them later replaces the
subscription (its backlog is lost — plan tag changes deliberately).

### spec.deletionPolicy

`string`

Deletion policy for the subscription — what happens when this
resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the subscription is deleted; its unacknowledged
               backlog is lost immediately
  "PREVENT" -- destroy FAILS; protects a consumer whose backlog
               must never be silently dropped
  "ABANDON" -- the subscription is removed from management but keeps
               accumulating and serving messages in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `delivery_method_mutual_exclusion`: only one of push_config, bigquery_config, or cloud_storage_config can be set; if none is set, the subscription uses pull delivery
- `bigquery_schema_mutual_exclusion`: only one of use_topic_schema and use_table_schema can be true in bigquery_config

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPubSubSubscription, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.subscription_id` | `string` | The fully qualified subscription ID. Format: projects/{project}/subscriptions/{name} This is the value downstream resources use to reference this subscription. |
| `status.outputs.subscription_name` | `string` | The short subscription name (same as the spec's subscription_name input). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.topic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.deadLetterPolicy.deadLetterTopic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.pushConfig.pushEndpoint` | GcpCloudRun | `status.outputs.url` |
| `spec.pushConfig.oidcToken.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.bigqueryConfig.table` | GcpBigQueryTable | `status.outputs.qualified_name` |
| `spec.bigqueryConfig.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.cloudStorageConfig.bucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.cloudStorageConfig.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.messageTransforms[].aiInference.endpoint` | GcpVertexAiEndpoint | `status.outputs.endpoint_id` |
| `spec.messageTransforms[].aiInference.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)
