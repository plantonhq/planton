# AwsEventBridgeRule

The **AwsEventBridgeRule** resource provides a standardized way to provision and manage AWS EventBridge rules with bundled targets through Planton. It supports event pattern matching, schedule-based triggering, input transformation, per-target dead letter queues, and retry policies.

## Spec Fields (80/20)

### Rule Configuration

- **event_bus_name**: Name of the event bus to attach this rule to. Defaults to "default" (the built-in AWS event bus). Accepts a `valueFrom` reference to an AwsEventBridgeBus resource. Changing this forces rule replacement.
- **description**: Human-readable description of the rule (max 512 characters).
- **event_pattern**: JSON event pattern for matching incoming events. Expressed as a structured object in YAML. Mutually exclusive with `schedule_expression`.
- **schedule_expression**: Cron or rate expression for time-based triggering. Mutually exclusive with `event_pattern`. Examples: `"rate(5 minutes)"`, `"cron(0 12 * * ? *)"`.
- **state**: Rule state — `"ENABLED"` or `"DISABLED"`. Defaults to `"ENABLED"` when not set.

### Targets

- **targets[].name**: User-assigned name for the target (max 64 chars, alphanumeric plus `-`, `_`, `.`). Used as the EventBridge `target_id` and Pulumi resource name.
- **targets[].arn**: ARN of the target resource (Lambda function, SQS queue, SNS topic, Step Functions state machine, etc.). Accepts `valueFrom` references.
- **targets[].role_arn**: IAM role for EventBridge to assume when invoking this target. Required for Step Functions, ECS, Kinesis, Batch. Not needed for Lambda, SQS, SNS (they use resource-based policies).

### Input Transformation (mutually exclusive)

- **targets[].input**: Constant JSON input to pass instead of the matched event.
- **targets[].input_path**: JSONPath expression to extract a portion of the event.
- **targets[].input_transformer**: Template-based transformation with `input_paths` (extraction) and `input_template` (assembly).

### Reliability

- **targets[].dead_letter_config.arn**: SQS queue ARN for events that fail delivery after all retries. Accepts `valueFrom` reference to AwsSqsQueue.
- **targets[].retry_policy.maximum_event_age_in_seconds**: How long to keep retrying (60-86400 seconds, default 86400).
- **targets[].retry_policy.maximum_retry_attempts**: Max retry count (0-185, default 185).

### SQS-Specific

- **targets[].sqs_config.message_group_id**: Message group ID for FIFO SQS queue targets.

## Stack Outputs

After provisioning, the AwsEventBridgeRule resource provides:

- **rule_arn**: The Amazon Resource Name (ARN) of the rule — used in IAM policies.
- **rule_name**: The name of the rule — used in EventBridge API calls.

## How It Works

When you define an AwsEventBridgeRule resource, Planton:

1. **Creates Rule**: Provisions an EventBridge rule with the name from `metadata.name`, attached to the specified event bus.
2. **Configures Matching**: Sets up either event pattern matching or schedule-based triggering.
3. **Creates Targets**: Provisions an EventTarget for each entry in `targets`, linked to the rule.
4. **Configures Reliability**: Applies dead letter queues and retry policies to individual targets.
5. **Applies Tags**: Tags the rule with Planton metadata (organization, environment, resource kind, resource ID).

## Use Cases

### Scheduled Lambda Invocation
Trigger a Lambda function on a recurring schedule (e.g., daily cleanup, hourly data sync). This is the serverless alternative to traditional cron jobs.

### Event-Driven Routing
Match specific events (EC2 state changes, S3 operations, custom application events) and route them to processing targets. The backbone of event-driven architectures.

### Fan-Out to Multiple Targets
Route a single event to multiple targets simultaneously — for example, sending an order event to both a processing Lambda and an analytics SQS queue.

### Event Transformation
Reshape event data before delivery using input transformers. Extract specific fields, add static context, or reformat the payload for the target's expected schema.

## Important Notes

### Rule vs Bus
- Rules are attached to a specific event bus. The default bus receives AWS service events automatically.
- For custom application events, create an AwsEventBridgeBus first, then attach rules to it.

### Target IAM Permissions
- **Lambda, SQS, SNS**: Use resource-based policies. EventBridge does NOT need a `role_arn` — but the target resource must grant EventBridge permission to invoke it.
- **Step Functions, ECS, Kinesis, Batch**: Require `role_arn`. EventBridge assumes this role to invoke the target.

### Event Pattern vs Schedule
- Exactly one of `event_pattern` or `schedule_expression` must be set.
- Event patterns match incoming events by structure (source, detail-type, detail fields).
- Schedule expressions fire on a time-based schedule (rate or cron).

### AWS Limits
- Maximum 5 targets per rule.
- Maximum 300 rules per event bus (soft limit, can be increased).

## Validation Rules

1. **Exactly one trigger**: Either `event_pattern` or `schedule_expression` must be set, not both.
2. **At least one target**: The `targets` list must contain at least one entry.
3. **State values**: Must be `"ENABLED"` or `"DISABLED"` when set.
4. **Target name format**: Max 64 chars, pattern `^[0-9A-Za-z_.-]+$`.
5. **Input mutual exclusion**: Only one of `input`, `input_path`, or `input_transformer` per target.
6. **Input template required**: When `input_transformer` is present, `input_template` is required.
7. **Retry policy ranges**: `maximum_event_age_in_seconds` (60-86400), `maximum_retry_attempts` (0-185).
8. **DLQ ARN required**: When `dead_letter_config` is present, `arn` is required.

## References

- [Amazon EventBridge Rules](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-rules.html)
- [Event Patterns](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-patterns.html)
- [Schedule Expressions](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-schedule-expressions.html)
- [EventBridge Targets](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-targets.html)
- [Retry Policy and DLQ](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-rule-dlq.html)
