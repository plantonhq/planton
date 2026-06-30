# OCI Queue Deployment Component

**Date**: February 21, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Implemented the OciQueue deployment component -- OCI's fully managed, serverless message queue for asynchronous communication between decoupled services. Supports configurable visibility timeouts, dead-letter queues, optional KMS encryption, large message support (up to 512 KB), and consumer group partitioned consumption. Second and final resource of Phase 8 (Messaging and Streaming), completing the phase.

## Problem Statement / Motivation

The Planton Oracle Cloud provider needs queue infrastructure support to enable event-driven architectures and async service communication patterns on OCI. OCI Queue is the platform's managed message queue offering, analogous to AWS SQS.

### Pain Points

- No managed message queue component existed in the OCI provider catalog
- Teams building event-driven architectures on OCI had no declarative way to provision queues
- Phase 8 (Messaging and Streaming) was incomplete with only OciStreamPool (Kafka-style) delivered

## Solution / What's New

A complete deployment component (`OciQueue`) with proto API definitions, Pulumi module (Go), and Terraform module (HCL), registered as CloudResourceKind 3371.

### Key Design Decisions

**Flattened Capabilities Model**: The OCI API models capabilities as a discriminated list with a `type` field (`CONSUMER_GROUPS`, `LARGE_MESSAGES`). The spec flattens this into `is_large_messages_enabled` (bool toggle) and `consumer_group_config` (message whose presence enables the capability). This produces cleaner YAML and provides type safety at the schema level. The IaC modules reconstruct the provider's list format internally.

**Excluded Imperative Controls**: `purge_trigger` and `purge_type` are imperative operational controls (trigger a purge by incrementing a counter), not declarative infrastructure configuration. They are excluded from the spec.

## Implementation Details

### Proto API (4 files)

- **spec.proto**: 9 fields, 1 nested message (`ConsumerGroupConfig`), no enums, no CEL rules
- **api.proto**: Standard KRM wiring (OciQueue, OciQueueStatus)
- **stack_input.proto**: OciQueueStackInput with target + provider config
- **stack_outputs.proto**: 2 outputs (`queue_id`, `messages_endpoint`)

### Spec Fields

| Field | Type | Notes |
|-------|------|-------|
| `compartment_id` | StringValueOrRef (required) | default_kind: OciCompartment |
| `custom_encryption_key_id` | StringValueOrRef | default_kind: OciKmsKey |
| `dead_letter_queue_delivery_count` | optional int32 | 0 disables DLQ |
| `retention_in_seconds` | optional int32 | ForceNew |
| `timeout_in_seconds` | optional int32 | Polling timeout |
| `visibility_in_seconds` | optional int32 | Visibility timeout |
| `channel_consumption_limit` | optional int32 | Per-channel resource cap |
| `is_large_messages_enabled` | optional bool | LARGE_MESSAGES capability |
| `consumer_group_config` | ConsumerGroupConfig | CONSUMER_GROUPS capability |

### Validation Tests

20 Ginkgo/Gomega tests (13 valid, 7 invalid scenarios) covering minimal configuration, encryption, DLQ, retention, timeout/visibility, channel limits, capabilities, consumer groups, valueFrom refs, and full configuration.

### Pulumi Module (5 files)

- `main.go`: Entry point with stack input loading
- `module/main.go`: Resources orchestrator with OCI provider setup
- `module/locals.go`: Locals struct with freeform tags from metadata labels
- `module/outputs.go`: Output constants (`queue_id`, `messages_endpoint`)
- `module/queue.go`: `queueResource()` creating `queue.NewQueue()` with conditional field assignment; `buildCapabilities()` helper reconstructing the provider's list format from flattened spec fields

### Terraform Module (5 files)

- `main.tf`: `oci_queue_queue.this` with dynamic capabilities block
- `locals.tf`: Freeform tags + capabilities list built from flattened spec fields via `concat()`
- `outputs.tf`: `queue_id` and `messages_endpoint`
- `variables.tf`: Metadata and spec type definitions
- `provider.tf`: OCI provider requirement (>= 5.0)

### Kind Registration

`OciQueue = 3371` registered under "Messaging and Streaming" section in `cloud_resource_kind.proto`, `kind_map_gen.go` regenerated.

## Benefits

- Enables declarative provisioning of OCI managed message queues
- Ergonomic YAML spec with flattened capabilities (vs. awkward discriminated list)
- Full composability via StringValueOrRef for compartment and KMS key references
- Completes Phase 8 (Messaging and Streaming): both OciStreamPool and OciQueue now available

## Impact

- **Users**: Can now provision OCI queues with dead-letter queues, encryption, large messages, and consumer groups through a single YAML manifest
- **Platform**: Phase 8 (Messaging and Streaming) is 100% complete (2/2 resources)
- **Infra Charts**: OCI Data Platform and Serverless Stack charts can now incorporate queue-based messaging

## Related Work

- **OciStreamPool** (R30): Kafka-compatible streaming, first resource of Phase 8
- **OciKmsKey** (R25): KMS encryption keys referenced via `custom_encryption_key_id`
- **OciCompartment** (R04): Compartment referenced via `compartment_id`

---

**Status**: Production Ready
