# AWS Kinesis Data Stream

Deploys an Amazon Kinesis Data Stream with configurable capacity mode (provisioned or on-demand), optional KMS encryption, extended data retention, and shard-level CloudWatch monitoring. The stream name is derived from `metadata.name` and cannot be changed after creation.

## What Gets Created

When you deploy an AwsKinesisStream resource, Planton provisions:

- **Kinesis Data Stream** — an `aws_kinesis_stream` resource with the specified capacity mode, shard count (provisioned only), retention period, and encryption settings
- **KMS Encryption** — enabled automatically when `kmsKeyId` is specified, using encryption type `KMS`
- **Enhanced Shard-Level Metrics** — per-shard CloudWatch metrics enabled when `shardLevelMetrics` entries are provided

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **A KMS key** if enabling server-side encryption (can use the Kinesis-owned key `alias/aws/kinesis` at no additional KMS cost)

## Quick Start

Create a file `kinesis-stream.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisStream
metadata:
  name: my-stream
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsKinesisStream.my-stream
spec:
  region: us-east-1
  streamMode: ON_DEMAND
```

Deploy:

```shell
planton apply -f kinesis-stream.yaml
```

This creates an on-demand Kinesis stream with 24-hour default retention, no encryption, and stream-level metrics only.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the Kinesis stream will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `streamMode` | `string` | Capacity mode: `"PROVISIONED"` or `"ON_DEMAND"`. Provisioned requires explicit shard count; on-demand auto-scales up to 200 MB/s write. | Must be one of the two valid values |
| `shardCount` | `int` | Number of shards. Each shard provides 1 MB/s write and 2 MB/s read. | Required when `streamMode` is `"PROVISIONED"` (>= 1). Must be 0 when `"ON_DEMAND"`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `retentionPeriodHours` | `int` | `24` | Hours that data remains accessible. Range: 24–8760 (1 day to 365 days). Extended retention incurs additional cost. |
| `kmsKeyId` | `string` | — | KMS key for server-side encryption. Accepts a key ID, key ARN, or alias (e.g., `alias/aws/kinesis`). Can reference an AwsKmsKey resource via `valueFrom`. |
| `maxRecordSizeInKib` | `int` | `1024` | Maximum single record size in KiB. Range: 1024–10240 (1–10 MiB). Note: deferred in current IaC modules pending provider upgrade. |
| `shardLevelMetrics` | `string[]` | `[]` | Shard-level CloudWatch metrics to enable. Valid values: `IncomingBytes`, `IncomingRecords`, `OutgoingBytes`, `OutgoingRecords`, `WriteProvisionedThroughputExceeded`, `ReadProvisionedThroughputExceeded`, `IteratorAgeMilliseconds`. |
| `enforceConsumerDeletion` | `bool` | `false` | When `true`, deregisters all enhanced fan-out consumers before deleting the stream. |

## Examples

### Provisioned Stream with Encryption

A provisioned stream with 4 shards, KMS encryption via the Kinesis-owned key, and 48-hour retention:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisStream
metadata:
  name: order-events
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.AwsKinesisStream.order-events
spec:
  region: us-east-1
  streamMode: PROVISIONED
  shardCount: 4
  retentionPeriodHours: 48
  kmsKeyId:
    value: alias/aws/kinesis
```

### On-Demand with Enhanced Monitoring

Production on-demand stream with 7-day retention, encryption, and all shard-level metrics for full observability:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisStream
metadata:
  name: analytics-pipeline
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsKinesisStream.analytics-pipeline
spec:
  region: us-east-1
  streamMode: ON_DEMAND
  retentionPeriodHours: 168
  kmsKeyId:
    value: arn:aws:kms:us-east-1:123456789012:key/mrk-abc123
  shardLevelMetrics:
    - IncomingBytes
    - IncomingRecords
    - OutgoingBytes
    - OutgoingRecords
    - WriteProvisionedThroughputExceeded
    - ReadProvisionedThroughputExceeded
    - IteratorAgeMilliseconds
  enforceConsumerDeletion: true
```

### Using Foreign Key References

Reference an Planton-managed KMS key instead of hardcoding the ARN:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisStream
metadata:
  name: audit-stream
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsKinesisStream.audit-stream
spec:
  region: us-east-1
  streamMode: ON_DEMAND
  retentionPeriodHours: 8760
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-encryption-key
      fieldPath: status.outputs.key_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `stream_arn` | `string` | ARN of the Kinesis stream, used for IAM policies, Firehose source configuration, and Lambda event source mappings |
| `stream_name` | `string` | Name of the Kinesis stream, used for Kinesis API calls (PutRecord, GetRecords) |

## Related Components

- [AwsKmsKey](/docs/catalog/aws/kms-key) — provides the encryption key referenced by `kmsKeyId`
- [AwsCloudwatchAlarm](/docs/catalog/aws/cloudwatch-alarm) — monitor `IteratorAgeMilliseconds` for consumer lag alerts
- [AwsLambda](/docs/catalog/aws/lambda) — commonly configured with Kinesis as an event source
- [AwsIamRole](/docs/catalog/aws/iam-role) — provides read/write permissions for producers and consumers
