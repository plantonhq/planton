# AwsLambda

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsLambdaSpec defines an AWS Lambda function: its code source (an S3
zip archive or an ECR container image), execution environment
(runtime, memory, timeout, ephemeral storage, environment variables),
VPC attachment, integrations (dead-letter queue, EFS mount, X-Ray
tracing, CloudWatch logging), and the function-scoped satellite
surface AWS models as separate resources but that are honestly part
of the function's own configuration: aliases (with weighted routing
and provisioned concurrency), the function URL, resource-policy
invoke permissions, the asynchronous-invocation config, recursion
detection, and runtime-update management.

The function name comes from metadata.name (create-time immutable in
AWS). Everything the function composes with attaches by reference:
the IAM execution role, the subnets and security groups of a VPC
attachment, the KMS key that encrypts environment variables, the
dead-letter queue, the EFS access point, and the CloudWatch log
group. Event sources (SQS, Kinesis, DynamoDB Streams, MSK) attach
through the separate AwsLambdaEventSourceMapping kind, which
references this function -- new mappings never modify the function.

Code-source guidance: S3 zip is the right default for most services
(fast cold starts, per-runtime tooling); a container image is the
right choice for large dependency trees (up to 10 GB), custom
OS-level packages, or teams standardized on image pipelines.

## Example

```yaml
# Canonical validated example: a zip-backed function exercising the full
# modern surface -- versioning with the $LATEST.PUBLISHED head, an alias
# with a qualified function URL and qualified satellites, per-qualifier
# scaling bounds, and the Managed Instances / durable-execution /
# per-tenant-isolation platform arms.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLambda
metadata:
  name: awslambda-demo
spec:
  region: us-west-2
  description: Example Lambda function
  roleArn:
    value: arn:aws:iam::123456789012:role/lambda-basic-exec
  s3:
    bucket:
      value: my-bucket
    key: lambda/hello.zip
  codeSha256: "q0Wep3z1sTImroLBJKcRuBIfqkKing9UVePUFmSpjKQ="
  runtime: nodejs22.x
  handler: index.handler
  architecture: x86_64
  memorySizeMb: 128
  timeoutSeconds: 10
  environment:
    EXAMPLE: "true"
  publish: true
  publishTo: LATEST_PUBLISHED
  managedInstances:
    capacityProviderArn: arn:aws:lambda:us-west-2:123456789012:capacity-provider:steady-fleet
    memoryGibPerVcpu: 4
    maxConcurrencyPerEnvironment: 8
  durableConfig:
    executionTimeoutSeconds: 86400
    retentionPeriodDays: 30
  tenantIsolationMode: PER_TENANT
  aliases:
    - name: live
      functionVersion: "1"
  functionUrl:
    authorizationType: AWS_IAM
    qualifier: live
  invokePermissions:
    - statementId: allow-url-account-callers
      principal: "123456789012"
      action: lambda:InvokeFunctionUrl
      functionUrlAuthType: AWS_IAM
      invokedViaFunctionUrl: true
      qualifier: live
    - statementId: allow-alexa-skill
      principal: alexa-appkit.amazon.com
      eventSourceToken: amzn1.ask.skill.12345678-1234-1234-1234-123456789012
  asyncInvokeConfig:
    maximumRetryAttempts: 1
    qualifier: live
  runtimeManagement:
    updateRuntimeOn: FunctionUpdate
    qualifier: live
  scalingConfigs:
    - qualifier: "$LATEST.PUBLISHED"
      minExecutionEnvironments: 1
      maxExecutionEnvironments: 20
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.s3` | `AwsLambdaS3Code` |  |  |  |
| `spec.s3.bucket` | `string \| valueFrom` | yes |  | AwsS3Bucket (`status.outputs.bucket_id`) |
| `spec.s3.key` | `string` | yes |  |  |
| `spec.s3.objectVersion` | `string` |  |  |  |
| `spec.imageUri` | `string` |  |  |  |
| `spec.sourceCodeHash` | `string` |  |  |  |
| `spec.codeSha256` | `string` |  |  |  |
| `spec.sourceKmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.runtime` | `string` |  |  |  |
| `spec.handler` | `string` |  |  |  |
| `spec.architecture` | `string` |  |  |  |
| `spec.memorySizeMb` | `int32` |  |  |  |
| `spec.timeoutSeconds` | `int32` |  |  |  |
| `spec.ephemeralStorageMb` | `int32` |  |  |  |
| `spec.environment` | `map<string, string>` |  |  |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.subnetIds` | `[]string \| valueFrom` |  |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroupIds` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.ipv6AllowedForDualStack` | `bool` |  |  |  |
| `spec.deadLetterTargetArn` | `string \| valueFrom` |  |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.tracingMode` | `string` |  |  |  |
| `spec.fileSystemConfig` | `AwsLambdaFileSystemConfig` |  |  |  |
| `spec.fileSystemConfig.accessPointArn` | `string \| valueFrom` | yes |  | AwsEfsAccessPoint (`status.outputs.access_point_arn`) |
| `spec.fileSystemConfig.localMountPath` | `string` | yes |  |  |
| `spec.imageConfig` | `AwsLambdaImageConfig` |  |  |  |
| `spec.imageConfig.entryPoint` | `[]string` |  |  |  |
| `spec.imageConfig.command` | `[]string` |  |  |  |
| `spec.imageConfig.workingDirectory` | `string` |  |  |  |
| `spec.layerArns` | `[]string \| valueFrom` |  |  |  |
| `spec.publish` | `bool` |  |  |  |
| `spec.publishTo` | `string` |  |  |  |
| `spec.reservedConcurrentExecutions` | `int32` |  |  |  |
| `spec.snapStart` | `bool` |  |  |  |
| `spec.managedInstances` | `AwsLambdaManagedInstances` |  |  |  |
| `spec.managedInstances.capacityProviderArn` | `string` | yes |  |  |
| `spec.managedInstances.memoryGibPerVcpu` | `double` |  |  |  |
| `spec.managedInstances.maxConcurrencyPerEnvironment` | `int32` |  |  |  |
| `spec.durableConfig` | `AwsLambdaDurableConfig` |  |  |  |
| `spec.durableConfig.executionTimeoutSeconds` | `int32` |  |  |  |
| `spec.durableConfig.retentionPeriodDays` | `int32` |  |  |  |
| `spec.tenantIsolationMode` | `string` |  |  |  |
| `spec.loggingConfig` | `AwsLambdaLoggingConfig` |  |  |  |
| `spec.loggingConfig.logFormat` | `string` |  |  |  |
| `spec.loggingConfig.applicationLogLevel` | `string` |  |  |  |
| `spec.loggingConfig.systemLogLevel` | `string` |  |  |  |
| `spec.loggingConfig.logGroup` | `string \| valueFrom` |  |  | AwsCloudwatchLogGroup (`status.outputs.log_group_name`) |
| `spec.codeSigningConfigArn` | `string` |  |  |  |
| `spec.aliases` | `[]AwsLambdaAlias` |  |  |  |
| `spec.aliases[].name` | `string` | yes |  |  |
| `spec.aliases[].description` | `string` |  |  |  |
| `spec.aliases[].functionVersion` | `string` | yes |  |  |
| `spec.aliases[].routingAdditionalVersionWeights` | `map<string, double>` |  |  |  |
| `spec.aliases[].provisionedConcurrentExecutions` | `int32` |  |  |  |
| `spec.functionUrl` | `AwsLambdaFunctionUrl` |  |  |  |
| `spec.functionUrl.authorizationType` | `string` | yes |  |  |
| `spec.functionUrl.invokeMode` | `string` |  |  |  |
| `spec.functionUrl.cors` | `AwsLambdaFunctionUrlCors` |  |  |  |
| `spec.functionUrl.cors.allowCredentials` | `bool` |  |  |  |
| `spec.functionUrl.cors.allowOrigins` | `[]string` |  |  |  |
| `spec.functionUrl.cors.allowMethods` | `[]string` |  |  |  |
| `spec.functionUrl.cors.allowHeaders` | `[]string` |  |  |  |
| `spec.functionUrl.cors.exposeHeaders` | `[]string` |  |  |  |
| `spec.functionUrl.cors.maxAgeSeconds` | `int32` |  |  |  |
| `spec.functionUrl.qualifier` | `string` |  |  |  |
| `spec.invokePermissions` | `[]AwsLambdaInvokePermission` |  |  |  |
| `spec.invokePermissions[].statementId` | `string` | yes |  |  |
| `spec.invokePermissions[].principal` | `string` | yes |  |  |
| `spec.invokePermissions[].action` | `string` |  |  |  |
| `spec.invokePermissions[].sourceArn` | `string` |  |  |  |
| `spec.invokePermissions[].sourceAccount` | `string` |  |  |  |
| `spec.invokePermissions[].principalOrgId` | `string` |  |  |  |
| `spec.invokePermissions[].functionUrlAuthType` | `string` |  |  |  |
| `spec.invokePermissions[].qualifier` | `string` |  |  |  |
| `spec.invokePermissions[].eventSourceToken` | `string` |  |  |  |
| `spec.invokePermissions[].invokedViaFunctionUrl` | `bool` |  |  |  |
| `spec.asyncInvokeConfig` | `AwsLambdaAsyncInvokeConfig` |  |  |  |
| `spec.asyncInvokeConfig.maximumRetryAttempts` | `int32` |  |  |  |
| `spec.asyncInvokeConfig.maximumEventAgeSeconds` | `int32` |  |  |  |
| `spec.asyncInvokeConfig.onSuccessDestinationArn` | `string \| valueFrom` |  |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.asyncInvokeConfig.onFailureDestinationArn` | `string \| valueFrom` |  |  | AwsSqsQueue (`status.outputs.queue_arn`) |
| `spec.asyncInvokeConfig.qualifier` | `string` |  |  |  |
| `spec.recursiveLoop` | `string` |  |  |  |
| `spec.runtimeManagement` | `AwsLambdaRuntimeManagement` |  |  |  |
| `spec.runtimeManagement.updateRuntimeOn` | `string` | yes |  |  |
| `spec.runtimeManagement.runtimeVersionArn` | `string` |  |  |  |
| `spec.runtimeManagement.qualifier` | `string` |  |  |  |
| `spec.scalingConfigs` | `[]AwsLambdaScalingConfig` |  |  |  |
| `spec.scalingConfigs[].qualifier` | `string` | yes |  |  |
| `spec.scalingConfigs[].minExecutionEnvironments` | `int32` |  |  |  |
| `spec.scalingConfigs[].maxExecutionEnvironments` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region the function is created in. Must match the region
of the S3 code bucket, any VPC subnets, and the EFS access point
it references.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Free-form description shown in the AWS Console -- operational
context for humans browsing the account. Up to 256 characters.

- rule: {"string":{"maxLen":"256"}}

### spec.roleArn

`string | valueFrom` · required

The IAM execution role the function assumes: it must trust
lambda.amazonaws.com and carry the policies the code needs
(CloudWatch Logs at minimum -- AWSLambdaBasicExecutionRole; plus
AWSLambdaVPCAccessExecutionRole for VPC attachment). Roles own
their policies -- this module never attaches policies to a role it
merely references. Reference an AwsIamRole role_arn output or pass
a literal role ARN.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.s3

`AwsLambdaS3Code`

The deployment package as a zip archive in S3. Requires runtime
and handler. Prefer a bucket in the function's own region --
cross-region pulls are slower and billed.

### spec.s3.bucket

`string | valueFrom` · required

The S3 bucket holding the deployment package. Must be in the
function's region. Reference an AwsS3Bucket bucket_id output or
pass a literal bucket name.

- references: AwsS3Bucket (`status.outputs.bucket_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsS3Bucket, name: <that resource's name>, fieldPath: status.outputs.bucket_id}} -- a bare string does not parse

### spec.s3.key

`string` · required

The object key (path) of the deployment package zip.

- rule: {"string":{"minLen":"1"}}

### spec.s3.objectVersion

`string`

Pin a specific object version (versioned buckets). Empty deploys
the current version; pair with source_code_hash for fully
declarative code rolls.

### spec.imageUri

`string`

The function code as a container image in ECR, e.g.
"123456789012.dkr.ecr.us-east-1.amazonaws.com/repo:tag". The image
defines the runtime and entrypoint, so runtime and handler must
stay empty; image_config below can override the entrypoint. Images
may be up to 10 GB -- the right choice for heavy dependency trees.

Lambda pulls only from a private ECR repository in the same account
and Region as the function; there is no registry-login field and AWS
accepts none. For an image that lives elsewhere (GHCR, Docker Hub,
another Region), push it to ECR or declare a pull-through cache rule on
AwsEcrRegistrySettings and point this URI at the cached repository.

### spec.sourceCodeHash

`string`

Base64-encoded SHA256 of the deployment package. Set it (usually
from your build pipeline) to make code updates declarative: a new
hash rolls the function, an unchanged hash is a no-op even when
the S3 object is rewritten in place. Leave empty to update only
when the S3 key or object version changes.

### spec.codeSha256

`string`

Base64-encoded SHA256 of the DEPLOYED package as AWS reports it
(the digest GetFunction returns). Set it to detect and roll
out-of-band code changes from the deployed artifact's own digest —
the deploy-side complement to source_code_hash, which hashes the
artifact you upload. Most configurations want source_code_hash;
use this when the artifact is published by a pipeline you don't
control and only the deployed digest is known.

### spec.sourceKmsKeyArn

`string | valueFrom`

The KMS key that encrypts the deployment package in S3 (bring-
your-own-key for the code artifact itself -- distinct from
kms_key_arn, which encrypts environment variables). Only
meaningful for zip deployments. Reference an AwsKmsKey key_arn
output or pass a literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.runtime

`string`

The language runtime for zip deployments, e.g. "nodejs22.x",
"python3.13", "java21", "dotnet8", "ruby3.4", "provided.al2023"
(custom runtimes / native binaries). AWS retires runtimes on its
own schedule and adds new ones frequently, so the accepted set is
validated by AWS at deploy time rather than frozen here. Required
for zip code; must stay empty for container images (the image
carries its own runtime).

### spec.handler

`string`

The function entrypoint for zip deployments. Format is
runtime-specific: "index.handler" (Node.js), "module.function"
(Python), "package.Class::method" (Java), "bootstrap" (custom
runtimes). Required for zip code; must stay empty for container
images (the image CMD/ENTRYPOINT defines it).

- rule: {"string":{"maxLen":"128"}}

### spec.architecture

`string`

The instruction-set architecture: "x86_64" or "arm64". Empty keeps
the AWS default (x86_64). arm64 (Graviton) is typically ~20%
cheaper per GB-second and often faster -- prefer it whenever your
runtime and native dependencies support it.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["x86_64","arm64"]}}

### spec.memorySizeMb

`int32`

Memory in MB: 128-10240 for standard functions, up to 32768 when
the function runs on Lambda Managed Instances (managed_instances)
-- AWS enforces the per-platform ceiling at deploy time. CPU and
network scale linearly with memory (a full vCPU arrives around
1769 MB), so raising memory is also how you buy CPU -- for
CPU-bound code a larger size often costs LESS overall by finishing
sooner. 0 keeps the AWS default (128 MB).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":32768,"gte":128}}

### spec.timeoutSeconds

`int32`

Maximum execution time per invocation in seconds, 1-900. Size it
slightly above the worst expected runtime; API-fronted functions
should stay well under their gateway's timeout. 0 keeps the AWS
default (3 seconds).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":900,"gte":1}}

### spec.ephemeralStorageMb

`int32`

Scratch space at /tmp in MB, 512-10240. Sized for workloads that
stage files locally (media processing, ML model unpacking). Billed
above the free 512 MB. 0 keeps the AWS default (512 MB).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":10240,"gte":512}}

### spec.environment

`map<string, string>`

Plain-configuration environment variables available to the code at
runtime. Never put secret material here -- environment variables
are visible to anyone who can read the function configuration.
Give the execution role access to SSM Parameter Store or Secrets
Manager and resolve secrets at runtime instead.

### spec.kmsKeyArn

`string | valueFrom`

The customer-managed KMS key that encrypts the environment
variables at rest (and SnapStart snapshots). Empty uses the
AWS-managed key. Reference an AwsKmsKey key_arn output or pass a
literal key ARN.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.subnetIds

`[]string | valueFrom`

Subnets the function's ENIs are created in -- typically private
subnets across at least two availability zones. Attaching to a VPC
gives the code access to private resources (databases, caches) and
removes default internet access (route through a NAT gateway to
restore it). Leave empty to run outside any VPC (the default, with
direct internet access). Reference AwsSubnet subnet_id outputs or
pass literal subnet IDs.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroupIds

`[]string | valueFrom`

Security groups attached to the function's ENIs. Outbound rules
must allow the services the code reaches (databases, AWS
endpoints, the internet via NAT). Required together with
subnet_ids when attaching to a VPC. Reference AwsSecurityGroup
security_group_id outputs or pass literal security group IDs.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.ipv6AllowedForDualStack

`bool`

Allow the VPC-attached function to make outbound IPv6 connections
over dual-stack subnets. Only meaningful with a VPC attachment.

### spec.deadLetterTargetArn

`string | valueFrom`

Where asynchronous invocations that exhaust their retries are
sent: an SQS queue or SNS topic ARN. The execution role needs
sqs:SendMessage / sns:Publish on the target. Reference an
AwsSqsQueue queue_arn output, or pass an SNS topic ARN as a
literal (or an explicit-kind reference to AwsSnsTopic topic_arn).
For finer-grained routing (separate success/failure destinations,
max event age), use async_invoke_config instead.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.tracingMode

`string`

AWS X-Ray tracing: "Active" (trace and sample invocations --
requires xray:PutTraceSegments on the execution role) or
"PassThrough" (only forward upstream trace headers). Empty keeps
the AWS default (PassThrough).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Active","PassThrough"]}}

### spec.fileSystemConfig

`AwsLambdaFileSystemConfig`

Mount an EFS access point into the execution environment --
durable shared storage across invocations and functions (ML
models, shared caches). Requires a VPC attachment reaching the
file system's mount targets.

### spec.fileSystemConfig.accessPointArn

`string | valueFrom` · required

The EFS ACCESS POINT ARN (not the file system ARN). Reference an
AwsEfsAccessPoint resource or pass a literal access point ARN.

- references: AwsEfsAccessPoint (`status.outputs.access_point_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsEfsAccessPoint, name: <that resource's name>, fieldPath: status.outputs.access_point_arn}} -- a bare string does not parse

### spec.fileSystemConfig.localMountPath

`string` · required

Where the file system appears inside the execution environment.
Must be under /mnt, e.g. "/mnt/data".

- rule: {"string":{"minLen":"1","pattern":"^/mnt/[a-zA-Z0-9-_.]+$"}}

### spec.imageConfig

`AwsLambdaImageConfig`

Container-image entrypoint overrides. Only meaningful with
image_uri.

### spec.imageConfig.entryPoint

`[]string`

Override the image ENTRYPOINT.

### spec.imageConfig.command

`[]string`

Override the image CMD (the handler argument for AWS base images).

### spec.imageConfig.workingDirectory

`string`

Override the image working directory.

### spec.layerArns

`[]string | valueFrom`

Lambda layer version ARNs merged into the execution environment,
in order (later layers shadow earlier ones), up to five. Layers
carry shared dependencies and tooling outside the deployment
package. Pass literal layer-version ARNs, e.g.
"arn:aws:lambda:us-west-2:123456789012:layer:shared-libs:3".

- rule: {"repeated":{"maxItems":"5"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.publish

`bool`

Publish a new immutable version on every code or configuration
change. Versions are what aliases route to -- enable this when
using aliases for traffic shifting or SnapStart (which only
applies to published versions).

### spec.publishTo

`string`

Maintain the "$LATEST.PUBLISHED" head pointer: "LATEST_PUBLISHED"
(the only value AWS currently accepts) keeps a moving qualifier
that always resolves to the newest published version -- the
addressable target scaling_configs and qualified invocations can
pin without naming version numbers. Empty leaves the head pointer
unmanaged. Rollout caveat (live-verified us-west-2, 2026-08-11):
regions/accounts where the $LATEST.PUBLISHED feature has not rolled
out reject the value at CreateFunction with
InvalidParameterValueException ("isn't a valid value for this
field") -- and AWS creates the function BEFORE rejecting this
parameter, so a failed create can leave a live function behind.
Mid-rollout is subtler (live-verified us-west-2, 2026-08-13):
CreateFunction can ACCEPT the value yet silently not maintain the
head -- the "$LATEST.PUBLISHED" qualifier answers
ResourceNotFoundException even after versions publish -- while
UpdateFunctionCode still rejects the same value. Acceptance at
create is NOT availability; the feature is available only where
the qualifier actually resolves after a publish. Set this field
only where that holds.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["LATEST_PUBLISHED"]}}

### spec.reservedConcurrentExecutions

`int32` · optional (explicit presence)

Reserved concurrency for this function. Unset: the function draws
from the account's unreserved pool (no dedicated cap). 0: all
invocations are throttled -- an operational kill switch. Positive:
that many concurrent executions are carved out of the account
pool, acting as both a guarantee and a ceiling.

- rule: {"int32":{"gte":0}}

### spec.snapStart

`bool`

SnapStart: resume new execution environments from a pre-initialized
snapshot instead of running init from scratch -- order-of-magnitude
cold-start reduction for JVM and other slow-init runtimes. Applies
to published versions only (enable publish and invoke through a
version or alias to benefit).

### spec.managedInstances

`AwsLambdaManagedInstances`

Run the function on a Lambda Managed Instances capacity provider --
dedicated EC2 capacity AWS manages on the function's behalf --
instead of the on-demand fleet. The platform for steady
high-throughput workloads, memory above 10 GB, and per-tenant
isolation.

### spec.managedInstances.capacityProviderArn

`string` · required

The Lambda capacity provider supplying the managed EC2 capacity.
Pass the capacity provider ARN as a literal, e.g.
"arn:aws:lambda:us-west-2:123456789012:capacity-provider:my-cp".

- rule: {"required":true}

### spec.managedInstances.memoryGibPerVcpu

`double`

Memory (GiB) provisioned per vCPU in each execution environment --
how compute-heavy vs memory-heavy the environments are sized.
0 keeps the AWS default sizing.

### spec.managedInstances.maxConcurrencyPerEnvironment

`int32`

Maximum concurrent invocations one execution environment may serve.
0 keeps the AWS default.

- rule: {"int32":{"gte":0}}

### spec.durableConfig

`AwsLambdaDurableConfig`

Durable execution: AWS checkpoints the function's progress so
long-running workflows survive interruption and resume where they
stopped, far beyond the classic 15-minute cap. ADDING OR REMOVING
this block REPLACES the function (an AWS constraint); the values
inside update in place.

### spec.durableConfig.executionTimeoutSeconds

`int32`

Maximum end-to-end time of one durable invocation in seconds,
1-31622400 (up to 366 days) -- checkpointing is what lets it far
exceed the classic 15-minute cap.

- rule: {"int32":{"lte":31622400,"gte":1}}

### spec.durableConfig.retentionPeriodDays

`int32`

How long (days, 1-90) AWS retains each durable execution's state
and history after it completes. 0 keeps the AWS default (14).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":90,"gte":1}}

### spec.tenantIsolationMode

`string`

Isolate execution environments per tenant: "PER_TENANT" (the only
mode AWS currently accepts) dedicates environments to the tenant id
callers pass at invoke time -- no cross-tenant reuse of warm
state. Create-time immutable: CHANGING it REPLACES the function.
Empty keeps the AWS default (environments shared across all
invocations).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["PER_TENANT"]}}

### spec.loggingConfig

`AwsLambdaLoggingConfig`

CloudWatch Logs delivery. Unset, AWS creates and writes to
"/aws/lambda/<function-name>" in plain-text format with the log
group owned by AWS (it survives function deletion). Configure to
switch to structured JSON, tune log levels, or write into a log
group you manage (retention, encryption, subscription filters).

- rule: application_log_level and system_log_level require log_format JSON -- plain-text logs carry no level field to filter on

### spec.loggingConfig.logFormat

`string`

"Text" (plain lines, the AWS default) or "JSON" (structured
records with built-in level filtering). Empty keeps the AWS
default. Level filtering below requires JSON.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Text","JSON"]}}

### spec.loggingConfig.applicationLogLevel

`string`

Minimum level of application (your code's) log records delivered:
"TRACE", "DEBUG", "INFO", "WARN", "ERROR", or "FATAL". JSON format
only. Empty delivers everything.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["TRACE","DEBUG","INFO","WARN","ERROR","FATAL"]}}

### spec.loggingConfig.systemLogLevel

`string`

Minimum level of system (Lambda platform) log records delivered:
"DEBUG", "INFO", or "WARN". JSON format only. Empty delivers
everything.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["DEBUG","INFO","WARN"]}}

### spec.loggingConfig.logGroup

`string | valueFrom`

Write into a log group you manage -- for retention policy,
KMS encryption, or subscription filters -- instead of the
AWS-created "/aws/lambda/<function-name>". Reference an
AwsCloudwatchLogGroup log_group_name output or pass a literal log
group name. The execution role needs logs:CreateLogStream and
logs:PutLogEvents on it.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_name}} -- a bare string does not parse

### spec.codeSigningConfigArn

`string`

A code-signing configuration ARN: AWS then rejects deployment
packages whose signatures don't match the trusted signing
profiles. Pass a literal ARN, e.g.
"arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-...".
Only meaningful for zip deployments.

### spec.aliases

`[]AwsLambdaAlias`

Named pointers to published versions, each optionally splitting
traffic between two versions (canary) and pre-warming provisioned
concurrency. Aliases are the stable invocation targets clients and
event sources should reference -- repointing an alias is how you
ship or roll back without touching callers. Each entry
materializes as its own alias resource keyed by name, so edits are
in-place.

- rule: AWS rejects provisioned concurrency on a weighted (canary) alias -- point the alias at a single version or drop the routing weights
- rule: AWS allows at most one additional version in an alias's routing config
- rule: routing weights are fractions of this alias's traffic -- each must be between 0.0 and 1.0
- rule: provisioned concurrency cannot target $LATEST -- point the alias at a published version

### spec.aliases[].name

`string` · required

The alias name, e.g. "live", "staging". Unique within the
function.

- rule: {"string":{"minLen":"1","maxLen":"128","pattern":"^[a-zA-Z0-9-_]+$"}}

### spec.aliases[].description

`string`

Free-form description of what this alias routes.

### spec.aliases[].functionVersion

`string` · required

The published version this alias points to, e.g. "1", or "$LATEST"
(unpublished head -- fine for dev aliases, avoid for production).
Numbering caveat (live-verified 2026-08-11): AWS never reuses
version numbers for a function NAME, even across delete/recreate --
a recreated function's first publish continues the old numbering,
so a literal pin like "1" that worked on the first deployment 404s
at CreateAlias ("Function not found ...:1") on a recreate. Pin
literal numbers only against a function whose publish history you
know; use "$LATEST" where the alias just needs to exist.

- rule: {"string":{"minLen":"1"}}

### spec.aliases[].routingAdditionalVersionWeights

`map<string, double>`

Canary routing: additional version(s) receiving a fraction of this
alias's traffic, as version -> weight (0.0-1.0). E.g. {"2": 0.1}
sends 10% of traffic to version 2 and 90% to function_version.
AWS allows at most one additional version.

### spec.aliases[].provisionedConcurrentExecutions

`int32` · optional (explicit presence)

Pre-warmed execution environments kept ready for this alias --
eliminates cold starts at the cost of paying for idle warmth.
Applied as a provisioned-concurrency config keyed by this alias.
Unset means no provisioned concurrency. AWS only allows provisioned
concurrency on an alias that points at exactly one published version:
not on a weighted (canary) alias, and not on $LATEST.

- rule: {"int32":{"gte":1}}

### spec.functionUrl

`AwsLambdaFunctionUrl`

A built-in HTTPS endpoint for the function -- the zero-
infrastructure alternative to an API gateway for simple HTTP
services and webhooks.

### spec.functionUrl.authorizationType

`string` · required

Who may invoke the URL: "AWS_IAM" (callers sign requests with
SigV4 -- the safe default) or "NONE" (public -- anyone with the
URL; AWS still requires an explicit public invoke permission,
which the module manages). Choose NONE only for genuinely public
webhooks and pair it with your own request validation.

- rule: {"required":true,"string":{"in":["AWS_IAM","NONE"]}}

### spec.functionUrl.invokeMode

`string`

"BUFFERED" (the default -- response returned whole, up to 6 MB) or
"RESPONSE_STREAM" (stream the response as it is produced, up to
20 MB soft cap -- for large payloads and time-to-first-byte).
Empty keeps the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BUFFERED","RESPONSE_STREAM"]}}

### spec.functionUrl.cors

`AwsLambdaFunctionUrlCors`

Cross-origin resource sharing for browser callers.

### spec.functionUrl.cors.allowCredentials

`bool`

Allow credentials (cookies, authorization headers) in cross-origin
requests.

### spec.functionUrl.cors.allowOrigins

`[]string`

Origins allowed to call the URL, e.g. "https://app.example.com";
"*" allows all.

### spec.functionUrl.cors.allowMethods

`[]string`

HTTP methods allowed, e.g. "GET", "POST"; "*" allows all.

### spec.functionUrl.cors.allowHeaders

`[]string`

Request headers allowed, e.g. "content-type", "authorization".

### spec.functionUrl.cors.exposeHeaders

`[]string`

Response headers exposed to browser scripts.

### spec.functionUrl.cors.maxAgeSeconds

`int32`

How long (seconds) browsers may cache the preflight response,
up to 86400 (24h). 0 keeps the AWS default (0 -- no caching).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":86400,"gte":1}}

### spec.functionUrl.qualifier

`string`

Attach the URL to one of this spec's aliases instead of $LATEST --
the URL then serves whatever version the alias routes (including
canary weights), so traffic shifting applies to URL callers too.
Must name an entry in aliases; empty serves the unpublished head.

- rule: {"string":{"maxLen":"128","pattern":"^[a-zA-Z0-9-_]*$"}}

### spec.invokePermissions

`[]AwsLambdaInvokePermission`

Resource-policy statements granting other principals and services
permission to invoke this function -- how S3 buckets, SNS topics,
EventBridge rules, and other accounts are authorized to call it.
Each entry materializes as its own permission statement keyed by
statement_id, so list edits add/remove statements in place.

- rule: function_url_auth_type only applies when action is lambda:InvokeFunctionUrl

### spec.invokePermissions[].statementId

`string` · required

Unique statement identifier within the function's policy, e.g.
"allow-s3-uploads-bucket". The per-name key each entry
materializes under.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^[a-zA-Z0-9-_]+$"}}

### spec.invokePermissions[].principal

`string` · required

Who is allowed to invoke: a service principal (e.g.
"s3.amazonaws.com", "sns.amazonaws.com", "events.amazonaws.com"),
an AWS account ID, or an IAM principal ARN.

- rule: {"string":{"minLen":"1"}}

### spec.invokePermissions[].action

`string`

The Lambda action granted. Empty keeps the sensible default
("lambda:InvokeFunction"); "lambda:InvokeFunctionUrl" grants URL
invocation instead.

### spec.invokePermissions[].sourceArn

`string`

Scope a service-principal grant to one source resource, e.g. the
S3 bucket ARN or SNS topic ARN that may invoke. Strongly
recommended for service principals -- without it ANY resource of
that service in ANY account with knowledge of the function name
could invoke (the confused-deputy problem).

### spec.invokePermissions[].sourceAccount

`string`

Scope a service-principal grant to sources owned by this account
ID -- the coarser companion to source_arn.

### spec.invokePermissions[].principalOrgId

`string`

Grant access to every account in an AWS Organization by org ID,
e.g. "o-a1b2c3d4e5".

### spec.invokePermissions[].functionUrlAuthType

`string`

Required auth type when granting "lambda:InvokeFunctionUrl":
"AWS_IAM" or "NONE" (the public-URL grant).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["AWS_IAM","NONE"]}}

### spec.invokePermissions[].qualifier

`string`

Scope the grant to one qualified ARN: a published version number
("3") or an alias name ("live"). The principal may then invoke only
that version/alias. Empty grants on the unqualified function.

- rule: {"string":{"maxLen":"128"}}

### spec.invokePermissions[].eventSourceToken

`string`

Token the caller must present with the invocation -- used by Alexa
Skills (the skill id) to pin the grant to one skill.

- rule: {"string":{"maxLen":"256"}}

### spec.invokePermissions[].invokedViaFunctionUrl

`bool`

Restrict the grant to invocations that arrive THROUGH the function
URL (rejects direct Invoke calls under this statement).

### spec.asyncInvokeConfig

`AwsLambdaAsyncInvokeConfig`

Delivery and retry behavior for asynchronous invocations (S3
events, SNS, EventBridge): retry count, event age, and on-success
/ on-failure destinations -- the richer successor to the
dead-letter queue.

### spec.asyncInvokeConfig.maximumRetryAttempts

`int32` · optional (explicit presence)

Retries after a failed asynchronous invocation, 0-2. Unset keeps
the AWS default (2).

- rule: {"int32":{"lte":2,"gte":0}}

### spec.asyncInvokeConfig.maximumEventAgeSeconds

`int32`

How long (seconds) an event may wait in the internal queue before
Lambda discards it, 60-21600. 0 keeps the AWS default (21600 --
6 hours).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":21600,"gte":60}}

### spec.asyncInvokeConfig.onSuccessDestinationArn

`string | valueFrom`

Where successful invocation records are delivered: an SQS queue,
SNS topic, EventBridge bus, or Lambda function ARN. Reference an
AwsSqsQueue queue_arn output (or an explicit-kind reference /
literal ARN for the other targets). The execution role needs send
rights on the destination.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.asyncInvokeConfig.onFailureDestinationArn

`string | valueFrom`

Where failed invocation records are delivered after retries are
exhausted -- same target types as on_success_destination_arn.

- references: AwsSqsQueue (`status.outputs.queue_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSqsQueue, name: <that resource's name>, fieldPath: status.outputs.queue_arn}} -- a bare string does not parse

### spec.asyncInvokeConfig.qualifier

`string`

Apply this config to one qualified scope instead of the whole
function: a published version number ("3") or an alias name
("live"). Async invocations through other qualifiers then keep the
AWS defaults. Empty applies at function scope.

- rule: {"string":{"maxLen":"128"}}

### spec.recursiveLoop

`string`

Recursive-loop detection: "Terminate" (the AWS default -- stop
runaway self-invocation loops automatically) or "Allow" (opt out
for workloads that legitimately self-invoke, e.g. intentional
fan-out through SQS). Empty keeps the AWS default.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["Allow","Terminate"]}}

### spec.runtimeManagement

`AwsLambdaRuntimeManagement`

How runtime patches roll out to the function: automatically, only
on function updates, or pinned to a specific runtime version.

- rule: update_runtime_on Manual requires runtime_version_arn (the exact runtime build to pin); other modes must leave it empty

### spec.runtimeManagement.updateRuntimeOn

`string` · required

"Auto" (the AWS default -- receive runtime patches as AWS releases
them), "FunctionUpdate" (patches apply only when the function is
next updated -- change-window control), or "Manual" (pin to the
exact runtime version in runtime_version_arn -- an emergency
rollback tool, not a steady state).

- rule: {"required":true,"string":{"in":["Auto","FunctionUpdate","Manual"]}}

### spec.runtimeManagement.runtimeVersionArn

`string`

The exact runtime version ARN to pin, required with (and only
meaningful for) "Manual". Obtain it from the function's runtime
update events or the Lambda console.

### spec.runtimeManagement.qualifier

`string`

Apply the policy to one qualified scope instead of the whole
function: a published version number ("3") or an alias name
("live"). Empty applies at function scope ($LATEST).

- rule: {"string":{"maxLen":"128"}}

### spec.scalingConfigs

`[]AwsLambdaScalingConfig`

Per-qualifier execution-environment scaling bounds for functions on
Lambda Managed Instances: pin a published version's (or the
"$LATEST.PUBLISHED" head's) environment fleet between a floor and a
ceiling. Each entry materializes as its own scaling-config resource
keyed by qualifier, so list edits update in place.

- rule: set at least one of min_execution_environments or max_execution_environments -- an empty scaling config is a reset, not a resource
- rule: max_execution_environments must be >= min_execution_environments

### spec.scalingConfigs[].qualifier

`string` · required

The qualifier the bounds apply to: a published version number
("3") or "$LATEST.PUBLISHED" (the newest published version --
maintained by publish_to). Alias names are NOT accepted here (an
AWS constraint specific to scaling configs).

- rule: {"required":true,"string":{"pattern":"^(\\$LATEST\\.PUBLISHED|[0-9]+)$"}}

### spec.scalingConfigs[].minExecutionEnvironments

`int32` · optional (explicit presence)

The floor of always-provisioned execution environments. Unset
leaves the floor to AWS.

- rule: {"int32":{"gte":0}}

### spec.scalingConfigs[].maxExecutionEnvironments

`int32` · optional (explicit presence)

The ceiling of execution environments AWS may scale to. Unset
leaves the ceiling to AWS.

- rule: {"int32":{"gte":1}}

## Validation Rules

- `exactly_one_code_source`: provide the function code as exactly one of s3 (zip archive) or image_uri (container image)
- `zip_requires_runtime_and_handler`: zip deployments (s3) require both runtime and handler -- they tell Lambda how to start your code
- `image_forbids_runtime_and_handler`: container-image deployments must leave runtime and handler empty -- the image defines both (use image_config to override the entrypoint)
- `image_config_requires_image`: image_config only applies to container-image deployments (image_uri)
- `source_kms_key_is_zip_only`: source_kms_key_arn encrypts the S3 deployment package and only applies to zip deployments
- `code_signing_is_zip_only`: code_signing_config_arn only applies to zip deployments -- container images are trusted through ECR
- `vpc_fields_travel_together`: VPC attachment requires both subnet_ids and security_group_ids; ipv6_allowed_for_dual_stack only applies with a VPC attachment
- `file_system_requires_vpc`: mounting an EFS access point requires a VPC attachment that reaches the file system's mount targets (subnet_ids + security_group_ids)
- `snap_start_requires_publish`: SnapStart only applies to published versions -- enable publish so versions exist to snapshot
- `aliases_require_publish`: aliases point at published versions -- enable publish so versions exist to route to
- `alias_names_unique`: each alias name must be unique -- aliases materialize per-name
- `permission_statement_ids_unique`: each invoke permission statement_id must be unique -- permissions materialize per statement
- `function_url_qualifier_names_alias`: function_url.qualifier must name one of this spec's aliases -- AWS only qualifies function URLs by alias
- `scaling_config_qualifiers_unique`: each scaling config qualifier must be unique -- configs materialize per qualifier

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsLambda, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.function_arn` | `string` | The function ARN -- the join key for event-source mappings, trigger configurations (Cognito, Firehose), resource policies, and IAM policy resources. |
| `status.outputs.function_name` | `string` | The function name -- what SDK invoke calls and CLI commands reference. |
| `status.outputs.invoke_arn` | `string` | The ARN AWS service integrations invoke through (the apigateway-shaped invocation ARN) -- what API Gateway integrations reference. |
| `status.outputs.qualified_arn` | `string` | The qualified ARN of the most recently published version (empty when publish is disabled). |
| `status.outputs.version` | `string` | The most recently published version number (empty when publish is disabled). |
| `status.outputs.function_url` | `string` | The HTTPS endpoint of the function URL (empty when no function URL is configured). |
| `status.outputs.alias_arns` | `map<string, string>` | ARNs of the function's aliases, keyed by alias name -- the stable invocation targets clients reference for traffic-shifted rollouts. Example valueFrom: status.outputs.alias_arns.live |
| `status.outputs.log_group_name` | `string` | The CloudWatch log group receiving the function's logs -- the AWS-default "/aws/lambda/<function-name>" or the custom group from logging_config.log_group. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.s3.bucket` | AwsS3Bucket | `status.outputs.bucket_id` |
| `spec.sourceKmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |
| `spec.subnetIds` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroupIds` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.deadLetterTargetArn` | AwsSqsQueue | `status.outputs.queue_arn` |
| `spec.fileSystemConfig.accessPointArn` | AwsEfsAccessPoint | `status.outputs.access_point_arn` |
| `spec.loggingConfig.logGroup` | AwsCloudwatchLogGroup | `status.outputs.log_group_name` |
| `spec.asyncInvokeConfig.onSuccessDestinationArn` | AwsSqsQueue | `status.outputs.queue_arn` |
| `spec.asyncInvokeConfig.onFailureDestinationArn` | AwsSqsQueue | `status.outputs.queue_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAppSyncApi | `spec.graphql.auth.lambda.authorizerUri` | `status.outputs.function_arn` |
| AwsAppSyncApi | `spec.graphql.additionalAuthProviders[].lambda.authorizerUri` | `status.outputs.function_arn` |
| AwsAppSyncApi | `spec.graphql.functions[].syncConfig.lambdaConflictHandlerArn` | `status.outputs.function_arn` |
| AwsAppSyncApi | `spec.graphql.resolvers[].syncConfig.lambdaConflictHandlerArn` | `status.outputs.function_arn` |
| AwsAppSyncApi | `spec.events.authProviders[].lambda.authorizerUri` | `status.outputs.function_arn` |
| AwsAppSyncApi | `spec.datasources[].lambda.functionArn` | `status.outputs.function_arn` |
| AwsBedrockAgent | `spec.promptOverride.overrideLambda` | `status.outputs.function_arn` |
| AwsBedrockAgent | `spec.actionGroups[].executor.lambda` | `status.outputs.function_arn` |
| AwsBedrockAgentCoreEvaluation | `spec.evaluators[].codeBased.lambdaArn` | `status.outputs.function_arn` |
| AwsBedrockAgentCoreGateway | `spec.interceptors[].lambdaArn` | `status.outputs.function_arn` |
| AwsBedrockAgentCoreGateway | `spec.targets[].backend.lambda.lambdaArn` | `status.outputs.function_arn` |
| AwsBedrockFlow | `spec.definition.nodes[].lambdaFunction.lambdaArn` | `status.outputs.function_arn` |
| AwsBedrockKnowledgeBase | `spec.dataSources[].vectorIngestion.customTransformation.lambdaArn` | `status.outputs.function_arn` |
| AwsClientVpn | `spec.clientConnectOptions.lambdaFunctionArn` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.preSignUp` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.preAuthentication` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.postAuthentication` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.postConfirmation` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.preTokenGeneration` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.preTokenGenerationConfig.lambdaArn` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.customMessage` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.userMigration` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.defineAuthChallenge` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.createAuthChallenge` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.verifyAuthChallengeResponse` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.customEmailSender.lambdaArn` | `status.outputs.function_arn` |
| AwsCognitoUserPool | `spec.lambdaConfig.customSmsSender.lambdaArn` | `status.outputs.function_arn` |
| AwsConfigRule | `spec.customLambda.functionArn` | `status.outputs.function_arn` |
| AwsEventBridgeScheduler | `spec.target.arn` | `status.outputs.function_arn` |
| AwsHttpApiGateway | `spec.routes[].integration.integrationUri` | `status.outputs.function_arn` |
| AwsHttpApiGateway | `spec.authorizers[].authorizerUri` | `status.outputs.function_arn` |
| AwsKinesisFirehose | `spec.extendedS3.processing.processors[].lambda.lambdaArn` | `status.outputs.function_arn` |
| AwsKinesisFirehose | `spec.opensearch.processing.processors[].lambda.lambdaArn` | `status.outputs.function_arn` |
| AwsKinesisFirehose | `spec.opensearchServerless.processing.processors[].lambda.lambdaArn` | `status.outputs.function_arn` |
| AwsKinesisFirehose | `spec.httpEndpoint.processing.processors[].lambda.lambdaArn` | `status.outputs.function_arn` |
| AwsKinesisFirehose | `spec.redshift.processing.processors[].lambda.lambdaArn` | `status.outputs.function_arn` |
| AwsKinesisFirehose | `spec.splunk.processing.processors[].lambda.lambdaArn` | `status.outputs.function_arn` |
| AwsKinesisFirehose | `spec.snowflake.processing.processors[].lambda.lambdaArn` | `status.outputs.function_arn` |
| AwsKinesisFirehose | `spec.iceberg.processing.processors[].lambda.lambdaArn` | `status.outputs.function_arn` |
| AwsLambdaEventSourceMapping | `spec.functionArn` | `status.outputs.function_arn` |
| AwsRestApiGateway | `spec.routes[].integration.uri` | `status.outputs.invoke_arn` |
| AwsRestApiGateway | `spec.authorizers[].lambdaInvokeUri` | `status.outputs.invoke_arn` |
| AwsS3Bucket | `spec.notification.lambdaFunctions[].lambdaFunctionArn` | `status.outputs.function_arn` |
| AwsSecretsManagerSecret | `spec.rotation.rotationLambdaArn` | `status.outputs.function_arn` |

## See Also

- [Overview](../README.md)
