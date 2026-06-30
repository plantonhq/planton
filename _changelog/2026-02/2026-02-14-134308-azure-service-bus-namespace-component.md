# Azure Service Bus Namespace Component (R21)

**Date**: February 14, 2026
**Type**: Feature
**Components**: Azure Provider, API Definitions, IaC Modules, Presets

## Summary

Added AzureServiceBusNamespace as a new Azure deployment component, providing enterprise message brokering with queues and topics. This is the 22nd Azure resource in the cloud provider expansion project (R21), covering the messaging category. The component bundles namespace + queues + topics per DD03 composite bundling rules, with 12 corrections applied from deep Terraform provider research.

## Problem Statement / Motivation

The Azure provider in Planton lacked messaging infrastructure capabilities. Enterprise workloads need reliable async messaging for decoupling microservices, event-driven architectures, and ordered processing workflows. Without Service Bus, teams had to manage messaging infrastructure outside of Planton's declarative model.

### Pain Points

- No declarative messaging infrastructure in Planton for Azure
- Teams managing Service Bus manually or via separate Terraform configs outside the infra chart DAG
- No ability to wire messaging connection strings into container apps or function apps via `valueFrom` references

## Solution / What's New

A complete AzureServiceBusNamespace deployment component with:

- **4 proto files**: spec, stack_outputs, api, stack_input with comprehensive validation (37 test cases)
- **Dual IaC modules**: Pulumi (Go) and Terraform (HCL) with full feature parity
- **3 message types**: AzureServiceBusNamespaceSpec, AzureServiceBusQueue, AzureServiceBusTopic
- **7 stack outputs**: namespace_id, namespace_name, endpoint, primary_connection_string, primary_key, queue_ids, topic_ids
- **3 presets**: standard-messaging, premium-enterprise, event-driven-microservices

### Corrections from T02 Spec

12 corrections were applied after deep research into the Terraform provider source (`servicebus_namespace_resource.go`, `servicebus_queue_resource.go`, `servicebus_topic_resource.go`):

1. Added `resource_group` (StringValueOrRef) -- missing from T02, per DD05
2. Added `region` (string) -- missing from T02, per established pattern
3. Changed SKU from proto enum to string+CEL ("Basic", "Standard", "Premium")
4. Renamed `enable_partitioning` to `partitioning_enabled` (provider-authentic)
5. Added `premium_messaging_partitions` (Premium-only, ForceNew)
6. Added `minimum_tls_version` (default "1.2")
7. Added `public_network_access_enabled` (default true)
8. Expanded queue fields: lock_duration, max_delivery_count, requires_duplicate_detection, requires_session, dead_lettering_on_message_expiration, forward_to, forward_dead_lettered_messages_to
9. Expanded topic fields: default_message_ttl, requires_duplicate_detection, support_ordering
10. Added `endpoint` output (Service Bus endpoint URL)
11. Added `queue_ids` and `topic_ids` map outputs (follows database_ids pattern)
12. Added namespace name validation CEL (`^[a-zA-Z][-a-zA-Z0-9]{4,48}[a-zA-Z0-9]$`)

### Key Design Decision: Topic Subscriptions Omitted

Topic subscriptions are deliberately omitted from this component. Subscriptions have a different lifecycle -- they are typically managed by consuming teams, not the infrastructure team that provisions the namespace. This follows DD03 bundling rules which scope the bundle to namespace + queues + topics.

### SDK Limitation: ZoneRedundant

The Pulumi Azure classic SDK v6.28.0 does not expose `zone_redundant` on `NamespaceArgs`. The field is available in the Terraform module. A comment in the Pulumi module marks this for future enablement.

## Implementation Details

### Component Structure

```
apis/dev/planton/provider/azure/azureservicebusnamespace/v1/
├── spec.proto                  # 3 messages, 11+11+6 fields
├── stack_outputs.proto         # 7 outputs including maps
├── api.proto                   # KRM wiring
├── stack_input.proto           # IaC module input
├── spec_test.go                # 37 validation tests
├── README.md                   # User-facing documentation
├── examples.md                 # 7 YAML examples
├── presets/                    # 3 preset configurations
├── iac/
│   ├── hack/manifest.yaml     # Test manifest
│   ├── pulumi/                # Go module (main, locals, outputs)
│   └── tf/                    # Terraform module (5 files)
```

### Queue Features (Production-Critical)

- **lock_duration**: ISO 8601 duration for message lock (default PT1M)
- **max_delivery_count**: Failed attempts before dead-lettering (default 10)
- **requires_duplicate_detection**: Idempotent messaging via MessageId tracking
- **requires_session**: FIFO ordered processing within session groups
- **forward_to / forward_dead_lettered_messages_to**: Message routing chains

### SKU Tiers

| Tier | Queues | Topics | Sessions | Max Msg Size | SLA |
|------|--------|--------|----------|-------------|-----|
| Basic | Yes | No | No | 256 KB | 99.9% |
| Standard | Yes | Yes | Yes | 256 KB | 99.95% |
| Premium | Yes | Yes | Yes | 100 MB | 99.95% |

## Benefits

- Declarative messaging infrastructure within the Planton ecosystem
- `valueFrom` connection string wiring into container apps, function apps, web apps
- Dual IaC (Pulumi + Terraform) with feature parity
- 37 validation tests ensuring spec correctness at build time
- 3 ready-to-deploy presets for common patterns

## Impact

- **Azure resource count**: 22 of 24 complete (R21 done, R22 + R23 remaining)
- **Messaging category**: First of two messaging resources (R22 AzureEventHubNamespace pending)
- **Infra chart enablement**: Connection strings can now be wired into app hosting charts

## Related Work

- Part of 20260212.05.sp.azure-resource-expansion project
- Follows patterns established by AzureCosmosdbAccount (bundled sub-resources) and AzureRedisCache (SKU-tiered resource)
- Next in queue: R22 AzureEventHubNamespace (enum 471, id_prefix azeh)

---

**Status**: Production Ready
**Build**: go build PASS, 37/37 tests green
