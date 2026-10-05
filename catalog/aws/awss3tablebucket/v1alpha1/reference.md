# AwsS3TableBucket

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsS3TableBucketSpec defines one S3 table bucket (S3 Tables) with
its full contents: namespaces, Iceberg tables, resource policies,
and replication - one declarative owner for the whole analytics
storage unit.

S3 Tables is purpose-built Apache Iceberg storage: tables are
first-class AWS resources (with ARNs and policies) that query
engines address through the catalog integration, and AWS maintains
them continuously (compaction, snapshot expiry, unreferenced-file
removal) per the maintenance dials below.

The bucket's name is metadata.name (3-63 lowercase letters,
digits, hyphens). Namespaces key their entries by name; tables key
theirs by name within a namespace. The table format argument is
module-pinned to ICEBERG - the provider's enum holds exactly that
one value.

## Example

```yaml
# Canonical AwsS3TableBucket example (hack/dev manifest and refgen
# Example source): an analytics table bucket with one namespace and a
# schema-bearing Iceberg table.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsS3TableBucket
metadata:
  name: analytics-lake
  id: analytics-lake
  org: test-org
  env: dev
spec:
  region: us-west-2
  forceDestroy: true
  namespaces:
    - name: analytics
      tables:
        - name: events
          icebergSchema:
            fields:
              - name: event_id
                type: string
                required: true
              - name: occurred_at
                type: timestamp
              - name: payload
                type: string
          compaction:
            targetFileSizeMb: 256
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.encryption` | `AwsS3TablesEncryption` |  |  |  |
| `spec.encryption.sseAlgorithm` | `string` |  |  |  |
| `spec.encryption.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.unreferencedFileRemoval` | `AwsS3TablesUnreferencedFileRemoval` |  |  |  |
| `spec.unreferencedFileRemoval.disabled` | `bool` |  |  |  |
| `spec.unreferencedFileRemoval.nonCurrentDays` | `int64` |  |  |  |
| `spec.unreferencedFileRemoval.unreferencedDays` | `int64` |  |  |  |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.resourcePolicy` | `string` |  |  |  |
| `spec.replication` | `AwsS3TablesReplication` |  |  |  |
| `spec.replication.role` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.replication.destinationTableBucketArns` | `[]string` | yes |  |  |
| `spec.namespaces` | `[]AwsS3TablesNamespace` |  |  |  |
| `spec.namespaces[].name` | `string` | yes |  |  |
| `spec.namespaces[].tables` | `[]AwsS3TablesTable` |  |  |  |
| `spec.namespaces[].tables[].name` | `string` | yes |  |  |
| `spec.namespaces[].tables[].icebergSchema` | `AwsS3TablesIcebergSchema` |  |  |  |
| `spec.namespaces[].tables[].icebergSchema.fields` | `[]AwsS3TablesSchemaField` | yes |  |  |
| `spec.namespaces[].tables[].icebergSchema.fields[].name` | `string` | yes |  |  |
| `spec.namespaces[].tables[].icebergSchema.fields[].type` | `string` | yes |  |  |
| `spec.namespaces[].tables[].icebergSchema.fields[].required` | `bool` |  |  |  |
| `spec.namespaces[].tables[].properties` | `map<string, string>` |  |  |  |
| `spec.namespaces[].tables[].encryption` | `AwsS3TablesEncryption` |  |  |  |
| `spec.namespaces[].tables[].encryption.sseAlgorithm` | `string` |  |  |  |
| `spec.namespaces[].tables[].encryption.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.namespaces[].tables[].compaction` | `AwsS3TablesCompaction` |  |  |  |
| `spec.namespaces[].tables[].compaction.disabled` | `bool` |  |  |  |
| `spec.namespaces[].tables[].compaction.targetFileSizeMb` | `int64` |  |  |  |
| `spec.namespaces[].tables[].snapshotManagement` | `AwsS3TablesSnapshotManagement` |  |  |  |
| `spec.namespaces[].tables[].snapshotManagement.disabled` | `bool` |  |  |  |
| `spec.namespaces[].tables[].snapshotManagement.maxSnapshotAgeHours` | `int64` |  |  |  |
| `spec.namespaces[].tables[].snapshotManagement.minSnapshotsToKeep` | `int64` |  |  |  |
| `spec.namespaces[].tables[].resourcePolicy` | `string` |  |  |  |
| `spec.namespaces[].tables[].replication` | `AwsS3TablesReplication` |  |  |  |
| `spec.namespaces[].tables[].replication.role` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.namespaces[].tables[].replication.destinationTableBucketArns` | `[]string` | yes |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the table bucket lives in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.encryption

`AwsS3TablesEncryption`

Encryption at rest for the bucket - the default new tables
inherit. Unset means S3-managed keys (SSE-S3).

- rule: kms_key_arn requires sse_algorithm aws:kms

### spec.encryption.sseAlgorithm

`string`

The algorithm: S3-managed keys (AES256) or KMS (aws:kms).

- rule: {"string":{"in":["AES256","aws:kms"]}}

### spec.encryption.kmsKeyArn

`string | valueFrom`

The KMS key for aws:kms. The S3 Tables maintenance service must
be granted use of it (kms:GenerateDataKey/Decrypt for
maintenance.s3tables.amazonaws.com) or compaction silently
stops. Reference an AwsKmsKey key_arn output or pass a literal
key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.unreferencedFileRemoval

`AwsS3TablesUnreferencedFileRemoval`

AWS's background cleanup of files no longer referenced by any
table snapshot. On by default at AWS; the dials tune how
aggressively orphans age out.

### spec.unreferencedFileRemoval.disabled

`bool`

Turn the cleanup off entirely (it is on at AWS by default).
Orphaned files then accumulate storage cost forever.

### spec.unreferencedFileRemoval.nonCurrentDays

`int64`

Days a file stays "noncurrent" before the cleanup marks it
unreferenced. 0 means AWS's default (10).

- rule: {"int64":{"gte":"0"}}

### spec.unreferencedFileRemoval.unreferencedDays

`int64`

Days an unreferenced file survives before deletion. 0 means
AWS's default (3).

- rule: {"int64":{"gte":"0"}}

### spec.forceDestroy

`bool`

Let destroy succeed on a non-empty bucket by deleting every
table and namespace first. Off by default: destroying data
should hurt. Config-only at AWS - imports never round-trip it.

### spec.resourcePolicy

`string`

The bucket's resource policy (JSON) - who can create/query
tables in it. Cross-account query access starts here.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.replication

`AwsS3TablesReplication`

Replicate the bucket's tables to other table buckets.

### spec.replication.role

`string | valueFrom` · required

The IAM role the replication service assumes. Reference an
AwsIamRole role_arn output or pass a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.replication.destinationTableBucketArns

`[]string` · required

The destination table buckets (1-5), by ARN.

- rule: {"repeated":{"minItems":"1","maxItems":"5","unique":true,"items":{"string":{"pattern":"^arn:aws[a-z-]*:s3tables:.+$"}}}}

### spec.namespaces

`[]AwsS3TablesNamespace`

The namespaces (logical databases) and their tables, keyed by
name.

- rule: table names must be unique within a namespace

### spec.namespaces[].name

`string` · required

The namespace name - what query engines address as the database.
Lowercase letters, digits, underscores; must start with a letter
or digit. Fixed for life.

- rule: {"string":{"minLen":"1","maxLen":"255","pattern":"^[a-z0-9][a-z0-9_]*$"}}

### spec.namespaces[].tables

`[]AwsS3TablesTable`

The namespace's tables, keyed by name.

### spec.namespaces[].tables[].name

`string` · required

The table name - what query engines address. Lowercase letters,
digits, underscores; must start with a letter or digit. Fixed
for life.

- rule: {"string":{"minLen":"1","maxLen":"255","pattern":"^[a-z0-9][a-z0-9_]*$"}}

### spec.namespaces[].tables[].icebergSchema

`AwsS3TablesIcebergSchema`

The table's Iceberg schema at creation - a birth certificate,
not a contract. CREATE-ONLY at AWS: the provider never reads it
back, and schema evolution happens through query engines (ALTER
TABLE). Both modules deliberately IGNORE post-create changes to
this field (editing it after the table exists is inert, never a
destructive replace - and imported tables plan zero-diff against
it). Leave it unset to create schema-less and define columns
through an engine.

- rule: schema field names must be unique

### spec.namespaces[].tables[].icebergSchema.fields

`[]AwsS3TablesSchemaField` · required

The columns (at least one).

- rule: {"repeated":{"minItems":"1"}}

### spec.namespaces[].tables[].icebergSchema.fields[].name

`string` · required

The column name.

- rule: {"string":{"minLen":"1"}}

### spec.namespaces[].tables[].icebergSchema.fields[].type

`string` · required

The Iceberg type: "int", "long", "float", "double", "string",
"boolean", "date", "time", "timestamp", "timestamptz", "uuid",
"binary", "decimal(p,s)", "fixed(n)", or a nested
struct/list/map expression.

- rule: {"string":{"minLen":"1"}}

### spec.namespaces[].tables[].icebergSchema.fields[].required

`bool`

Whether the column is NOT NULL (Iceberg "required"). Default:
nullable.

### spec.namespaces[].tables[].properties

`map<string, string>`

Iceberg table properties at creation ("format-version",
"write.parquet.compression-codec", ...). Create-only, like the
schema.

### spec.namespaces[].tables[].encryption

`AwsS3TablesEncryption`

Encryption for THIS table when it must differ from the bucket
default. Fixed for life.

- rule: kms_key_arn requires sse_algorithm aws:kms

### spec.namespaces[].tables[].encryption.sseAlgorithm

`string`

The algorithm: S3-managed keys (AES256) or KMS (aws:kms).

- rule: {"string":{"in":["AES256","aws:kms"]}}

### spec.namespaces[].tables[].encryption.kmsKeyArn

`string | valueFrom`

The KMS key for aws:kms. The S3 Tables maintenance service must
be granted use of it (kms:GenerateDataKey/Decrypt for
maintenance.s3tables.amazonaws.com) or compaction silently
stops. Reference an AwsKmsKey key_arn output or pass a literal
key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.namespaces[].tables[].compaction

`AwsS3TablesCompaction`

AWS's automatic file compaction for this table. On at AWS by
default.

### spec.namespaces[].tables[].compaction.disabled

`bool`

Turn compaction off (it is on at AWS by default). Small-file
buildup then degrades query performance until an engine compacts
manually.

### spec.namespaces[].tables[].compaction.targetFileSizeMb

`int64`

The target compacted file size in MiB. 0 means AWS's default
(512).

- rule: {"int64":{"gte":"0"}}

### spec.namespaces[].tables[].snapshotManagement

`AwsS3TablesSnapshotManagement`

AWS's automatic snapshot expiry for this table. On at AWS by
default.

### spec.namespaces[].tables[].snapshotManagement.disabled

`bool`

Turn snapshot expiry off (it is on at AWS by default). Snapshots
(and the files they pin) then accumulate forever.

### spec.namespaces[].tables[].snapshotManagement.maxSnapshotAgeHours

`int64`

Hours a snapshot stays eligible for time travel before expiry.
0 means AWS's default (120).

- rule: {"int64":{"gte":"0"}}

### spec.namespaces[].tables[].snapshotManagement.minSnapshotsToKeep

`int64`

The minimum number of snapshots always kept regardless of age.
0 means AWS's default (1).

- rule: {"int64":{"gte":"0"}}

### spec.namespaces[].tables[].resourcePolicy

`string`

The table's own resource policy (JSON) - per-table access grants
on top of the bucket policy.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.namespaces[].tables[].replication

`AwsS3TablesReplication`

Replicate THIS table to destination table buckets (table-level
replication, independent of the bucket-level rule).

### spec.namespaces[].tables[].replication.role

`string | valueFrom` · required

The IAM role the replication service assumes. Reference an
AwsIamRole role_arn output or pass a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.namespaces[].tables[].replication.destinationTableBucketArns

`[]string` · required

The destination table buckets (1-5), by ARN.

- rule: {"repeated":{"minItems":"1","maxItems":"5","unique":true,"items":{"string":{"pattern":"^arn:aws[a-z-]*:s3tables:.+$"}}}}

## Validation Rules

- `spec.namespace_names_unique`: namespace names must be unique

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsS3TableBucket, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.table_bucket_arn` | `string` | The table bucket's ARN - what catalog integrations, policies, and replication destinations reference, and the provider's import ID. |
| `status.outputs.owner_account_id` | `string` | The AWS account that owns the bucket. |
| `status.outputs.table_arns` | `map<string, string>` | Each table's ARN, keyed "namespace//table" (the module's instance key) - what per-table policies and table-level replication reference. |
| `status.outputs.table_warehouse_locations` | `map<string, string>` | Each table's warehouse location (s3://... metadata root), keyed "namespace//table" - what query engines configured manually point at. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.encryption.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.replication.role` | AwsIamRole | `status.outputs.role_arn` |
| `spec.namespaces[].tables[].encryption.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.namespaces[].tables[].replication.role` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
