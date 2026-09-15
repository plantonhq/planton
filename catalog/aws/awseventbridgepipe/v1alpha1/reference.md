# AwsEventBridgePipe

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

AwsEventBridgePipeSpec defines one EventBridge Pipe: a managed
point-to-point integration that reads from ONE source, optionally
filters and enriches events in flight, and delivers to ONE target -
no event bus in between.

The pipe's name in AWS is metadata.name. The SOURCE is fixed for
life (changing it replaces the pipe, resetting consumer positions);
the TARGET swaps in place. The execution role must allow the pipe to
read the source and write the target - its trust policy allows
pipes.amazonaws.com.

Source and target ARNs range over many services (SQS, Kinesis,
DynamoDB streams, Kafka, MQ on the source side; 16+ services on the
target side), so no single kind dominates - references on those
fields carry NO default kind, and a valueFrom in a manifest must
state its kind explicitly.

Pipes bill per event processed - an idle pipe on an empty source
costs nothing at rest. desired_state STOPPED pauses consumption
without losing the pipe (creates and state flips can take minutes;
the provider waits up to 30).

## Example

```yaml
# Canonical AwsEventBridgePipe example (hack/dev manifest and refgen
# Example source): an SQS-to-SQS pipe with an order filter, batch
# tuning, and an input template.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEventBridgePipe
metadata:
  name: orders-pipe
  id: orders-pipe
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Move order events from intake to processing
  source:
    value: arn:aws:sqs:us-west-2:123456789012:orders-intake
  sourceParameters:
    filterCriteria:
      filters:
        - pattern: '{"body":{"type":["order"]}}'
    sqs:
      batchSize: 10
      maximumBatchingWindowInSeconds: 5
  target:
    value: arn:aws:sqs:us-west-2:123456789012:orders-processing
  targetParameters:
    inputTemplate: '{"orderId": <$.body.orderId>, "receivedAt": <aws.pipes.event.ingestion-time>}'
  roleArn:
    value: arn:aws:iam::123456789012:role/pipe-exec
  desiredState: RUNNING
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.source` | `string \| valueFrom` | yes |  |  |
| `spec.sourceParameters` | `AwsEventBridgePipeSourceParameters` |  |  |  |
| `spec.sourceParameters.filterCriteria` | `AwsEventBridgePipeFilterCriteria` |  |  |  |
| `spec.sourceParameters.filterCriteria.filters` | `[]AwsEventBridgePipeFilter` | yes |  |  |
| `spec.sourceParameters.filterCriteria.filters[].pattern` | `string` | yes |  |  |
| `spec.sourceParameters.sqs` | `AwsEventBridgePipeSqsSourceParameters` |  |  |  |
| `spec.sourceParameters.sqs.batchSize` | `int32` |  |  |  |
| `spec.sourceParameters.sqs.maximumBatchingWindowInSeconds` | `int32` |  |  |  |
| `spec.sourceParameters.kinesis` | `AwsEventBridgePipeKinesisSourceParameters` |  |  |  |
| `spec.sourceParameters.kinesis.startingPosition` | `string` |  |  |  |
| `spec.sourceParameters.kinesis.startingPositionTimestamp` | `string` |  |  |  |
| `spec.sourceParameters.kinesis.batchSize` | `int32` |  |  |  |
| `spec.sourceParameters.kinesis.maximumBatchingWindowInSeconds` | `int32` |  |  |  |
| `spec.sourceParameters.kinesis.maximumRecordAgeInSeconds` | `int32` |  |  |  |
| `spec.sourceParameters.kinesis.maximumRetryAttempts` | `int32` |  |  |  |
| `spec.sourceParameters.kinesis.onPartialBatchItemFailure` | `string` |  |  |  |
| `spec.sourceParameters.kinesis.parallelizationFactor` | `int32` |  |  |  |
| `spec.sourceParameters.kinesis.deadLetterQueueArn` | `string \| valueFrom` |  |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.sourceParameters.dynamodb` | `AwsEventBridgePipeDynamoDbSourceParameters` |  |  |  |
| `spec.sourceParameters.dynamodb.startingPosition` | `string` |  |  |  |
| `spec.sourceParameters.dynamodb.batchSize` | `int32` |  |  |  |
| `spec.sourceParameters.dynamodb.maximumBatchingWindowInSeconds` | `int32` |  |  |  |
| `spec.sourceParameters.dynamodb.maximumRecordAgeInSeconds` | `int32` |  |  |  |
| `spec.sourceParameters.dynamodb.maximumRetryAttempts` | `int32` |  |  |  |
| `spec.sourceParameters.dynamodb.onPartialBatchItemFailure` | `string` |  |  |  |
| `spec.sourceParameters.dynamodb.parallelizationFactor` | `int32` |  |  |  |
| `spec.sourceParameters.dynamodb.deadLetterQueueArn` | `string \| valueFrom` |  |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.sourceParameters.msk` | `AwsEventBridgePipeMskSourceParameters` |  |  |  |
| `spec.sourceParameters.msk.topicName` | `string` | yes |  |  |
| `spec.sourceParameters.msk.consumerGroupId` | `string` |  |  |  |
| `spec.sourceParameters.msk.startingPosition` | `string` |  |  |  |
| `spec.sourceParameters.msk.batchSize` | `int32` |  |  |  |
| `spec.sourceParameters.msk.maximumBatchingWindowInSeconds` | `int32` |  |  |  |
| `spec.sourceParameters.msk.credentials` | `AwsEventBridgePipeMskCredentials` |  |  |  |
| `spec.sourceParameters.msk.credentials.clientCertificateTlsAuth` | `string` |  |  |  |
| `spec.sourceParameters.msk.credentials.saslScram512Auth` | `string` |  |  |  |
| `spec.sourceParameters.selfManagedKafka` | `AwsEventBridgePipeSelfManagedKafkaSourceParameters` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.topicName` | `string` | yes |  |  |
| `spec.sourceParameters.selfManagedKafka.additionalBootstrapServers` | `[]string` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.consumerGroupId` | `string` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.startingPosition` | `string` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.batchSize` | `int32` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.maximumBatchingWindowInSeconds` | `int32` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.credentials` | `AwsEventBridgePipeSelfManagedKafkaCredentials` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.credentials.basicAuth` | `string` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.credentials.clientCertificateTlsAuth` | `string` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.credentials.saslScram256Auth` | `string` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.credentials.saslScram512Auth` | `string` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.serverRootCaCertificate` | `string` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.vpc` | `AwsEventBridgePipeSelfManagedKafkaVpc` |  |  |  |
| `spec.sourceParameters.selfManagedKafka.vpc.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.sourceParameters.selfManagedKafka.vpc.securityGroups` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.sourceParameters.activemq` | `AwsEventBridgePipeActiveMqSourceParameters` |  |  |  |
| `spec.sourceParameters.activemq.queueName` | `string` | yes |  |  |
| `spec.sourceParameters.activemq.basicAuthCredentials` | `string` |  |  |  |
| `spec.sourceParameters.activemq.batchSize` | `int32` |  |  |  |
| `spec.sourceParameters.activemq.maximumBatchingWindowInSeconds` | `int32` |  |  |  |
| `spec.sourceParameters.rabbitmq` | `AwsEventBridgePipeRabbitMqSourceParameters` |  |  |  |
| `spec.sourceParameters.rabbitmq.queueName` | `string` | yes |  |  |
| `spec.sourceParameters.rabbitmq.virtualHost` | `string` |  |  |  |
| `spec.sourceParameters.rabbitmq.basicAuthCredentials` | `string` |  |  |  |
| `spec.sourceParameters.rabbitmq.batchSize` | `int32` |  |  |  |
| `spec.sourceParameters.rabbitmq.maximumBatchingWindowInSeconds` | `int32` |  |  |  |
| `spec.enrichment` | `string \| valueFrom` |  |  |  |
| `spec.enrichmentParameters` | `AwsEventBridgePipeEnrichmentParameters` |  |  |  |
| `spec.enrichmentParameters.inputTemplate` | `string` |  |  |  |
| `spec.enrichmentParameters.httpParameters` | `AwsEventBridgePipeHttpParameters` |  |  |  |
| `spec.enrichmentParameters.httpParameters.headerParameters` | `map<string, string>` |  |  |  |
| `spec.enrichmentParameters.httpParameters.pathParameterValue` | `string` |  |  |  |
| `spec.enrichmentParameters.httpParameters.queryStringParameters` | `map<string, string>` |  |  |  |
| `spec.target` | `string \| valueFrom` | yes |  |  |
| `spec.targetParameters` | `AwsEventBridgePipeTargetParameters` |  |  |  |
| `spec.targetParameters.inputTemplate` | `string` |  |  |  |
| `spec.targetParameters.sqs` | `AwsEventBridgePipeSqsTargetParameters` |  |  |  |
| `spec.targetParameters.sqs.messageGroupId` | `string` |  |  |  |
| `spec.targetParameters.sqs.messageDeduplicationId` | `string` |  |  |  |
| `spec.targetParameters.kinesis` | `AwsEventBridgePipeKinesisTargetParameters` |  |  |  |
| `spec.targetParameters.kinesis.partitionKey` | `string` | yes |  |  |
| `spec.targetParameters.lambda` | `AwsEventBridgePipeLambdaTargetParameters` |  |  |  |
| `spec.targetParameters.lambda.invocationType` | `string` |  |  |  |
| `spec.targetParameters.stepFunction` | `AwsEventBridgePipeStepFunctionTargetParameters` |  |  |  |
| `spec.targetParameters.stepFunction.invocationType` | `string` |  |  |  |
| `spec.targetParameters.ecsTask` | `AwsEventBridgePipeEcsTaskTargetParameters` |  |  |  |
| `spec.targetParameters.ecsTask.taskDefinitionArn` | `string \| valueFrom` | yes |  | AwsEcsTaskDefinition (`status.outputs.task_definition_arn`) |
| `spec.targetParameters.ecsTask.taskCount` | `int32` |  |  |  |
| `spec.targetParameters.ecsTask.launchType` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.capacityProviderStrategy` | `[]AwsEventBridgePipeCapacityProviderStrategy` |  |  |  |
| `spec.targetParameters.ecsTask.capacityProviderStrategy[].capacityProvider` | `string` | yes |  |  |
| `spec.targetParameters.ecsTask.capacityProviderStrategy[].base` | `int32` |  |  |  |
| `spec.targetParameters.ecsTask.capacityProviderStrategy[].weight` | `int32` |  |  |  |
| `spec.targetParameters.ecsTask.networkConfiguration` | `AwsEventBridgePipeEcsNetworkConfiguration` |  |  |  |
| `spec.targetParameters.ecsTask.networkConfiguration.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.targetParameters.ecsTask.networkConfiguration.securityGroups` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.targetParameters.ecsTask.networkConfiguration.assignPublicIp` | `bool` |  |  |  |
| `spec.targetParameters.ecsTask.group` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.platformVersion` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides` | `AwsEventBridgePipeEcsTaskOverrides` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides` | `[]AwsEventBridgePipeEcsContainerOverride` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].name` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].command` | `[]string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].cpu` | `int32` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].memory` | `int32` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].memoryReservation` | `int32` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].environment` | `[]AwsEventBridgePipeEcsEnvironmentVariable` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].environment[].name` | `string` | yes |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].environment[].value` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].environmentFiles` | `[]AwsEventBridgePipeEcsEnvironmentFile` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].environmentFiles[].type` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].environmentFiles[].value` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].resourceRequirements` | `[]AwsEventBridgePipeEcsResourceRequirement` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].resourceRequirements[].type` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.containerOverrides[].resourceRequirements[].value` | `string` | yes |  |  |
| `spec.targetParameters.ecsTask.overrides.cpu` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.memory` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.ephemeralStorageSizeInGib` | `int32` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.executionRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.targetParameters.ecsTask.overrides.taskRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.targetParameters.ecsTask.overrides.inferenceAcceleratorOverrides` | `[]AwsEventBridgePipeEcsInferenceAcceleratorOverride` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.inferenceAcceleratorOverrides[].deviceName` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.overrides.inferenceAcceleratorOverrides[].deviceType` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.placementConstraints` | `[]AwsEventBridgePipePlacementConstraint` |  |  |  |
| `spec.targetParameters.ecsTask.placementConstraints[].type` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.placementConstraints[].expression` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.placementStrategy` | `[]AwsEventBridgePipePlacementStrategy` |  |  |  |
| `spec.targetParameters.ecsTask.placementStrategy[].type` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.placementStrategy[].field` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.propagateTags` | `string` |  |  |  |
| `spec.targetParameters.ecsTask.tags` | `map<string, string>` |  |  |  |
| `spec.targetParameters.ecsTask.enableEcsManagedTags` | `bool` |  |  |  |
| `spec.targetParameters.ecsTask.enableExecuteCommand` | `bool` |  |  |  |
| `spec.targetParameters.ecsTask.referenceId` | `string` |  |  |  |
| `spec.targetParameters.batchJob` | `AwsEventBridgePipeBatchJobTargetParameters` |  |  |  |
| `spec.targetParameters.batchJob.jobDefinition` | `string` | yes |  |  |
| `spec.targetParameters.batchJob.jobName` | `string` | yes |  |  |
| `spec.targetParameters.batchJob.arraySize` | `int32` |  |  |  |
| `spec.targetParameters.batchJob.retryAttempts` | `int32` |  |  |  |
| `spec.targetParameters.batchJob.parameters` | `map<string, string>` |  |  |  |
| `spec.targetParameters.batchJob.dependsOn` | `[]AwsEventBridgePipeBatchJobDependency` |  |  |  |
| `spec.targetParameters.batchJob.dependsOn[].jobId` | `string` |  |  |  |
| `spec.targetParameters.batchJob.dependsOn[].type` | `string` |  |  |  |
| `spec.targetParameters.batchJob.containerOverrides` | `AwsEventBridgePipeBatchContainerOverrides` |  |  |  |
| `spec.targetParameters.batchJob.containerOverrides.command` | `[]string` |  |  |  |
| `spec.targetParameters.batchJob.containerOverrides.environment` | `[]AwsEventBridgePipeEcsEnvironmentVariable` |  |  |  |
| `spec.targetParameters.batchJob.containerOverrides.environment[].name` | `string` | yes |  |  |
| `spec.targetParameters.batchJob.containerOverrides.environment[].value` | `string` |  |  |  |
| `spec.targetParameters.batchJob.containerOverrides.instanceType` | `string` |  |  |  |
| `spec.targetParameters.batchJob.containerOverrides.resourceRequirements` | `[]AwsEventBridgePipeBatchResourceRequirement` |  |  |  |
| `spec.targetParameters.batchJob.containerOverrides.resourceRequirements[].type` | `string` |  |  |  |
| `spec.targetParameters.batchJob.containerOverrides.resourceRequirements[].value` | `string` | yes |  |  |
| `spec.targetParameters.redshiftData` | `AwsEventBridgePipeRedshiftDataTargetParameters` |  |  |  |
| `spec.targetParameters.redshiftData.database` | `string` | yes |  |  |
| `spec.targetParameters.redshiftData.sqls` | `[]string` | yes |  |  |
| `spec.targetParameters.redshiftData.dbUser` | `string` |  |  |  |
| `spec.targetParameters.redshiftData.secretManagerArn` | `string` |  |  |  |
| `spec.targetParameters.redshiftData.statementName` | `string` |  |  |  |
| `spec.targetParameters.redshiftData.withEvent` | `bool` |  |  |  |
| `spec.targetParameters.sagemakerPipeline` | `AwsEventBridgePipeSageMakerPipelineTargetParameters` |  |  |  |
| `spec.targetParameters.sagemakerPipeline.pipelineParameters` | `[]AwsEventBridgePipeSageMakerPipelineParameter` |  |  |  |
| `spec.targetParameters.sagemakerPipeline.pipelineParameters[].name` | `string` | yes |  |  |
| `spec.targetParameters.sagemakerPipeline.pipelineParameters[].value` | `string` | yes |  |  |
| `spec.targetParameters.eventbridgeEventBus` | `AwsEventBridgePipeEventBusTargetParameters` |  |  |  |
| `spec.targetParameters.eventbridgeEventBus.detailType` | `string` |  |  |  |
| `spec.targetParameters.eventbridgeEventBus.source` | `string` |  |  |  |
| `spec.targetParameters.eventbridgeEventBus.endpointId` | `string` |  |  |  |
| `spec.targetParameters.eventbridgeEventBus.resources` | `[]string` |  |  |  |
| `spec.targetParameters.eventbridgeEventBus.time` | `string` |  |  |  |
| `spec.targetParameters.cloudwatchLogs` | `AwsEventBridgePipeCloudWatchLogsTargetParameters` |  |  |  |
| `spec.targetParameters.cloudwatchLogs.logStreamName` | `string` |  |  |  |
| `spec.targetParameters.cloudwatchLogs.timestamp` | `string` |  |  |  |
| `spec.targetParameters.http` | `AwsEventBridgePipeHttpParameters` |  |  |  |
| `spec.targetParameters.http.headerParameters` | `map<string, string>` |  |  |  |
| `spec.targetParameters.http.pathParameterValue` | `string` |  |  |  |
| `spec.targetParameters.http.queryStringParameters` | `map<string, string>` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.desiredState` | `string` |  |  |  |
| `spec.kmsKeyIdentifier` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.logConfiguration` | `AwsEventBridgePipeLogConfiguration` |  |  |  |
| `spec.logConfiguration.level` | `string` |  |  |  |
| `spec.logConfiguration.includeExecutionData` | `bool` |  |  |  |
| `spec.logConfiguration.cloudwatchLogs` | `AwsEventBridgePipeCloudWatchLogsLogDestination` |  |  |  |
| `spec.logConfiguration.cloudwatchLogs.logGroupArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.logConfiguration.firehose` | `AwsEventBridgePipeFirehoseLogDestination` |  |  |  |
| `spec.logConfiguration.firehose.deliveryStreamArn` | `string \| valueFrom` | yes |  | AwsKinesisFirehose (`status.outputs.delivery_stream_arn`) |
| `spec.logConfiguration.s3` | `AwsEventBridgePipeS3LogDestination` |  |  |  |
| `spec.logConfiguration.s3.bucketName` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.logConfiguration.s3.bucketOwner` | `string` |  |  |  |
| `spec.logConfiguration.s3.outputFormat` | `string` |  |  |  |
| `spec.logConfiguration.s3.prefix` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the pipe lives in. Example: "us-east-1".

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

What this pipe moves and why.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512"}}

### spec.source

`string | valueFrom` · required

The source's ARN: an SQS queue, Kinesis stream, DynamoDB stream,
MSK topic, self-managed Kafka cluster, or ActiveMQ/RabbitMQ
broker. FIXED FOR LIFE - changing the source replaces the whole
pipe. References here carry no default kind: state the kind
explicitly in valueFrom.

Containment-exempt: a pipe is a managed integration that READS its
source, transforms, and WRITES its target; none of the three ends is
where the pipe lives. Whichever of them is a container (an ECS
cluster, a Redshift cluster, an event bus), the pipe stands between
its ends on a diagram, never inside one of them.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.sourceParameters

`AwsEventBridgePipeSourceParameters`

Source-family tuning (batching, positions, credentials) plus the
event filter. The family block must match the source's service;
AWS applies service defaults when unset.

- rule: configure at most one source family block (sqs, kinesis, dynamodb, msk, self_managed_kafka, activemq, rabbitmq) - it must match the source arn's service

### spec.sourceParameters.filterCriteria

`AwsEventBridgePipeFilterCriteria`

Event filter: only events matching at least one pattern reach the
enrichment/target (and only matching events bill). At most 5
patterns, each in EventBridge pattern syntax.

### spec.sourceParameters.filterCriteria.filters

`[]AwsEventBridgePipeFilter` · required

The filter patterns (OR-ed; at most 5), each an EventBridge event
pattern JSON document. Example: {"body":{"type":["order"]}}.

- rule: {"repeated":{"minItems":"1","maxItems":"5"}}

### spec.sourceParameters.filterCriteria.filters[].pattern

`string` · required

The EventBridge event pattern (JSON, up to 4096 characters).

- rule: {"string":{"minLen":"1","maxLen":"4096"}}

### spec.sourceParameters.sqs

`AwsEventBridgePipeSqsSourceParameters`

SQS source tuning.

### spec.sourceParameters.sqs.batchSize

`int32` · optional (explicit presence)

Events per batch (1-10000; FIFO queues cap at 10). Unset keeps
AWS's default (10).

- rule: {"int32":{"lte":10000,"gte":1}}

### spec.sourceParameters.sqs.maximumBatchingWindowInSeconds

`int32` · optional (explicit presence)

How long to gather a batch before invoking (0-300 seconds).

- rule: {"int32":{"lte":300,"gte":0}}

### spec.sourceParameters.kinesis

`AwsEventBridgePipeKinesisSourceParameters`

Kinesis stream source tuning.

- rule: starting_position AT_TIMESTAMP requires starting_position_timestamp; TRIM_HORIZON/LATEST forbid it

### spec.sourceParameters.kinesis.startingPosition

`string`

Where a NEW pipe starts reading. FIXED FOR LIFE - changing it
replaces the pipe.

- rule: {"string":{"in":["TRIM_HORIZON","LATEST","AT_TIMESTAMP"]}}

### spec.sourceParameters.kinesis.startingPositionTimestamp

`string`

The instant to start from (RFC3339), with AT_TIMESTAMP. Fixed for
life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.sourceParameters.kinesis.batchSize

`int32` · optional (explicit presence)

Events per batch (1-10000). Unset keeps AWS's default (100).

- rule: {"int32":{"lte":10000,"gte":1}}

### spec.sourceParameters.kinesis.maximumBatchingWindowInSeconds

`int32` · optional (explicit presence)

How long to gather a batch before invoking (0-300 seconds).

- rule: {"int32":{"lte":300,"gte":0}}

### spec.sourceParameters.kinesis.maximumRecordAgeInSeconds

`int32` · optional (explicit presence)

Discard records older than this (60-604800 seconds; -1 never
discards). Unset keeps AWS's default (-1).

- rule: {"int32":{"lte":604800,"gte":-1}}

### spec.sourceParameters.kinesis.maximumRetryAttempts

`int32` · optional (explicit presence)

Retry a failed batch at most this many times (-1 retries until
the records expire). Unset keeps AWS's default (-1).

- rule: {"int32":{"lte":10000,"gte":-1}}

### spec.sourceParameters.kinesis.onPartialBatchItemFailure

`string`

On a partial batch failure, AUTOMATIC_BISECT splits the batch and
retries the halves. Unset retries the whole batch.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AUTOMATIC_BISECT"]}}

### spec.sourceParameters.kinesis.parallelizationFactor

`int32` · optional (explicit presence)

Concurrent batches per shard (1-10). Unset keeps AWS's default
(1).

- rule: {"int32":{"lte":10,"gte":1}}

### spec.sourceParameters.kinesis.deadLetterQueueArn

`string | valueFrom`

Where records land after retries are exhausted (an SQS queue or
SNS topic ARN). Reference an AwsSqsQueue queue_arn output or pass
a literal ARN.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.sourceParameters.dynamodb

`AwsEventBridgePipeDynamoDbSourceParameters`

DynamoDB stream source tuning.

### spec.sourceParameters.dynamodb.startingPosition

`string`

Where a NEW pipe starts reading. FIXED FOR LIFE - changing it
replaces the pipe.

- rule: {"string":{"in":["TRIM_HORIZON","LATEST"]}}

### spec.sourceParameters.dynamodb.batchSize

`int32` · optional (explicit presence)

Events per batch (1-10000). Unset keeps AWS's default (100).

- rule: {"int32":{"lte":10000,"gte":1}}

### spec.sourceParameters.dynamodb.maximumBatchingWindowInSeconds

`int32` · optional (explicit presence)

How long to gather a batch before invoking (0-300 seconds).

- rule: {"int32":{"lte":300,"gte":0}}

### spec.sourceParameters.dynamodb.maximumRecordAgeInSeconds

`int32` · optional (explicit presence)

Discard records older than this (60-604800 seconds; -1 never
discards). Unset keeps AWS's default (-1).

- rule: {"int32":{"lte":604800,"gte":-1}}

### spec.sourceParameters.dynamodb.maximumRetryAttempts

`int32` · optional (explicit presence)

Retry a failed batch at most this many times (-1 retries until
the records expire). Unset keeps AWS's default (-1).

- rule: {"int32":{"lte":10000,"gte":-1}}

### spec.sourceParameters.dynamodb.onPartialBatchItemFailure

`string`

On a partial batch failure, AUTOMATIC_BISECT splits the batch and
retries the halves. Unset retries the whole batch.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AUTOMATIC_BISECT"]}}

### spec.sourceParameters.dynamodb.parallelizationFactor

`int32` · optional (explicit presence)

Concurrent batches per shard (1-10). Unset keeps AWS's default
(1).

- rule: {"int32":{"lte":10,"gte":1}}

### spec.sourceParameters.dynamodb.deadLetterQueueArn

`string | valueFrom`

Where records land after retries are exhausted (an SQS queue or
SNS topic ARN). Reference an AwsSqsQueue queue_arn output or pass
a literal ARN.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.sourceParameters.msk

`AwsEventBridgePipeMskSourceParameters`

Amazon MSK source tuning.

### spec.sourceParameters.msk.topicName

`string` · required

The topic to read (1-249 characters, Kafka topic charset). FIXED
FOR LIFE - changing it replaces the pipe.

- rule: {"string":{"minLen":"1","maxLen":"249","pattern":"^[^.]([0-9A-Za-z_.-]+)$"}}

### spec.sourceParameters.msk.consumerGroupId

`string`

The Kafka consumer group id (up to 200 characters). Unset lets
AWS generate one.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"200"}}

### spec.sourceParameters.msk.startingPosition

`string`

Where a NEW pipe starts reading. Fixed for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["TRIM_HORIZON","LATEST"]}}

### spec.sourceParameters.msk.batchSize

`int32` · optional (explicit presence)

Events per batch (1-10000). Unset keeps AWS's default (100).

- rule: {"int32":{"lte":10000,"gte":1}}

### spec.sourceParameters.msk.maximumBatchingWindowInSeconds

`int32` · optional (explicit presence)

How long to gather a batch before invoking (0-300 seconds).

- rule: {"int32":{"lte":300,"gte":0}}

### spec.sourceParameters.msk.credentials

`AwsEventBridgePipeMskCredentials`

Broker credentials, as the ARN of the Secrets Manager secret
holding them - a reference, never the credential value itself.

- rule: set exactly one of client_certificate_tls_auth and sasl_scram_512_auth

### spec.sourceParameters.msk.credentials.clientCertificateTlsAuth

`string`

Secrets Manager secret ARN holding the client TLS certificate.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws.*:secretsmanager:.*$"}}

### spec.sourceParameters.msk.credentials.saslScram512Auth

`string`

Secrets Manager secret ARN holding the SASL/SCRAM-512 credentials.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws.*:secretsmanager:.*$"}}

### spec.sourceParameters.selfManagedKafka

`AwsEventBridgePipeSelfManagedKafkaSourceParameters`

Self-managed Apache Kafka source tuning.

### spec.sourceParameters.selfManagedKafka.topicName

`string` · required

The topic to read (1-249 characters, Kafka topic charset). FIXED
FOR LIFE.

- rule: {"string":{"minLen":"1","maxLen":"249","pattern":"^[^.]([0-9A-Za-z_.-]+)$"}}

### spec.sourceParameters.selfManagedKafka.additionalBootstrapServers

`[]string`

Extra bootstrap "host:port" endpoints beyond the source ARN's (at
most 2). Fixed for life.

- rule: {"repeated":{"maxItems":"2","items":{"string":{"minLen":"1","maxLen":"300"}}}}

### spec.sourceParameters.selfManagedKafka.consumerGroupId

`string`

The Kafka consumer group id (up to 200 characters). FIXED FOR
LIFE on self-managed Kafka (unlike MSK).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"200"}}

### spec.sourceParameters.selfManagedKafka.startingPosition

`string`

Where a NEW pipe starts reading. Fixed for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["TRIM_HORIZON","LATEST"]}}

### spec.sourceParameters.selfManagedKafka.batchSize

`int32` · optional (explicit presence)

Events per batch (1-10000). Unset keeps AWS's default (100).

- rule: {"int32":{"lte":10000,"gte":1}}

### spec.sourceParameters.selfManagedKafka.maximumBatchingWindowInSeconds

`int32` · optional (explicit presence)

How long to gather a batch before invoking (0-300 seconds).

- rule: {"int32":{"lte":300,"gte":0}}

### spec.sourceParameters.selfManagedKafka.credentials

`AwsEventBridgePipeSelfManagedKafkaCredentials`

Broker credentials, as the ARN of the Secrets Manager secret
holding them - a reference, never the credential value itself.

- rule: set exactly one of basic_auth, client_certificate_tls_auth, sasl_scram_256_auth, and sasl_scram_512_auth

### spec.sourceParameters.selfManagedKafka.credentials.basicAuth

`string`

Secrets Manager secret ARN holding SASL/PLAIN credentials.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws.*:secretsmanager:.*$"}}

### spec.sourceParameters.selfManagedKafka.credentials.clientCertificateTlsAuth

`string`

Secrets Manager secret ARN holding the client TLS certificate.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws.*:secretsmanager:.*$"}}

### spec.sourceParameters.selfManagedKafka.credentials.saslScram256Auth

`string`

Secrets Manager secret ARN holding SASL/SCRAM-256 credentials.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws.*:secretsmanager:.*$"}}

### spec.sourceParameters.selfManagedKafka.credentials.saslScram512Auth

`string`

Secrets Manager secret ARN holding SASL/SCRAM-512 credentials.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws.*:secretsmanager:.*$"}}

### spec.sourceParameters.selfManagedKafka.serverRootCaCertificate

`string`

Secrets Manager secret ARN holding the brokers' root CA
certificate, for private CAs.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws.*$"}}

### spec.sourceParameters.selfManagedKafka.vpc

`AwsEventBridgePipeSelfManagedKafkaVpc`

VPC placement for reaching brokers on private networks.

### spec.sourceParameters.selfManagedKafka.vpc.subnets

`[]string | valueFrom` · required

The subnets the client uses (at most 16). Reference AwsSubnet
subnet_id outputs or pass literal subnet-... ids.

Containment-exempt: the pipe is a managed service that reaches INTO
the VPC to poll the brokers; it is not deployed into the subnets. On a
diagram the pipe stands outside the network with a line in -- the same
verdict a Lambda function's VPC subnets carry.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1","maxItems":"16"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.sourceParameters.selfManagedKafka.vpc.securityGroups

`[]string | valueFrom`

The security groups attached to the client (at most 5). Reference
AwsSecurityGroup security_group_id outputs or pass literal sg-...
ids.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.sourceParameters.activemq

`AwsEventBridgePipeActiveMqSourceParameters`

ActiveMQ broker source tuning.

### spec.sourceParameters.activemq.queueName

`string` · required

The queue to read (1-1000 characters). FIXED FOR LIFE.

- rule: {"string":{"minLen":"1","maxLen":"1000"}}

### spec.sourceParameters.activemq.basicAuthCredentials

`string`

Secrets Manager secret ARN holding the broker's basic-auth
credentials - a reference, never the credential value itself.
REQUIRED by AWS for ActiveMQ.

- rule: {"string":{"pattern":"^arn:aws.*:secretsmanager:.*$"}}

### spec.sourceParameters.activemq.batchSize

`int32` · optional (explicit presence)

Events per batch (1-10000). Unset keeps AWS's default (100).

- rule: {"int32":{"lte":10000,"gte":1}}

### spec.sourceParameters.activemq.maximumBatchingWindowInSeconds

`int32` · optional (explicit presence)

How long to gather a batch before invoking (0-300 seconds).

- rule: {"int32":{"lte":300,"gte":0}}

### spec.sourceParameters.rabbitmq

`AwsEventBridgePipeRabbitMqSourceParameters`

RabbitMQ broker source tuning.

### spec.sourceParameters.rabbitmq.queueName

`string` · required

The queue to read (1-1000 characters). FIXED FOR LIFE.

- rule: {"string":{"minLen":"1","maxLen":"1000"}}

### spec.sourceParameters.rabbitmq.virtualHost

`string`

The RabbitMQ virtual host (unset uses "/"). Fixed for life.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"200"}}

### spec.sourceParameters.rabbitmq.basicAuthCredentials

`string`

Secrets Manager secret ARN holding the broker's basic-auth
credentials - a reference, never the credential value itself.
REQUIRED by AWS for RabbitMQ.

- rule: {"string":{"pattern":"^arn:aws.*:secretsmanager:.*$"}}

### spec.sourceParameters.rabbitmq.batchSize

`int32` · optional (explicit presence)

Events per batch (1-10000). Unset keeps AWS's default (100).

- rule: {"int32":{"lte":10000,"gte":1}}

### spec.sourceParameters.rabbitmq.maximumBatchingWindowInSeconds

`int32` · optional (explicit presence)

How long to gather a batch before invoking (0-300 seconds).

- rule: {"int32":{"lte":300,"gte":0}}

### spec.enrichment

`string | valueFrom`

The enrichment step's ARN: a Lambda function, Step Functions
express state machine, or API destination that transforms each
batch before delivery. Unset skips enrichment. References here
carry no default kind: state the kind explicitly in valueFrom.
Access, not placement, like `source`.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.enrichmentParameters

`AwsEventBridgePipeEnrichmentParameters`

HTTP shaping (for API-destination enrichment) and the input
template applied before the enrichment call.

### spec.enrichmentParameters.inputTemplate

`string`

The input template applied to each event BEFORE the enrichment
call (up to 8192 characters; static text with <$.json.path>
placeholders).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"8192"}}

### spec.enrichmentParameters.httpParameters

`AwsEventBridgePipeHttpParameters`

HTTP shaping when the enrichment is an API destination.

### spec.enrichmentParameters.httpParameters.headerParameters

`map<string, string>`

Headers added to the request.

### spec.enrichmentParameters.httpParameters.pathParameterValue

`string`

The value filling the destination URL's "*" path wildcard. (AWS's
API models this as a list but accepts exactly ONE entry.)

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.enrichmentParameters.httpParameters.queryStringParameters

`map<string, string>`

Query-string parameters added to the request.

### spec.target

`string | valueFrom` · required

The target's ARN: an ECS cluster, Batch job queue, Lambda
function, Step Functions state machine, Kinesis stream, SQS queue,
Redshift cluster, SageMaker pipeline, CloudWatch log group,
EventBridge bus, or API destination. Swaps in place. References
here carry no default kind: state the kind explicitly in
valueFrom. Access, not placement, like `source`.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.targetParameters

`AwsEventBridgePipeTargetParameters`

Target-family invocation shaping plus the input template applied
before delivery. The family block must match the target's service.

- rule: configure at most one target family block (sqs, kinesis, lambda, step_function, ecs_task, batch_job, redshift_data, sagemaker_pipeline, eventbridge_event_bus, cloudwatch_logs, http) - it must match the target arn's service

### spec.targetParameters.inputTemplate

`string`

The input template applied to each event BEFORE delivery (up to
8192 characters; static text with <$.json.path> placeholders).
Removing it later genuinely clears it at AWS.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"8192"}}

### spec.targetParameters.sqs

`AwsEventBridgePipeSqsTargetParameters`

SQS SendMessage shaping (FIFO group/dedup ids).

### spec.targetParameters.sqs.messageGroupId

`string`

The FIFO message group id (up to 100 characters). FIFO queues
only.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"100"}}

### spec.targetParameters.sqs.messageDeduplicationId

`string`

The FIFO deduplication id (up to 100 characters; supports
<$.json.path> dynamic values). FIFO queues without
content-based deduplication only.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"100"}}

### spec.targetParameters.kinesis

`AwsEventBridgePipeKinesisTargetParameters`

Kinesis PutRecord shaping (partition key).

### spec.targetParameters.kinesis.partitionKey

`string` · required

The partition key (up to 256 characters; supports <$.json.path>
dynamic values).

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.targetParameters.lambda

`AwsEventBridgePipeLambdaTargetParameters`

Lambda invocation type.

### spec.targetParameters.lambda.invocationType

`string`

REQUEST_RESPONSE waits for the function result (failures retry);
FIRE_AND_FORGET does not.

- rule: {"string":{"in":["REQUEST_RESPONSE","FIRE_AND_FORGET"]}}

### spec.targetParameters.stepFunction

`AwsEventBridgePipeStepFunctionTargetParameters`

Step Functions invocation type.

### spec.targetParameters.stepFunction.invocationType

`string`

REQUEST_RESPONSE waits for the execution to finish (EXPRESS state
machines only); FIRE_AND_FORGET starts it and moves on.

- rule: {"string":{"in":["REQUEST_RESPONSE","FIRE_AND_FORGET"]}}

### spec.targetParameters.ecsTask

`AwsEventBridgePipeEcsTaskTargetParameters`

ECS RunTask shaping (target arn: the cluster).

### spec.targetParameters.ecsTask.taskDefinitionArn

`string | valueFrom` · required

The task definition to run. Reference an AwsEcsTaskDefinition
task_definition_arn output or pass a literal ARN.

- references: AwsEcsTaskDefinition (`status.outputs.task_definition_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEcsTaskDefinition, name: <that resource's name>, fieldPath: status.outputs.task_definition_arn}} -- a bare string does not parse

### spec.targetParameters.ecsTask.taskCount

`int32` · optional (explicit presence)

How many tasks each event batch launches. Unset keeps AWS's
default (1).

- rule: {"int32":{"gte":1}}

### spec.targetParameters.ecsTask.launchType

`string`

FARGATE, EC2, or EXTERNAL. Unset lets the cluster's default
capacity providers decide (mutually exclusive with
capacity_provider_strategy at AWS).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["EC2","FARGATE","EXTERNAL"]}}

### spec.targetParameters.ecsTask.capacityProviderStrategy

`[]AwsEventBridgePipeCapacityProviderStrategy`

Capacity provider strategy entries (at most 6). Mutually exclusive
with launch_type at AWS.

- rule: {"repeated":{"maxItems":"6"}}

### spec.targetParameters.ecsTask.capacityProviderStrategy[].capacityProvider

`string` · required

The capacity provider's name (e.g. "FARGATE", "FARGATE_SPOT", or a
custom provider).

- rule: {"string":{"minLen":"1","maxLen":"255"}}

### spec.targetParameters.ecsTask.capacityProviderStrategy[].base

`int32`

How many tasks run on this provider before the weights apply
(0-100000).

- rule: {"int32":{"lte":100000,"gte":0}}

### spec.targetParameters.ecsTask.capacityProviderStrategy[].weight

`int32`

The relative share of remaining tasks on this provider (0-1000).

- rule: {"int32":{"lte":1000,"gte":0}}

### spec.targetParameters.ecsTask.networkConfiguration

`AwsEventBridgePipeEcsNetworkConfiguration`

VPC networking for awsvpc-mode tasks (required for Fargate).

### spec.targetParameters.ecsTask.networkConfiguration.subnets

`[]string | valueFrom` · required

The subnets tasks launch into (at most 16). Reference AwsSubnet
subnet_id outputs or pass literal subnet-... ids.

Containment-exempt: these subnets place the ECS TASKS the pipe
launches, not the pipe itself, which is a managed service outside the
network. On a diagram the pipe stands beside the VPC with a line in --
the verdict an EventBridge rule's ECS target subnets already carry.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"repeated":{"minItems":"1","maxItems":"16"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.targetParameters.ecsTask.networkConfiguration.securityGroups

`[]string | valueFrom`

The security groups attached to tasks (at most 5). Unset uses the
VPC's default. Reference AwsSecurityGroup security_group_id
outputs or pass literal sg-... ids.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.targetParameters.ecsTask.networkConfiguration.assignPublicIp

`bool`

Assign public IPs to tasks (public-subnet Fargate tasks that pull
images over the internet need this).

### spec.targetParameters.ecsTask.group

`string`

The ECS task group name (up to 255 characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.targetParameters.ecsTask.platformVersion

`string`

The Fargate platform version (e.g. "LATEST", "1.4.0"). Fargate
only.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.targetParameters.ecsTask.overrides

`AwsEventBridgePipeEcsTaskOverrides`

Per-launch overrides of the task definition (containers, sizing,
roles).

### spec.targetParameters.ecsTask.overrides.containerOverrides

`[]AwsEventBridgePipeEcsContainerOverride`

Per-container overrides.

### spec.targetParameters.ecsTask.overrides.containerOverrides[].name

`string`

The container's name (as in the task definition).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.targetParameters.ecsTask.overrides.containerOverrides[].command

`[]string`

The command override.

### spec.targetParameters.ecsTask.overrides.containerOverrides[].cpu

`int32` · optional (explicit presence)

The CPU units override.

### spec.targetParameters.ecsTask.overrides.containerOverrides[].memory

`int32` · optional (explicit presence)

The hard memory limit override (MiB).

### spec.targetParameters.ecsTask.overrides.containerOverrides[].memoryReservation

`int32` · optional (explicit presence)

The soft memory limit override (MiB).

### spec.targetParameters.ecsTask.overrides.containerOverrides[].environment

`[]AwsEventBridgePipeEcsEnvironmentVariable`

Environment variable overrides.

### spec.targetParameters.ecsTask.overrides.containerOverrides[].environment[].name

`string` · required

The variable's name.

- rule: {"string":{"minLen":"1"}}

### spec.targetParameters.ecsTask.overrides.containerOverrides[].environment[].value

`string`

The variable's value.

### spec.targetParameters.ecsTask.overrides.containerOverrides[].environmentFiles

`[]AwsEventBridgePipeEcsEnvironmentFile`

Environment files (S3 ARNs of .env objects).

### spec.targetParameters.ecsTask.overrides.containerOverrides[].environmentFiles[].type

`string`

The only type AWS supports is "s3".

- rule: {"string":{"in":["s3"]}}

### spec.targetParameters.ecsTask.overrides.containerOverrides[].environmentFiles[].value

`string`

The S3 object ARN of the .env file.

- rule: {"string":{"pattern":"^arn:aws.*$"}}

### spec.targetParameters.ecsTask.overrides.containerOverrides[].resourceRequirements

`[]AwsEventBridgePipeEcsResourceRequirement`

Resource requirement overrides (GPU / InferenceAccelerator).

### spec.targetParameters.ecsTask.overrides.containerOverrides[].resourceRequirements[].type

`string`

"GPU" or "InferenceAccelerator".

- rule: {"string":{"in":["GPU","InferenceAccelerator"]}}

### spec.targetParameters.ecsTask.overrides.containerOverrides[].resourceRequirements[].value

`string` · required

The requirement's value (GPU count, or the accelerator device
name).

- rule: {"string":{"minLen":"1"}}

### spec.targetParameters.ecsTask.overrides.cpu

`string`

The task-level CPU override (e.g. "1 vCPU").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.targetParameters.ecsTask.overrides.memory

`string`

The task-level memory override (e.g. "2048").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.targetParameters.ecsTask.overrides.ephemeralStorageSizeInGib

`int32` · optional (explicit presence)

The ephemeral storage size in GiB (21-200). Fargate only.

- rule: {"int32":{"lte":200,"gte":21}}

### spec.targetParameters.ecsTask.overrides.executionRoleArn

`string | valueFrom`

The execution role override. Reference an AwsIamRole role_arn
output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.targetParameters.ecsTask.overrides.taskRoleArn

`string | valueFrom`

The task role override. Reference an AwsIamRole role_arn output or
pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.targetParameters.ecsTask.overrides.inferenceAcceleratorOverrides

`[]AwsEventBridgePipeEcsInferenceAcceleratorOverride`

Inference accelerator overrides.

### spec.targetParameters.ecsTask.overrides.inferenceAcceleratorOverrides[].deviceName

`string`

The accelerator device name (as in the task definition).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.targetParameters.ecsTask.overrides.inferenceAcceleratorOverrides[].deviceType

`string`

The accelerator device type.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.targetParameters.ecsTask.placementConstraints

`[]AwsEventBridgePipePlacementConstraint`

Placement constraints (at most 10). EC2 launch type only.

- rule: {"repeated":{"maxItems":"10"}}

### spec.targetParameters.ecsTask.placementConstraints[].type

`string`

"distinctInstance" (spread tasks across instances) or "memberOf"
(cluster-query-language expression).

- rule: {"string":{"in":["distinctInstance","memberOf"]}}

### spec.targetParameters.ecsTask.placementConstraints[].expression

`string`

The cluster query language expression, for type memberOf.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"2000"}}

### spec.targetParameters.ecsTask.placementStrategy

`[]AwsEventBridgePipePlacementStrategy`

Placement strategies (at most 5). EC2 launch type only.

- rule: {"repeated":{"maxItems":"5"}}

### spec.targetParameters.ecsTask.placementStrategy[].type

`string`

"random", "spread", or "binpack".

- rule: {"string":{"in":["random","spread","binpack"]}}

### spec.targetParameters.ecsTask.placementStrategy[].field

`string`

The field the strategy applies to (e.g. "instanceId" or
"attribute:ecs.availability-zone" for spread, "cpu"/"memory" for
binpack).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"255"}}

### spec.targetParameters.ecsTask.propagateTags

`string`

Propagate tags from the task definition to launched tasks. The
only value AWS accepts is "TASK_DEFINITION"; unset propagates
nothing.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["TASK_DEFINITION"]}}

### spec.targetParameters.ecsTask.tags

`map<string, string>`

Tags applied to every launched task.

### spec.targetParameters.ecsTask.enableEcsManagedTags

`bool`

Use ECS managed tags on launched tasks.

### spec.targetParameters.ecsTask.enableExecuteCommand

`bool`

Enable ECS Exec on launched tasks (interactive debugging).

### spec.targetParameters.ecsTask.referenceId

`string`

The reference id attached to each RunTask call (up to 1024
characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"1024"}}

### spec.targetParameters.batchJob

`AwsEventBridgePipeBatchJobTargetParameters`

Batch SubmitJob shaping (target arn: the job queue).

### spec.targetParameters.batchJob.jobDefinition

`string` · required

The job definition (name, name:revision, or ARN).

- rule: {"string":{"minLen":"1"}}

### spec.targetParameters.batchJob.jobName

`string` · required

The submitted job's name (up to 128 characters).

- rule: {"string":{"minLen":"1","maxLen":"128"}}

### spec.targetParameters.batchJob.arraySize

`int32` · optional (explicit presence)

Array job sizing (2-10000 child jobs). Unset submits a regular
job.

- rule: {"int32":{"lte":10000,"gte":2}}

### spec.targetParameters.batchJob.retryAttempts

`int32` · optional (explicit presence)

Retry attempts override (1-10). Unset keeps the job definition's.

- rule: {"int32":{"lte":10,"gte":1}}

### spec.targetParameters.batchJob.parameters

`map<string, string>`

Job parameter placeholders (name → value), substituted into the
job definition.

### spec.targetParameters.batchJob.dependsOn

`[]AwsEventBridgePipeBatchJobDependency`

Jobs this submission depends on (at most 20).

- rule: {"repeated":{"maxItems":"20"}}

### spec.targetParameters.batchJob.dependsOn[].jobId

`string`

The job id this submission waits on.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.targetParameters.batchJob.dependsOn[].type

`string`

"SEQUENTIAL" or "N_TO_N" (array jobs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SEQUENTIAL","N_TO_N"]}}

### spec.targetParameters.batchJob.containerOverrides

`AwsEventBridgePipeBatchContainerOverrides`

The container override for the submitted job.

### spec.targetParameters.batchJob.containerOverrides.command

`[]string`

The command override.

### spec.targetParameters.batchJob.containerOverrides.environment

`[]AwsEventBridgePipeEcsEnvironmentVariable`

Environment variable overrides.

### spec.targetParameters.batchJob.containerOverrides.environment[].name

`string` · required

The variable's name.

- rule: {"string":{"minLen":"1"}}

### spec.targetParameters.batchJob.containerOverrides.environment[].value

`string`

The variable's value.

### spec.targetParameters.batchJob.containerOverrides.instanceType

`string`

The instance type override (multi-node parallel jobs only).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.targetParameters.batchJob.containerOverrides.resourceRequirements

`[]AwsEventBridgePipeBatchResourceRequirement`

Resource requirement overrides ("GPU", "MEMORY", "VCPU").

### spec.targetParameters.batchJob.containerOverrides.resourceRequirements[].type

`string`

"GPU", "MEMORY", or "VCPU".

- rule: {"string":{"in":["GPU","MEMORY","VCPU"]}}

### spec.targetParameters.batchJob.containerOverrides.resourceRequirements[].value

`string` · required

The requirement's value (count, MiB, or vCPUs).

- rule: {"string":{"minLen":"1"}}

### spec.targetParameters.redshiftData

`AwsEventBridgePipeRedshiftDataTargetParameters`

Redshift Data API statement shaping (target arn: the cluster).

- rule: set exactly one of db_user (temporary credentials) and secret_manager_arn (stored credentials)

### spec.targetParameters.redshiftData.database

`string` · required

The database to run statements in (1-64 characters).

- rule: {"string":{"minLen":"1","maxLen":"64"}}

### spec.targetParameters.redshiftData.sqls

`[]string` · required

The SQL statements to run (each up to 100000 characters).

- rule: {"repeated":{"minItems":"1","items":{"string":{"minLen":"1","maxLen":"100000"}}}}

### spec.targetParameters.redshiftData.dbUser

`string`

The database user for temporary-credential auth (1-128
characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128"}}

### spec.targetParameters.redshiftData.secretManagerArn

`string`

Secrets Manager secret ARN holding the database credentials - a
reference, never the credential value itself.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^arn:aws.*$"}}

### spec.targetParameters.redshiftData.statementName

`string`

The statement name, for Data API auditing (1-500 characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"500"}}

### spec.targetParameters.redshiftData.withEvent

`bool`

Send an event back to EventBridge when the statement finishes.

### spec.targetParameters.sagemakerPipeline

`AwsEventBridgePipeSageMakerPipelineTargetParameters`

SageMaker StartPipelineExecution shaping.

- rule: pipeline_parameters entries must have unique names

### spec.targetParameters.sagemakerPipeline.pipelineParameters

`[]AwsEventBridgePipeSageMakerPipelineParameter`

Name/value parameters passed to the pipeline execution (at most
200).

- rule: {"repeated":{"maxItems":"200"}}

### spec.targetParameters.sagemakerPipeline.pipelineParameters[].name

`string` · required

The parameter's name (supports <$.json.path> dynamic values).

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.targetParameters.sagemakerPipeline.pipelineParameters[].value

`string` · required

The parameter's value.

- rule: {"string":{"minLen":"1","maxLen":"1024"}}

### spec.targetParameters.eventbridgeEventBus

`AwsEventBridgePipeEventBusTargetParameters`

EventBridge PutEvents shaping (target arn: the bus).

### spec.targetParameters.eventbridgeEventBus.detailType

`string`

The event's detail-type (what rules match on; up to 128
characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"128"}}

### spec.targetParameters.eventbridgeEventBus.source

`string`

The event's source (e.g. "com.acme.orders"; up to 256 characters).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"256"}}

### spec.targetParameters.eventbridgeEventBus.endpointId

`string`

The global-endpoint id for multi-region buses (e.g.
"abcde.veo"). Unset for regular buses.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"50","pattern":"^[0-9A-Za-z-]+[.][0-9A-Za-z-]+$"}}

### spec.targetParameters.eventbridgeEventBus.resources

`[]string`

ARNs the event claims involvement of (rule matching on resources;
at most 10).

- rule: {"repeated":{"maxItems":"10","items":{"string":{"pattern":"^arn:aws.*$"}}}}

### spec.targetParameters.eventbridgeEventBus.time

`string`

The event's timestamp, as a <$.json.path> expression into the
source event. Unset stamps delivery time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"256","pattern":"^\\$(\\.[\\w/_-]+(\\[(\\d+|\\*)\\])*)*$"}}

### spec.targetParameters.cloudwatchLogs

`AwsEventBridgePipeCloudWatchLogsTargetParameters`

CloudWatch Logs PutLogEvents shaping (target arn: the log group).

### spec.targetParameters.cloudwatchLogs.logStreamName

`string`

The log stream to write to (up to 256 characters). Unset lets the
pipe manage streams.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"256"}}

### spec.targetParameters.cloudwatchLogs.timestamp

`string`

The log event's timestamp, as a <$.json.path> expression into the
source event (e.g. "$.detail.timestamp"). Unset stamps delivery
time.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"256","pattern":"^\\$(\\.[\\w/_-]+(\\[(\\d+|\\*)\\])*)*$"}}

### spec.targetParameters.http

`AwsEventBridgePipeHttpParameters`

API-destination HTTP shaping.

### spec.targetParameters.http.headerParameters

`map<string, string>`

Headers added to the request.

### spec.targetParameters.http.pathParameterValue

`string`

The value filling the destination URL's "*" path wildcard. (AWS's
API models this as a list but accepts exactly ONE entry.)

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.targetParameters.http.queryStringParameters

`map<string, string>`

Query-string parameters added to the request.

### spec.roleArn

`string | valueFrom` · required

The role the pipe assumes to read the source, call the enrichment,
and write the target. Trust policy allows pipes.amazonaws.com.
Reference an AwsIamRole role_arn output or pass a literal ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.desiredState

`string`

RUNNING consumes and delivers; STOPPED pauses consumption without
deleting the pipe (stream positions are kept). Unset keeps AWS's
default (RUNNING).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["RUNNING","STOPPED"]}}

### spec.kmsKeyIdentifier

`string | valueFrom`

Customer-managed KMS key that encrypts pipe data at rest (unset
uses AWS-owned keys). Reference an AwsKmsKey key_arn output or
pass a literal key id/ARN/alias.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.logConfiguration

`AwsEventBridgePipeLogConfiguration`

Execution logging: level plus one or more destinations (CloudWatch
Logs, Firehose, S3). Unset disables pipe logging.

- rule: configure at least one destination (cloudwatch_logs, firehose, s3) - a level with nowhere to write logs nothing

### spec.logConfiguration.level

`string`

How much to log: OFF, ERROR, INFO, or TRACE (TRACE includes
payloads when include_execution_data is set - mind the
sensitivity).

- rule: {"string":{"in":["OFF","ERROR","INFO","TRACE"]}}

### spec.logConfiguration.includeExecutionData

`bool`

Include event payloads in execution logs. Payloads may carry
sensitive data - enable deliberately.

### spec.logConfiguration.cloudwatchLogs

`AwsEventBridgePipeCloudWatchLogsLogDestination`

Write execution logs to a CloudWatch log group.

### spec.logConfiguration.cloudwatchLogs.logGroupArn

`string | valueFrom` · required

The log group's ARN. Reference an AwsCloudwatchLogGroup
log_group_arn output or pass a literal ARN.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.logConfiguration.firehose

`AwsEventBridgePipeFirehoseLogDestination`

Write execution logs to a Firehose delivery stream.

### spec.logConfiguration.firehose.deliveryStreamArn

`string | valueFrom` · required

The delivery stream's ARN. Reference an AwsKinesisFirehose
delivery_stream_arn output or pass a literal ARN.

- references: AwsKinesisFirehose (`status.outputs.delivery_stream_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKinesisFirehose, name: <that resource's name>, fieldPath: status.outputs.delivery_stream_arn}} -- a bare string does not parse

### spec.logConfiguration.s3

`AwsEventBridgePipeS3LogDestination`

Write execution logs to an S3 bucket.

### spec.logConfiguration.s3.bucketName

`string | valueFrom` · required

The receiving bucket's name. Reference an AwsS3Bucket bucket_id
output or pass a literal bucket name.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.logConfiguration.s3.bucketOwner

`string`

The bucket owner's AWS account id (cross-account bucket
protection).

- rule: {"string":{"pattern":"^\\d{12}$"}}

### spec.logConfiguration.s3.outputFormat

`string`

The log object format. Unset keeps AWS's default (json).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["json","plain","w3c"]}}

### spec.logConfiguration.s3.prefix

`string`

The key prefix under the bucket.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

## Validation Rules

- `spec.enrichment_parameters_require_enrichment`: enrichment_parameters shape the enrichment invocation - configure enrichment or drop them

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsEventBridgePipe, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.pipe_arn` | `string` | The pipe's ARN. |
| `status.outputs.pipe_name` | `string` | The pipe's name (metadata.name echoed back - the provider's import ID). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.sourceParameters.kinesis.deadLetterQueueArn` | AwsSqsQueue | `status.outputs.queue_arn` |
| `spec.sourceParameters.dynamodb.deadLetterQueueArn` | AwsSqsQueue | `status.outputs.queue_arn` |
| `spec.sourceParameters.selfManagedKafka.vpc.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.sourceParameters.selfManagedKafka.vpc.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.targetParameters.ecsTask.taskDefinitionArn` | AwsEcsTaskDefinition | `status.outputs.task_definition_arn` |
| `spec.targetParameters.ecsTask.networkConfiguration.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.targetParameters.ecsTask.networkConfiguration.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.targetParameters.ecsTask.overrides.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.targetParameters.ecsTask.overrides.taskRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.kmsKeyIdentifier` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.logConfiguration.cloudwatchLogs.logGroupArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.logConfiguration.firehose.deliveryStreamArn` | AwsKinesisFirehose | `status.outputs.delivery_stream_arn` |
| `spec.logConfiguration.s3.bucketName` | AwsS3Bucket | `status.outputs.bucket_id` |

## See Also

- [Overview](../README.md)
