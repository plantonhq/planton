# AwsStepFunction Pulumi Module — Architecture Overview

## Resource Graph

```
AwsStepFunctionStackInput
  └── AwsStepFunction (target)
        ├── spec.definition → json.Marshal → sfn.StateMachine.Definition
        ├── spec.role_arn → sfn.StateMachine.RoleArn
        ├── spec.type → sfn.StateMachine.Type (default: STANDARD)
        ├── spec.tracing_enabled → sfn.StateMachine.TracingConfiguration
        ├── spec.logging → sfn.StateMachine.LoggingConfiguration
        │     ├── level → Level
        │     ├── include_execution_data → IncludeExecutionData
        │     └── log_destination (+ ":*") → LogDestination
        └── spec.encryption → sfn.StateMachine.EncryptionConfiguration
              ├── kms_key_id → KmsKeyId
              ├── (auto) type = CUSTOMER_MANAGED_KMS_KEY
              └── kms_data_key_reuse_period_seconds → KmsDataKeyReusePeriodSeconds
```

## Data Flow

1. `LoadStackInput` deserializes the protobuf-encoded stack input
2. `initializeLocals` extracts target, spec, and computes AWS tags
3. `stateMachine` creates the single `sfn.StateMachine` resource:
   - Serializes `definition` Struct → JSON
   - Defaults `type` to `STANDARD` if empty
   - Conditionally adds tracing, logging, and encryption blocks
   - Auto-appends `:*` to log destination ARN
4. Exports `state_machine_arn` and `state_machine_name`

## Error Handling

All errors are wrapped with context using `github.com/pkg/errors`. The module fails fast on any resource creation error.
