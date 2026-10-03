# AWS Kinesis Data Stream

Deploys a Kinesis Data Stream in either On-Demand or Provisioned capacity mode, with configurable data retention, KMS encryption, enhanced shard-level CloudWatch metrics, and consumer deletion behavior. The capacity mode is the load-bearing choice — it sets the pricing model and scaling behavior — and the KMS key accepts ValueFromRef wiring for composing with an AwsKmsKey in the same InfraChart. The stream name derives from `metadata.name` and cannot change after creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kinesis Data Stream** -- a real-time data stream named from your manifest's `metadata.name`, configured with the specified capacity mode (On-Demand or Provisioned)
- **Stream Mode Configuration** -- On-Demand mode auto-scales shards to match throughput (up to 200 MB/s write); Provisioned mode creates the exact number of shards specified in `shardCount`
- **KMS Encryption** -- configured only when `kmsKeyId` is provided; encrypts data records at rest using the specified KMS key. When absent, encryption is disabled
- **Enhanced Shard-Level Metrics** -- enabled only when `shardLevelMetrics` entries are provided; publishes per-shard CloudWatch metrics for monitoring individual shard performance
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A KMS key** (optional) -- required when encrypting data at rest. Accepts a key ID, key ARN, or alias (e.g., `alias/aws/kinesis` for the AWS-managed Kinesis key). Provide the value directly or reference an AwsKmsKey Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS Kinesis Data Stream**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **On-Demand Minimal** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsKinesisStream
metadata:
  name: clickstream-events
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  streamMode: ON_DEMAND
```

```shell
planton apply -f kinesis-stream.yaml
```

This creates an On-Demand stream with AWS-managed shard scaling, default 24-hour retention, and no encryption. An Infra Job tracks the provisioning in real time.

### InfraChart

When using KMS encryption, use ValueFromRef to wire the stream to a KMS key deployed in the same InfraPipeline:

```yaml
spec:
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: streaming-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the KMS key first, then provisions the Kinesis stream with KMS encryption using the resolved key ARN.

## Key Configuration

These are the most important decisions when configuring a Kinesis Data Stream. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Capacity mode** -- Set `streamMode` to `ON_DEMAND` for automatic shard management (up to 200 MB/s write, 400 MB/s read) with pay-per-GB pricing. Set to `PROVISIONED` with a specific `shardCount` for predictable throughput (1 MB/s write and 2 MB/s read per shard) with pay-per-shard-hour pricing. This is a required field with no default.

**Warm throughput** -- On-Demand streams scale reactively; set `warmThroughputMibPs` to keep write capacity for a known burst (a launch, a scheduled batch) pre-provisioned so scaling lag never throttles it. Billed per MiB/s-hour on top of on-demand charges. Mutually exclusive with `shardCount` -- warm throughput applies to On-Demand streams only.

**Data retention** -- Set `retentionPeriodHours` between 24 and 8760 (1 day to 365 days), or leave it unset for the AWS default of 24 hours. Extended retention enables reprocessing scenarios and late-arriving consumers but incurs additional cost.

**Record size** -- Set `maxRecordSizeInKib` between 1024 and 10240 (1 MiB to 10 MiB), or leave it unset for the AWS default of 1 MiB. Larger records suit aggregated events and rich payloads but consume proportionally more shard capacity per write.

**Encryption** -- Provide `kmsKeyId` to encrypt data records at rest. Use `alias/aws/kinesis` for the AWS-managed key (no additional cost beyond KMS API calls) or a customer-managed key ARN for key rotation control and CloudTrail audit trails.

**Enhanced monitoring** -- Configure `shardLevelMetrics` with specific metric names (e.g., `IteratorAgeMilliseconds`, `WriteProvisionedThroughputExceeded`) to identify hot shards and capacity bottlenecks, or the `ALL` shorthand for every metric. Enhanced metrics incur additional CloudWatch cost per metric per shard.

**Cross-account access** -- The optional `resourcePolicy` field attaches a resource-based IAM policy keyed by the stream ARN, granting another account's principals PutRecord/GetRecords without role assumption. Consumer-scoped grants (SubscribeToShard) belong on the AwsKinesisStreamConsumer's own `resourcePolicy`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `stream_arn` | Amazon Resource Name of the stream | Firehose source configuration, Lambda event source mappings, IAM policies |
| `stream_name` | Kinesis stream name | Application SDK calls (PutRecord, GetRecords), CloudWatch alarm dimensions |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**On-Demand minimal** -- Auto-scaling capacity with no encryption and default retention. The simplest starting point for variable or unpredictable workloads. Start from the **On-Demand Minimal** preset.

**Provisioned with encryption** -- Fixed shard count with 48-hour retention and AWS-managed Kinesis key encryption. Suitable for steady workloads where throughput is predictable and cost optimization matters. Start from the **Provisioned Encrypted** preset.

**Production analytics** -- On-Demand mode with 7-day retention, KMS encryption, and all shard-level metrics enabled. Designed for production analytics pipelines where reprocessing, compliance, and per-shard observability are required. Start from the **Production Analytics** preset.

## Works With

- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for data-at-rest encryption
- [**AWS Kinesis Stream Consumer**](/infra-catalog/aws-kinesis-stream-consumer) -- registers an enhanced fan-out consumer with a dedicated 2 MB/s read pipe per shard on this stream