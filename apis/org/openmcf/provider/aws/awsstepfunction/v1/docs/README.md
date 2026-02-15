# AwsStepFunction: Architecture and Design

## Overview

The AwsStepFunction component wraps the `aws_sfn_state_machine` Terraform resource (and its Pulumi equivalent `sfn.StateMachine`) into a single OpenMCF deployment component. It provisions a Step Functions state machine with optional CloudWatch logging, X-Ray tracing, and customer-managed KMS encryption.

## Resource Mapping

| OpenMCF Component | AWS Resource | Count |
|-------------------|-------------|-------|
| AwsStepFunction | `aws_sfn_state_machine` | 1 |

This is a single-resource component. The state machine itself is the only AWS resource created.

## Architecture

```
AwsStepFunction
└── aws_sfn_state_machine
    ├── definition (ASL JSON, serialized from YAML Struct)
    ├── role_arn (IAM execution role)
    ├── logging_configuration (→ CloudWatch Logs)
    ├── tracing_configuration (→ AWS X-Ray)
    └── encryption_configuration (→ KMS)
```

## Design Decisions

### 1. Definition as google.protobuf.Struct (Not String)

The `definition` field uses `google.protobuf.Struct` instead of a plain string. This enables users to write ASL definitions as native YAML in their spec files, which is significantly more readable than embedding JSON strings. The IaC modules serialize the Struct to JSON before passing it to the AWS API.

ASL key casing (StartAt, States, Type, Resource, etc.) is preserved through `protobuf.Struct` serialization because Struct stores keys as-is without any case transformation.

This is consistent with how SQS `policy`, SNS `filter_policy`, and EventBridge `event_pattern` are handled across other OpenMCF AWS components.

### 2. Encryption Type Auto-Inferred

The TF provider exposes an `encryption_configuration.type` field with values `AWS_OWNED_KMS_KEY` and `CUSTOMER_MANAGED_KMS_KEY`. Instead of requiring users to set this redundant field, the OpenMCF spec omits it. The IaC module infers the type automatically:

- When `encryption.kms_key_id` is provided → `CUSTOMER_MANAGED_KMS_KEY`
- When the encryption block is omitted → AWS-owned keys (default)

This eliminates a source of configuration errors where users might set a KMS key but forget to set the type.

### 3. Tracing as Flat Boolean

The TF/Pulumi providers wrap tracing in a nested block: `tracing_configuration { enabled = true }`. Since there is only one field in the block, the OpenMCF spec flattens it to a simple `tracing_enabled` boolean on the top-level spec. The IaC modules translate this to the nested structure required by the provider.

### 4. Log Destination `:*` Suffix Auto-Appended

AWS requires the CloudWatch Log Group ARN for Step Functions logging to end with `:*`. This is a provider quirk that should not burden users. The IaC modules automatically append `:*` if the ARN does not already end with it, so users can reference a log group ARN directly via `StringValueOrRef` without worrying about the suffix.

### 5. Version Publishing Omitted (v1)

The TF provider supports `publish` and `version_description` for state machine versioning (blue-green deployments with aliases). This is omitted from v1 because:

- Fewer than 20% of Step Functions users use versioning
- It adds complexity to both the spec and IaC modules
- It can be added in a future version without breaking changes

### 6. State Machine Type is Force-New

Changing the `type` field (STANDARD ↔ EXPRESS) requires replacing the state machine. This is an AWS constraint, not an OpenMCF limitation. The `type` field is documented accordingly, and users should be aware that changing it will destroy and recreate the resource.

## Dependencies

### Upstream (Resources This Component Depends On)

| Resource | Field | Relationship |
|----------|-------|-------------|
| AwsIamRole | `role_arn` | Required — execution role |
| AwsKmsKey | `encryption.kms_key_id` | Optional — customer-managed encryption |
| AwsCloudwatchLogGroup | `logging.log_destination` | Optional — execution logging |

### Downstream (Resources That Reference This Component)

| Resource | Output Used | Use Case |
|----------|-------------|----------|
| AwsEventBridgeRule | `state_machine_arn` | EventBridge target |
| AwsHttpApiGateway | `state_machine_arn` | API Gateway integration |
| AwsIamRole | `state_machine_arn` | IAM policy for invocation |

## Infra Chart Composition

Step Functions typically appears in infra charts at the orchestration layer:

```
Layer 0: VPC, IAM Roles, KMS Keys
Layer 1: Lambda Functions, SQS Queues, SNS Topics, CloudWatch Log Groups
Layer 2: Step Functions State Machine (references Layer 1 resources in ASL definition)
Layer 3: EventBridge Rules, API Gateway (triggers the state machine)
```

In a serverless-api chart, Step Functions coordinates Lambda functions into request-processing pipelines. In an event-driven chart, it orchestrates event handling with branching, retries, and error workflows.

## Validation Strategy

- **Field-level**: `buf.validate` for required fields (`definition`, `role_arn`)
- **CEL validations**: Cross-field constraints (type values, logging level values, destination requirement, encryption reuse period range)
- **Provider-side**: AWS validates the ASL definition during create/update via `ValidateStateMachineDefinition` API
