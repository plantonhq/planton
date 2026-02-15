# AwsStepFunction Resource Kind (R06)

**Date**: February 15, 2026
**Type**: Feature
**Component**: AWS Step Functions State Machine
**Enum**: AwsStepFunction = 241

## Summary

Added AwsStepFunction as the sixth new AWS resource kind in the cloud provider expansion project. This component wraps `aws_sfn_state_machine` for orchestrating serverless workflows using Amazon States Language (ASL), with support for STANDARD and EXPRESS execution modes, CloudWatch logging, X-Ray tracing, and customer-managed KMS encryption.

## What Was Delivered

### Proto API (4 files)
- `spec.proto` — 7 top-level fields, 3 nested messages, 5 CEL validations
  - `google.protobuf.Struct` for ASL definition (native YAML authoring)
  - `StringValueOrRef` for role_arn (→ AwsIamRole), kms_key_id (→ AwsKmsKey), log_destination
  - Nested: `AwsStepFunctionLoggingConfig`, `AwsStepFunctionEncryptionConfig`
- `stack_outputs.proto` — state_machine_arn, state_machine_name
- `api.proto` — KRM envelope with metadata/spec/status
- `stack_input.proto` — target + provider_config

### Validation Tests (26 tests, all passing)
- 13 happy path (minimal, STANDARD, EXPRESS, logging levels, encryption, production-ready)
- 13 failure scenarios (missing required fields, invalid types, invalid levels, range violations, missing destination)

### Pulumi Module (4 files, ~150 lines)
- `main.go` — Entry point with provider setup
- `locals.go` — Tags and target reference
- `outputs.go` — Output key constants
- `state_machine.go` — Single resource with all config blocks
  - ASL Struct → JSON serialization
  - Log destination `:*` suffix auto-append
  - Encryption type auto-inference

### Terraform Module (5 files, ~80 lines)
- Dynamic blocks for tracing, logging, and encryption
- Feature parity with Pulumi module

### Documentation
- `README.md` — Spec reference, examples, type comparison table
- `examples.md` — 7 examples (hello world, Lambda task, multi-step pipeline, EXPRESS, production, parallel, infra-chart)
- `docs/README.md` — Architecture deep-dive, design decisions, dependency graph
- `catalog-page.md` — Catalog entry with quick start

### Presets (3)
- `01-standard-workflow` — Minimal STANDARD with single Lambda task
- `02-express-workflow` — EXPRESS with event routing (Choice state)
- `03-production-workflow` — Full observability and encryption with valueFrom references

## Key Design Decisions

- **`definition` as `google.protobuf.Struct`** — Native YAML authoring, consistent with SQS policy / EventBridge event_pattern pattern. ASL key casing preserved through serialization.
- **Encryption type auto-inferred** — `CUSTOMER_MANAGED_KMS_KEY` when kms_key_id is set, AWS-owned otherwise. Users don't need to set the redundant `type` field.
- **Tracing as flat boolean** — `tracing_enabled` instead of nested `tracing_configuration { enabled }`. Single bool doesn't justify a nested message.
- **Log destination `:*` suffix auto-appended** — IaC module handles this AWS quirk transparently.
- **Version publishing omitted (v1)** — Niche use case (<20%), 80/20 rule. Can be added later without breaking changes.

## Files Created/Changed

- 37 files in `apis/org/openmcf/provider/aws/awsstepfunction/v1/`
- `apis/org/openmcf/shared/cloudresourcekind/cloud_resource_kind.proto` (enum addition)
- `_changelog/2026-02/2026-02-15-095143-aws-step-function-resource-kind.md`
