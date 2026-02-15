# AwsEventBridgeBus — Research Documentation

## Overview

Amazon EventBridge is a serverless event bus service that connects applications using events. It ingests events from AWS services, SaaS applications, and custom sources, then routes them to targets based on rules. EventBridge is the evolution of CloudWatch Events, providing the same core functionality with expanded features for SaaS integrations and schema management.

## Architecture

EventBridge operates around three core concepts: **event buses**, **rules**, and **targets**.

**Event buses** receive events. Every AWS account has a default event bus that automatically receives events from AWS services (EC2 state changes, S3 object operations, etc.). Custom event buses isolate application-defined events from the default bus, enabling independent access control, encryption, and monitoring.

**Rules** match incoming events against patterns and route matched events to one or more targets. Rules are attached to a specific bus.

**Targets** are AWS services that process events (Lambda, SQS, SNS, Step Functions, Kinesis, etc.).

### Custom vs Default Bus

The default bus is shared across all AWS services in an account. Custom buses provide:
- **Isolation**: Application events are separated from AWS service events.
- **Access control**: Fine-grained IAM policies per bus.
- **Encryption**: Customer-managed KMS keys for event encryption.
- **Partner integration**: SaaS partners deliver events to dedicated custom buses.

### Partner Event Sources

AWS EventBridge integrates with 30+ SaaS providers (Datadog, PagerDuty, Zendesk, etc.) via partner event sources. When you create a partner integration, AWS creates an event source in your account. You then create a custom bus with the same name as the event source, and the partner's events flow to your bus.

## Design Decisions

### Why StringValueOrRef for kms_key_identifier

The KMS key identifier field uses `StringValueOrRef` to enable infra-chart composability. In a typical infra chart, a KMS key is created as a separate resource and its ARN is wired into downstream resources. The `valueFrom` reference creates a dependency edge in the deployment DAG, ensuring the KMS key is provisioned before the bus.

### Why StringValueOrRef for dead_letter_config.arn

The DLQ ARN uses `StringValueOrRef` to enable the common pattern of defining both the bus and its DLQ in the same infra chart. The DLQ (an SQS queue) is deployed first, and the bus's `deadLetterConfig.arn` references the queue's output ARN via `valueFrom`.

### Why string + CEL for log levels (not proto enums)

Log levels (`OFF`, `ERROR`, `INFO`, `TRACE`) and include_detail values (`NONE`, `FULL`) use plain strings with CEL `in` validation rather than protobuf enums. This keeps the values provider-authentic (matching the exact AWS API strings) and avoids proto enum prefix conventions.

### Why dead_letter_config and log_config are included

The T02 planning guidance listed only `description`, `kms_key_identifier`, and `event_source_name`. Deep research into the Terraform provider (`aws_cloudwatch_event_bus`) revealed two additional nested blocks:

1. **`dead_letter_config`**: Bus-level DLQ for events that fail delivery to any rule target. This is a production best practice for event-driven architectures where event loss is unacceptable.
2. **`log_config`**: Event delivery logging with configurable verbosity. Essential for debugging event routing and monitoring delivery failures.

Both are simple nested messages that add significant production value without introducing complexity.

### Why bus policy is NOT bundled

The `aws_cloudwatch_event_bus_policy` is a separate Terraform resource (unlike SQS where policy is an attribute of the queue). Bundling it would add complexity, and most custom buses work fine with the default same-account access policy. Bus policies are a niche use case (cross-account event delivery) that can be added in a future iteration.

### Deliberately Omitted for v1

- **Bus policy**: Resource-based IAM policy for the bus. Separate TF resource, niche use case.
- **Event archive**: Replaying historical events. Complex feature with its own lifecycle.
- **Schema discovery**: EventBridge Schema Registry integration. Separate concern.

## Terraform Provider Reference

The primary Terraform resource is `aws_cloudwatch_event_bus` from the `hashicorp/aws` provider.

Key attributes:
- `name` (ForceNew) — bus name is immutable
- `event_source_name` (ForceNew) — partner source is immutable
- `kms_key_identifier` — KMS key ARN, key ID, key alias, or key alias ARN
- `dead_letter_config` — nested block with `arn` (SQS queue ARN)
- `log_config` — nested block with `level` and `include_detail`
- `description` — up to 512 characters

Related resources:
- `aws_cloudwatch_event_bus_policy` — resource-based IAM policy (not bundled)
- `aws_cloudwatch_event_rule` — rules that match and route events (separate component: AwsEventBridgeRule)
- `aws_cloudwatch_event_target` — targets attached to rules

## Pulumi Resource Reference

The Pulumi resource is `cloudwatch.EventBus` from `pulumi-aws/sdk/v7/go/aws/cloudwatch`. Input properties map directly to Terraform attributes with camelCase naming. The bus name and ARN are the primary outputs.
