# AwsKinesisStream

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsKinesisStreamSpec defines the desired configuration for an Amazon Kinesis Data Stream.

Kinesis Data Streams is a real-time data streaming service that captures gigabytes of
data per second from hundreds of thousands of sources. Common sources include website
click-streams, database event streams, financial transactions, social media feeds, IT
logs, and location-tracking events.

Two capacity modes are available:
- PROVISIONED: You specify the number of shards. Each shard provides 1 MB/s write and
  2 MB/s read capacity. You pay per shard-hour. Use this when you can predict throughput.
- ON_DEMAND: AWS automatically manages shards to accommodate up to 200 MB/s write and
  400 MB/s read. You pay per GB of data written and read. Use this when throughput is
  unpredictable or bursty.

Notes:
- The stream name (from metadata.name) cannot be changed after creation (ForceNew).
- Encryption at rest is supported via a customer-managed KMS key. When kms_key_id is
  set, the IaC modules automatically enable KMS encryption; when absent, encryption is
  disabled (NONE). You can also use the Kinesis-owned key by passing "alias/aws/kinesis".
- Enhanced shard-level CloudWatch metrics are available for production observability.
- Credentials, region, and deployment workflow live outside this spec in stack inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsKinesisStream
metadata:
  name: test-stream
spec:
  region: us-west-2
  streamMode: ON_DEMAND
  retentionPeriodHours: 48
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.streamMode` | `string` |  |  |  |
| `spec.shardCount` | `int32` |  |  |  |
| `spec.retentionPeriodHours` | `int32` |  |  |  |
| `spec.kmsKeyId` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.maxRecordSizeInKib` | `int32` |  |  |  |
| `spec.shardLevelMetrics` | `[]string` |  |  |  |
| `spec.enforceConsumerDeletion` | `bool` |  |  |  |
| `spec.warmThroughputMibPs` | `int32` |  |  |  |
| `spec.resourcePolicy` | `object` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.streamMode

`string`

Capacity mode for the stream. This is a fundamental design choice that
determines pricing model, scaling behavior, and operational overhead.

Valid values:
- "PROVISIONED": You manage shard count explicitly. Each shard provides
  1 MB/s ingestion and 2 MB/s consumption. Cost-effective for steady,
  predictable workloads.
- "ON_DEMAND": AWS auto-scales shards to match throughput. Supports up to
  200 MB/s write and 400 MB/s read. Best for variable or unpredictable
  workloads. No capacity planning required.

This field is required. There is no default -- you must make an explicit
choice because the two modes have fundamentally different cost and
operational characteristics.

### spec.shardCount

`int32`

Number of open shards in the stream. Each shard provides 1 MB/s write
(1,000 records/s) and 2 MB/s read capacity.

Required when stream_mode is "PROVISIONED" (must be >= 1).
Must be 0 (or omitted) when stream_mode is "ON_DEMAND" because AWS
manages shard count automatically.

Can be updated after creation to scale provisioned streams. AWS uses
uniform scaling (UpdateShardCount with UNIFORM_SCALING strategy).

### spec.retentionPeriodHours

`int32`

Duration in hours that data records remain accessible after being added to
the stream. After the retention period expires, records are no longer
accessible via GetRecords.

Range: 24–8760 (1 day to 365 days). AWS default: 24 hours.
Leave at 0 to use the AWS default.

Increasing retention is useful for reprocessing scenarios and late-arriving
consumers. Note: extended retention (beyond 24h) incurs additional cost.

### spec.kmsKeyId

`string | valueFrom`

Customer-managed KMS key for server-side encryption of data at rest. When
set, the IaC modules automatically configure KMS encryption (encryption_type
= "KMS"). When absent, encryption is disabled (encryption_type = "NONE").

Accepts a KMS key ID, key ARN, alias name (e.g., "alias/aws/kinesis"), or
alias ARN. Also accepts a reference to an AwsKmsKey resource.

Encryption can be enabled or disabled after stream creation (not ForceNew).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.maxRecordSizeInKib

`int32`

Maximum size of a single data record in KiB (kibibytes). Records exceeding
this limit are rejected by the PutRecord/PutRecords API.

Range: 1024–10240 (1 MiB to 10 MiB). AWS default: 1024 (1 MiB).
Leave at 0 to use the AWS default.

Larger record sizes are useful for aggregated events, rich JSON payloads,
or binary data. Note: larger records consume more shard capacity.

### spec.shardLevelMetrics

`[]string`

Shard-level CloudWatch metrics to enable. By default, Kinesis only provides
stream-level metrics. Enabling shard-level metrics allows monitoring of
individual shard performance, which is critical for identifying hot shards
and capacity bottlenecks in production.

Valid values (one or more):
- "IncomingBytes" -- bytes written per shard
- "IncomingRecords" -- records written per shard
- "OutgoingBytes" -- bytes read per shard
- "OutgoingRecords" -- records read per shard
- "WriteProvisionedThroughputExceeded" -- throttled write requests
- "ReadProvisionedThroughputExceeded" -- throttled read requests
- "IteratorAgeMilliseconds" -- consumer lag per shard
- "ALL" -- shorthand that enables every shard-level metric above

Enhanced metrics incur additional CloudWatch cost per metric per shard.
Leave empty to use stream-level metrics only (no additional cost).

### spec.enforceConsumerDeletion

`bool`

When true, all registered enhanced fan-out consumers are automatically
deregistered before the stream is deleted, preventing deletion errors.
When false (default), deleting a stream with active consumers will fail.

This is an operational setting that only affects stream deletion. It has
no impact on the running stream.

### spec.warmThroughputMibPs

`int32`

Pre-provisioned warm write throughput in MiB/s for ON_DEMAND streams.
On-demand streams normally scale up in response to observed traffic;
warm throughput keeps capacity for a known burst level ready in advance,
so a sudden spike (a product launch, a scheduled batch) is absorbed
without ProvisionedThroughputExceeded throttling while scaling catches
up. Billed per MiB/s-hour on top of on-demand data charges.

Mutually exclusive with shard_count (warm throughput is meaningless on
PROVISIONED streams, where capacity IS the shard count). Leave at 0 to
let on-demand scaling manage capacity reactively.

### spec.resourcePolicy

`object`

Resource-based access policy for the stream, as a standard IAM policy
document. The primary use is cross-account access: granting another
account's principals PutRecord/GetRecords on this stream without role
assumption. AWS models this as a separate resource-policy API keyed by
the stream ARN; it is folded here because the policy has no identity of
its own and follows the stream's lifecycle.

## Validation Rules

- `stream_mode_required`: stream_mode is required and must be 'PROVISIONED' or 'ON_DEMAND'
- `shard_count_required_for_provisioned`: shard_count must be at least 1 when stream_mode is 'PROVISIONED'
- `shard_count_forbidden_for_on_demand`: shard_count must be 0 when stream_mode is 'ON_DEMAND' (AWS manages shards automatically)
- `retention_period_range`: retention_period_hours must be between 24 and 8760 (1 day to 365 days) when set
- `max_record_size_range`: max_record_size_in_kib must be between 1024 and 10240 (1 MiB to 10 MiB) when set
- `shard_level_metrics_valid`: each shard_level_metrics value must be one of: IncomingBytes, IncomingRecords, OutgoingBytes, OutgoingRecords, WriteProvisionedThroughputExceeded, ReadProvisionedThroughputExceeded, IteratorAgeMilliseconds, ALL
- `warm_throughput_conflicts_with_shard_count`: warm_throughput_mib_ps cannot be set together with shard_count (warm throughput applies to ON_DEMAND streams only)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsKinesisStream, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.stream_arn` | `string` | The Amazon Resource Name (ARN) of the Kinesis stream. This is the primary identifier used for IAM policies, cross-service permissions, and as a source/target reference in other resources (e.g., Firehose source, Lambda event source mapping, EventBridge target). |
| `status.outputs.stream_name` | `string` | The name of the Kinesis stream. Used for Kinesis API calls (PutRecord, GetRecords, etc.) and for human-readable identification. The stream name is unique within an AWS account and region. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockAgentCoreMemory | `spec.kinesisDelivery.dataStreamArn` | `status.outputs.stream_arn` |
| AwsCloudwatchLogDelivery | `spec.crossAccountDestination.targetArn` | `status.outputs.stream_arn` |
| AwsDynamodb | `spec.kinesisStreamingDestination.streamArn` | `status.outputs.stream_arn` |
| AwsEventBridgeScheduler | `spec.target.arn` | `status.outputs.stream_arn` |
| AwsKinesisFirehose | `spec.kinesisStreamSource.streamArn` | `status.outputs.stream_arn` |
| AwsKinesisStreamConsumer | `spec.streamArn` | `status.outputs.stream_arn` |

## See Also

- [Overview](../README.md)
