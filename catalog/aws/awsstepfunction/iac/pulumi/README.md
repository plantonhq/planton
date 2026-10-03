# AwsStepFunction — Pulumi Module

## Overview

This Pulumi module provisions an AWS Step Functions state machine with optional CloudWatch logging, X-Ray tracing, and customer-managed KMS encryption.

## Module Structure

```
module/
  main.go           — Entry point: provider setup, orchestrate resource creation
  locals.go         — Locals struct with AWS tags, target reference
  outputs.go        — Output key constants
  state_machine.go  — Core resource: creates sfn.StateMachine with all config blocks
```

## IaC Inputs

The module reads `AwsStepFunctionIacInput` which contains:
- `target` — The fully-specified `AwsStepFunction` resource
- `provider_config` — Optional AWS credentials/region override

## Outputs

| Key | Description |
|-----|-------------|
| `state_machine_arn` | ARN of the state machine |
| `state_machine_name` | Name of the state machine |
| `state_machine_version_arn` | ARN of the most recently published version (empty unless `publish` is true) |
| `revision_id` | Revision identifier of the current definition |
| `status` | Lifecycle status reported by AWS |
| `creation_date` | RFC3339 creation timestamp |
| `alias_arns` | Alias ARNs keyed by alias name (empty map without aliases) |

## Local Development

```bash
# Build
cd module && go build ./...

# Debug with test manifest
./debug.sh preview
./debug.sh up
./debug.sh destroy
```

## Key Implementation Notes

- **Definition serialization**: The ASL definition arrives as a `google.protobuf.Struct` and is serialized to JSON using `json.Marshal(spec.Definition.AsMap())`.
- **Versioning**: `publish = true` publishes an immutable version on create and on every configuration change; the latest version's ARN is exported.
- **Log destination suffix**: The module auto-appends `:*` to CloudWatch Log Group ARNs that don't already end with it (AWS requirement for Step Functions).
- **Encryption type inference**: When `encryption.kms_key_id` is provided, the module sets `type = "CUSTOMER_MANAGED_KMS_KEY"` automatically.
- **Type default**: When `spec.type` is empty, the module defaults to `"STANDARD"`.
