# AWS EventBridge Pipe

Deploys an EventBridge Pipe — the managed point-to-point integration that reads from one source, optionally filters and enriches each event, and delivers to one target, replacing the "Lambda that moves messages from A to B". Sources span SQS, Kinesis, DynamoDB streams, MSK, self-managed Kafka, ActiveMQ, and RabbitMQ; targets span sixteen-plus services from Lambda and Step Functions to ECS RunTask, event buses, and API destinations. Pipes bill per event processed, so an idle pipe on a quiet source costs nothing at rest.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EventBridge Pipe** — named after `metadata.name`, wired from the source ARN to the target ARN under the execution role, consuming immediately unless `desiredState` says otherwise
- **Source tuning** — the family block matching the source's service: batching and windows for SQS, starting positions and retry/bisect/DLQ policy for Kinesis and DynamoDB streams, topics and Secrets Manager credential references for Kafka, queue names for MQ brokers
- **Event filter** — configured only when `sourceParameters.filterCriteria` is set: up to 5 OR-ed EventBridge patterns; only matching events reach the enrichment and target
- **Enrichment step** — configured only when `enrichment` is set: a Lambda function, Step Functions express state machine, or API destination that transforms each batch before delivery
- **Target shaping** — the family block matching the target's service (FIFO ids for SQS, partition keys for Kinesis, RunTask shape for ECS, SubmitJob shape for Batch, and so on), plus optional input templates before enrichment and delivery
- **Execution logging** — configured only when `logConfiguration` is set: a level (ERROR, INFO, TRACE) and one or more destinations (CloudWatch Logs, Firehose, S3)

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with EventBridge Pipes permissions and `iam:PassRole` on the execution role. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **The source and target resources** — the pipe wires existing ARNs; deploy them first (or in the same InfraChart).
- **An execution role trusting `pipes.amazonaws.com`** — with permissions to read the source, call the enrichment, and write the target. Goes in `roleArn`.
- **For Kafka and MQ sources** — broker credentials stored in Secrets Manager; the spec takes the secret's ARN, never the credential value.
- **For self-managed Kafka on private networks** — subnets and security groups for the pipe's Kafka client (`sourceParameters.selfManagedKafka.vpc`).

## Deploy

### Console

Open the deployment store, find **AWS EventBridge Pipe**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, then the pipe's shape: source and its family tuning, the optional filter and enrichment, the target and its shaping, and the execution role. Start from the **Queue-to-Queue with a Filter** preset in the [Presets](#presets) tab for the classic SQS-to-SQS shape, or the **DynamoDB Stream to Lambda** preset for change-data capture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgePipe
metadata:
  name: orders-pipe
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  description: Move order events from intake to processing
  source:
    valueFrom:
      kind: AwsSqsQueue
      name: orders-intake
      fieldPath: status.outputs.queue_arn
  sourceParameters:
    filterCriteria:
      filters:
        - pattern: '{"body":{"type":["order"]}}'
    sqs:
      batchSize: 10
      maximumBatchingWindowInSeconds: 5
  target:
    valueFrom:
      kind: AwsSqsQueue
      name: orders-processing
      fieldPath: status.outputs.queue_arn
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: orders-pipe-exec
      fieldPath: status.outputs.role_arn
```

```shell
planton apply -f eventbridge-pipe.yaml
```

This creates a running pipe that drains order-typed messages from the intake queue into the processing queue in batches of 10, under the referenced execution role. An Infra Job tracks the provisioning in real time.

### InfraChart

When the pipe deploys alongside its queues and role in one chart, wire all three edges via ValueFromRef — and note that `source`, `target`, and `enrichment` carry no default kind, so each valueFrom states its kind explicitly:

```yaml
spec:
  region: us-east-1
  source:
    valueFrom:
      kind: AwsSqsQueue
      name: orders-intake
      fieldPath: status.outputs.queue_arn
  target:
    valueFrom:
      kind: AwsLambda
      name: orders-processor
      fieldPath: status.outputs.function_arn
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: orders-pipe-exec
      fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the queue, function, and role first, then wires the pipe across them.

## Key Configuration

These are the most important decisions when configuring a pipe. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The source is a marriage, the target is a job** — Changing `source` (or its family's position fields: `startingPosition`, `topicName`, `queueName`) replaces the whole pipe and resets consumer positions; changing `target` swaps in place. Design pipes around their source and move targets freely. Stream starting positions (`TRIM_HORIZON`, `LATEST`, `AT_TIMESTAMP`) are likewise fixed for life — pick them deliberately.

**Match the family block to the ARN's service** — A `kinesis` block on an SQS source fails at AWS, not at validation, because the ARN may be an unresolved reference at validation time. The spec's CELs stop double-blocks; keeping the family aligned with the source and target services is yours.

**desiredState is the money lever** — Pipes bill per event processed, and `desiredState: STOPPED` halts consumption without losing the pipe or its stream positions. Pause through incident triage or cost freezes instead of deleting. Creates and state flips ride a provisioning state machine — minutes, not seconds; the provider waits up to 30.

**Filters are the cheapest optimization** — Only events matching `filterCriteria` reach the enrichment and target, and only matching events bill. A tight pattern (up to 5, OR-ed) is both a correctness and a cost control.

**Failure policy lives on stream sources** — For Kinesis and DynamoDB streams, `maximumRetryAttempts`, `maximumRecordAgeInSeconds`, `onPartialBatchItemFailure: AUTOMATIC_BISECT`, and `deadLetterQueueArn` decide whether a poison record blocks the shard or lands in a DLQ. The AWS defaults retry until records expire — set a DLQ before the first bad record, not after.

**Credentials never enter the manifest** — Kafka and MQ sources authenticate through Secrets Manager secret ARNs; the spec's patterns reject anything that is not a `secretsmanager` ARN. Rotate the secret in Secrets Manager and the pipe follows without a deploy.

**TRACE logging can leak payloads** — `logConfiguration.level: TRACE` with `includeExecutionData: true` writes event payloads into the log destination. Enable it deliberately, and point it at a destination with matching access controls.

## Outputs and Dependencies

### What This Kind Consumes

`source`, `target`, and `enrichment` range over many services, so they carry no default kind — a valueFrom on them states its kind explicitly. The common wirings:

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsSqsQueue** | `source` / `target` | `status.outputs.queue_arn` |
| **AwsLambda** | `target` / `enrichment` | `status.outputs.function_arn` |
| **AwsDynamodb** | `source` | `status.outputs.stream_arn` |
| **AwsKinesisStream** | `source` / `target` | `status.outputs.stream_arn` |
| **AwsSqsQueue** | `sourceParameters.kinesis.deadLetterQueueArn` | `status.outputs.queue_arn` |
| **AwsKmsKey** | `kmsKeyIdentifier` | `status.outputs.key_arn` |
| **AwsCloudwatchLogGroup** | `logConfiguration.cloudwatchLogs.logGroupArn` | `status.outputs.log_group_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `pipe_arn` | The pipe's ARN | IAM policies scoping who may start, stop, or describe the pipe |

`pipe_name` is also echoed back (it equals `metadata.name`) — it is the provider's import ID, not a composition input.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Queue-to-queue with a filter** — Drain one SQS queue into another, keeping only matching events and reshaping each message with an input template. No polling Lambda to own, and non-matching events never bill. Start from the **Queue-to-Queue with a Filter** preset.

**Change-data capture from DynamoDB** — Read the table's stream from `LATEST`, bisect failing batches automatically, dead-letter what still fails, and invoke the consumer function synchronously so failures retry. The starting position is fixed for life: `TRIM_HORIZON` replays retained history, `LATEST` starts now. Start from the **DynamoDB Stream to Lambda** preset.

**Enrich before delivery** — Insert a Lambda, express state machine, or API destination between source and target to transform each batch in flight. The trade against enriching in the consumer: the enrichment runs per batch on the pipe's bill and its failures follow the source's retry policy, but every downstream target sees the enriched shape.

**Stream into an event bus** — Point a Kinesis or DynamoDB stream at an EventBridge bus target with `eventbridgeEventBus.detailType` and `source` set, fanning point-to-point traffic out to rule-based routing when one consumer becomes many.

## Works With

- [**AWS SQS Queue**](/infra-catalog/aws-sqs-queue) — the classic source and target, and the DLQ for stream sources
- [**AWS Lambda**](/infra-catalog/aws-lambda) — target or enrichment step, invoked per batch
- [**AWS DynamoDB**](/infra-catalog/aws-dynamodb) — its stream feeds the pipe for change-data capture
- [**AWS Kinesis Data Stream**](/infra-catalog/aws-kinesis-stream) — high-throughput source or target with partition-key shaping
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) — the pipes-trusting execution role wired via `roleArn`
- [**AWS EventBridge Bus**](/infra-catalog/aws-event-bridge-bus) — target that fans pipe traffic out to rule-based routing
- [**AWS EventBridge API Destination**](/infra-catalog/aws-event-bridge-api-destination) — authenticated HTTP target or enrichment endpoint
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption of pipe data at rest via `kmsKeyIdentifier`
