# AlicloudRocketmqInstance Pulumi Module Overview

## Architecture

The module creates a RocketMQ 5.x instance and then iterates over two bundled
collections — topics and consumer groups — to create child resources. Hidden
provider fields (`commodity_code`, `service_code`, `internet_spec`, `flow_out_type`)
are derived from user-facing spec fields in `locals.go`.

```
AlicloudRocketmqInstanceStackInput
├── spec.region → alicloud.Provider
├── spec.* → rocketmq.RocketMQInstance
│   ├── networkInfo.vpcInfo (from spec.vpcId, vswitchId, securityGroupId)
│   ├── networkInfo.internetInfo (from spec.internetInfo)
│   └── productInfo (from spec.msgProcessSpec, productInfo)
├── spec.topics[] → rocketmq.RocketMQTopic (parent: instance)
│   └── per topic: topicName, messageType, remark, maxSendTps
└── spec.consumerGroups[] → rocketmq.ConsumerGroup (parent: instance)
    └── per group: consumerGroupId, deliveryOrderType, remark,
                   maxReceiveTps, consumeRetryPolicy
```

## Key Components

### Controller (`module/main.go`)

Entry point for resource creation. Orchestrates:
1. Creates the Alicloud provider scoped to `spec.region`
2. Builds the `RocketMQInstanceArgs` with VPC info, internet info, and optional product info
3. Creates the instance resource
4. Loops over `spec.topics` calling `topic()` for each, collecting ID outputs
5. Loops over `spec.consumerGroups` calling `consumerGroup()` for each, collecting ID outputs
6. Exports all five stack outputs

Also contains `extractEndpointUrl()` which searches the instance's computed
`network_info.endpoints` array for a matching endpoint type (`TCP_VPC` or
`TCP_INTERNET`) and returns its URL.

### Locals (`module/locals.go`)

Handles all input transformation and default resolution:

- **`instanceName()`** — returns `spec.instanceName` if set, else `metadata.name`
- **`paymentType()`** — returns `spec.paymentType` if set, else `"PayAsYouGo"`
- **`commodityCode()`** — derives billing code: `ons_rmqsrvlesspost_public_cn` for serverless, `ons_rmqsub_public_cn` for Subscription, `ons_rmqpost_public_cn` for PayAsYouGo
- **`internetSpec()`** — returns `"enable"` or `"disable"` based on `spec.internetInfo.enabled`
- **`flowOutType()`** — returns `"uninvolved"` when internet disabled, else the configured billing type or `"payByTraffic"`
- **`messageType()`** — returns topic's message type or `"NORMAL"`
- **`retryPolicy()`** — returns consumer group's retry policy or `"DefaultRetryPolicy"`

Tags are computed from metadata (name, id, org, env, resource_kind) merged with user-provided `spec.tags`.

### Outputs (`module/outputs.go`)

Defines five output constants matching `stack_outputs.proto`:

| Constant | Value | Source |
|----------|-------|--------|
| `OpInstanceId` | `instance_id` | `instance.ID()` |
| `OpTcpEndpoint` | `tcp_endpoint` | Extracted from endpoints (`TCP_VPC`) |
| `OpInternetEndpoint` | `internet_endpoint` | Extracted from endpoints (`TCP_INTERNET`) |
| `OpTopicIds` | `topic_ids` | Map of `topicName → topic.ID()` |
| `OpConsumerGroupIds` | `consumer_group_ids` | Map of `consumerGroupId → consumerGroup.ID()` |

### Topics (`module/topics.go`)

Creates a single `rocketmq.RocketMQTopic` resource. Called once per entry in
`spec.topics[]`. Each topic is a Pulumi child of the instance resource.

### Consumer Groups (`module/consumer_groups.go`)

Creates a single `rocketmq.ConsumerGroup` resource with its `ConsumeRetryPolicy`.
Called once per entry in `spec.consumerGroups[]`. Each consumer group is a
Pulumi child of the instance resource.

## Design Decisions

- **Parent-child relationships**: Topics and consumer groups use `pulumi.Parent(instance)` so they are logically grouped under the instance in the Pulumi state and automatically deleted when the instance is destroyed.
- **No explicit dependency ordering between topics/groups**: Topics and consumer groups are independent of each other and can be created in parallel within the Pulumi engine.
- **Commodity code hidden**: The `commodityCode()` derivation prevents users from needing to know billing-specific internal codes that vary by architecture and payment type.
- **Endpoint extraction from computed attributes**: The instance's endpoints are computed server-side after creation, so `extractEndpointUrl()` uses `ApplyT` to extract them from the Pulumi output.

## Customization Guide

| Goal | File to Modify |
|------|---------------|
| Add a new spec field to the instance | `module/main.go` (add to `instanceArgs`) |
| Change default behavior for optional fields | `module/locals.go` (modify helper functions) |
| Add a new output | `module/outputs.go` (add constant) + `module/main.go` (add `ctx.Export`) |
| Add a new bundled sub-resource type | New file in `module/` following the `topics.go` pattern |
| Change tag computation | `module/locals.go` (`initializeLocals` function) |
