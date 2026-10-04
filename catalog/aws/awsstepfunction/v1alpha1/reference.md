# AwsStepFunction

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsStepFunctionSpec defines the desired configuration for an AWS Step Functions
state machine.

Step Functions orchestrates distributed workflows by coordinating AWS services
(Lambda, SQS, SNS, DynamoDB, ECS, etc.) into serverless state machines defined
in Amazon States Language (ASL).

Two state machine types are supported:
- STANDARD: Long-running workflows (up to 1 year) with exactly-once execution,
  full execution history, and visual debugging in the AWS console.
- EXPRESS: High-volume, short-duration workflows (up to 5 minutes) with
  at-most-once execution semantics, optimized for event processing pipelines.

Notes:
- The `type` cannot be changed after creation (forces replacement).
- The `definition` field accepts ASL as native YAML. The IaC modules serialize
  it to JSON before passing it to the AWS API. ASL key casing (StartAt, States,
  Type, etc.) is preserved through serialization.
- The state machine name is taken from metadata.name. AWS restricts names to
  1-80 characters matching [0-9A-Za-z_-] (no spaces or dots).
- There is deliberately no description field: the CreateStateMachine API has
  no description input (the AWS console derives one from the definition's
  Comment field), so a spec field would be silently dropped.
- Credentials, region, and deployment workflow live outside this spec in IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsStepFunction
metadata:
  name: test-workflow
  org: test-org
  env: dev
  id: test-workflow-dev
spec:
  region: us-west-2
  type: STANDARD
  publish: true
  aliases:
    - name: live
      description: Follows the version published by the current deployment
  roleArn:
    value: arn:aws:iam::123456789012:role/StepFunctionsExecRole
  definition:
    StartAt: Hello
    States:
      Hello:
        Type: Pass
        Result: Hello, World!
        End: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.type` | `string` |  |  |  |
| `spec.definition` | `object` | yes |  |  |
| `spec.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.publish` | `bool` |  |  |  |
| `spec.aliases` | `[]AwsStepFunctionAlias` |  |  |  |
| `spec.aliases[].name` | `string` |  |  |  |
| `spec.aliases[].description` | `string` |  |  |  |
| `spec.logging` | `AwsStepFunctionLoggingConfig` |  |  |  |
| `spec.logging.level` | `string` |  |  |  |
| `spec.logging.includeExecutionData` | `bool` |  |  |  |
| `spec.logging.logDestination` | `string \| valueFrom` |  |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.tracingEnabled` | `bool` |  |  |  |
| `spec.encryption` | `AwsStepFunctionEncryptionConfig` |  |  |  |
| `spec.encryption.kmsKeyId` | `string \| valueFrom` | yes |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.encryption.kmsDataKeyReusePeriodSeconds` | `int32` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.type

`string`

State machine type. Determines execution semantics and pricing model.
- "STANDARD": Long-running, exactly-once, full execution history.
- "EXPRESS": High-volume, short-duration, at-most-once.
Cannot be changed after creation (forces replacement).
When omitted the IaC module defaults to "STANDARD".

### spec.definition

`object` · required

State machine definition in Amazon States Language (ASL). Write the
definition as native YAML; the IaC module serializes it to JSON for the
AWS API. ASL key casing (StartAt, States, Type, Resource, etc.) is
preserved through serialization.

Maximum size: 1,048,576 bytes (1 MB) after JSON serialization.

Use AWS Step Functions Workflow Studio, CDK, or the ASL specification
to author complex workflows, then express the result as YAML here.

- rule: {"required":true}

### spec.roleArn

`string | valueFrom` · required

IAM execution role ARN. The role must have a trust policy for
states.amazonaws.com and policies granting access to all services
invoked by the workflow (Lambda:InvokeFunction, SQS:SendMessage, etc.).

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.publish

`bool`

Publish a version of the state machine on every create and on every
configuration update. Published versions are immutable snapshots
(definition + role + logging/tracing/encryption at publish time) addressed
by the version ARN exported in outputs. Versions are the foundation
for alias-based traffic shifting and safe rollbacks: point consumers at a
version ARN (or an alias routing between two versions) instead of the
mutable state machine ARN. When false (the default), executions always run
the latest saved revision.

### spec.aliases

`[]AwsStepFunctionAlias`

Named aliases for the state machine. Each alias points at the version
published by this deployment (requires publish: true) and exposes a
stable ARN consumers can invoke while the underlying version advances
on every configuration change. Aliases are keyed by name: renaming an
entry replaces that alias without touching its siblings.

Weighted canary routing between two specific versions is an imperative
deployment-shift operation (the routing weights change DURING a
rollout, not in a declarative snapshot) and is deliberately not
modeled; each alias here routes 100% of traffic to the version this
deployment published.

### spec.aliases[].name

`string`

Alias name. 1-80 characters matching [0-9A-Za-z_-]. The name keys the
provider resource: renaming replaces this alias without touching
siblings.

- rule: {"string":{"pattern":"^[0-9A-Za-z_-]{1,80}$"}}

### spec.aliases[].description

`string`

Optional human-readable description of the alias (up to 256 characters).

- rule: {"string":{"maxLen":"256"}}

### spec.logging

`AwsStepFunctionLoggingConfig`

Logging configuration for execution history events. When omitted, no
logging configuration is sent (new state machines default to level OFF).
To explicitly turn logging OFF on a state machine that previously had it
on, keep this block with level: "OFF" — removing the block entirely
sends nothing and the provider keeps the last applied logging state.
Logging is supported for both STANDARD and EXPRESS state machines.

- rule: logging.level must be 'ALL', 'ERROR', 'FATAL', or 'OFF' when set

### spec.logging.level

`string`

Logging level. Determines which execution history events are logged.
- "ALL": Log all event types (recommended for development and debugging).
- "ERROR": Log only error events (recommended for production).
- "FATAL": Log only fatal errors.
- "OFF": Disable logging.

### spec.logging.includeExecutionData

`bool`

Whether to include execution input and output data in log entries. When
true, the full JSON payloads passed between states are logged. Useful for
debugging but may increase log volume and expose sensitive data.

### spec.logging.logDestination

`string | valueFrom`

CloudWatch Logs log group ARN for log delivery. Accepts a direct ARN or a
reference to a CloudWatch Log Group resource.

Note: AWS requires the ARN to end with ":*". The IaC module automatically
appends this suffix if not present, so you can reference a log group ARN
directly without worrying about the suffix.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.tracingEnabled

`bool` · optional (explicit presence)

Enable AWS X-Ray tracing for the state machine. When true, Step Functions
sends trace data to X-Ray for visualizing request flows. Ensure the
execution role has xray:PutTraceSegments and xray:PutTelemetryRecords
permissions.

Tri-state: leave unset to keep the AWS default (tracing off, and no
tracing configuration is sent). Set explicitly to true or false to pin
the state — an explicit false is what turns tracing OFF on a state
machine that previously had it on (simply removing a true value sends
nothing, and the provider keeps the last applied state).

### spec.encryption

`AwsStepFunctionEncryptionConfig`

Encryption configuration for data at rest. When omitted, AWS uses
AWS-owned keys (default, no additional cost). Provide this block to
use a customer-managed KMS key for encrypting state machine data,
execution history, and input/output payloads.

One-way in practice: once a customer-managed key has been applied,
REMOVING this block does not revert the state machine to AWS-owned
keys — the provider suppresses the block's removal and keeps the last
applied key. Reverting requires an out-of-band update today.

- rule: kms_data_key_reuse_period_seconds must be between 60 and 900 when set

### spec.encryption.kmsKeyId

`string | valueFrom` · required

Customer-managed KMS key ARN for encrypting state machine data, execution
history, and input/output payloads. The key must be a symmetric encryption
key in the same region as the state machine. For cross-account access, use
the full key ARN (not alias).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.encryption.kmsDataKeyReusePeriodSeconds

`int32`

Duration in seconds for which Step Functions reuses a data encryption key
before calling KMS GenerateDataKey again. Higher values reduce KMS API
costs but increase the window for key reuse.
Range: 60–900 seconds. AWS default: 300 (5 minutes).
Leave at 0 to use the AWS default.

## Validation Rules

- `type_valid_values`: type must be 'STANDARD' or 'EXPRESS' when set
- `logging_destination_required_when_enabled`: logging.log_destination is required when logging.level is not 'OFF'
- `aliases_require_publish`: aliases require publish: true (an alias routes to the published version)
- `alias_names_unique`: aliases[].name must be unique
- `role_arn_format`: role_arn literal value must be an IAM role ARN (arn:...)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsStepFunction, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.state_machine_arn` | `string` | The Amazon Resource Name (ARN) of the state machine. This is the primary identifier used for invoking the state machine and for cross-service references (EventBridge targets, API Gateway integrations, IAM policies). |
| `status.outputs.state_machine_name` | `string` | The name of the state machine. Useful for dashboards, monitoring, and human-readable references in logs and alerts. |
| `status.outputs.state_machine_version_arn` | `string` | The ARN of the most recently published version of the state machine (e.g. "...:stateMachine:orders:3"). Populated only when spec.publish is true. Point consumers (EventBridge targets, aliases) at this ARN to pin them to an immutable snapshot instead of the mutable state machine. |
| `status.outputs.revision_id` | `string` | The revision identifier of the current state machine definition. AWS assigns a new revision id on every definition or configuration change, whether or not a version is published. Useful for change auditing. |
| `status.outputs.status` | `string` | Lifecycle status of the state machine as reported by AWS (e.g. "ACTIVE", "DELETING"). |
| `status.outputs.creation_date` | `string` | RFC3339 timestamp of when the state machine was created. |
| `status.outputs.alias_arns` | `map<string, string>` | Alias ARNs keyed by alias name (spec.aliases[].name), e.g. "prod" -> "...:stateMachine:orders:prod". Each alias routes to the version this deployment published; consumers invoke the alias ARN to follow deployments without repointing. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.logging.logDestination` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.encryption.kmsKeyId` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsEventBridgeScheduler | `spec.target.arn` | `status.outputs.state_machine_arn` |

## See Also

- [Overview](../README.md)
