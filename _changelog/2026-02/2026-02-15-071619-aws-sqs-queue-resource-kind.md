# AwsSqsQueue Resource Kind — First AWS Expansion Component

**Date**: February 15, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi Module, Terraform Module, Documentation

## Summary

Added the AwsSqsQueue resource kind (enum 225) as the first new AWS component in the cloud-provider-expansion project. The component supports both Standard and FIFO queue types with dead letter queue routing, dual encryption modes (SSE-SQS and SSE-KMS), IAM access policies via `google.protobuf.Struct`, and comprehensive delivery tuning — backed by both Pulumi and Terraform modules with full feature parity.

## Problem Statement / Motivation

Planton's AWS coverage stood at 25 resource kinds, lacking foundational messaging services. SQS is the most fundamental building block for event-driven architectures, microservice decoupling, and serverless workflows on AWS. Without it, infra charts for serverless-api, event-driven, and microservices patterns couldn't express the message queuing layer.

### Pain Points

- No SQS support meant users couldn't model message-driven architectures declaratively
- Infra charts for event-driven patterns (Lambda -> SQS -> Lambda) were impossible
- Dead letter queue patterns — critical for production resilience — had no representation

## Solution / What's New

A complete deployment component following Planton's ideal state checklist:

### Proto API (4 files + tests)

- **spec.proto**: 14 fields covering both Standard and FIFO queue types, organized into delivery settings, FIFO-specific settings, dead letter config, encryption, and access policy sections
- **9 CEL validations**: FIFO-only field guards, encryption mutual exclusion, range validations, KMS dependency checks
- **25 spec tests**: All passing, covering every validation rule

### Key Design Decisions

- **`google.protobuf.Struct` for IAM policy**: First usage in Planton. Enables native YAML authoring of IAM policy documents instead of JSON-in-YAML. The middleware serialization path will be built to support this going forward.
- **String + CEL for FIFO fields**: `deduplication_scope` and `fifo_throughput_limit` use plain strings with CEL `in` validation instead of proto enums, keeping values provider-authentic (`"messageGroup"` not `DEDUPLICATION_SCOPE_MESSAGE_GROUP`).
- **`max_message_size` up to 1 MB**: The Terraform provider now validates up to 1,048,576 bytes (matching AWS's expanded limit), while the AWS default remains 256 KB.
- **Self-referential StringValueOrRef**: `dead_letter_config.target_arn` references `AwsSqsQueue` → `status.outputs.queue_arn`, enabling infra-chart DLQ patterns where both queues are defined in the same chart.

### IaC Modules

- **Pulumi**: 4 Go files (main.go, locals.go, outputs.go, queue.go) — clean, Terraform-readable style
- **Terraform**: 5 HCL files with feature parity (main.tf, locals.tf, outputs.tf, variables.tf, provider.tf)
- Both modules auto-append `.fifo` to FIFO queue names and serialize `google.protobuf.Struct` policy to JSON

### Documentation and Presets

- Production-quality README with field reference, use cases, and validation rules
- Examples covering: minimal, DLQ pattern, FIFO, KMS encryption, SNS fan-out with IAM policy
- Catalog page for the component registry
- Research docs with design rationale
- 2 presets: `01-standard-queue` (long polling, SSE-SQS) and `02-fifo-with-deduplication` (high-throughput FIFO with DLQ)

## Implementation Details

### Component File Tree

```
apis/dev/planton/provider/aws/awssqsqueue/v1/
├── spec.proto, api.proto, stack_input.proto, stack_outputs.proto
├── spec_test.go (25 tests)
├── README.md, examples.md, catalog-page.md
├── docs/README.md
├── presets/ (2 YAML + 2 MD)
├── iac/pulumi/module/ (main.go, locals.go, outputs.go, queue.go)
├── iac/pulumi/ (main.go, Pulumi.yaml, Makefile, debug.sh)
├── iac/tf/ (main.tf, locals.tf, outputs.tf, variables.tf, provider.tf)
└── iac/hack/manifest.yaml
```

### Infra Chart Composability

- **Upstream**: AwsKmsKey (optional, for SSE-KMS encryption)
- **Downstream**: AwsLambda (event source), AwsSnsTopic (subscription endpoint), AwsEventBridgeRule (target)
- **Self-referential**: DLQ pattern — one AwsSqsQueue references another AwsSqsQueue's `queue_arn` output

## Benefits

- **First step in AWS expansion**: Opens the path for the remaining 31 new AWS resource kinds
- **Messaging foundation**: Enables serverless-api, event-driven, and microservices infra charts
- **DLQ pattern**: Production-resilience pattern available from day one
- **Dual encryption**: Compliance-ready with both SSE-SQS (zero cost) and SSE-KMS (audit trail)
- **google.protobuf.Struct precedent**: Establishes the pattern for IAM policies across all future components

## Impact

- AWS resource coverage: 25 → 26 kinds
- New enum: `AwsSqsQueue = 225` in `cloud_resource_kind.proto`
- 44 files changed, ~2960 lines added
- All 25 validation tests passing

## Related Work

- Parent project: `20260212.01.planton-cloud-provider-expansion`
- Sub-project: `20260215.02.sp.aws-resource-expansion` (R01 of 32)
- Next in queue: R02 AwsSnsTopic (messaging fan-out), R03-R04 EventBridge (event routing)

---

**Status**: Production Ready
