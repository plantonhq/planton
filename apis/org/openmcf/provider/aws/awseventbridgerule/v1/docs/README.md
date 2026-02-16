# AwsEventBridgeRule — Research Documentation

## Overview

Amazon EventBridge rules are the routing layer of the EventBridge service. Rules evaluate incoming events against patterns (or fire on a schedule) and deliver matched events to one or more targets. Rules are the glue between event sources and event consumers.

## Architecture

### Rule Execution Flow

1. An event arrives on an event bus (default or custom).
2. EventBridge evaluates ALL rules attached to that bus against the event.
3. Rules with matching event patterns fire simultaneously.
4. Each matching rule delivers the event to all of its targets in parallel.
5. If input transformation is configured, the event is reshaped before delivery.
6. If delivery fails, EventBridge retries according to the target's retry policy.
7. After all retries are exhausted, the event goes to the target's DLQ (if configured).

### Rule Types

**Event Pattern Rules** match events by their structure. Patterns can filter on:
- `source` — who sent the event (e.g., "aws.ec2", "com.myapp.orders")
- `detail-type` — the event type (e.g., "EC2 Instance State-change Notification")
- `detail` — nested event data with prefix matching, numeric matching, etc.

**Schedule Rules** fire on a time-based schedule:
- `rate(value unit)` — periodic firing (e.g., "rate(5 minutes)", "rate(1 hour)")
- `cron(min hour dom month dow year)` — cron expression with 6 fields

### Targets

EventBridge supports 20+ target types. Each target is a separate AWS resource (`aws_cloudwatch_event_target` in Terraform, `cloudwatch.EventTarget` in Pulumi) linked to the rule by name.

Common target types:
- **Lambda** (~50% of all EB targets) — direct invocation, no role needed
- **SQS** (~20%) — enqueue event as message, no role needed
- **SNS** (~10%) — publish event to topic, no role needed
- **Step Functions** (~8%) — start execution, requires role
- **CloudWatch Log Group** (~5%) — log event, requires role
- **API Gateway / HTTP** (~5%) — invoke REST endpoint
- **ECS** (~2%) — run task, requires role with complex config

## Design Decisions

### Why Bundle Targets with the Rule

Targets are bundled with the rule (as a `repeated` field) rather than being separate OpenMCF components because:

1. **A rule without targets is useless** — it matches events but does nothing.
2. **Follows the AwsSnsTopic precedent** — SNS subscriptions are bundled with the topic for the same reason.
3. **Simplifies infra chart templates** — one resource = one complete routing configuration.
4. **Independent lifecycle is rare** — targets rarely change independently from the rule they belong to.

In Terraform, rules and targets are separate resources (`aws_cloudwatch_event_rule` and `aws_cloudwatch_event_target`). The Pulumi module creates both from the single spec, using the target `name` as the Pulumi resource name and `target_id`.

### Why google.protobuf.Struct for Event Patterns

Event patterns are JSON structures with nested arrays and objects. Using `google.protobuf.Struct` allows users to write patterns in native YAML without embedded JSON strings. The IaC module serializes the Struct to JSON before passing it to the EventBridge API.

This follows the precedent set by:
- AwsSqsQueue's `policy` field (IAM policy as Struct)
- AwsSnsTopic's `filter_policy` field (subscription filter as Struct)

### Why Mutual Exclusivity for Event Pattern and Schedule

The AWS documentation states that `event_pattern` and `schedule_expression` are mutually exclusive. While the Terraform provider does not enforce this with `ConflictsWith`, we enforce it via CEL validation because:

1. Having both is confusing — the rule fires on schedule AND when events match.
2. If a user needs both behaviors, they should create two separate rules (clearer intent, independent lifecycle).
3. The T02 planning guidance explicitly states "not both, not neither."

### Why No Map-Keyed Target Outputs

Unlike AwsSnsTopic which exports `subscription_arns` (subscription ARNs are useful for downstream resources), EventBridge target resources don't produce useful outputs for downstream consumption. The Terraform `aws_cloudwatch_event_target` resource ID is an internal identifier, not an AWS ARN. The target's `arn` is the USER's resource ARN (e.g., the Lambda function ARN), which they already know.

### Why Only SQS-Specific Config (No ECS, Batch, etc.)

The Terraform provider supports 12 target-specific configuration blocks. We include only `sqs_config` (for FIFO queue `message_group_id`) because:

1. **80/20 rule**: Lambda, SQS, SNS, and Step Functions cover ~90% of EventBridge targets. They need only `arn` + optional `role_arn`.
2. **Complexity budget**: ECS target alone has ~15 fields. Including all blocks would make the spec enormous.
3. **SQS FIFO is essential**: Without `message_group_id`, EventBridge cannot deliver to FIFO queues at all. This is a blocker, not a convenience feature.
4. **Escape hatch**: Users needing ECS/Batch/Kinesis targets can use Terraform or Pulumi directly.

### Deliberately Omitted for v1

- **Rule-level `role_arn`**: For cross-account event delivery. Niche (<5% of rules).
- **`ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS` state**: Very niche CloudTrail-specific state.
- **`force_destroy`**: For managed rules. OpenMCF manages the lifecycle.
- **Target-specific blocks**: `ecs_target`, `batch_target`, `kinesis_target`, `redshift_target`, `http_target`, `appsync_target`, `sagemaker_pipeline_target`, `run_command_targets`.
- **`target_id` auto-generation**: Derived from target `name` field.

## Terraform Provider Reference

**Rule**: `aws_cloudwatch_event_rule` from `hashicorp/aws`.
- `name` (ForceNew) — rule name
- `event_bus_name` (ForceNew, default "default")
- `event_pattern` (JSON, max 4096 chars)
- `schedule_expression` (max 256 chars)
- `state` (ENABLED / DISABLED / ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS)
- `role_arn` — rule-level IAM role (cross-account)
- `description` (max 512 chars)

**Target**: `aws_cloudwatch_event_target` from `hashicorp/aws`.
- `rule` (required, ForceNew)
- `arn` (required)
- `target_id` (optional, auto-generated)
- `event_bus_name` (must match rule)
- `role_arn` — target-level IAM role
- `input` / `input_path` / `input_transformer` (mutually exclusive)
- `dead_letter_config` (SQS ARN)
- `retry_policy` (max age + max attempts)
- Target-specific blocks: sqs_target, ecs_target, batch_target, kinesis_target, http_target, redshift_target, appsync_target, sagemaker_pipeline_target, run_command_targets

## Pulumi Resource Reference

**Rule**: `cloudwatch.EventRule` from `pulumi-aws/sdk/v7/go/aws/cloudwatch`.
**Target**: `cloudwatch.EventTarget` from `pulumi-aws/sdk/v7/go/aws/cloudwatch`.

Input properties map directly to Terraform attributes with camelCase naming. The rule ARN and name are the primary outputs.
