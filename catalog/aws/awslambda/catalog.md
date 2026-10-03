# AWS Lambda

Deploys a complete AWS Lambda function: the code source (an S3 zip archive or an ECR container image), the execution environment (memory, timeout, ephemeral storage, environment variables), VPC attachment with an optional EFS mount, and the function-scoped surface AWS models as separate resources but that are honestly part of the function's own configuration — aliases with canary traffic shifting and provisioned concurrency, the function URL, resource-policy invoke permissions, the asynchronous-invocation config, recursion detection, and runtime patch management. Event sources (SQS, Kinesis, DynamoDB Streams, Kafka) attach through the separate AwsLambdaEventSourceMapping Infra Component, which references this function.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Lambda Function** -- configured with the specified code source, runtime or image, memory, timeout, architecture, and every environment setting in the spec
- **Aliases** (when defined) -- one alias resource per entry, keyed by name, each optionally splitting traffic between two published versions and carrying provisioned concurrency
- **Function URL** (when configured) -- the built-in HTTPS endpoint with its authorization mode and CORS rules
- **Invoke Permissions** (when defined) -- one resource-policy statement per entry, keyed by statement ID
- **Async Invocation Config** (when configured) -- retry, event-age, and success/failure destination settings
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically

The function's **name comes from `metadata.name`** — create-time immutable in AWS. Its CloudWatch log group is AWS-created (`/aws/lambda/<function-name>`) unless `loggingConfig.logGroup` points at a log group you manage.

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An IAM execution role** trusting `lambda.amazonaws.com` with `AWSLambdaBasicExecutionRole` attached (add `AWSLambdaVPCAccessExecutionRole` for VPC attachment). Provide the ARN directly or reference an AwsIamRole Infra Component.
- **A deployment artifact** -- a zip archive in an S3 bucket in the function's region, or a container image in ECR.
- **VPC subnets and security groups** (optional) -- only when the function must reach private resources. Both travel together, and attachment removes default internet access (route through a NAT gateway to restore it).
- **KMS keys** (optional) -- one for environment variables at rest (`kmsKeyArn`), one for the deployment package in S3 (`sourceKmsKeyArn`).

## Deploy

### Console

Open the deployment store, find **AWS Lambda**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Zip-Based Lambda Function** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsLambda
metadata:
  name: payment-processor
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  roleArn:
    value: "arn:aws:iam::123456789012:role/lambda-exec-role"
  s3:
    bucket:
      value: my-deploy-bucket
    key: functions/payment-processor.zip
  runtime: nodejs22.x
  handler: index.handler
  memorySizeMb: 512
  timeoutSeconds: 30
```

```shell
planton apply -f lambda.yaml
```

This creates the function from the S3 deployment package. An Infra Job tracks the provisioning and streams progress in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the function to resources deployed in the same InfraPipeline:

```yaml
spec:
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: lambda-exec-role
      fieldPath: status.outputs.role_arn
  s3:
    bucket:
      valueFrom:
        kind: AwsS3Bucket
        name: lambda-artifacts
        fieldPath: status.outputs.bucket_id
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-b
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: lambda-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: lambda-env-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the referenced resources first, then provisions the function with the resolved values.

## Key Configuration

These are the most important decisions when configuring AWS Lambda. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Code source** -- Exactly one of `s3` (a zip archive, with `runtime` and `handler` required) or `imageUri` (an ECR image carrying its own runtime and entrypoint; `imageConfig` can override the entrypoint without rebuilding). S3 zip is the right default for most services; images suit heavy dependency trees (up to 10 GB).

**Memory buys CPU** -- `memorySizeMb` (128–10240) is Lambda's only sizing dial: CPU and network scale linearly with it, and a full vCPU arrives around 1769 MB. For CPU-bound code, a larger size often costs less overall by finishing sooner.

**Architecture** -- Empty keeps the AWS default (`x86_64`). Set `arm64` for ~20% better price-performance when your runtime and native dependencies support Graviton.

**VPC attachment travels as a unit** -- `subnetIds` and `securityGroupIds` go together, and the optional `fileSystemConfig` (EFS) requires them. Attachment removes default internet access.

**The deploy model** -- Enable `publish` to freeze an immutable version on every change, then define `aliases` (e.g. `live`) as the stable invocation targets. Repointing an alias ships or rolls back without touching callers; a canary weight splits an alias's traffic with one additional version. `snapStart` (JVM cold-start elimination) applies to published versions only. `publishTo: LATEST_PUBLISHED` additionally maintains the `$LATEST.PUBLISHED` head — a moving qualifier that always resolves to the newest published version, which `scalingConfigs` and qualified invocations can pin without naming version numbers. The head is a newer AWS capability still rolling out: regions/accounts without it reject the value at create (`InvalidParameterValueException`) — and AWS creates the function before rejecting the parameter, so clean up the half-created function if a deploy fails on this field. The function URL and the async/runtime-management satellites each take an optional `qualifier` to serve an alias (or version) instead of the unpublished head.

**The execution platform** -- Three arms move the function off the default on-demand fleet: `managedInstances` runs it on a Lambda Managed Instances capacity provider (dedicated EC2 capacity AWS manages — steady high throughput, memory beyond 10 GB), `tenantIsolationMode: PER_TENANT` dedicates execution environments to the tenant id callers pass at invoke time (create-time immutable), and `durableConfig` makes invocations durable — AWS checkpoints progress so week-long workflows survive interruption and resume (adding or removing the block replaces the function; per-qualifier `scalingConfigs` pin the environment fleet between a floor and a ceiling).

**Reserved concurrency is tri-state** -- Unset draws from the account pool; a positive value is both a guarantee and a ceiling; an explicit `0` throttles every invocation — the operational kill switch.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsS3Bucket** | `s3.bucket` | `status.outputs.bucket_id` |
| **AwsKmsKey** (optional) | `kmsKeyArn`, `sourceKmsKeyArn` | `status.outputs.key_arn` |
| **AwsSubnet** (optional) | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsSqsQueue** (optional) | `deadLetterTargetArn`, `asyncInvokeConfig.*DestinationArn` | `status.outputs.queue_arn` |
| **AwsEfsAccessPoint** (optional) | `fileSystemConfig.accessPointArn` | `status.outputs.access_point_arn` |
| **AwsCloudwatchLogGroup** (optional) | `loggingConfig.logGroup` | `status.outputs.log_group_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `function_arn` | The function ARN | Event source mappings, trigger configs, IAM policies |
| `function_name` | The function name | SDK invoke calls, CLI commands, CloudWatch alarms |
| `invoke_arn` | The apigateway-shaped invocation ARN | API Gateway integrations |
| `qualified_arn` | ARN of the most recently published version | Version-pinned invocation (empty when publish is off) |
| `version` | The most recently published version number | Alias definitions, deploy automation |
| `function_url` | The HTTPS endpoint | Webhook registration (empty when no URL is configured) |
| `alias_arns` | Alias ARNs keyed by alias name | Stable invocation targets, e.g. `status.outputs.alias_arns.live` |
| `log_group_name` | The CloudWatch log group receiving the logs | Log subscriptions, metric filters |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Zip-based event handler** -- S3 code source, a current runtime, explicit sizing, and a failure destination on the async config. The standard configuration for API endpoints, webhook handlers, and queue workers. Start from the **Zip-Based Lambda Function** preset.

**Container-based function** -- ECR image code source for functions with large dependency trees or image-standardized build pipelines. Start from the **Container-Based Lambda Function** preset.

**Traffic-shifted production service** -- `publish` on, a `live` alias with a canary weight for progressive rollouts, and provisioned concurrency on the steady alias when cold starts matter.

## Works With

- [**AWS Lambda Event Source Mapping**](/infra-catalog/aws-lambda-event-source-mapping) -- wires SQS queues, Kinesis/DynamoDB streams, and Kafka topics into this function
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the execution role
- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- holds the zip deployment package
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides subnets for VPC-attached functions
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls network access for the function's ENIs
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- encrypts environment variables and the deployment artifact
- [**AWS SQS Queue**](/infra-catalog/aws-sqs-queue) -- receives dead-letter and async failure records
- [**AWS Elastic File System**](/infra-catalog/aws-elastic-file-system) -- durable shared storage mounted into the execution environment
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- a managed destination for the function's logs
