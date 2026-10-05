# AwsLambdaEventSourceMapping

Lambda event source mapping resource for Planton. Provisions the managed poller that reads records from an SQS queue, a Kinesis stream, a DynamoDB stream, an MSK topic, a self-managed Kafka cluster, an Amazon MQ broker, or a DocumentDB change stream, batches them, and invokes a Lambda function.

## When to use

- You want a queue or stream to drive Lambda invocations without managing polling yourself.
- You need independent control over consumption (pause, batching, filters, failure routing) separate from the function or the source.
- You are wiring the messaging graph into compute: the mapping is its own node, not a setting on the function.

## Prerequisites

| Prerequisite | Why | Planton Resource |
|---|---|---|
| Lambda function | The mapping invokes this function | `AwsLambda` |
| Event source | The queue, stream, cluster, or broker the poller reads from | `AwsSqsQueue`, `AwsKinesisStream`, `AwsDynamodb`, `AwsMskCluster`, etc. |

## API envelope

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLambdaEventSourceMapping
metadata:
  name: <resource-id>
spec: { ... }
```

## Spec fields reference

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `region` | string | **yes** | — | AWS region (must match function and source). |
| `functionArn` | StringValueOrRef | **yes** | — | Lambda function to invoke. Default ref: `AwsLambda` / `function_arn`. |
| `eventSourceArn` | StringValueOrRef | one of | — | AWS event source ARN. Default ref: `AwsSqsQueue` / `queue_arn`. **ForceNew**. |
| `selfManagedKafka` | object | one of | — | Non-MSK Kafka bootstrap servers. **ForceNew**. |
| `disabled` | bool | no | `false` | Pause consumption without deleting the mapping. |
| `batchSize` | int32 | no | source default | Records per invocation batch. |
| `maximumBatchingWindowSeconds` | int32 | no | `0` | Seconds to gather records before invoking (0–300). |
| `filters` | list | no | — | EventBridge-style patterns (max 10). |
| `functionResponseTypes` | list | no | — | `ReportBatchItemFailures` for partial-batch retry. |
| `startingPosition` | string | stream/Kafka | — | `TRIM_HORIZON`, `LATEST`, or `AT_TIMESTAMP`. **ForceNew**. |
| `scalingMaxConcurrency` | int32 | no | AWS default | SQS-only per-mapping concurrency throttle (minimum 2; ceiling is the function's own concurrency). |

See `spec.proto` for the full surface including Kafka schema registry, provisioned pollers, MQ, and DocumentDB options.

## Outputs

| Output | Description |
|---|---|
| `uuid` | Server-assigned mapping UUID -- AWS API identity. |
| `mapping_arn` | The mapping ARN. |
| `function_arn` | Function ARN as resolved by AWS. |
| `state` | Mapping state at deploy time (e.g. `Enabled`, `Disabled`). |

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLambdaEventSourceMapping
metadata:
  name: order-worker
spec:
  region: us-west-2
  functionArn:
    valueFrom:
      kind: AwsLambda
      name: order-processor
      fieldPath: status.outputs.function_arn
  eventSourceArn:
    valueFrom:
      kind: AwsSqsQueue
      name: order-events
      fieldPath: status.outputs.queue_arn
  functionResponseTypes:
    - ReportBatchItemFailures
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
