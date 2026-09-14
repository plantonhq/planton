# GcpPubSubTopic

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpPubSubTopicSpec defines the configuration for a GCP Pub/Sub topic.

A topic is a named resource to which messages are sent by publishers and
from which messages are pulled by subscribers via subscriptions. Topics
are the foundation of event-driven architectures in GCP -- they decouple
message producers from consumers and support one-to-many message distribution.

Important behavioral notes:

  - The topic_name field is immutable after creation. Changing it requires
    destroying and recreating the topic (and all its subscriptions).

  - Message retention (message_retention_duration) is independent of
    subscription-level retention. When set, any subscription can seek to
    a timestamp within the retention window.

  - Message storage policy regions are enforced at publish time. Publishers
    in non-allowed regions have their messages routed to an allowed region.
    When enforce_in_transit is true, publish calls from non-allowed regions
    are rejected entirely.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubTopic
metadata:
  name: test-pubsub-topic
spec:
  projectId:
    value: "test-project"
  topicName: test-events-topic
  # Explicit destroy behavior: DELETE removes the topic with the stack
  # (its subscriptions survive but drain to empty).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.topicName` | `string` | yes |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.messageRetentionDuration` | `string` |  |  |  |
| `spec.messageStoragePolicy` | `GcpPubSubTopicMessageStoragePolicy` |  |  |  |
| `spec.messageStoragePolicy.allowedPersistenceRegions` | `[]string` | yes |  |  |
| `spec.messageStoragePolicy.enforceInTransit` | `bool` |  |  |  |
| `spec.schemaSettings` | `GcpPubSubTopicSchemaSettings` |  |  |  |
| `spec.schemaSettings.schema` | `string \| valueFrom` | yes |  | GcpPubSubSchema (`status.outputs.schema_id`) |
| `spec.schemaSettings.encoding` | `string` |  |  |  |
| `spec.schemaSettings.firstRevisionId` | `string \| valueFrom` |  |  | GcpPubSubSchema (`status.outputs.revision_id`) |
| `spec.schemaSettings.lastRevisionId` | `string \| valueFrom` |  |  | GcpPubSubSchema (`status.outputs.revision_id`) |
| `spec.ingestionDataSourceSettings` | `GcpPubSubTopicIngestionDataSourceSettings` |  |  |  |
| `spec.ingestionDataSourceSettings.awsKinesis` | `GcpPubSubTopicIngestionAwsKinesis` |  |  |  |
| `spec.ingestionDataSourceSettings.awsKinesis.streamArn` | `string` | yes |  |  |
| `spec.ingestionDataSourceSettings.awsKinesis.consumerArn` | `string` | yes |  |  |
| `spec.ingestionDataSourceSettings.awsKinesis.awsRoleArn` | `string` | yes |  |  |
| `spec.ingestionDataSourceSettings.awsKinesis.gcpServiceAccount` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.ingestionDataSourceSettings.awsMsk` | `GcpPubSubTopicIngestionAwsMsk` |  |  |  |
| `spec.ingestionDataSourceSettings.awsMsk.clusterArn` | `string` | yes |  |  |
| `spec.ingestionDataSourceSettings.awsMsk.topic` | `string` | yes |  |  |
| `spec.ingestionDataSourceSettings.awsMsk.awsRoleArn` | `string` | yes |  |  |
| `spec.ingestionDataSourceSettings.awsMsk.gcpServiceAccount` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.ingestionDataSourceSettings.azureEventHubs` | `GcpPubSubTopicIngestionAzureEventHubs` |  |  |  |
| `spec.ingestionDataSourceSettings.azureEventHubs.resourceGroup` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.azureEventHubs.namespace` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.azureEventHubs.eventHub` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.azureEventHubs.clientId` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.azureEventHubs.tenantId` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.azureEventHubs.subscriptionId` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.azureEventHubs.gcpServiceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.ingestionDataSourceSettings.cloudStorage` | `GcpPubSubTopicIngestionCloudStorage` |  |  |  |
| `spec.ingestionDataSourceSettings.cloudStorage.bucket` | `string \| valueFrom` | yes |  | GcpGcsBucket (`status.outputs.bucket_id`) |
| `spec.ingestionDataSourceSettings.cloudStorage.matchGlob` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.cloudStorage.minimumObjectCreateTime` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.cloudStorage.avroFormat` | `GcpPubSubTopicIngestionCloudStorageAvroFormat` |  |  |  |
| `spec.ingestionDataSourceSettings.cloudStorage.pubsubAvroFormat` | `GcpPubSubTopicIngestionCloudStoragePubsubAvroFormat` |  |  |  |
| `spec.ingestionDataSourceSettings.cloudStorage.textFormat` | `GcpPubSubTopicIngestionCloudStorageTextFormat` |  |  |  |
| `spec.ingestionDataSourceSettings.cloudStorage.textFormat.delimiter` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.confluentCloud` | `GcpPubSubTopicIngestionConfluentCloud` |  |  |  |
| `spec.ingestionDataSourceSettings.confluentCloud.bootstrapServer` | `string` | yes |  |  |
| `spec.ingestionDataSourceSettings.confluentCloud.topic` | `string` | yes |  |  |
| `spec.ingestionDataSourceSettings.confluentCloud.identityPoolId` | `string` | yes |  |  |
| `spec.ingestionDataSourceSettings.confluentCloud.gcpServiceAccount` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.ingestionDataSourceSettings.confluentCloud.clusterId` | `string` |  |  |  |
| `spec.ingestionDataSourceSettings.platformLogsSettings` | `GcpPubSubTopicIngestionPlatformLogsSettings` |  |  |  |
| `spec.ingestionDataSourceSettings.platformLogsSettings.severity` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.messageTransforms` | `[]GcpPubSubTopicMessageTransform` |  |  |  |
| `spec.messageTransforms[].javascriptUdf` | `GcpPubSubTopicMessageTransformJavascriptUdf` |  |  |  |
| `spec.messageTransforms[].javascriptUdf.functionName` | `string` | yes |  |  |
| `spec.messageTransforms[].javascriptUdf.code` | `string` | yes |  |  |
| `spec.messageTransforms[].disabled` | `bool` |  |  |  |
| `spec.messageTransforms[].aiInference` | `GcpPubSubTopicMessageTransformAiInference` |  |  |  |
| `spec.messageTransforms[].aiInference.endpoint` | `string \| valueFrom` | yes |  | GcpVertexAiEndpoint (`status.outputs.endpoint_id`) |
| `spec.messageTransforms[].aiInference.serviceAccountEmail` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.messageTransforms[].aiInference.unstructuredInference` | `GcpPubSubTopicMessageTransformAiInferenceUnstructuredInference` |  |  |  |
| `spec.messageTransforms[].aiInference.unstructuredInference.parameters` | `map<string, string>` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the topic will be created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.topicName

`string` · required

Name of the Pub/Sub topic.
Must be 3-255 characters, start with a letter, and contain only letters,
numbers, hyphens, underscores, periods, tildes, plus signs, and percent
signs. Names beginning with "goog" are reserved by Google and rejected
at create time. Immutable after creation.

- rule: topic names beginning with 'goog' are reserved by Google — choose a different name
- rule: {"required":true,"string":{"minLen":"3","maxLen":"255","pattern":"^[a-zA-Z][a-zA-Z0-9\\-_\\.~+%]*$"}}

### spec.kmsKeyName

`string | valueFrom`

Cloud KMS key for encrypting messages at rest (CMEK).
Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
The Pub/Sub service account must have roles/cloudkms.cryptoKeyEncrypterDecrypter
on this key. If not set, messages are encrypted with Google-managed keys.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.messageRetentionDuration

`string`

Minimum duration to retain a message after it is published to the topic.
When set, messages published in the last message_retention_duration are always
available to subscribers, and any attached subscription can seek to a timestamp
within the retention window.
Format: duration string (e.g., "604800s" for 7 days).
Range: 600s (10 minutes) to 2678400s (31 days).
If not set, message retention is controlled by individual subscriptions.

### spec.messageStoragePolicy

`GcpPubSubTopicMessageStoragePolicy`

Policy constraining the set of GCP regions where messages may be stored.
When not set, no regional constraints are applied.

### spec.messageStoragePolicy.allowedPersistenceRegions

`[]string` · required

A list of GCP region IDs where messages may be persisted in storage.
Messages published by publishers running in non-allowed regions will be
routed for storage in one of the allowed regions.
Must contain at least one region when the policy is specified.

- rule: {"repeated":{"minItems":"1"}}

### spec.messageStoragePolicy.enforceInTransit

`bool`

When true, allowed_persistence_regions is also used to enforce in-transit
guarantees for messages. Pub/Sub will fail publish operations on this topic
and subscribe operations on any subscription attached to this topic in any
region not listed in allowed_persistence_regions.

### spec.schemaSettings

`GcpPubSubTopicSchemaSettings`

Schema validation settings for messages published to the topic.
When set, all published messages are validated against the specified schema.

### spec.schemaSettings.schema

`string | valueFrom` · required

The Pub/Sub schema that published messages must conform to.
Accepts a literal fully qualified path (projects/{project}/schemas/{schema})
or a reference to a GcpPubSubSchema resource. If the schema is later
deleted, the topic validates against the "_deleted-schema_" sentinel and
publishes fail — destroy topics (or recreate them without validation)
before destroying the schema they reference.

- references: GcpPubSubSchema (`status.outputs.schema_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubSchema, name: <that resource's name>, fieldPath: status.outputs.schema_id}} -- a bare string does not parse

### spec.schemaSettings.encoding

`string`

The encoding of messages validated against the schema.
Valid values: "JSON" or "BINARY".

- rule: encoding must be JSON or BINARY

### spec.schemaSettings.firstRevisionId

`string | valueFrom`

The minimum (inclusive) schema revision accepted for validating
messages. When empty, any revision created before last_revision_id
(or any revision at all, if that is also empty) is accepted.
Accepts a literal revision ID or a reference to a GcpPubSubSchema
resource — its revision_id output is the revision the schema's
current definition committed, so referencing it pins the topic to
the contract this deploy produced.

- references: GcpPubSubSchema (`status.outputs.revision_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubSchema, name: <that resource's name>, fieldPath: status.outputs.revision_id}} -- a bare string does not parse

### spec.schemaSettings.lastRevisionId

`string | valueFrom`

The maximum (inclusive) schema revision accepted for validating
messages. When empty, any revision created after first_revision_id
(or any revision at all, if that is also empty) is accepted. Pinning
BOTH bounds to the same revision freezes the contract exactly.
Accepts a literal revision ID or a reference to a GcpPubSubSchema
resource (its revision_id output).

- references: GcpPubSubSchema (`status.outputs.revision_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubSchema, name: <that resource's name>, fieldPath: status.outputs.revision_id}} -- a bare string does not parse

### spec.ingestionDataSourceSettings

`GcpPubSubTopicIngestionDataSourceSettings`

Settings for ingesting data from external sources into this topic.
Supports AWS Kinesis, AWS MSK, Azure Event Hubs, Cloud Storage, and
Confluent Cloud. Typically one data source is configured per topic.

### spec.ingestionDataSourceSettings.awsKinesis

`GcpPubSubTopicIngestionAwsKinesis`

Ingest from Amazon Kinesis Data Streams.

### spec.ingestionDataSourceSettings.awsKinesis.streamArn

`string` · required

The ARN of the Kinesis data stream to ingest from.

- rule: {"required":true}

### spec.ingestionDataSourceSettings.awsKinesis.consumerArn

`string` · required

The ARN of the Kinesis consumer to use for Enhanced Fan-Out delivery.

- rule: {"required":true}

### spec.ingestionDataSourceSettings.awsKinesis.awsRoleArn

`string` · required

The ARN of the AWS IAM role used for cross-account access to the Kinesis stream.

- rule: {"required":true}

### spec.ingestionDataSourceSettings.awsKinesis.gcpServiceAccount

`string | valueFrom` · required

The GCP service account used for Federated Identity authentication with AWS.
Accepts a literal email or a reference to a GcpServiceAccount resource.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.ingestionDataSourceSettings.awsMsk

`GcpPubSubTopicIngestionAwsMsk`

Ingest from Amazon Managed Streaming for Apache Kafka (MSK).

### spec.ingestionDataSourceSettings.awsMsk.clusterArn

`string` · required

The ARN of the MSK cluster to ingest from.

- rule: {"required":true}

### spec.ingestionDataSourceSettings.awsMsk.topic

`string` · required

The name of the Kafka topic in MSK to ingest from.

- rule: {"required":true}

### spec.ingestionDataSourceSettings.awsMsk.awsRoleArn

`string` · required

The ARN of the AWS IAM role used for cross-account access to the MSK cluster.

- rule: {"required":true}

### spec.ingestionDataSourceSettings.awsMsk.gcpServiceAccount

`string | valueFrom` · required

The GCP service account used for Federated Identity authentication with AWS.
Accepts a literal email or a reference to a GcpServiceAccount resource.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.ingestionDataSourceSettings.azureEventHubs

`GcpPubSubTopicIngestionAzureEventHubs`

Ingest from Azure Event Hubs.

### spec.ingestionDataSourceSettings.azureEventHubs.resourceGroup

`string`

The Azure resource group containing the Event Hubs namespace.

### spec.ingestionDataSourceSettings.azureEventHubs.namespace

`string`

The Azure Event Hubs namespace.

### spec.ingestionDataSourceSettings.azureEventHubs.eventHub

`string`

The name of the Event Hub to ingest from.

### spec.ingestionDataSourceSettings.azureEventHubs.clientId

`string`

The Azure AD application client ID for authentication.

### spec.ingestionDataSourceSettings.azureEventHubs.tenantId

`string`

The Azure AD tenant ID.

### spec.ingestionDataSourceSettings.azureEventHubs.subscriptionId

`string`

The Azure subscription ID.

### spec.ingestionDataSourceSettings.azureEventHubs.gcpServiceAccount

`string | valueFrom`

The GCP service account used for Federated Identity authentication with Azure.
Accepts a literal email or a reference to a GcpServiceAccount resource.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.ingestionDataSourceSettings.cloudStorage

`GcpPubSubTopicIngestionCloudStorage`

Ingest from Google Cloud Storage.

### spec.ingestionDataSourceSettings.cloudStorage.bucket

`string | valueFrom` · required

The name of the Cloud Storage bucket to ingest from (without "gs://" prefix).
See: https://cloud.google.com/storage/docs/buckets#naming

- references: GcpGcsBucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGcsBucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.ingestionDataSourceSettings.cloudStorage.matchGlob

`string`

Glob pattern used to match objects that will be ingested.
If unset, all objects in the bucket will be ingested.

### spec.ingestionDataSourceSettings.cloudStorage.minimumObjectCreateTime

`string`

Only ingest objects with a creation timestamp equal to or later than this value.
Format: RFC 3339 (e.g., "2024-01-01T00:00:00Z").
If unset, all objects are eligible for ingestion regardless of creation time.

### spec.ingestionDataSourceSettings.cloudStorage.avroFormat

`GcpPubSubTopicIngestionCloudStorageAvroFormat`

Read Cloud Storage data in Avro binary format. The bytes of each object
are set to the data field of a Pub/Sub message.

### spec.ingestionDataSourceSettings.cloudStorage.pubsubAvroFormat

`GcpPubSubTopicIngestionCloudStoragePubsubAvroFormat`

Read Cloud Storage data written via Cloud Storage subscriptions.
Restores the data and attributes of the originally exported Pub/Sub messages.

### spec.ingestionDataSourceSettings.cloudStorage.textFormat

`GcpPubSubTopicIngestionCloudStorageTextFormat`

Read Cloud Storage data in text format. Each line of text (as defined by
the delimiter) becomes the data field of a Pub/Sub message.

### spec.ingestionDataSourceSettings.cloudStorage.textFormat.delimiter

`string`

The line delimiter. Defaults to newline ("\n") when not set.

### spec.ingestionDataSourceSettings.confluentCloud

`GcpPubSubTopicIngestionConfluentCloud`

Ingest from Confluent Cloud.

### spec.ingestionDataSourceSettings.confluentCloud.bootstrapServer

`string` · required

The Confluent Cloud bootstrap server address. Format: host:port.

- rule: {"required":true}

### spec.ingestionDataSourceSettings.confluentCloud.topic

`string` · required

The name of the Confluent Cloud topic to ingest from.

- rule: {"required":true}

### spec.ingestionDataSourceSettings.confluentCloud.identityPoolId

`string` · required

The Workload Identity Pool ID used for Federated Identity authentication
with Confluent Cloud.

- rule: {"required":true}

### spec.ingestionDataSourceSettings.confluentCloud.gcpServiceAccount

`string | valueFrom` · required

The GCP service account used for Federated Identity authentication
with Confluent Cloud. Accepts a literal email or a reference to a
GcpServiceAccount resource.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.ingestionDataSourceSettings.confluentCloud.clusterId

`string`

The Confluent Cloud cluster ID. Optional.

### spec.ingestionDataSourceSettings.platformLogsSettings

`GcpPubSubTopicIngestionPlatformLogsSettings`

Platform logging settings for the ingestion pipeline.

### spec.ingestionDataSourceSettings.platformLogsSettings.severity

`string`

The minimum severity level of platform logs that will be written.
Valid values: "DISABLED", "DEBUG", "INFO", "WARNING", "ERROR".
If unset, no platform logs will be generated.

- rule: severity must be one of: DISABLED, DEBUG, INFO, WARNING, ERROR

### spec.labels

`map<string, string>`

User-defined labels attached to the topic, for cost attribution and
fleet queries. Merged with Planton's platform labels (which win on
key conflicts).

### spec.messageTransforms

`[]GcpPubSubTopicMessageTransform`

Ordered pipeline of transforms applied to every published message
before it is stored — redact, normalize, or filter at the topic
boundary instead of in every subscriber. Transforms run in list
order; a transform returning null drops the message.

- rule: each transform step is exactly one of: javascript_udf or ai_inference

### spec.messageTransforms[].javascriptUdf

`GcpPubSubTopicMessageTransformJavascriptUdf`

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

`GcpPubSubTopicMessageTransformAiInference`

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

`GcpPubSubTopicMessageTransformAiInferenceUnstructuredInference`

Configuration for making inferences using arbitrary JSON payloads
(rather than a model-specific request schema).

### spec.messageTransforms[].aiInference.unstructuredInference.parameters

`map<string, string>`

A parameters object included in each inference request (e.g. model
temperature or system-prompt knobs the endpoint understands). Combined
with the message data to form the request body.

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the topic for org-policy and IAM
conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
Create-time only: changing them later replaces the topic (and detaches
every subscription with it — plan tag changes deliberately).

### spec.deletionPolicy

`string`

Deletion policy for the topic — what happens when this resource is
destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the topic is deleted. Its subscriptions survive but
               receive no new messages and drain to empty
  "PREVENT" -- destroy FAILS; protects the topic an event pipeline
               publishes into
  "ABANDON" -- the topic is removed from management but left serving
               in GCP (publishers and subscriptions keep working)

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPubSubTopic, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.topic_id` | `string` | The fully qualified topic ID. Format: projects/{project}/topics/{name} This is the value downstream resources (subscriptions, schedulers, Cloud Functions) use to reference this topic. |
| `status.outputs.topic_name` | `string` | The short topic name (same as the spec's topic_name input). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.schemaSettings.schema` | GcpPubSubSchema | `status.outputs.schema_id` |
| `spec.schemaSettings.firstRevisionId` | GcpPubSubSchema | `status.outputs.revision_id` |
| `spec.schemaSettings.lastRevisionId` | GcpPubSubSchema | `status.outputs.revision_id` |
| `spec.ingestionDataSourceSettings.awsKinesis.gcpServiceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.ingestionDataSourceSettings.awsMsk.gcpServiceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.ingestionDataSourceSettings.azureEventHubs.gcpServiceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.ingestionDataSourceSettings.cloudStorage.bucket` | GcpGcsBucket | `status.outputs.bucket_id` |
| `spec.ingestionDataSourceSettings.confluentCloud.gcpServiceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.messageTransforms[].aiInference.endpoint` | GcpVertexAiEndpoint | `status.outputs.endpoint_id` |
| `spec.messageTransforms[].aiInference.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpCloudFunction | `spec.trigger.eventTrigger.pubsubTopic` | `status.outputs.topic_id` |
| GcpCloudSchedulerJob | `spec.pubsubTarget.topicName` | `status.outputs.topic_id` |
| GcpEventarcMessageBus | `spec.pipelines[].destination.topic` | `status.outputs.topic_id` |
| GcpEventarcTrigger | `spec.transportPubsubTopic` | `status.outputs.topic_id` |
| GcpGcsBucket | `spec.notifications[].topic` | `status.outputs.topic_id` |
| GcpGkeCluster | `spec.notificationPubsub.topic` | `status.outputs.topic_id` |
| GcpLoggingSink | `spec.destination.pubsubTopic` | `status.outputs.topic_id` |
| GcpPubSubSubscription | `spec.topic` | `status.outputs.topic_id` |
| GcpPubSubSubscription | `spec.deadLetterPolicy.deadLetterTopic` | `status.outputs.topic_id` |
| GcpSecretManagerSecret | `spec.topics` | `status.outputs.topic_id` |

## See Also

- [Overview](../README.md)
