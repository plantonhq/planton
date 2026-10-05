# AWS Lambda Event Source Mapping

Creates the managed poller that reads records from an event source — an SQS queue, a Kinesis stream, a DynamoDB stream, an MSK topic, a self-managed Kafka cluster, an Amazon MQ broker, or a DocumentDB change stream — batches them, and invokes a Lambda function. The mapping is its own node in the resource graph, not a setting of the function: a function may have many mappings, a mapping can be repointed to a different function in place, and pausing consumption, tuning batching, and re-driving failures all happen here without touching the function or the source.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Event Source Mapping** -- the AWS-managed poller (a server-assigned UUID) connecting the source to the function, with the batching, filtering, error-handling, and scaling settings in the spec

Everything else is referenced, never modified: the Lambda function, the queue/stream/cluster being consumed, the KMS key encrypting filter criteria, and the failure destination.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A Lambda function** -- reference an AwsLambda Infra Component or pass a literal function ARN. No invoke permission is needed: the poller invokes through the function's execution role.
- **The execution role can READ the source** -- the function's execution role needs the source-side permissions (`sqs:ReceiveMessage`/`DeleteMessage`/`GetQueueAttributes` for SQS, the Kinesis/DynamoDB stream read actions, or the MSK/Kafka cluster access) — the mapping never grants them.
- **The event source** -- the queue, stream, cluster, broker, or DocumentDB cluster to consume. Create-time immutable: a different source is a different mapping.

## Deploy

### Console

Open the deployment store, find **AWS Lambda Event Source Mapping**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields — the source type you pick shapes every later step. Start from the **SQS Queue Worker** or **Kinesis Stream Consumer** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLambdaEventSourceMapping
metadata:
  name: orders-queue-consumer
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  functionArn:
    valueFrom:
      kind: AwsLambda
      name: orders-worker
      fieldPath: status.outputs.function_arn
  eventSourceArn:
    valueFrom:
      kind: AwsSqsQueue
      name: orders-queue
      fieldPath: status.outputs.queue_arn
  batchSize: 25
  maximumBatchingWindowSeconds: 5
  functionResponseTypes:
    - ReportBatchItemFailures
```

```shell
planton apply -f mapping.yaml
```

This wires the queue into the function with partial-batch failure reporting. An Infra Job tracks the provisioning and streams progress in real time.

### InfraChart

The mapping is the edge that completes the serverless story in a multi-resource environment — deploy the queue, the function, and the mapping in one InfraPipeline:

```yaml
spec:
  functionArn:
    valueFrom:
      kind: AwsLambda
      name: orders-worker
      fieldPath: status.outputs.function_arn
  eventSourceArn:
    valueFrom:
      kind: AwsKinesisStream
      name: clickstream
      fieldPath: status.outputs.stream_arn
  startingPosition: TRIM_HORIZON
```

The InfraPipeline resolves the dependency graph, deploys the function and the source first, then wires them together.

## Key Configuration

These are the most important decisions when configuring an event source mapping. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The source type shapes everything** -- exactly one of `eventSourceArn` (SQS, Kinesis, DynamoDB Streams, MSK, MQ, DocumentDB) or `selfManagedKafka`. Stream and Kafka sources require `startingPosition` (`TRIM_HORIZON` to process the backlog, `LATEST` for only new records); Kafka sources require `topics`; MQ requires `mqQueue`; SQS takes none of these.

**Batching is the latency-vs-cost dial** -- the defaults (10 records for SQS, 100 for streams, no batching window) optimize latency. For throughput and cost, raise `batchSize` together with `maximumBatchingWindowSeconds` so the poller can actually fill the larger batches.

**Report partial failures** -- `functionResponseTypes: [ReportBatchItemFailures]` retries only the failed records of a batch instead of the whole batch — the right setting for almost every SQS and stream consumer (the code must return the `batchItemFailures` shape).

**Filter before you pay** -- `filters` discard non-matching records BEFORE invocation, so function time is never billed for events the code would ignore. Up to 10 EventBridge-style patterns, OR-ed together.

**Stream error handling** -- `bisectBatchOnFunctionError` isolates poison records by splitting failing batches; `maximumRetryAttempts` and `maximumRecordAgeSeconds` (both with `-1` = the never-give-up AWS default) bound retries; `onFailureDestinationArn` receives discarded batches' metadata.

**Pause without deleting** -- `disabled: true` stops consumption while retaining the tracked position; re-enabling resumes where it stopped.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsLambda** | `functionArn` | `status.outputs.function_arn` |
| **AwsSqsQueue** | `eventSourceArn` (default kind) | `status.outputs.queue_arn` |
| **AwsKinesisStream** | `eventSourceArn` (explicit kind) | `status.outputs.stream_arn` |
| **AwsDynamodb** | `eventSourceArn` (explicit kind) | `status.outputs.stream_arn` |
| **AwsMskCluster** | `eventSourceArn` (explicit kind) | `status.outputs.cluster_arn` |
| **AwsKmsKey** (optional) | `kmsKeyArn` | `status.outputs.key_arn` |
| **AwsSqsQueue** (optional) | `onFailureDestinationArn` | `status.outputs.queue_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `uuid` | The server-assigned mapping identifier | CLI operations, AWS API calls |
| `mapping_arn` | The mapping ARN | IAM policies, audit |
| `function_arn` | The function the mapping invokes (as resolved by AWS) | Verification, audit |
| `state` | The mapping state (Enabled, Disabled, ...) | Operational dashboards |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**SQS worker** -- the everyday queue consumer: partial-batch failure reporting on, batching tuned for throughput, and optionally a per-mapping concurrency cap (`scalingMaxConcurrency`) to protect a downstream database. Start from the **SQS Queue Worker** preset.

**Kinesis consumer** -- `TRIM_HORIZON` to process the backlog, `parallelizationFactor` to multiply per-shard throughput while preserving per-key ordering, bisect-on-error plus a failure destination for poison records. Start from the **Kinesis Stream Consumer** preset.

**Kafka consumer** -- MSK by reference (or self-managed brokers with source-access credentials from Secrets Manager), topics and an optional pinned consumer group, provisioned pollers when throughput must be predictable -- and a `pollerGroupName` to share one provisioned fleet across many mappings (the cost lever for many low-traffic topics). Start from the **MSK Consumer with Shared Provisioned Pollers** preset.

## Works With

- [**AWS Lambda**](/infra-catalog/aws-lambda) -- the function every mapping invokes
- [**AWS SQS Queue**](/infra-catalog/aws-sqs-queue) -- the most common source, and the usual failure destination
- [**AWS Kinesis Data Stream**](/infra-catalog/aws-kinesis-stream) -- ordered, replayable stream consumption
- [**AWS DynamoDB**](/infra-catalog/aws-dynamodb) -- change-data capture from table streams
- [**AWS MSK Cluster**](/infra-catalog/aws-msk-cluster) -- managed Kafka topic consumption
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- customer-managed encryption for filter criteria
