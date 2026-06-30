# AwsSnsTopic Resource Kind

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, Pulumi CLI Integration

## Summary

Added the AwsSnsTopic resource kind (R02, enum 226) to Planton, providing a standardized way to provision AWS SNS topics with bundled subscriptions, KMS encryption, IAM access policies, message filtering, and subscription dead letter queues. This is the second AWS resource kind forged in the cloud provider expansion project, establishing the pub/sub messaging pattern alongside the AwsSqsQueue (R01) completed earlier.

## Problem Statement / Motivation

Planton's AWS coverage lacked a native pub/sub messaging resource. SNS is foundational to event-driven architectures on AWS — it powers fan-out notifications, cross-service event distribution, alarm routing, and serverless event pipelines. Without AwsSnsTopic, infra charts could not express the common pattern of "publish to one topic, deliver to many subscribers."

### Pain Points

- No way to declare SNS topics in Planton manifests
- No infra-chart composability for fan-out patterns (SNS → multiple SQS queues)
- No cross-resource dependency wiring for subscription endpoints (SQS queue ARNs, Lambda function ARNs)
- Missing the second half of the messaging pair (SQS queue → consumer, SNS topic → fan-out)

## Solution / What's New

A complete AwsSnsTopic deployment component following the established forge pattern:

### Proto API (4 files)

- **spec.proto**: 10 top-level fields, 3 nested messages (AwsSnsTopicSubscription, AwsSnsSubscriptionRedriveConfig), 8 CEL validations across spec and subscription levels
- **api.proto**: Kubernetes-style resource envelope (api_version, kind, metadata, spec, status)
- **stack_input.proto**: Input envelope with target resource and AWS provider config
- **stack_outputs.proto**: topic_arn, topic_name, subscription_arns (map<string, string>)

### Bundled Subscriptions

Subscriptions are defined inline with the topic as a repeated field. Each subscription has:
- Protocol (9 valid values: sqs, lambda, http, https, email, email-json, sms, firehose, application)
- Endpoint as StringValueOrRef (polymorphic — no default_kind since target varies by protocol)
- Filter policy (google.protobuf.Struct) with scope (MessageAttributes or MessageBody)
- Raw message delivery flag
- Subscription-level dead letter queue (redrive_config → AwsSqsQueue)
- Firehose role ARN (StringValueOrRef → AwsIamRole)

### IaC Modules

- **Pulumi module**: 5 Go files (main.go, locals.go, outputs.go, topic.go, subscription.go) — topic creation with conditional FIFO settings, subscription iteration with filter policy serialization and redrive policy JSON construction
- **Terraform module**: 5 HCL files (main.tf, variables.tf, outputs.tf, provider.tf, locals.tf) — topic + subscriptions via for_each keyed by subscription name

### Documentation

- User-facing README.md with spec field reference, use cases, and validation rules
- examples.md with 7 realistic examples (minimal, KMS-encrypted, fan-out with filtering, Lambda with DLQ, FIFO, multi-protocol, IAM policy)
- catalog-page.md for the web catalog
- docs/README.md with research documentation covering architecture, design decisions, and omissions

### Presets

- 01-standard-topic: Minimal Standard topic with SHA256 signatures
- 02-fifo-with-deduplication: FIFO topic with content-based dedup and high-throughput mode
- 03-fanout-to-sqs: Fan-out pattern with two filtered SQS subscriptions using valueFrom references

## Implementation Details

### Spec Design Highlights

- **FIFO fields use provider-authentic naming**: `fifo_throughput_scope` with values "Topic" / "MessageGroup" (distinct from SQS's `fifo_throughput_limit` with "perQueue" / "perMessageGroupId")
- **Subscription `name` field**: Added as a user-assigned key for the `subscription_arns` output map, following the AwsS3ObjectSet pattern for map-keyed outputs
- **Polymorphic endpoint**: StringValueOrRef without default_kind — target resource varies by protocol
- **Subscription-level CEL validations**: Protocol value validation, filter_policy_scope requires filter_policy, subscription_role_arn required for firehose protocol
- **google.protobuf.Struct for policy and filter_policy**: Native YAML authoring experience without JSON escaping

### Deliberately Omitted for v1

- Per-protocol delivery status logging (15 fields, <20% usage)
- archive_policy (FIFO-only, new feature, niche)
- Subscription delivery_policy and replay_policy (niche)
- Subscription confirmation timeout/auto-confirm (HTTP/S only, niche)

### Key Patterns from AwsSqsQueue (R01) Reused

- FIFO name auto-suffix (`.fifo`)
- AWS tags from metadata
- google.protobuf.Struct → JSON serialization
- StringValueOrRef `.GetValue()` pattern
- Zero-means-default for numeric fields

### New Pattern Introduced

- **Map-keyed subscription outputs**: Using `pulumi.StringMap` for `subscription_arns` (inspired by AwsS3ObjectSet's `object_etags` pattern)
- **Subscription iteration**: Creating multiple `sns.TopicSubscription` resources from a repeated proto field

## Benefits

- Enables fan-out notification patterns in infra charts
- Completes the messaging pair: AwsSqsQueue (point-to-point) + AwsSnsTopic (pub/sub)
- Subscription filtering reduces unnecessary message delivery
- Cross-resource wiring via StringValueOrRef creates proper dependency edges in the infra chart DAG
- Map-keyed subscription ARN outputs enable fine-grained downstream references

## Impact

- **New resource kind**: AwsSnsTopic (enum 226, id_prefix awssns)
- **Files created**: ~45 files across proto, Go, HCL, YAML, and Markdown
- **Validation tests**: 34 tests covering spec-level and subscription-level validations (all passing)
- **Infra chart readiness**: Topic can be composed with AwsSqsQueue, AwsLambda, AwsKmsKey, AwsIamRole, and AwsEventBridgeRule

## Related Work

- **AwsSqsQueue (R01)**: Completed earlier in this session. AwsSnsTopic references AwsSqsQueue for subscription endpoints and subscription DLQs.
- **AwsEventBridgeRule (R04)**: Upcoming. Will reference AwsSnsTopic as a rule target.
- **Infra charts**: The serverless-api and event-driven charts will compose AwsSnsTopic with AwsSqsQueue and AwsLambda.

---

**Status**: Production Ready
**Timeline**: Single session (R02 of AWS expansion project)
