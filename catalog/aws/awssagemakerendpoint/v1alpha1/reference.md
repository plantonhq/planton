# AwsSagemakerEndpoint

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsSagemakerEndpointSpec defines the desired configuration for an
Amazon SageMaker AI real-time inference endpoint TOGETHER WITH its
endpoint configuration (the immutable capacity/variant definition the
endpoint points at). The endpoint's AWS name derives from
metadata.name.

AWS's own update model: an endpoint configuration is immutable, so
changing any variant/capture/async setting rolls a NEW configuration
and repoints the endpoint at it (UpdateEndpoint, optionally shaped by
`deployment`). The modules own that choreography - configurations are
name-suffixed and created before the old one is destroyed, so the
endpoint never references a deleted configuration.

## Example

```yaml
# Canonical AwsSagemakerEndpoint example (hack/dev manifest and refgen
# Example source): an instance-backed endpoint exercising every arm -
# weighted variants, managed instance scaling, routing, core dumps,
# data capture, async inference, KMS, and a canary blue/green
# deployment policy with alarm rollback. Literal ARNs stand in for
# composed references so the offline `tofu plan` renders every arm.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSagemakerEndpoint
metadata:
  name: churn-scoring
  id: churn-scoring
  org: test-org
  env: dev
spec:
  region: us-west-2
  kmsKeyArn:
    value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
  productionVariants:
    - variantName: primary
      model:
        value: churn-scoring-model
      instanceType: ml.m5.large
      initialInstanceCount: 2
      initialVariantWeight: 0.9
      routingStrategy: LEAST_OUTSTANDING_REQUESTS
      volumeSizeGb: 64
      containerStartupHealthCheckTimeoutSeconds: 300
      modelDataDownloadTimeoutSeconds: 600
      enableSsmAccess: true
      managedInstanceScaling:
        status: ENABLED
        minInstanceCount: 1
        maxInstanceCount: 4
      coreDump:
        destinationS3Uri: s3://my-dumps/churn-scoring/
        kmsKeyArn:
          value: arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
    - variantName: canary
      model:
        value: churn-scoring-model-next
      instanceType: ml.m5.large
      initialInstanceCount: 1
      initialVariantWeight: 0.1
  asyncInference:
    outputS3Path: s3://my-bucket/async-out/
    failureS3Path: s3://my-bucket/async-fail/
    maxConcurrentInvocationsPerInstance: 4
    successTopicArn:
      value: arn:aws:sns:us-west-2:123456789012:inference-ok
    errorTopicArn:
      value: arn:aws:sns:us-west-2:123456789012:inference-err
    includeInferenceResponseIn:
      - ERROR_NOTIFICATION_TOPIC
  dataCapture:
    destinationS3Uri: s3://my-bucket/capture/
    initialSamplingPercentage: 25
    captureModes:
      - Input
      - Output
    enableCapture: true
    jsonContentTypes:
      - application/json
  deployment:
    blueGreen:
      trafficRoutingType: CANARY
      waitIntervalSeconds: 300
      canarySize:
        type: CAPACITY_PERCENT
        value: 20
      terminationWaitSeconds: 120
      maximumExecutionTimeoutSeconds: 3600
    autoRollbackAlarmNames:
      - churn-scoring-5xx
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.productionVariants` | `[]AwsSagemakerEndpointVariant` | yes |  |  |
| `spec.productionVariants[].variantName` | `string` |  |  |  |
| `spec.productionVariants[].model` | `string \| valueFrom` |  |  | AwsSagemakerModel (`status.outputs.model_name`) |
| `spec.productionVariants[].instanceType` | `string` |  |  |  |
| `spec.productionVariants[].initialInstanceCount` | `int32` |  |  |  |
| `spec.productionVariants[].initialVariantWeight` | `float` |  |  |  |
| `spec.productionVariants[].serverless` | `AwsSagemakerEndpointServerlessConfig` |  |  |  |
| `spec.productionVariants[].serverless.maxConcurrency` | `int32` |  |  |  |
| `spec.productionVariants[].serverless.memorySizeMb` | `int32` |  |  |  |
| `spec.productionVariants[].serverless.provisionedConcurrency` | `int32` |  |  |  |
| `spec.productionVariants[].managedInstanceScaling` | `AwsSagemakerEndpointManagedInstanceScaling` |  |  |  |
| `spec.productionVariants[].managedInstanceScaling.status` | `string` |  |  |  |
| `spec.productionVariants[].managedInstanceScaling.minInstanceCount` | `int32` |  |  |  |
| `spec.productionVariants[].managedInstanceScaling.maxInstanceCount` | `int32` |  |  |  |
| `spec.productionVariants[].routingStrategy` | `string` |  |  |  |
| `spec.productionVariants[].volumeSizeGb` | `int32` |  |  |  |
| `spec.productionVariants[].containerStartupHealthCheckTimeoutSeconds` | `int32` |  |  |  |
| `spec.productionVariants[].modelDataDownloadTimeoutSeconds` | `int32` |  |  |  |
| `spec.productionVariants[].enableSsmAccess` | `bool` |  |  |  |
| `spec.productionVariants[].inferenceAmiVersion` | `string` |  |  |  |
| `spec.productionVariants[].acceleratorType` | `string` |  |  |  |
| `spec.productionVariants[].coreDump` | `AwsSagemakerEndpointCoreDump` |  |  |  |
| `spec.productionVariants[].coreDump.destinationS3Uri` | `string` | yes |  |  |
| `spec.productionVariants[].coreDump.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.productionVariants[].mlCapacityReservationArn` | `string` |  |  |  |
| `spec.shadowVariants` | `[]AwsSagemakerEndpointVariant` |  |  |  |
| `spec.shadowVariants[].variantName` | `string` |  |  |  |
| `spec.shadowVariants[].model` | `string \| valueFrom` |  |  | AwsSagemakerModel (`status.outputs.model_name`) |
| `spec.shadowVariants[].instanceType` | `string` |  |  |  |
| `spec.shadowVariants[].initialInstanceCount` | `int32` |  |  |  |
| `spec.shadowVariants[].initialVariantWeight` | `float` |  |  |  |
| `spec.shadowVariants[].serverless` | `AwsSagemakerEndpointServerlessConfig` |  |  |  |
| `spec.shadowVariants[].serverless.maxConcurrency` | `int32` |  |  |  |
| `spec.shadowVariants[].serverless.memorySizeMb` | `int32` |  |  |  |
| `spec.shadowVariants[].serverless.provisionedConcurrency` | `int32` |  |  |  |
| `spec.shadowVariants[].managedInstanceScaling` | `AwsSagemakerEndpointManagedInstanceScaling` |  |  |  |
| `spec.shadowVariants[].managedInstanceScaling.status` | `string` |  |  |  |
| `spec.shadowVariants[].managedInstanceScaling.minInstanceCount` | `int32` |  |  |  |
| `spec.shadowVariants[].managedInstanceScaling.maxInstanceCount` | `int32` |  |  |  |
| `spec.shadowVariants[].routingStrategy` | `string` |  |  |  |
| `spec.shadowVariants[].volumeSizeGb` | `int32` |  |  |  |
| `spec.shadowVariants[].containerStartupHealthCheckTimeoutSeconds` | `int32` |  |  |  |
| `spec.shadowVariants[].modelDataDownloadTimeoutSeconds` | `int32` |  |  |  |
| `spec.shadowVariants[].enableSsmAccess` | `bool` |  |  |  |
| `spec.shadowVariants[].inferenceAmiVersion` | `string` |  |  |  |
| `spec.shadowVariants[].acceleratorType` | `string` |  |  |  |
| `spec.shadowVariants[].coreDump` | `AwsSagemakerEndpointCoreDump` |  |  |  |
| `spec.shadowVariants[].coreDump.destinationS3Uri` | `string` | yes |  |  |
| `spec.shadowVariants[].coreDump.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.shadowVariants[].mlCapacityReservationArn` | `string` |  |  |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.executionRoleArn` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.asyncInference` | `AwsSagemakerEndpointAsyncInference` |  |  |  |
| `spec.asyncInference.outputS3Path` | `string` | yes |  |  |
| `spec.asyncInference.failureS3Path` | `string` |  |  |  |
| `spec.asyncInference.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.asyncInference.maxConcurrentInvocationsPerInstance` | `int32` |  |  |  |
| `spec.asyncInference.successTopicArn` | `string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.asyncInference.errorTopicArn` | `string \| valueFrom` |  |  | AwsSnsTopic (`status.outputs.topic_arn`) |
| `spec.asyncInference.includeInferenceResponseIn` | `[]string` |  |  |  |
| `spec.dataCapture` | `AwsSagemakerEndpointDataCapture` |  |  |  |
| `spec.dataCapture.destinationS3Uri` | `string` | yes |  |  |
| `spec.dataCapture.initialSamplingPercentage` | `int32` |  |  |  |
| `spec.dataCapture.captureModes` | `[]string` | yes |  |  |
| `spec.dataCapture.enableCapture` | `bool` |  |  |  |
| `spec.dataCapture.csvContentTypes` | `[]string` |  |  |  |
| `spec.dataCapture.jsonContentTypes` | `[]string` |  |  |  |
| `spec.dataCapture.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.deployment` | `AwsSagemakerEndpointDeployment` |  |  |  |
| `spec.deployment.blueGreen` | `AwsSagemakerEndpointBlueGreenPolicy` |  |  |  |
| `spec.deployment.blueGreen.trafficRoutingType` | `string` |  |  |  |
| `spec.deployment.blueGreen.waitIntervalSeconds` | `int32` |  |  |  |
| `spec.deployment.blueGreen.canarySize` | `AwsSagemakerEndpointCapacitySize` |  |  |  |
| `spec.deployment.blueGreen.canarySize.type` | `string` |  |  |  |
| `spec.deployment.blueGreen.canarySize.value` | `int32` |  |  |  |
| `spec.deployment.blueGreen.linearStepSize` | `AwsSagemakerEndpointCapacitySize` |  |  |  |
| `spec.deployment.blueGreen.linearStepSize.type` | `string` |  |  |  |
| `spec.deployment.blueGreen.linearStepSize.value` | `int32` |  |  |  |
| `spec.deployment.blueGreen.terminationWaitSeconds` | `int32` |  |  |  |
| `spec.deployment.blueGreen.maximumExecutionTimeoutSeconds` | `int32` |  |  |  |
| `spec.deployment.rolling` | `AwsSagemakerEndpointRollingPolicy` |  |  |  |
| `spec.deployment.rolling.maximumBatchSize` | `AwsSagemakerEndpointCapacitySize` | yes |  |  |
| `spec.deployment.rolling.maximumBatchSize.type` | `string` |  |  |  |
| `spec.deployment.rolling.maximumBatchSize.value` | `int32` |  |  |  |
| `spec.deployment.rolling.waitIntervalSeconds` | `int32` |  |  |  |
| `spec.deployment.rolling.rollbackMaximumBatchSize` | `AwsSagemakerEndpointCapacitySize` |  |  |  |
| `spec.deployment.rolling.rollbackMaximumBatchSize.type` | `string` |  |  |  |
| `spec.deployment.rolling.rollbackMaximumBatchSize.value` | `int32` |  |  |  |
| `spec.deployment.rolling.maximumExecutionTimeoutSeconds` | `int32` |  |  |  |
| `spec.deployment.autoRollbackAlarmNames` | `[]string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the endpoint will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.productionVariants

`[]AwsSagemakerEndpointVariant` · required

The models served by this endpoint - each variant hosts one model
with its own capacity (1-10). One variant is the common case;
multiple variants split traffic by weight (A/B testing).

- rule: {"repeated":{"minItems":"1","maxItems":"10"}}
- rule: serverless variants cannot set instance_type, initial_instance_count, managed_instance_scaling, volume_size_gb, accelerator_type, inference_ami_version, core_dump, or ml_capacity_reservation_arn
- rule: either instance_type or serverless must be set

### spec.productionVariants[].variantName

`string`

Variant name (letters, digits, hyphens; max 63). Optional - the
modules default it deterministically per position ("variant-0",
"variant-1", ...) so plans stay stable; AWS's console convention
for a single variant is "AllTraffic".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[0-9A-Za-z-]+$"}}

### spec.productionVariants[].model

`string | valueFrom`

The SageMaker model this variant serves. Omit ONLY for
inference-component endpoints (components attach models later) -
then the spec-level execution_role_arn is required.

- references: AwsSagemakerModel (`status.outputs.model_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSagemakerModel, name: <that resource's name>, fieldPath: status.outputs.model_name}} -- a bare string does not parse

### spec.productionVariants[].instanceType

`string`

Instance type for dedicated capacity (an "ml.*" type, e.g.
"ml.m5.large", "ml.g5.xlarge"). AWS's accepted set grows with every
release - the value passes through to the API, which rejects
unknown types. Required unless `serverless` is set. NOTE the
per-type "for endpoint usage" Service Quota: on a fresh AWS
account it defaults to ZERO for nearly every instance family
(ml.m5.large and ml.c6i.large included) - CreateEndpoint fails
with ResourceLimitExceeded until a quota increase is granted. The
entry-level exceptions whose default is 2 are ml.t2.* (x86) and
ml.m6g.large (Graviton - the container image must be arm64);
`serverless` needs no per-type quota (default: 5 endpoints per
region). Check with: aws service-quotas list-service-quotas
--service-code sagemaker.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^ml\\.[a-z0-9]+([.-][a-z0-9]+)*$"}}

### spec.productionVariants[].initialInstanceCount

`int32` · optional (explicit presence)

Number of instances to launch initially (>= 1). With
`managed_instance_scaling` or application auto-scaling this is the
starting point, not a fixed size.

- rule: {"int32":{"gte":1}}

### spec.productionVariants[].initialVariantWeight

`float` · optional (explicit presence)

Relative traffic share among variants (>= 0; AWS defaults to 1.0).
Traffic splits proportionally to weight/sum(weights); 0 sends no
traffic while keeping the variant deployed.

- rule: {"float":{"gte":0}}

### spec.productionVariants[].serverless

`AwsSagemakerEndpointServerlessConfig`

Serverless compute instead of dedicated instances: SageMaker scales
capacity with traffic and bills per inference (nothing while idle).
Excludes
every instance-based setting on this variant.

- rule: provisioned_concurrency must not exceed max_concurrency

### spec.productionVariants[].serverless.maxConcurrency

`int32`

Maximum concurrent invocations the variant processes (1-200).

- rule: {"int32":{"lte":200,"gte":1}}

### spec.productionVariants[].serverless.memorySizeMb

`int32`

Memory per invocation environment in MB - one of 1024, 2048, 3072,
4096, 5120, 6144 (CPU scales with memory).

- rule: {"int32":{"in":[1024,2048,3072,4096,5120,6144]}}

### spec.productionVariants[].serverless.provisionedConcurrency

`int32` · optional (explicit presence)

Pre-warmed concurrency held ready to serve without cold starts
(1-200; must not exceed max_concurrency). Billed while provisioned.

- rule: {"int32":{"lte":200,"gte":1}}

### spec.productionVariants[].managedInstanceScaling

`AwsSagemakerEndpointManagedInstanceScaling`

Endpoint-managed instance range (min/max autoscaling handled by
SageMaker itself, no Application Auto Scaling policy needed).

- rule: min_instance_count must not exceed max_instance_count

### spec.productionVariants[].managedInstanceScaling.status

`string`

"ENABLED" or "DISABLED". Omitted = AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

### spec.productionVariants[].managedInstanceScaling.minInstanceCount

`int32` · optional (explicit presence)

Instances retained when scaling down (>= 0; 0 lets the endpoint
scale to zero where the instance family supports it).

- rule: {"int32":{"gte":0}}

### spec.productionVariants[].managedInstanceScaling.maxInstanceCount

`int32` · optional (explicit presence)

Instances provisioned at peak (>= 1).

- rule: {"int32":{"gte":1}}

### spec.productionVariants[].routingStrategy

`string`

How invocations route across instances:
"LEAST_OUTSTANDING_REQUESTS" (favor free capacity) or "RANDOM".
Omitted = AWS default (RANDOM).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["LEAST_OUTSTANDING_REQUESTS","RANDOM"]}}

### spec.productionVariants[].volumeSizeGb

`int32` · optional (explicit presence)

ML storage volume attached to each instance, in GB (1-512). Only
instance families with EBS volumes accept it (nitro-local-storage
families reject it).

- rule: {"int32":{"lte":512,"gte":1}}

### spec.productionVariants[].containerStartupHealthCheckTimeoutSeconds

`int32` · optional (explicit presence)

Seconds the inference container may take to pass its startup health
check (60-3600). Raise for large models with slow load times.

- rule: {"int32":{"lte":3600,"gte":60}}

### spec.productionVariants[].modelDataDownloadTimeoutSeconds

`int32` · optional (explicit presence)

Seconds allowed to download and extract model data from S3 to the
instance (60-3600). Raise for multi-GB artifacts.

- rule: {"int32":{"lte":3600,"gte":60}}

### spec.productionVariants[].enableSsmAccess

`bool`

Allow AWS Systems Manager sessions into the variant's instances
(debugging). Ignored on inference-component endpoints.

### spec.productionVariants[].inferenceAmiVersion

`string`

Pin the preconfigured inference AMI (GPU/Neuron driver line).
Omitted = AWS picks. Values at the current provider line:
"al2-ami-sagemaker-inference-gpu-2",
"al2-ami-sagemaker-inference-gpu-2-1",
"al2-ami-sagemaker-inference-gpu-3-1",
"al2-ami-sagemaker-inference-neuron-2",
"al2023-ami-sagemaker-inference-gpu-4-1".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["al2-ami-sagemaker-inference-gpu-2","al2-ami-sagemaker-inference-gpu-2-1","al2-ami-sagemaker-inference-gpu-3-1","al2-ami-sagemaker-inference-neuron-2","al2023-ami-sagemaker-inference-gpu-4-1"]}}

### spec.productionVariants[].acceleratorType

`string`

Elastic Inference accelerator attached to the variant (EI is
deprecated by AWS - existing workloads only). Values:
ml.eia1.medium/large/xlarge, ml.eia2.medium/large/xlarge.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ml.eia1.medium","ml.eia1.large","ml.eia1.xlarge","ml.eia2.medium","ml.eia2.large","ml.eia2.xlarge"]}}

### spec.productionVariants[].coreDump

`AwsSagemakerEndpointCoreDump`

Write core dumps from a crashed model container to S3.

### spec.productionVariants[].coreDump.destinationS3Uri

`string` · required

S3 destination for core dumps. Example: "s3://my-dumps/endpoint/"

- rule: {"string":{"minLen":"1","maxLen":"512","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.productionVariants[].coreDump.kmsKeyArn

`string | valueFrom`

KMS key encrypting the dumps at rest in S3.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.productionVariants[].mlCapacityReservationArn

`string`

Launch this variant's instances ONLY into the named ML capacity
reservation (the modules send AWS's capacity-reservations-only
preference alongside - its single legal value).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.shadowVariants

`[]AwsSagemakerEndpointVariant`

Shadow variants receive a copy of production traffic without
returning responses to callers (shadow testing). AWS allows exactly
one production and one shadow variant when shadow testing.

- rule: {"repeated":{"maxItems":"10"}}
- rule: serverless variants cannot set instance_type, initial_instance_count, managed_instance_scaling, volume_size_gb, accelerator_type, inference_ami_version, core_dump, or ml_capacity_reservation_arn
- rule: either instance_type or serverless must be set

### spec.shadowVariants[].variantName

`string`

Variant name (letters, digits, hyphens; max 63). Optional - the
modules default it deterministically per position ("variant-0",
"variant-1", ...) so plans stay stable; AWS's console convention
for a single variant is "AllTraffic".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"63","pattern":"^[0-9A-Za-z-]+$"}}

### spec.shadowVariants[].model

`string | valueFrom`

The SageMaker model this variant serves. Omit ONLY for
inference-component endpoints (components attach models later) -
then the spec-level execution_role_arn is required.

- references: AwsSagemakerModel (`status.outputs.model_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSagemakerModel, name: <that resource's name>, fieldPath: status.outputs.model_name}} -- a bare string does not parse

### spec.shadowVariants[].instanceType

`string`

Instance type for dedicated capacity (an "ml.*" type, e.g.
"ml.m5.large", "ml.g5.xlarge"). AWS's accepted set grows with every
release - the value passes through to the API, which rejects
unknown types. Required unless `serverless` is set. NOTE the
per-type "for endpoint usage" Service Quota: on a fresh AWS
account it defaults to ZERO for nearly every instance family
(ml.m5.large and ml.c6i.large included) - CreateEndpoint fails
with ResourceLimitExceeded until a quota increase is granted. The
entry-level exceptions whose default is 2 are ml.t2.* (x86) and
ml.m6g.large (Graviton - the container image must be arm64);
`serverless` needs no per-type quota (default: 5 endpoints per
region). Check with: aws service-quotas list-service-quotas
--service-code sagemaker.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^ml\\.[a-z0-9]+([.-][a-z0-9]+)*$"}}

### spec.shadowVariants[].initialInstanceCount

`int32` · optional (explicit presence)

Number of instances to launch initially (>= 1). With
`managed_instance_scaling` or application auto-scaling this is the
starting point, not a fixed size.

- rule: {"int32":{"gte":1}}

### spec.shadowVariants[].initialVariantWeight

`float` · optional (explicit presence)

Relative traffic share among variants (>= 0; AWS defaults to 1.0).
Traffic splits proportionally to weight/sum(weights); 0 sends no
traffic while keeping the variant deployed.

- rule: {"float":{"gte":0}}

### spec.shadowVariants[].serverless

`AwsSagemakerEndpointServerlessConfig`

Serverless compute instead of dedicated instances: SageMaker scales
capacity with traffic and bills per inference (nothing while idle).
Excludes
every instance-based setting on this variant.

- rule: provisioned_concurrency must not exceed max_concurrency

### spec.shadowVariants[].serverless.maxConcurrency

`int32`

Maximum concurrent invocations the variant processes (1-200).

- rule: {"int32":{"lte":200,"gte":1}}

### spec.shadowVariants[].serverless.memorySizeMb

`int32`

Memory per invocation environment in MB - one of 1024, 2048, 3072,
4096, 5120, 6144 (CPU scales with memory).

- rule: {"int32":{"in":[1024,2048,3072,4096,5120,6144]}}

### spec.shadowVariants[].serverless.provisionedConcurrency

`int32` · optional (explicit presence)

Pre-warmed concurrency held ready to serve without cold starts
(1-200; must not exceed max_concurrency). Billed while provisioned.

- rule: {"int32":{"lte":200,"gte":1}}

### spec.shadowVariants[].managedInstanceScaling

`AwsSagemakerEndpointManagedInstanceScaling`

Endpoint-managed instance range (min/max autoscaling handled by
SageMaker itself, no Application Auto Scaling policy needed).

- rule: min_instance_count must not exceed max_instance_count

### spec.shadowVariants[].managedInstanceScaling.status

`string`

"ENABLED" or "DISABLED". Omitted = AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ENABLED","DISABLED"]}}

### spec.shadowVariants[].managedInstanceScaling.minInstanceCount

`int32` · optional (explicit presence)

Instances retained when scaling down (>= 0; 0 lets the endpoint
scale to zero where the instance family supports it).

- rule: {"int32":{"gte":0}}

### spec.shadowVariants[].managedInstanceScaling.maxInstanceCount

`int32` · optional (explicit presence)

Instances provisioned at peak (>= 1).

- rule: {"int32":{"gte":1}}

### spec.shadowVariants[].routingStrategy

`string`

How invocations route across instances:
"LEAST_OUTSTANDING_REQUESTS" (favor free capacity) or "RANDOM".
Omitted = AWS default (RANDOM).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["LEAST_OUTSTANDING_REQUESTS","RANDOM"]}}

### spec.shadowVariants[].volumeSizeGb

`int32` · optional (explicit presence)

ML storage volume attached to each instance, in GB (1-512). Only
instance families with EBS volumes accept it (nitro-local-storage
families reject it).

- rule: {"int32":{"lte":512,"gte":1}}

### spec.shadowVariants[].containerStartupHealthCheckTimeoutSeconds

`int32` · optional (explicit presence)

Seconds the inference container may take to pass its startup health
check (60-3600). Raise for large models with slow load times.

- rule: {"int32":{"lte":3600,"gte":60}}

### spec.shadowVariants[].modelDataDownloadTimeoutSeconds

`int32` · optional (explicit presence)

Seconds allowed to download and extract model data from S3 to the
instance (60-3600). Raise for multi-GB artifacts.

- rule: {"int32":{"lte":3600,"gte":60}}

### spec.shadowVariants[].enableSsmAccess

`bool`

Allow AWS Systems Manager sessions into the variant's instances
(debugging). Ignored on inference-component endpoints.

### spec.shadowVariants[].inferenceAmiVersion

`string`

Pin the preconfigured inference AMI (GPU/Neuron driver line).
Omitted = AWS picks. Values at the current provider line:
"al2-ami-sagemaker-inference-gpu-2",
"al2-ami-sagemaker-inference-gpu-2-1",
"al2-ami-sagemaker-inference-gpu-3-1",
"al2-ami-sagemaker-inference-neuron-2",
"al2023-ami-sagemaker-inference-gpu-4-1".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["al2-ami-sagemaker-inference-gpu-2","al2-ami-sagemaker-inference-gpu-2-1","al2-ami-sagemaker-inference-gpu-3-1","al2-ami-sagemaker-inference-neuron-2","al2023-ami-sagemaker-inference-gpu-4-1"]}}

### spec.shadowVariants[].acceleratorType

`string`

Elastic Inference accelerator attached to the variant (EI is
deprecated by AWS - existing workloads only). Values:
ml.eia1.medium/large/xlarge, ml.eia2.medium/large/xlarge.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ml.eia1.medium","ml.eia1.large","ml.eia1.xlarge","ml.eia2.medium","ml.eia2.large","ml.eia2.xlarge"]}}

### spec.shadowVariants[].coreDump

`AwsSagemakerEndpointCoreDump`

Write core dumps from a crashed model container to S3.

### spec.shadowVariants[].coreDump.destinationS3Uri

`string` · required

S3 destination for core dumps. Example: "s3://my-dumps/endpoint/"

- rule: {"string":{"minLen":"1","maxLen":"512","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.shadowVariants[].coreDump.kmsKeyArn

`string | valueFrom`

KMS key encrypting the dumps at rest in S3.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.shadowVariants[].mlCapacityReservationArn

`string`

Launch this variant's instances ONLY into the named ML capacity
reservation (the modules send AWS's capacity-reservations-only
preference alongside - its single legal value).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.kmsKeyArn

`string | valueFrom`

KMS key encrypting the ML storage volumes attached to the
endpoint's instances. NOT usable with nitro-local-storage instance
families (ml.g5/g6/p4d/p5 and similar encrypt locally by default
and reject a custom key) or serverless-only endpoints.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.executionRoleArn

`string | valueFrom`

IAM role for the endpoint configuration - REQUIRED when any variant
omits `model` (inference-component endpoints, where components bring
their own models later); otherwise the models' roles apply.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.asyncInference

`AwsSagemakerEndpointAsyncInference`

Queue requests and deliver responses to S3 instead of returning
them synchronously (large payloads, long-running inference).

### spec.asyncInference.outputS3Path

`string` · required

S3 location for inference responses. Example:
"s3://my-bucket/async-out/"

- rule: {"string":{"minLen":"1","maxLen":"512","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.asyncInference.failureS3Path

`string`

S3 location for FAILED inference responses (kept apart from
successes).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"maxLen":"512","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.asyncInference.kmsKeyArn

`string | valueFrom`

KMS key encrypting the async output objects.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.asyncInference.maxConcurrentInvocationsPerInstance

`int32` · optional (explicit presence)

Cap concurrent requests sent to one model container. Omitted = AWS
picks an optimal value (1-1000).

- rule: {"int32":{"lte":1000,"gte":1}}

### spec.asyncInference.successTopicArn

`string | valueFrom`

SNS topic notified on successful inference.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.asyncInference.errorTopicArn

`string | valueFrom`

SNS topic notified on failed inference.

- references: AwsSnsTopic (`status.outputs.topic_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSnsTopic, name: <that resource's name>, fieldPath: status.outputs.topic_arn}} -- a bare string does not parse

### spec.asyncInference.includeInferenceResponseIn

`[]string`

Which notifications carry the full inference response inline:
any of "SUCCESS_NOTIFICATION_TOPIC", "ERROR_NOTIFICATION_TOPIC".

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["SUCCESS_NOTIFICATION_TOPIC","ERROR_NOTIFICATION_TOPIC"]}}}}

### spec.dataCapture

`AwsSagemakerEndpointDataCapture`

Capture request/response payloads to S3 (the Model Monitor feed).

### spec.dataCapture.destinationS3Uri

`string` · required

S3 location for captured data. Example: "s3://my-bucket/capture/"

- rule: {"string":{"minLen":"1","maxLen":"512","pattern":"^(https|s3)://([^/]+)/?(.*)$"}}

### spec.dataCapture.initialSamplingPercentage

`int32`

Percentage of traffic to capture (0-100).

- rule: {"int32":{"lte":100,"gte":0}}

### spec.dataCapture.captureModes

`[]string` · required

Capture request payloads ("Input"), responses ("Output"), or both
in one entry ("InputAndOutput"). 1-2 entries.

- rule: {"repeated":{"minItems":"1","maxItems":"2","unique":true,"items":{"string":{"in":["Input","Output","InputAndOutput"]}}}}

### spec.dataCapture.enableCapture

`bool`

Start with capture ON (flip without replacing the configuration is
not possible - the configuration is immutable, so the modules roll
a new one).

### spec.dataCapture.csvContentTypes

`[]string`

Request content types captured as CSV (e.g. "text/csv"; max 10).
When either content-type list is set, AWS requires at least one
entry across the two.

- rule: {"repeated":{"maxItems":"10","items":{"string":{"minLen":"1","maxLen":"256","pattern":"^[0-9A-Za-z](-*[0-9A-Za-z])*\\/[0-9A-Za-z](-*[0-9A-Za-z.])*$"}}}}

### spec.dataCapture.jsonContentTypes

`[]string`

Request content types captured as JSON (e.g. "application/json";
max 10).

- rule: {"repeated":{"maxItems":"10","items":{"string":{"minLen":"1","maxLen":"256","pattern":"^[0-9A-Za-z](-*[0-9A-Za-z])*\\/[0-9A-Za-z](-*[0-9A-Za-z.])*$"}}}}

### spec.dataCapture.kmsKeyArn

`string | valueFrom`

KMS key encrypting captured data in S3.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.deployment

`AwsSagemakerEndpointDeployment`

How UpdateEndpoint rolls new capacity: blue/green with traffic
shifting (AWS default when omitted) or rolling batches, plus
CloudWatch-alarm auto-rollback. NOTE: AWS rejects a ROLLING policy
on a single-instance fleet at Create/UpdateEndpoint ("Cannot update
endpoint with single instance using RollingUpdatePolicy",
live-caught 2026-08-25) - a one-instance endpoint uses blue/green
or omits deployment; the CEL below front-loads the confirmed
single-variant case.

- rule: exactly one of blue_green and rolling must be set

### spec.deployment.blueGreen

`AwsSagemakerEndpointBlueGreenPolicy`

Blue/green: a full new fleet is provisioned and traffic shifts per
the routing configuration. Exactly one of `blue_green` and
`rolling`.

- rule: canary_size requires traffic_routing_type CANARY
- rule: linear_step_size requires traffic_routing_type LINEAR

### spec.deployment.blueGreen.trafficRoutingType

`string`

"ALL_AT_ONCE" (flip everything after one bake), "CANARY" (a canary
batch first, then the rest), or "LINEAR" (equal steps).

- rule: {"string":{"in":["ALL_AT_ONCE","CANARY","LINEAR"]}}

### spec.deployment.blueGreen.waitIntervalSeconds

`int32`

Seconds between traffic-shift steps (0-3600) - the bake time per
step while rollback alarms are watched.

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.deployment.blueGreen.canarySize

`AwsSagemakerEndpointCapacitySize`

Size of the first (canary) traffic batch - CANARY only; at most 50%
of the variant's capacity.

### spec.deployment.blueGreen.canarySize.type

`string`

"INSTANCE_COUNT" or "CAPACITY_PERCENT".

- rule: {"string":{"in":["INSTANCE_COUNT","CAPACITY_PERCENT"]}}

### spec.deployment.blueGreen.canarySize.value

`int32`

The count or percentage (>= 1).

- rule: {"int32":{"gte":1}}

### spec.deployment.blueGreen.linearStepSize

`AwsSagemakerEndpointCapacitySize`

Size of each LINEAR step - LINEAR only; 10-50% of capacity.

### spec.deployment.blueGreen.linearStepSize.type

`string`

"INSTANCE_COUNT" or "CAPACITY_PERCENT".

- rule: {"string":{"in":["INSTANCE_COUNT","CAPACITY_PERCENT"]}}

### spec.deployment.blueGreen.linearStepSize.value

`int32`

The count or percentage (>= 1).

- rule: {"int32":{"gte":1}}

### spec.deployment.blueGreen.terminationWaitSeconds

`int32` · optional (explicit presence)

Extra seconds the OLD fleet is kept after traffic fully shifts
(0-3600; AWS default 0) - the rollback safety window.

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.deployment.blueGreen.maximumExecutionTimeoutSeconds

`int32` · optional (explicit presence)

Hard ceiling for the whole deployment (600-14400 seconds; must
exceed wait interval + termination wait).

- rule: {"int32":{"lte":14400,"gte":600}}

### spec.deployment.rolling

`AwsSagemakerEndpointRollingPolicy`

Rolling: instances are replaced in batches in place (no parallel
fleet cost). Exactly one of `blue_green` and `rolling`.

### spec.deployment.rolling.maximumBatchSize

`AwsSagemakerEndpointCapacitySize` · required

Batch size per rolling step (5-50% of the fleet when expressed as
CAPACITY_PERCENT).

- rule: {"required":true}

### spec.deployment.rolling.maximumBatchSize.type

`string`

"INSTANCE_COUNT" or "CAPACITY_PERCENT".

- rule: {"string":{"in":["INSTANCE_COUNT","CAPACITY_PERCENT"]}}

### spec.deployment.rolling.maximumBatchSize.value

`int32`

The count or percentage (>= 1).

- rule: {"int32":{"gte":1}}

### spec.deployment.rolling.waitIntervalSeconds

`int32`

Seconds between batches (0-3600) - the bake time while rollback
alarms are watched.

- rule: {"int32":{"lte":3600,"gte":0}}

### spec.deployment.rolling.rollbackMaximumBatchSize

`AwsSagemakerEndpointCapacitySize`

Batch size used when rolling BACK to the old fleet (AWS defaults to
100% - one-shot rollback).

### spec.deployment.rolling.rollbackMaximumBatchSize.type

`string`

"INSTANCE_COUNT" or "CAPACITY_PERCENT".

- rule: {"string":{"in":["INSTANCE_COUNT","CAPACITY_PERCENT"]}}

### spec.deployment.rolling.rollbackMaximumBatchSize.value

`int32`

The count or percentage (>= 1).

- rule: {"int32":{"gte":1}}

### spec.deployment.rolling.maximumExecutionTimeoutSeconds

`int32` · optional (explicit presence)

Hard ceiling for the whole deployment (600-14400 seconds).

- rule: {"int32":{"lte":14400,"gte":600}}

### spec.deployment.autoRollbackAlarmNames

`[]string`

CloudWatch alarms watched during deployment - any alarm firing
rolls the endpoint back (1-10 alarm names).

- rule: {"repeated":{"maxItems":"10","unique":true,"items":{"string":{"minLen":"1"}}}}

## Validation Rules

- `variant_names_unique`: variant names must be unique across production_variants and shadow_variants
- `shadow_one_each_side`: shadow testing requires exactly one production variant and one shadow variant
- `role_required_without_models`: execution_role_arn is required when any variant omits model
- `rolling_requires_multi_instance_fleet`: a rolling deployment policy cannot manage a single-instance endpoint - use two or more instances, a blue_green policy, or omit deployment

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsSagemakerEndpoint, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.endpoint_name` | `string` | The endpoint name (the AWS identity clients invoke). |
| `status.outputs.endpoint_arn` | `string` | The Amazon Resource Name of the endpoint. |
| `status.outputs.endpoint_config_name` | `string` | The name of the endpoint configuration currently in service (the modules roll a new suffixed configuration on capacity changes). |
| `status.outputs.endpoint_config_arn` | `string` | The Amazon Resource Name of that endpoint configuration. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.productionVariants[].model` | AwsSagemakerModel | `status.outputs.model_name` |
| `spec.productionVariants[].coreDump.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.shadowVariants[].model` | AwsSagemakerModel | `status.outputs.model_name` |
| `spec.shadowVariants[].coreDump.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.executionRoleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.asyncInference.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.asyncInference.successTopicArn` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.asyncInference.errorTopicArn` | AwsSnsTopic | `status.outputs.topic_arn` |
| `spec.dataCapture.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## See Also

- [Overview](../README.md)
