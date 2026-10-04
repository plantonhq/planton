# AwsSagemakerFeatureGroup

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSagemakerFeatureGroupSpec defines the desired configuration for an
Amazon SageMaker Feature Store feature group - the schema'd store ML
features are written to and served from. The group's AWS name derives
from metadata.name.

A feature group has an ONLINE store (low-latency reads for serving),
an OFFLINE store (an S3/Glue table for training), or both - at least
one. Everything except the online store's TTL and the throughput
settings is create-time only: changing it replaces the group (AWS's
contract).

## Example

```yaml
# Canonical AwsSagemakerFeatureGroup example (hack/dev manifest and
# refgen Example source): a dual-store group exercising every arm -
# online store with InMemory storage, KMS, and TTL; offline store with
# Iceberg, KMS, and a named Glue catalog entry; a vector feature;
# provisioned throughput. Literal ARNs stand in for composed references
# so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerFeatureGroup
metadata:
  name: customer-features
  id: customer-features
  org: test-org
  env: dev
spec:
  region: us-west-2
  recordIdentifierFeatureName: customer_id
  eventTimeFeatureName: event_time
  description: Customer features for churn scoring
  roleArn:
    value: arn:aws:iam::123456789012:role/sagemaker-execution
  featureDefinitions:
    - name: customer_id
      type: String
    - name: event_time
      type: Fractional
    - name: lifetime_value
      type: Fractional
    - name: embedding
      type: Fractional
      collectionType: Vector
      vectorDimension: 256
  onlineStore:
    enabled: true
    kmsKeyArn:
      value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
    storageType: InMemory
    ttl:
      unit: Days
      value: 30
  offlineStore:
    s3Uri: s3://my-feature-store/customers/
    kmsKeyArn:
      value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
    tableFormat: Iceberg
    disableGlueTableCreation: false
    dataCatalog:
      catalog: AwsDataCatalog
      database: feature_store
      tableName: customer_features
  throughput:
    mode: Provisioned
    provisionedReadCapacityUnits: 100
    provisionedWriteCapacityUnits: 50
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.recordIdentifierFeatureName` | `string` | yes |  |  |
| `spec.eventTimeFeatureName` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.featureDefinitions` | `[]AwsSagemakerFeatureGroupFeature` | yes |  |  |
| `spec.featureDefinitions[].name` | `string` | yes |  |  |
| `spec.featureDefinitions[].type` | `string` |  |  |  |
| `spec.featureDefinitions[].collectionType` | `string` |  |  |  |
| `spec.featureDefinitions[].vectorDimension` | `int32` |  |  |  |
| `spec.onlineStore` | `AwsSagemakerFeatureGroupOnlineStore` |  |  |  |
| `spec.onlineStore.enabled` | `bool` |  |  |  |
| `spec.onlineStore.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.onlineStore.storageType` | `string` |  |  |  |
| `spec.onlineStore.ttl` | `AwsSagemakerFeatureGroupTtl` |  |  |  |
| `spec.onlineStore.ttl.unit` | `string` |  |  |  |
| `spec.onlineStore.ttl.value` | `int32` |  |  |  |
| `spec.offlineStore` | `AwsSagemakerFeatureGroupOfflineStore` |  |  |  |
| `spec.offlineStore.s3Uri` | `string` | yes |  |  |
| `spec.offlineStore.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.offlineStore.disableGlueTableCreation` | `bool` |  |  |  |
| `spec.offlineStore.tableFormat` | `string` |  |  |  |
| `spec.offlineStore.dataCatalog` | `AwsSagemakerFeatureGroupDataCatalog` |  |  |  |
| `spec.offlineStore.dataCatalog.catalog` | `string` | yes |  |  |
| `spec.offlineStore.dataCatalog.database` | `string` | yes |  |  |
| `spec.offlineStore.dataCatalog.tableName` | `string` | yes |  |  |
| `spec.throughput` | `AwsSagemakerFeatureGroupThroughput` |  |  |  |
| `spec.throughput.mode` | `string` |  |  |  |
| `spec.throughput.provisionedReadCapacityUnits` | `int32` |  |  |  |
| `spec.throughput.provisionedWriteCapacityUnits` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the feature group will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.recordIdentifierFeatureName

`string` · required

The feature whose value uniquely identifies a record (must be one
of `feature_definitions`; 1-64 characters).

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z]([-_]*[0-9A-Za-z])*$"}}

### spec.eventTimeFeatureName

`string` · required

The feature carrying each record's event time (must be one of
`feature_definitions`; Fractional epoch seconds or String
ISO-8601). The TTL clock and offline-store partitioning key.

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z]([-_]*[0-9A-Za-z])*$"}}

### spec.description

`string`

Free-form description (max 128 characters). Changing it replaces
the group (provider-enforced).

- rule: {"string":{"maxLen":"128"}}

### spec.roleArn

`string | valueFrom` · required

IAM role used to persist data into the offline store's S3 location.
The role must trust sagemaker.amazonaws.com and be able to write
the bucket. When offline_store is configured, CreateFeatureGroup
validates the role AGAINST THE BUCKET at create time - it assumes
the role and calls s3:GetBucketAcl, and writes carry
s3:PutObjectAcl (the verbs AWS's own AmazonSageMakerFeatureStoreAccess
managed policy grants). A role that can only read/write objects
fails the create with ValidationException "Invalid S3Uri" wrapping
the S3 AccessDenied (live-verified 2026-08-25).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.featureDefinitions

`[]AwsSagemakerFeatureGroupFeature` · required

The group's schema: every feature records may carry (1-2500).

- rule: {"repeated":{"minItems":"1","maxItems":"2500"}}
- rule: vector_dimension is required when collection_type is Vector and forbidden otherwise

### spec.featureDefinitions[].name

`string` · required

Feature name (1-64 characters; letters, digits, hyphen,
underscore). AWS reserves "is_deleted", "write_time", and
"api_invocation_time".

- rule: {"string":{"minLen":"1","maxLen":"64","pattern":"^[0-9A-Za-z]([-_]*[0-9A-Za-z])*$","notIn":["is_deleted","write_time","api_invocation_time"]}}

### spec.featureDefinitions[].type

`string`

Value type: "Integral", "Fractional", or "String".

- rule: {"string":{"in":["Integral","Fractional","String"]}}

### spec.featureDefinitions[].collectionType

`string`

Collection shape for multi-valued features: "List", "Set", or
"Vector" (embeddings - requires `vector_dimension`). Omitted =
scalar. Collections require an online store with InMemory storage.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["List","Set","Vector"]}}

### spec.featureDefinitions[].vectorDimension

`int32` · optional (explicit presence)

Number of dimensions of a Vector feature (1-8192). Required exactly
when collection_type is Vector.

- rule: {"int32":{"lte":8192,"gte":1}}

### spec.onlineStore

`AwsSagemakerFeatureGroupOnlineStore`

Low-latency store for real-time serving.

### spec.onlineStore.enabled

`bool`

Turn the online store on. Changing it replaces the group.

### spec.onlineStore.kmsKeyArn

`string | valueFrom`

KMS key encrypting the online store at rest.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.onlineStore.storageType

`string`

Storage tier: "Standard" (default) or "InMemory" (lowest latency;
required for collection-typed features). Changing it replaces the
group.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Standard","InMemory"]}}

### spec.onlineStore.ttl

`AwsSagemakerFeatureGroupTtl`

Hard-delete records this long after their event time
(ExpiresAt = EventTime + ttl). The ONLY online-store setting that
updates in place.

### spec.onlineStore.ttl.unit

`string`

"Seconds", "Minutes", "Hours", "Days", or "Weeks".

- rule: {"string":{"in":["Seconds","Minutes","Hours","Days","Weeks"]}}

### spec.onlineStore.ttl.value

`int32`

The duration count (>= 1).

- rule: {"int32":{"gte":1}}

### spec.offlineStore

`AwsSagemakerFeatureGroupOfflineStore`

S3/Glue store for training datasets and point-in-time queries.

### spec.offlineStore.s3Uri

`string` · required

S3 URI records are written under. Example: "s3://my-features/"

- rule: {"string":{"minLen":"1","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.offlineStore.kmsKeyArn

`string | valueFrom`

KMS key encrypting offline-store objects.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.offlineStore.disableGlueTableCreation

`bool`

Skip the automatic Glue Data Catalog table (register your own
catalog entry instead - pair with `data_catalog`).

### spec.offlineStore.tableFormat

`string`

Table format: "Glue" (default - a standard Glue table) or
"Iceberg" (Apache Iceberg - faster queries, compaction).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Glue","Iceberg"]}}

### spec.offlineStore.dataCatalog

`AwsSagemakerFeatureGroupDataCatalog`

Name your own Glue catalog entry (with
disable_glue_table_creation) or override the generated one's
location.

### spec.offlineStore.dataCatalog.catalog

`string` · required

Glue catalog name.

- rule: {"string":{"minLen":"1"}}

### spec.offlineStore.dataCatalog.database

`string` · required

Glue database name.

- rule: {"string":{"minLen":"1"}}

### spec.offlineStore.dataCatalog.tableName

`string` · required

Glue table name.

- rule: {"string":{"minLen":"1"}}

### spec.throughput

`AwsSagemakerFeatureGroupThroughput`

Read/write billing mode - on-demand (default) or provisioned
capacity. The one surface besides TTL that updates in place.

- rule: provisioned capacity units require mode Provisioned

### spec.throughput.mode

`string`

"OnDemand" (default - pay per request) or "Provisioned" (fixed
capacity - pair with the capacity units).

- rule: {"string":{"in":["OnDemand","Provisioned"]}}

### spec.throughput.provisionedReadCapacityUnits

`int32` · optional (explicit presence)

Provisioned read capacity units (0-10000000). Only with Provisioned
mode.

- rule: {"int32":{"lte":10000000,"gte":0}}

### spec.throughput.provisionedWriteCapacityUnits

`int32` · optional (explicit presence)

Provisioned write capacity units (0-10000000). Only with
Provisioned mode.

- rule: {"int32":{"lte":10000000,"gte":0}}

## Validation Rules

- `at_least_one_store`: at least one of online_store (enabled) and offline_store must be configured
- `feature_names_unique`: feature_definitions entries must have unique names
- `record_identifier_defined`: record_identifier_feature_name must be one of feature_definitions
- `event_time_defined`: event_time_feature_name must be one of feature_definitions

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerFeatureGroup, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.feature_group_name` | `string` | The feature group name (the AWS identity ingestion and serving calls use). |
| `status.outputs.feature_group_arn` | `string` | The Amazon Resource Name of the feature group. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.onlineStore.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.offlineStore.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
