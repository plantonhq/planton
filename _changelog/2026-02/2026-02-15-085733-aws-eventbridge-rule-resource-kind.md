# AWS EventBridge Rule Resource Kind

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Documentation

## Summary

Added the AwsEventBridgeRule (R04) deployment component to Planton — the fourth new AWS resource kind in the cloud provider expansion project. EventBridge rules are the routing layer of event-driven architectures, matching incoming events by pattern or schedule and delivering them to targets like Lambda, SQS, SNS, and Step Functions.

## Problem Statement / Motivation

EventBridge rules are the core routing mechanism for event-driven architectures on AWS. Without them, EventBridge buses receive events but cannot route them to processing targets. The existing AwsEventBridgeBus (R03) component creates custom buses, but buses need rules to become useful.

### Pain Points

- No way to define EventBridge routing rules through Planton
- Infra charts for event-driven architectures were blocked without rule support
- Manual Terraform/Pulumi code was required for event routing configuration

## Solution / What's New

A complete deployment component covering the EventBridge rule lifecycle with bundled targets. The component follows the established AwsSnsTopic pattern of bundling child resources (targets) with the parent resource (rule).

### Key Design Decisions

- **Bundled targets**: Targets are defined inline with the rule (like SNS subscriptions) because a rule without targets is functionally useless
- **google.protobuf.Struct for event patterns**: Users write event patterns in native YAML instead of embedded JSON strings
- **Mutual exclusivity enforced**: event_pattern and schedule_expression are mutually exclusive via CEL validation
- **80/20 target support**: Covers Lambda, SQS, SNS, Step Functions, and CloudWatch Log targets (90%+ of real-world usage) without the complexity of ECS/Batch/Kinesis-specific blocks

### Surprise Findings (Not in Planning Guidance)

Deep research into the Terraform provider revealed capabilities not in the T02 planning phase:

1. **Per-target dead letter config** — essential for production reliability
2. **Per-target retry policy** — controls retry behavior (max age, max attempts)
3. **SQS-specific config** — `message_group_id` for FIFO queue targets
4. **Rule-level vs target-level role_arn distinction** — rule-level is for cross-account delivery (niche), target-level is for assumed-role invocation (common)

## Implementation Details

### Proto API (4 files)

- `spec.proto`: 6 rule-level fields, 5 nested messages (AwsEventBridgeTarget, AwsEventBridgeInputTransformer, AwsEventBridgeTargetDeadLetterConfig, AwsEventBridgeTargetRetryPolicy, AwsEventBridgeTargetSqsConfig), 9 CEL validations
- StringValueOrRef for `event_bus_name` (→ AwsEventBridgeBus), target `arn` (polymorphic), target `role_arn` (→ AwsIamRole), DLQ `arn` (→ AwsSqsQueue)
- google.protobuf.Struct for `event_pattern`

### Validation Tests (36 specs)

Comprehensive coverage: happy path (8), rule-level CEL (4), field constraints (2), target validations (4), input mutual exclusion (6), input transformer (2), dead letter config (2), retry policy ranges (5), SQS config (1), edge cases (2).

### Pulumi Module (5 files)

- `main.go`: Entry point with AWS provider setup, orchestrates rule and target creation
- `locals.go`: Locals struct, AWS tags, Struct-to-JSON serialization helper
- `outputs.go`: Output constants (rule_arn, rule_name)
- Rule creation with event pattern serialization, schedule expression, state
- Target iteration creating EventTarget per spec entry with input transformation, DLQ, retry policy, SQS config

### Terraform Module (5 files)

- Rule resource with event pattern (jsonencode), schedule expression, state
- Target resources via `for_each` keyed by target name
- Dynamic blocks for input_transformer, dead_letter_config, retry_policy, sqs_target
- Full feature parity with Pulumi module

### Documentation

- README.md: Complete field documentation, use cases, validation rules, references
- examples.md: 7 examples (scheduled Lambda, EC2 events to SQS, custom bus with transformer, fan-out, Step Functions, daily cron, FIFO SQS)
- catalog-page.md: User-facing catalog page with quick start and configuration reference
- docs/README.md: Research documentation with design decisions and architecture

### Presets (3)

1. **01-schedule-lambda**: Scheduled rule targeting Lambda (simplest cron replacement)
2. **02-event-pattern-sqs**: Event pattern matching EC2 state changes with DLQ
3. **03-multi-target-fanout**: Custom bus with 2 targets, input transformer, retry policy

## Benefits

- **Complete event routing**: Bus + Rule covers the full EventBridge lifecycle
- **Production-ready reliability**: Per-target DLQ and retry policy out of the box
- **Infra chart composability**: StringValueOrRef enables wiring rules to buses, Lambda functions, SQS queues, and IAM roles in dependency-aware templates
- **80/20 coverage**: Supports 90%+ of real-world EventBridge target types without spec bloat

## Impact

- **New component**: `apis/dev/planton/provider/aws/awseventbridgerule/v1/` (~48 files, ~3200 lines)
- **Enum registration**: AwsEventBridgeRule = 228 in cloud_resource_kind.proto
- **Infra charts**: Unblocks event-driven architecture charts that require rule routing
- **Downstream references**: Rules reference AwsEventBridgeBus (bus_name), AwsLambda (function_arn), AwsSqsQueue (queue_arn), AwsSnsTopic (topic_arn), AwsIamRole (role_arn)

## Related Work

- **R01 AwsSqsQueue** (2026-02-15): Foundational messaging resource, commonly used as rule target and DLQ
- **R02 AwsSnsTopic** (2026-02-15): Pub/sub messaging, bundled subscriptions pattern used as reference
- **R03 AwsEventBridgeBus** (2026-02-15): Parent resource — buses receive events, rules route them

---

**Status**: Production Ready
**Timeline**: Single session implementation
