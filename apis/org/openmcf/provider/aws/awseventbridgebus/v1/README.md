# AwsEventBridgeBus

The **AwsEventBridgeBus** resource provides a standardized way to provision and manage AWS EventBridge custom event buses through OpenMCF. It supports KMS encryption, dead letter queue routing for undeliverable events, configurable logging, and partner event source integration.

## Spec Fields (80/20)

### Essential Fields (80% Use Case)

- **description**: Human-readable description of the event bus (max 512 characters).
- **kms_key_identifier**: KMS key for encrypting events on this bus. Accepts a literal key ARN/ID/alias or a `valueFrom` reference to an AwsKmsKey resource. When omitted, events are encrypted with an AWS-owned key at no additional cost.

### Dead Letter Queue (Production Best Practice)

- **dead_letter_config.arn**: ARN of the SQS queue to use as the bus-level DLQ. Events that fail delivery to any rule target on this bus are routed here. Accepts a literal value or a `valueFrom` reference to an AwsSqsQueue.

### Logging (Observability)

- **log_config.level**: Logging level — `"OFF"`, `"ERROR"`, `"INFO"`, or `"TRACE"`. Required when `log_config` is set.
- **log_config.include_detail**: Whether to include event detail in logs — `"NONE"` or `"FULL"`. Optional; defaults to `"NONE"`.

### Partner Event Sources (Niche)

- **event_source_name**: Partner event source name for SaaS integrations (e.g., Datadog, PagerDuty). Must match the pattern `aws.partner/{partner}/{...}`. The bus name (`metadata.name`) must match this value. Immutable after creation.

## Stack Outputs

After provisioning, the AwsEventBridgeBus resource provides the following outputs:

- **bus_name**: The name of the event bus — the primary identifier used in EventBridge API calls and rule configurations.
- **bus_arn**: The Amazon Resource Name (ARN) of the event bus — used for IAM policies, cross-account event delivery, and as a reference in other resources.

## How It Works

When you define an AwsEventBridgeBus resource, OpenMCF:

1. **Creates Custom Bus**: Provisions an EventBridge custom event bus with the name from `metadata.name`.
2. **Configures Encryption**: Applies the specified KMS key when `kms_key_identifier` is set.
3. **Sets Up Dead Letter Queue**: Configures bus-level DLQ routing when `dead_letter_config` is provided.
4. **Enables Logging**: Configures CloudWatch Logs delivery when `log_config` is provided.
5. **Applies Tags**: Tags the bus with OpenMCF metadata (organization, environment, resource kind, resource ID).

## Use Cases

### Application Event Bus
Create a dedicated bus to isolate your application's events from the default bus. This allows fine-grained access control and prevents cross-application interference.

### Encrypted Event Bus for Compliance
Use a customer-managed KMS key to encrypt events at rest. Required for compliance frameworks that mandate customer-controlled encryption keys (SOC 2, HIPAA, PCI DSS).

### Production Event Bus with DLQ
Attach a dead letter queue to catch events that fail delivery to rule targets. Essential for event-driven architectures where message loss is unacceptable.

### Partner Event Source Integration
Create a bus to receive events from SaaS partners (Datadog, Zendesk, PagerDuty) via the EventBridge partner integration. Set `event_source_name` to the partner's source name.

## Important Notes

### Custom vs Default Bus
- Every AWS account has a "default" event bus that receives events from AWS services automatically. This component creates **custom** buses — you cannot use it to manage the default bus.
- Custom buses receive only events that are explicitly published to them via `PutEvents` API calls or partner integrations.

### Bus Name Immutability
- The bus name (from `metadata.name`) is immutable. Changing it forces bus replacement (delete + recreate).
- For partner event buses, `metadata.name` must match `event_source_name` exactly.

### Bus-Level vs Rule-Level DLQ
- The `dead_letter_config` on this resource is the **bus-level** DLQ — it catches events that fail delivery to any rule target on this bus.
- Individual EventBridge rules can also have their own DLQ configuration for rule-specific failure handling.

### Resource-Based Policy
- Bus access policies (who can put events on this bus) are managed via a separate resource (`AwsEventBridgeBusPolicy`, not yet implemented). For most use cases, the default policy (same-account access) is sufficient.

## Validation Rules

The API enforces these validations:

1. **Description length**: Maximum 512 characters.
2. **Event source name pattern**: Must match `aws.partner/{partner}/{...}` when set.
3. **Log config level**: Must be one of `OFF`, `ERROR`, `INFO`, `TRACE`.
4. **Log config include_detail**: Must be one of `NONE`, `FULL` when set.
5. **Dead letter config ARN**: Required when `dead_letter_config` is present.

## References

- [Amazon EventBridge User Guide](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-what-is.html)
- [Custom Event Buses](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-create-event-bus.html)
- [EventBridge Encryption](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-encryption-at-rest.html)
- [EventBridge Dead-Letter Queues](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-rule-dlq.html)
- [Partner Event Sources](https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-saas.html)
