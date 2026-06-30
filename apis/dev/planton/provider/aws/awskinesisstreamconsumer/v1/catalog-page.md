# AWS Kinesis Stream Consumer

Registers an Amazon Kinesis enhanced fan-out consumer for a Kinesis Data Stream, providing dedicated 2 MB/s read throughput per shard independent of all other consumers. The consumer name is derived from `metadata.name` and both the name and stream binding are immutable after creation.

## What Gets Created

When you deploy an AwsKinesisStreamConsumer resource, Planton provisions:

- **Kinesis Stream Consumer** — an `aws_kinesis_stream_consumer` resource registered with the specified stream, providing dedicated 2 MB/s per shard via SubscribeToShard (HTTP/2 push delivery)

No additional sub-resources are created. This is a single-resource component.

## Prerequisites

- **AWS credentials** configured via environment variables or Planton provider config
- **An existing Kinesis Data Stream** — the consumer registers with a stream. The stream can be managed by Planton (AwsKinesisStream) or pre-existing.
- **IAM permissions** — `kinesis:RegisterStreamConsumer` and `kinesis:DeregisterStreamConsumer`

## Quick Start

Create a file `kinesis-stream-consumer.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: my-consumer
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.AwsKinesisStreamConsumer.my-consumer
spec:
  region: us-east-1
  streamArn:
    value: arn:aws:kinesis:us-east-1:123456789012:stream/my-stream
```

Deploy:

```shell
planton apply -f kinesis-stream-consumer.yaml
```

This registers an enhanced fan-out consumer with the specified stream. The consumer gets dedicated 2 MB/s read throughput per shard.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the consumer will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |
| `streamArn` | `StringValueOrRef` | ARN of the parent Kinesis Data Stream to register with. ForceNew — changing the stream forces consumer replacement. | Required. Accepts literal ARN or `valueFrom` reference to AwsKinesisStream. |

### Optional Fields

This component has no optional fields. The consumer name is derived from `metadata.name`. Tags are derived from `metadata.labels`.

## Examples

### Consumer with Direct Stream ARN

Register a consumer with an existing stream using a literal ARN:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: dashboard-consumer
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsKinesisStreamConsumer.dashboard-consumer
spec:
  region: us-east-1
  streamArn:
    value: arn:aws:kinesis:us-east-1:123456789012:stream/order-events
```

### Consumer with Stream Reference

Reference an Planton-managed Kinesis stream. The platform resolves the ARN at deployment time:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: analytics-consumer
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.AwsKinesisStreamConsumer.analytics-consumer
spec:
  region: us-east-1
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: clickstream-events
      fieldPath: status.outputs.stream_arn
```

### Multiple Consumers on the Same Stream

Deploy multiple consumers for different processing pipelines, each with dedicated throughput:

```yaml
# Consumer 1: Real-time dashboard
apiVersion: aws.planton.dev/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: dashboard-consumer
spec:
  region: us-east-1
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: order-events
      fieldPath: status.outputs.stream_arn
---
# Consumer 2: Audit trail
apiVersion: aws.planton.dev/v1
kind: AwsKinesisStreamConsumer
metadata:
  name: audit-consumer
spec:
  region: us-east-1
  streamArn:
    valueFrom:
      kind: AwsKinesisStream
      name: order-events
      fieldPath: status.outputs.stream_arn
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `consumer_arn` | `string` | ARN of the registered consumer, used for Lambda event source mappings with enhanced fan-out and IAM policies |
| `consumer_name` | `string` | Name of the consumer (matches `metadata.name`) |
| `stream_arn` | `string` | ARN of the parent Kinesis Data Stream (echoed for convenience) |
| `creation_timestamp` | `string` | RFC3339 timestamp of consumer registration |

## Related Components

- [AwsKinesisStream](/docs/catalog/aws/kinesis-stream) — the parent stream this consumer registers with
- [AwsLambda](/docs/catalog/aws/lambda) — commonly configured with the `consumer_arn` output for enhanced fan-out event source mappings
- [AwsIamRole](/docs/catalog/aws/iam-role) — provides `kinesis:SubscribeToShard` permissions for consumer applications
