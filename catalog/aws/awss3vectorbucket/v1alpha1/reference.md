# AwsS3VectorBucket

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsS3VectorBucketSpec defines one S3 vector bucket with its vector
indexes - purpose-built storage for AI embeddings, queried by
similarity instead of key. The natural backend for Bedrock
knowledge bases (AwsBedrockKnowledgeBase's s3_vectors arm points at
an index defined here) and for any RAG stack that wants
pay-per-query vector search without running a vector database.

The bucket's name is metadata.name (3-63 lowercase letters,
digits, hyphens). Indexes key their entries by name, and EVERY
index property is fixed for life - an index is replaced, not
edited, so dimension must match the embedding model before the
first vector lands. The index data_type argument is module-pinned
to float32 - the provider's enum holds exactly that one value.

## Example

```yaml
# Canonical AwsS3VectorBucket example (hack/dev manifest and refgen
# Example source): embedding storage for a Bedrock knowledge base -
# one index sized for Titan Text v2 (1024 dimensions).
apiVersion: aws.planton.dev/v1alpha1
kind: AwsS3VectorBucket
metadata:
  name: kb-vectors
  id: kb-vectors
  org: test-org
  env: dev
spec:
  region: us-west-2
  forceDestroy: true
  indexes:
    - name: kb-embeddings
      dimension: 1024
      distanceMetric: cosine
      nonFilterableMetadataKeys:
        - source_text
        - source_uri
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.encryption` | `AwsS3VectorsEncryption` |  |  |  |
| `spec.encryption.sseType` | `string` |  |  |  |
| `spec.encryption.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.policy` | `string` |  |  |  |
| `spec.indexes` | `[]AwsS3VectorsIndex` |  |  |  |
| `spec.indexes[].name` | `string` | yes |  |  |
| `spec.indexes[].dimension` | `int64` |  |  |  |
| `spec.indexes[].distanceMetric` | `string` |  |  |  |
| `spec.indexes[].nonFilterableMetadataKeys` | `[]string` |  |  |  |
| `spec.indexes[].encryption` | `AwsS3VectorsEncryption` |  |  |  |
| `spec.indexes[].encryption.sseType` | `string` |  |  |  |
| `spec.indexes[].encryption.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |

## Field Details

### spec.region

`string` · required

The AWS region the vector bucket lives in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.encryption

`AwsS3VectorsEncryption`

Encryption at rest. Unset means S3-managed keys (SSE-S3). Fixed
for life of the bucket.

- rule: kms_key_arn requires sse_type aws:kms

### spec.encryption.sseType

`string`

The encryption type: S3-managed keys (AES256) or KMS (aws:kms).

- rule: {"string":{"in":["AES256","aws:kms"]}}

### spec.encryption.kmsKeyArn

`string | valueFrom`

The KMS key for aws:kms. Reference an AwsKmsKey key_arn output
or pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.forceDestroy

`bool`

Let destroy succeed on a non-empty bucket by deleting every
index and vector first. Off by default: destroying embeddings
that took real money to compute should hurt. Config-only at AWS
- imports never round-trip it.

### spec.policy

`string`

The bucket's resource policy (JSON) - who can put/query vectors.
Bedrock knowledge bases in other accounts get their access here.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.indexes

`[]AwsS3VectorsIndex`

The vector indexes, keyed by name.

### spec.indexes[].name

`string` · required

The index name - what Bedrock knowledge bases and query calls
address. 3-63 lowercase letters, digits, hyphens. Fixed for
life.

- rule: {"string":{"minLen":"3","maxLen":"63","pattern":"^[a-z0-9][a-z0-9-]*[a-z0-9]$"}}

### spec.indexes[].dimension

`int64`

The vector dimension (1-4096). MUST equal the embedding model's
output dimension (Titan Text v2: 1024/512/256; Cohere: 1024) -
a mismatched put is rejected vector-by-vector.

- rule: {"int64":{"lte":"4096","gte":"1"}}

### spec.indexes[].distanceMetric

`string`

How similarity is scored: "cosine" (angle - the usual choice for
normalized text embeddings) or "euclidean" (distance).  Must
match what the embedding model was trained for.

- rule: {"string":{"in":["cosine","euclidean"]}}

### spec.indexes[].nonFilterableMetadataKeys

`[]string`

Metadata keys stored WITH vectors but excluded from filterable
metadata (1-10 keys). Bulky payload fields (the source text
chunk, URLs) belong here - filterable metadata has a hard size
budget per vector.

- rule: {"repeated":{"maxItems":"10","unique":true}}

### spec.indexes[].encryption

`AwsS3VectorsEncryption`

Encryption for THIS index when it must differ from the bucket
default. Fixed for life.

- rule: kms_key_arn requires sse_type aws:kms

### spec.indexes[].encryption.sseType

`string`

The encryption type: S3-managed keys (AES256) or KMS (aws:kms).

- rule: {"string":{"in":["AES256","aws:kms"]}}

### spec.indexes[].encryption.kmsKeyArn

`string | valueFrom`

The KMS key for aws:kms. Reference an AwsKmsKey key_arn output
or pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

## Validation Rules

- `spec.index_names_unique`: index names must be unique

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsS3VectorBucket, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.vector_bucket_arn` | `string` | The vector bucket's ARN - what policies and Bedrock knowledge bases reference, and the provider's import ID. |
| `status.outputs.index_arns` | `map<string, string>` | Each index's ARN, keyed by index name - what a Bedrock knowledge base's s3_vectors arm points at. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.encryption.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.indexes[].encryption.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockKnowledgeBase | `spec.storage.s3Vectors.vectorBucketArn` | `status.outputs.vector_bucket_arn` |

## See Also

- [Overview](../README.md)
