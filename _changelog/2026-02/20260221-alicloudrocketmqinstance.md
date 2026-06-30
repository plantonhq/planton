# AliCloudRocketmqInstance Component Added

**Date**: 2026-02-21
**Component**: AliCloudRocketmqInstance
**Enum**: 3120
**ID Prefix**: acrmq

## Summary

Added the AliCloudRocketmqInstance deployment component -- manages a RocketMQ 5.x instance with bundled topics and consumer groups in Alibaba Cloud. RocketMQ is a distributed messaging and streaming platform supporting normal, FIFO, delayed, and transactional messages. The component uses the newer RocketMQ 5.x API (2022-08-01) rather than the legacy ONS API, providing VPC-integrated instances with configurable throughput tiers, billing modes, and optional internet access.

## What Was Created

### API Definition
- `apis/dev/planton/provider/alicloud/alicloudrocketmqinstance/v1/` -- Full proto API (spec, api, stack_input, stack_outputs)
- Registered `AliCloudRocketmqInstance = 3120` in `CloudResourceKind` enum under a new Messaging category
- Nested messages: `AliCloudRocketmqTopic`, `AliCloudRocketmqConsumerGroup`, `AliCloudRocketmqConsumeRetryPolicy`, `AliCloudRocketmqInternetInfo`, `AliCloudRocketmqProductInfo`

### IaC Modules
- **Pulumi** (Go): Creates alicloud provider, RocketMQ instance with nested network_info (VPC + internet) and product_info blocks, then iterates topics and consumer groups as child resources. Extracts VPC and internet endpoints from computed `network_info.endpoints`.
- **Terraform** (HCL): `alicloud_rocketmq_instance` with dynamic `product_info` and `vswitches` blocks, `alicloud_rocketmq_topic` and `alicloud_rocketmq_consumer_group` via `for_each`. Endpoint extraction via `try()` on computed endpoints list.

### Tests
- Ginkgo/Gomega spec validation tests: 24 specs covering valid inputs (minimal, professional with topics, subscription billing, internet access, product info with encryption, consumer groups with retry, transaction topics) and invalid inputs (wrong api_version, wrong kind, missing metadata, missing spec, empty region, invalid series_code, invalid sub_series_code, missing vpc_id, invalid payment_type, invalid message_type, empty topic_name, empty consumer_group_id, invalid delivery_order_type, invalid auto_renew_period, max_retry_times exceeds 1000, invalid flow_out_type, invalid retry_policy)

### Documentation
- README.md with edition matrix, bundled resource rationale, and build/test commands
- examples.md with 3 YAML examples (minimal dev, production with topics/consumer groups, enterprise with subscription/internet/encryption)
- catalog-page.md with complete field reference tables for all nested structures

## Design Decisions (Deviations from T02)

- **RocketMQ v5 instead of ONS**: T02 designed around the legacy ONS API (`alicloud_ons_instance`). Switched to RocketMQ 5.x (`alicloud_rocketmq_instance`) which provides VPC integration, edition tiers, billing configuration, and richer consumer group semantics. This is a significant scope increase but the right choice for a production platform.
- **Flattened VPC networking**: The provider's deeply nested `network_info.vpc_info` block is flattened -- `vpc_id` and `vswitch_id` are promoted to spec root with `StringValueOrRef`, consistent with every other AliCloud component.
- **Internet access as optional nested message**: `internet_info` stays nested because `flow_out_type` is conditionally relevant. When omitted, defaults to internet disabled (`internet_spec: "disable"`, `flow_out_type: "uninvolved"`).
- **`msg_process_spec` at top level**: Promoted from `product_info` to spec root as the primary sizing knob.
- **`service_code` hardcoded**: Always "rmq" for RocketMQ. Not exposed to users.
- **`commodity_code` derived**: Computed from `payment_type` + `sub_series_code`. Not exposed.
- **Consumer group retry policy optional**: TF provider requires `consume_retry_policy`, but spec makes it optional and defaults to `DefaultRetryPolicy` in IaC code for simpler consumer group definitions.
- **Excluded `alicloud_rocketmq_account` and `alicloud_rocketmq_acl`**: ACL/auth has independent lifecycle and is a security-sensitive concern better managed separately.
- **Consumer groups instead of groups**: Uses the v5 `consumer_group` (with retry policies, delivery order, TPS limits) instead of the legacy ONS `group` (basic group_type).

## Verification

- `go build ./...` -- PASS
- `go vet ./...` -- PASS
- `go test ./...` -- PASS (24/24 specs)
- `terraform init` -- PASS (alicloud provider v1.271.0)
- `terraform validate` -- PASS
