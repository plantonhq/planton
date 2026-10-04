# AWS Step Functions

Deploys a Step Functions state machine that orchestrates distributed workflows using Amazon States Language (ASL) definitions expressed as native YAML. The kind supports both STANDARD (long-running, exactly-once) and EXPRESS (high-volume, short-duration) state machine types, with configurable CloudWatch Logs logging, X-Ray tracing, and customer-managed KMS encryption. The execution role, log group, and KMS key all accept ValueFromRef wiring, so a state machine composes with AwsIamRole, AwsCloudwatchLogGroup, and AwsKmsKey resources in the same InfraChart. The state machine type is a one-way door — it cannot be changed after creation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Step Functions State Machine** -- a state machine configured with the specified type, ASL definition (serialized from YAML to JSON), and IAM execution role
- **Logging Configuration** -- configured only when `logging` is provided with a level other than OFF; sends execution history events to the specified CloudWatch Logs log group
- **X-Ray Tracing** -- enabled only when `tracingEnabled` is true; sends trace data for visualizing request flows across coordinated services
- **Encryption Configuration** -- configured only when `encryption` is provided; uses a customer-managed KMS key for encrypting state machine data, execution history, and input/output payloads
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An IAM execution role** with a trust policy for `states.amazonaws.com` and policies granting access to all services invoked by the workflow (Lambda:InvokeFunction, SQS:SendMessage, SNS:Publish, etc.). Provide the ARN directly or reference an AwsIamRole Infra Component via ValueFromRef.
- **A CloudWatch log group** (optional) -- required when enabling execution logging. Provide the ARN directly or reference an AwsCloudwatchLogGroup Infra Component via ValueFromRef.
- **A KMS key** (optional) -- required when using customer-managed encryption. The key must be a symmetric encryption key in the same region. Provide the ARN directly or reference an AwsKmsKey Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS Step Functions**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Workflow** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsStepFunction
metadata:
  name: order-processor
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  roleArn:
    value: "arn:aws:iam::123456789012:role/step-functions-role"
  definition:
    StartAt: ProcessOrder
    States:
      ProcessOrder:
        Type: Task
        Resource: "arn:aws:lambda:us-west-2:123456789012:function:process-order"
        End: true
```

```shell
planton apply -f step-function.yaml
```

This creates a STANDARD state machine with a single Lambda task. No logging, tracing, or encryption is configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the state machine to an IAM role, log group, and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  roleArn:
    valueFrom:
      kind: AwsIamRole
      name: step-functions-role
      fieldPath: status.outputs.role_arn
  logging:
    level: ALL
    includeExecutionData: true
    logDestination:
      valueFrom:
        kind: AwsCloudwatchLogGroup
        name: workflow-logs
        fieldPath: status.outputs.log_group_arn
  encryption:
    kmsKeyId:
      valueFrom:
        kind: AwsKmsKey
        name: workflow-encryption-key
        fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the IAM role, log group, and KMS key first, then provisions the state machine with the resolved values.

## Key Configuration

These are the most important decisions when configuring Step Functions. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**State machine type** -- STANDARD (default) supports workflows up to 1 year with exactly-once execution semantics and full execution history in the console. EXPRESS supports workflows up to 5 minutes with at-most-once semantics, optimized for high-volume event processing. The type cannot be changed after creation.

**ASL definition** -- Write the workflow definition as native YAML in the `definition` field; the IaC module serializes it to JSON. ASL key casing (StartAt, States, Type, Resource) is preserved through serialization. Maximum 1 MB after JSON serialization.

**Logging** -- Set `logging.level` to ALL for full execution visibility (recommended for development) or ERROR for production (captures failures only). When enabled, `logging.logDestination` must reference a CloudWatch log group ARN. Set `logging.includeExecutionData` to true to include input/output payloads in log entries.

**Encryption** -- By default, AWS uses AWS-owned keys. Provide `encryption.kmsKeyId` with a customer-managed KMS key for compliance requirements. Adjust `encryption.kmsDataKeyReusePeriodSeconds` (60--900, default 300) to balance KMS API costs against key reuse window.

**Versioning** -- Set `publish: true` to publish an immutable version on every create and configuration update (definition + role + logging/tracing/encryption frozen at publish time). The newest version's ARN is exported as `state_machine_version_arn` — pin EventBridge targets or aliases to it for safe rollouts and rollbacks. When false (the default), executions always run the latest saved revision.

**Aliases** -- `aliases` entries (requires `publish: true`) each create a named alias routing 100% of traffic to the version this deployment published; consumers invoke the stable alias ARN (exported in `alias_arns` keyed by name) and follow deployments without repointing. Weighted canary routing between two specific versions is an imperative deployment-shift operation and is deliberately not modeled.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsIamRole** | `roleArn` | `status.outputs.role_arn` |
| **AwsCloudwatchLogGroup** (optional) | `logging.logDestination` | `status.outputs.log_group_arn` |
| **AwsKmsKey** (optional) | `encryption.kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `state_machine_arn` | Amazon Resource Name of the state machine | EventBridge rule targets, API Gateway integrations, IAM policies |
| `state_machine_name` | Name of the state machine | Dashboards, monitoring, human-readable log references |
| `state_machine_version_arn` | ARN of the most recently published version (populated only when `publish` is true) | Version-pinned EventBridge targets and aliases for safe rollouts |
| `alias_arns` | Alias ARNs keyed by alias name (populated only when `aliases` are defined) | Stable invocation targets that follow deployments, e.g. `status.outputs.alias_arns.live` |
| `revision_id` | Revision identifier of the current definition (changes on every update) | Change auditing |
| `status` | Lifecycle status reported by AWS (e.g. ACTIVE) | Health dashboards |
| `creation_date` | RFC3339 creation timestamp | Auditing and inventory |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard workflow** -- STANDARD type with a single Lambda task. The starting point for long-running business process automation, ETL pipelines, and workflows requiring exactly-once execution and full console debugging. Start from the **Standard Workflow** preset.

**Express workflow** -- EXPRESS type with a Choice-state routing pattern and error-level logging. Suited for high-volume, short-duration event processing pipelines (EventBridge to Step Functions to Lambda). Start from the **Express Workflow** preset.

**Production workflow** -- STANDARD type with ALL-level logging, X-Ray tracing, customer-managed KMS encryption, and ValueFromRef wiring for IAM role, log group, and KMS key. The recommended configuration for business-critical production workflows. Start from the **Production Workflow** preset.

## Works With

- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- provides the execution role for the state machine to invoke AWS services
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- provides the logging destination for execution history events
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for encrypting state machine data and execution history