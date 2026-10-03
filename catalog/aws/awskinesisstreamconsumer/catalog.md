# AWS Kinesis Stream Consumer

Registers an enhanced fan-out consumer for an Amazon Kinesis Data Stream, providing dedicated 2 MB/s read throughput per shard independent of all other consumers. The stream binding accepts ValueFromRef wiring to an AwsKinesisStream in the same InfraChart, and both the consumer name and the stream binding are immutable after creation — a rename or a re-point is a replacement.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kinesis Stream Consumer** -- a registered enhanced fan-out consumer on the specified stream, with dedicated 2 MB/s per shard via SubscribeToShard (HTTP/2 push delivery with ~70ms propagation delay)
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A Kinesis Data Stream** -- the consumer registers with an existing stream. Provide the stream ARN directly or reference an AwsKinesisStream Infra Component via ValueFromRef.
- **IAM permissions** -- the deploying role needs `kinesis:RegisterStreamConsumer` and `kinesis:DeregisterStreamConsumer` on the target stream.

## Deploy

### Console

Open the deployment store, find **AWS Kinesis Stream Consumer**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Basic Consumer** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsKinesisStreamConsumer
metadata:
  name: analytics-consumer
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  streamArn:
    value: "arn:aws:kinesis:us-east-1:123456789012:stream/order-events"
```

```shell
planton apply -f kinesis-consumer.yaml
```

This registers an enhanced fan-out consumer with the specified Kinesis stream. The consumer gets dedicated 2 MB/s read throughput per shard. No additional configuration is needed -- AWS manages all internals. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the consumer to a Kinesis stream deployed in the same InfraPipeline:

```yaml
spec:
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: order-events
      fieldPath: status.outputs.stream_arn
```

The InfraPipeline resolves the dependency graph, deploys the Kinesis stream first, then registers the consumer with the resolved stream ARN.

## Key Configuration

These are the most important decisions when configuring a Kinesis stream consumer. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Consumer name** -- Derived from `metadata.name`. The name is immutable after creation -- renaming requires replacing the consumer. Choose a descriptive name that identifies the consuming application (e.g., `analytics-consumer`, `audit-trail-consumer`).

**Stream binding** -- The `streamArn` field is ForceNew. Changing the stream forces consumer replacement (deregister + re-register). A consumer can only be registered with one stream at a time. Up to 20 consumers can be registered per stream (soft limit).

**Cross-account access** -- The optional `resourcePolicy` field attaches a resource-based IAM policy to the consumer, keyed by the consumer ARN. Its primary use is granting another account's principals `kinesis:SubscribeToShard` and `kinesis:DescribeStreamConsumer` on this consumer without role assumption. Most consumers need no policy -- same-account readers authorize through their own IAM identities. Stream-level grants (PutRecord, GetRecords) belong on the stream's own `resourcePolicy`, not here.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKinesisStream** | `streamArn` | `status.outputs.stream_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `consumer_arn` | ARN of the registered consumer | Lambda event source mappings with enhanced fan-out, IAM policies |
| `consumer_name` | Consumer name (matches metadata.name) | Human-readable identification, ListStreamConsumers filtering |
| `stream_arn` | ARN of the parent Kinesis Data Stream | Downstream resource stream discovery without separate lookup |
| `creation_timestamp` | RFC3339 timestamp of consumer registration | Operational visibility, debugging registration order |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Basic consumer** -- Register an enhanced fan-out consumer with an existing stream using a direct ARN. Suitable for standalone consumer registration and quick prototyping. Start from the **Basic Consumer** preset.

**Stream reference consumer** -- Register a consumer using a ValueFromRef to a Planton-managed Kinesis stream. The platform resolves the stream ARN at deployment time and creates a dependency edge in the InfraPipeline DAG. Start from the **Stream Reference (valueFrom)** preset.

## Works With

- [**AWS Kinesis Data Stream**](/infra-catalog/aws-kinesis-stream) -- provides the parent stream this consumer registers with for dedicated throughput