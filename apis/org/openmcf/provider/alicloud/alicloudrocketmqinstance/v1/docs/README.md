# Alibaba Cloud RocketMQ Instance: From Manual Queues to Declarative Messaging Infrastructure

## Introduction

Message queuing is the nervous system of distributed applications. When services need to communicate asynchronously—processing orders, streaming events, coordinating microservices—they rely on a message broker to decouple producers from consumers. On Alibaba Cloud, RocketMQ is the native distributed messaging platform, originally built to handle the extreme scale of Alibaba's own infrastructure (100 billion+ messages daily during Singles' Day).

RocketMQ 5.x represents a significant architectural leap from the legacy ONS (Open Notification Service) API. It introduces VPC-native instances with configurable throughput tiers, a cleaner resource model (instances, topics, consumer groups), and a modern 2022-08-01 API that replaces the fragmented ONS endpoint. Despite this modernization, deploying a production-ready RocketMQ instance remains a multi-step orchestration problem: the instance itself, its network placement, topics with the right message types, consumer groups with appropriate retry policies, and internet access configuration all need to be coordinated correctly.

This document examines the full deployment landscape for RocketMQ 5.x—from console provisioning to control-plane automation—and explains how OpenMCF's `AliCloudRocketmqInstance` component abstracts this complexity into a single declarative manifest that handles the 80% use case while exposing the 20% of knobs that production workloads actually need.

## Historical Context: ONS to RocketMQ 5.x

### The ONS Era (Legacy)

The original Alibaba Cloud messaging service was exposed through the ONS (Open Notification Service) API. In Terraform, this meant using `alicloud_ons_instance`, `alicloud_ons_topic`, and `alicloud_ons_group` resources. The ONS API had several limitations:

- **Flat instance model**: Instances were created with minimal configuration (just a name and remark), with no control over deployment architecture, throughput tiers, or network placement.
- **Topic message types as integers**: Message types were specified as opaque integers (`0` for normal, `1` for ordered, `2` for transaction, `4` for delay, `5` for FIFO), making configuration error-prone and manifests unreadable.
- **No VPC integration**: ONS instances lacked native VPC endpoint configuration. Network access was controlled through separate mechanisms.
- **Consumer groups as "GID" strings**: The `alicloud_ons_group` resource required a `group_id` prefixed with `GID_`, an implementation detail that leaked into the user-facing API.

### The RocketMQ 5.x API (2022-08-01)

The RocketMQ 5.x API (`alicloud_rocketmq_instance`, `alicloud_rocketmq_topic`, `alicloud_rocketmq_consumer_group`) is a clean-room redesign:

- **Edition-based instances**: Controlled by `series_code` (standard, professional, ultimate) and `sub_series_code` (cluster_ha, single_node, serverless). These determine feature sets, throughput ceilings, and deployment architectures.
- **VPC-native networking**: Instances are placed directly in a VPC with VSwitch and security group configuration via a `network_info` block.
- **Typed message types**: Topics use human-readable strings (`NORMAL`, `FIFO`, `DELAY`, `TRANSACTION`) instead of opaque integers.
- **Consumer group retry policies**: Each consumer group can have an explicit retry policy (exponential backoff or fixed interval) with dead-letter topic routing.
- **Internet access control**: Optional public endpoint with configurable bandwidth billing.

The 5.x API is a strict improvement. OpenMCF targets it exclusively—there is no reason to support the legacy ONS API for new deployments.

## The RocketMQ Deployment Landscape

### Level 0: Manual Provisioning via Alibaba Cloud Console

The Alibaba Cloud console provides a wizard for creating RocketMQ 5.x instances through the "Message Queue for Apache RocketMQ" service page.

**Workflow**:
1. Select edition (Standard/Professional/Ultimate) and architecture (Single Node/Cluster HA/Serverless)
2. Configure VPC, VSwitch, and security group for network placement
3. Choose billing (Pay-As-You-Go or Subscription with period)
4. Optionally enable internet access with bandwidth billing
5. After instance creation, navigate to "Topics" tab to create topics one at a time
6. Navigate to "Groups" tab to create consumer groups one at a time
7. For each consumer group, configure retry policies separately

**Common Mistakes**:

1. **Wrong edition for the workload**: Selecting "standard/single_node" for production and discovering it lacks HA failover only during an incident. The edition is ForceNew—changing it requires destroying and recreating the instance, losing all message history.

2. **Missing VSwitch**: Creating an instance in a VPC without specifying a VSwitch, leading to the instance being placed in an arbitrary availability zone. For serverless instances, the documentation recommends at least two VSwitches for availability, but the console doesn't enforce this.

3. **Topic message type mismatch**: Creating a topic with `NORMAL` type and then attempting to publish `FIFO` messages to it. Message type is ForceNew on topics—the only fix is to delete and recreate the topic, which means losing all unconsumed messages.

4. **Missing retry policies**: Accepting the default retry policy (exponential backoff, 16 retries) for all consumer groups, even when business logic requires a fixed retry interval or different max retry count. By the time this surfaces in production, messages are cycling through 16 retries with exponential delays before hitting the dead-letter queue.

5. **No dead-letter topic**: Failing to configure a dead-letter target topic for consumer groups. When messages exhaust all retries, they are silently discarded rather than being routed to a topic for manual investigation.

**Verdict**: Acceptable for exploration and learning the RocketMQ 5.x resource model. **Unacceptable for production** due to the number of coordinated resources (instance + topics + consumer groups + retry policies) that must be configured consistently.

### Level 1: Scripted Provisioning with Alibaba Cloud CLI

The `aliyun` CLI provides imperative commands for RocketMQ management:

```bash
# Create instance
aliyun rocketmq CreateInstance \
  --SeriesCode professional \
  --SubSeriesCode cluster_ha \
  --PaymentType PayAsYouGo \
  --NetworkInfo.VpcInfo.VpcId vpc-xxx \
  --NetworkInfo.InternetInfo.InternetSpec disable

# Create topic
aliyun rocketmq CreateTopic \
  --InstanceId rmq-xxx \
  --TopicName order-events \
  --MessageType NORMAL

# Create consumer group
aliyun rocketmq CreateConsumerGroup \
  --InstanceId rmq-xxx \
  --ConsumerGroupId GID_order_processor \
  --ConsumeRetryPolicy.RetryPolicy DefaultRetryPolicy \
  --ConsumeRetryPolicy.MaxRetryTimes 16
```

**The Orchestration Problem**: A complete RocketMQ deployment requires a strict sequence of API calls:
1. Create the instance (async—must poll for `Running` status)
2. Wait for the instance to become available (can take minutes)
3. Create each topic (referencing the instance ID)
4. Create each consumer group with its retry policy (referencing the instance ID)

Scripts must handle the asynchronous nature of instance creation, retry on transient failures, and track which resources were already created for idempotency. This is significant engineering effort for what should be a declarative operation.

**Verdict**: Suitable for one-off automation or CI/CD scripts that manage individual RocketMQ resources. Not suitable for managing the full lifecycle of an instance with its bundled topics and consumer groups.

### Level 2: Infrastructure as Code (Terraform/OpenTofu)

Terraform provides the most mature IaC path for RocketMQ 5.x through the `alicloud` provider's `alicloud_rocketmq_*` resources.

**Resource Granularity**: The Terraform provider models RocketMQ as three separate resources:

```hcl
resource "alicloud_rocketmq_instance" "main" {
  instance_name   = "production-mq"
  series_code     = "professional"
  sub_series_code = "cluster_ha"
  service_code    = "rmq"
  payment_type    = "PayAsYouGo"
  commodity_code  = "ons_rmqpost_public_cn"

  network_info {
    vpc_info {
      vpc_id = "vpc-xxx"
      vswitches {
        vswitch_id = "vsw-xxx"
      }
    }
    internet_info {
      internet_spec = "disable"
      flow_out_type = "uninvolved"
    }
  }
}

resource "alicloud_rocketmq_topic" "order_events" {
  instance_id  = alicloud_rocketmq_instance.main.id
  topic_name   = "order-events"
  message_type = "NORMAL"
}

resource "alicloud_rocketmq_consumer_group" "order_processor" {
  instance_id       = alicloud_rocketmq_instance.main.id
  consumer_group_id = "GID_order_processor"

  consume_retry_policy {
    retry_policy  = "DefaultRetryPolicy"
    max_retry_times = 16
  }
}
```

**Implementation Details Hidden from Users**:

The Terraform provider requires several implementation details that users shouldn't need to know:

- `service_code`: Always `"rmq"` for RocketMQ 5.x instances. There is no alternative value.
- `commodity_code`: Varies by billing and architecture (`ons_rmqpost_public_cn` for PayAsYouGo, `ons_rmqsub_public_cn` for Subscription, `ons_rmqsrvlesspost_public_cn` for Serverless). Users must look up the correct code.
- `network_info.internet_info.flow_out_type`: Must be `"uninvolved"` when internet is disabled, `"payByTraffic"` or `"payByBandwidth"` when enabled. This is a provider-level enum that doesn't appear in the RocketMQ product documentation.
- `network_info.internet_info.internet_spec`: Must be `"enable"` or `"disable"` (string, not boolean).

These are the kind of details that trip up even experienced Terraform users and that a higher-level abstraction should hide.

**State Management**: Standard Terraform concerns apply—remote state backend, state locking, workspace management for multi-environment deployments.

**Verdict**: The standard approach for teams already using Terraform. However, the deeply nested `network_info` block and the hidden `commodity_code`/`service_code` fields make raw Terraform modules more complex than necessary for the common case.

### Level 3: Infrastructure as Code (Pulumi)

Pulumi provides equivalent functionality through the `pulumi-alicloud` SDK:

```go
instance, err := rocketmq.NewRocketMQInstance(ctx, "production-mq",
    &rocketmq.RocketMQInstanceArgs{
        InstanceName:  pulumi.String("production-mq"),
        SeriesCode:    pulumi.String("professional"),
        SubSeriesCode: pulumi.String("cluster_ha"),
        ServiceCode:   pulumi.String("rmq"),
        PaymentType:   pulumi.String("PayAsYouGo"),
        NetworkInfo: rocketmq.RocketMQInstanceNetworkInfoArgs{
            VpcInfo: rocketmq.RocketMQInstanceNetworkInfoVpcInfoArgs{
                VpcId: pulumi.String("vpc-xxx"),
            },
            InternetInfo: rocketmq.RocketMQInstanceNetworkInfoInternetInfoArgs{
                InternetSpec: pulumi.String("disable"),
                FlowOutType:  pulumi.String("uninvolved"),
            },
        },
    })
```

**Advantages over Terraform**:
- Type safety catches field name typos at compile time
- Programmatic loops for creating multiple topics and consumer groups without `for_each` quirks
- Parent-child relationships between instance and topics/groups are explicit in the Pulumi resource graph

**Same hidden complexity**: The `service_code`, `commodity_code`, `internet_spec`, and `flow_out_type` fields are equally opaque in Pulumi. A higher-level abstraction is needed regardless of which IaC tool is used underneath.

**Verdict**: Preferred for teams using Go, TypeScript, or Python for infrastructure. Same abstraction gap as Terraform for RocketMQ-specific details.

### Level 4: Control Planes (Crossplane, OpenMCF)

Control planes move beyond one-shot IaC applies to continuously reconciled infrastructure.

**Crossplane**: Can manage Alibaba Cloud RocketMQ through its provider-alicloud, but requires composing three separate Managed Resources (Instance, Topic, ConsumerGroup) with Crossplane Compositions to bundle them. This is significant boilerplate for a resource that naturally forms a composite.

**OpenMCF**: Provides a single `AliCloudRocketmqInstance` API resource that bundles the instance, topics, and consumer groups into one manifest. The control plane reconciles all three resource types as a unit, ensuring consistency.

**Verdict**: Control planes are the future for production infrastructure management. OpenMCF's composite bundling of RocketMQ resources is the right abstraction for this naturally-grouped resource.

## Comparative Analysis

| Aspect | Console | CLI | Terraform | Pulumi | OpenMCF |
|--------|---------|-----|-----------|--------|---------|
| Resource bundling | Manual coordination | Script orchestration | `for_each` loops | Programmatic loops | Native composite |
| Hidden fields (`commodity_code`, etc.) | Wizard handles | User must know | User must know | User must know | Abstracted away |
| Network config | Wizard-guided | Deeply nested JSON | Deeply nested HCL | Deeply nested Go | Flattened `vpcId`/`vswitchId` |
| Message type safety | Dropdown | String parameter | String parameter | String parameter | Proto enum with CEL validation |
| Retry policy management | Per-group UI | Per-group CLI call | Per-group resource | Per-group resource | Inline in consumer group definition |
| State management | None | None | Remote backend required | Pulumi Service or backend | Built-in |
| Drift detection | Manual audit | None | `terraform plan` | `pulumi preview` | Continuous reconciliation |

## The OpenMCF Approach

### 80/20 Design Decisions

**RocketMQ 5.x only (not ONS)**: The component targets the modern `alicloud_rocketmq_*` resources exclusively. The legacy ONS API (`alicloud_ons_*`) is not supported. This is a deliberate choice—the 5.x API is strictly superior, and supporting both would double the maintenance surface for zero user benefit.

**Composite bundling (DD07)**: Topics and consumer groups are bundled into the instance spec because they are meaningless without a parent instance. You cannot create a topic or consumer group without an instance ID. This follows the same pattern as `AliCloudRdsInstance` (which bundles databases and accounts) and `AliCloudLogProject` (which bundles log stores).

**ACL excluded**: RocketMQ ACL accounts and permission rules are intentionally not bundled. Security configuration has an independent lifecycle—IAM policies change more frequently than messaging topology, and a single instance may have ACL rules managed by different teams. This is consistent with how `AliCloudRamRole` separates identity from policy.

**Flattened network configuration**: The provider's deeply nested `network_info.vpc_info.vpc_id` / `network_info.vpc_info.vswitches[].vswitch_id` structure is flattened to top-level `vpc_id` and `vswitch_id` fields. This is consistent with every other AliCloud networking component in OpenMCF (VPC, VSwitch, SecurityGroup, NatGateway, etc.) and makes YAML manifests dramatically simpler.

**Hidden implementation details**: The `service_code` (always `"rmq"`), `commodity_code` (derived from `payment_type` and `sub_series_code`), `internet_spec` (derived from `internet_info.enabled`), and `flow_out_type` (derived from internet enablement and billing choice) are all computed internally. Users never see these fields.

**Internet access as an optional nested message**: The `internet_info` block is kept as a nested message (rather than flattened) because its fields are conditionally relevant—`flow_out_type` and `flow_out_bandwidth` only matter when internet is enabled. Flattening them would create confusing top-level fields that are ignored in the common case (VPC-only access).

### API Design

The `AliCloudRocketmqInstanceSpec` message structure:

```
AliCloudRocketmqInstanceSpec
├── region (required)
├── series_code (required: standard | professional | ultimate)
├── sub_series_code (required: cluster_ha | single_node | serverless)
├── vpc_id (required, StringValueOrRef)
├── instance_name (optional, defaults to metadata.name)
├── remark (optional)
├── payment_type (optional, default: PayAsYouGo)
├── period / period_unit / auto_renew / auto_renew_period (Subscription fields)
├── vswitch_id (optional, StringValueOrRef)
├── security_group_id (optional)
├── internet_info (optional nested message)
│   ├── enabled (default: false)
│   ├── flow_out_type (default: payByTraffic)
│   └── flow_out_bandwidth (for payByBandwidth)
├── msg_process_spec (optional, throughput tier)
├── product_info (optional nested message)
│   ├── message_retention_time
│   ├── send_receive_ratio
│   ├── auto_scaling
│   ├── trace_on
│   ├── storage_encryption (ForceNew)
│   └── storage_secret_key
├── ip_whitelists (optional)
├── resource_group_id (optional)
├── tags (optional)
├── topics[] (composite bundled)
│   ├── topic_name (required)
│   ├── message_type (default: NORMAL)
│   ├── remark
│   └── max_send_tps
└── consumer_groups[] (composite bundled)
    ├── consumer_group_id (required)
    ├── delivery_order_type (Concurrently | Orderly)
    ├── remark
    ├── max_receive_tps
    └── consume_retry_policy
        ├── retry_policy (default: DefaultRetryPolicy)
        ├── max_retry_times (default: 16)
        └── dead_letter_target_topic
```

### Foreign Key References

The `vpc_id` and `vswitch_id` fields use `StringValueOrRef`, enabling declarative cross-resource dependencies:

```yaml
vpcId:
  valueFrom:
    name: my-vpc        # references AliCloudVpc resource
vswitchId:
  valueFrom:
    name: my-vswitch    # references AliCloudVswitch resource
```

This allows the control plane to resolve VPC and VSwitch IDs from other managed resources at deployment time, rather than requiring hardcoded IDs.

## Implementation Landscape

### Pulumi Module Architecture

The Pulumi module is organized into five files:

| File | Responsibility |
|------|---------------|
| `main.go` | Orchestrates instance creation, iterates over topics and consumer groups |
| `locals.go` | Computes tags, derives `instanceName`, `paymentType`, `commodityCode`, `internetSpec`, `flowOutType`, `messageType`, `retryPolicy` |
| `outputs.go` | Defines output constant names (`instance_id`, `tcp_endpoint`, etc.) |
| `topics.go` | Creates individual `rocketmq.RocketMQTopic` resources as children of the instance |
| `consumer_groups.go` | Creates individual `rocketmq.ConsumerGroup` resources with retry policies as children of the instance |

**Key implementation patterns**:

- **Parent-child relationships**: Topics and consumer groups are created with `pulumi.Parent(instance)`, establishing explicit dependency and enabling cascading deletes.
- **Computed commodity code**: The `commodityCode()` function in `locals.go` derives the correct billing code from `sub_series_code` and `payment_type`, hiding a common source of user confusion.
- **Endpoint extraction**: The `extractEndpointUrl()` function in `main.go` searches the instance's computed `network_info.endpoints` array for `TCP_VPC` and `TCP_INTERNET` endpoint types, extracting their URLs for stack outputs.
- **Optional field handling**: Helper functions `optionalBool()`, `optionalInt()`, and `optionalString()` handle nil-to-Pulumi-type conversions cleanly.

### Terraform Module Architecture

The Terraform module mirrors the Pulumi structure:

| File | Responsibility |
|------|---------------|
| `main.tf` | `alicloud_rocketmq_instance` resource with dynamic `product_info` block |
| `topics.tf` | `alicloud_rocketmq_topic` resources via `for_each` over `local.topics_map` |
| `consumer_groups.tf` | `alicloud_rocketmq_consumer_group` resources via `for_each` over `local.consumer_groups_map` |
| `locals.tf` | Tag computation, `commodity_code` derivation, `internet_spec`/`flow_out_type` logic, collection-to-map conversions |
| `variables.tf` | Input variables with validation rules for `series_code`, `sub_series_code`, `payment_type` |
| `outputs.tf` | Instance ID, TCP/internet endpoints (extracted via comprehension over `endpoints`), topic/consumer group ID maps |

**Key Terraform patterns**:

- **`for_each` with name-keyed maps**: Topics are keyed by `topic_name` and consumer groups by `consumer_group_id`, ensuring stable resource addresses even when the list order changes.
- **Dynamic `product_info` block**: Only rendered when `msg_process_spec` is non-empty or `product_info` is non-null, avoiding empty blocks that cause provider errors.
- **Endpoint extraction via comprehension**: Uses `[for ep in ... : ep.endpoint_url if ep.endpoint_type == "TCP_VPC"][0]` wrapped in `try()` for safe extraction.

### Resources Created

A complete deployment creates:

1. **`alicloud_rocketmq_instance`** (or `rocketmq.RocketMQInstance`) — The managed RocketMQ 5.x instance with VPC networking, billing configuration, and optional internet access.
2. **`alicloud_rocketmq_topic`** (or `rocketmq.RocketMQTopic`) × N — One per entry in `spec.topics[]`, with message type and optional throughput limits.
3. **`alicloud_rocketmq_consumer_group`** (or `rocketmq.ConsumerGroup`) × M — One per entry in `spec.consumer_groups[]`, with delivery order type, retry policy, and optional dead-letter topic.

## Production Best Practices

### Edition Selection

| Workload | Recommended Edition | Architecture | Why |
|----------|-------------------|--------------|-----|
| Dev/test | standard | single_node | Lowest cost, no HA needed |
| Light production | standard | cluster_ha | HA failover at standard throughput |
| Production | professional | cluster_ha | Higher throughput ceiling, production SLA |
| Auto-scaling production | professional | serverless | Scales with demand, no capacity planning |
| Mission-critical | ultimate | cluster_ha | Highest throughput, lowest latency guarantee |

**Critical**: `series_code` and `sub_series_code` are both **ForceNew**. Changing either field requires destroying the instance and recreating it, which means losing all unconsumed messages and resetting all consumer group offsets. Choose the right edition from the start, or plan for a migration.

### Message Type Selection

| Message Type | Ordering | Latency | Use Case |
|-------------|----------|---------|----------|
| `NORMAL` | None guaranteed | Lowest | Event notifications, log ingestion, fire-and-forget |
| `FIFO` | Strict per message group | Higher | Payment processing, state machine transitions |
| `DELAY` | None (delivered after delay) | Configurable | Scheduled notifications, timeout handling |
| `TRANSACTION` | None | Higher (two-phase) | Distributed transactions requiring commit/rollback |

**Critical**: Message type is **ForceNew** on topics. Publishing a FIFO message to a NORMAL topic (or vice versa) will fail at the SDK level. Plan message types before creating topics.

### Consumer Group Retry Policies

**DefaultRetryPolicy (exponential backoff)**: Recommended for most workloads. Retries start at 1 second and increase exponentially. After 16 retries (default), the message goes to the dead-letter topic. Total retry duration is approximately 4 hours.

**FixedRetryPolicy**: Use when business logic requires consistent retry intervals (e.g., retry every 5 seconds for payment processing). Set `max_retry_times` based on the maximum acceptable delay.

**Dead-letter topics**: Always configure `dead_letter_target_topic` for consumer groups that process business-critical messages. Without it, messages that exhaust all retries are silently dropped. Create the dead-letter topic in the same instance's `topics[]` list with `NORMAL` message type.

### Security Configuration

- **VPC-only access (default)**: When `internet_info` is omitted or `enabled: false`, the instance is only accessible from within the VPC. This is the recommended configuration for backend services.
- **IP whitelisting**: Use `ip_whitelists` to restrict access to specific CIDR ranges within the VPC, especially in shared VPC environments.
- **Security groups**: Attach a security group via `security_group_id` to control network-level access to the instance's VPC endpoint.
- **Encryption at rest**: Enable `product_info.storage_encryption` with a KMS key (`storage_secret_key`) for compliance-sensitive workloads. This is **ForceNew**—it cannot be enabled after instance creation.

### Monitoring and Observability

- **Message tracing**: Enable `product_info.trace_on` for production instances to track message flow from producer to consumer. Essential for debugging delivery failures and latency issues.
- **CloudMonitor integration**: RocketMQ 5.x instances automatically report metrics to Alibaba Cloud CloudMonitor, including message production/consumption rates, consumer lag, and instance health.
- **SLS integration**: For detailed message auditing, connect the instance to an `AliCloudLogProject` via Alibaba Cloud's log service integration (configured outside this component).

### Cost Optimization

- **Pay-As-You-Go for variable workloads**: Default billing mode. No commitment, pay for actual usage. Best for development, testing, and workloads with unpredictable traffic patterns.
- **Subscription for steady-state production**: Prepaid billing with 1/2/3/6/12-month or 1/2/3-year terms at significant discounts. Use `auto_renew` to prevent accidental expiration.
- **Serverless for spiky workloads**: The `serverless` sub-series auto-scales throughput, eliminating over-provisioning costs for workloads with wide traffic variance.
- **Message retention**: The `product_info.message_retention_time` field controls how long messages are stored (in hours). Longer retention increases storage costs. Default retention is typically 72 hours; set it to the minimum your consumers need.

## Conclusion

RocketMQ 5.x on Alibaba Cloud is a capable distributed messaging platform, but deploying it correctly requires coordinating multiple resources (instance, topics, consumer groups) with careful attention to edition selection, message types, retry policies, and network configuration. The raw Terraform and Pulumi providers expose every knob but also expose implementation details (`commodity_code`, `service_code`, nested `network_info`) that most users should never need to touch.

OpenMCF's `AliCloudRocketmqInstance` component addresses this by:
- **Bundling** instance, topics, and consumer groups into a single manifest
- **Hiding** computed fields like `commodity_code`, `service_code`, `internet_spec`, and `flow_out_type`
- **Flattening** the nested `network_info` structure to simple `vpc_id` and `vswitch_id` fields
- **Validating** enum fields (`series_code`, `sub_series_code`, `message_type`, `delivery_order_type`, `retry_policy`) with CEL expressions at the API layer
- **Defaulting** sensibly (PayAsYouGo, NORMAL message type, DefaultRetryPolicy, internet disabled)

The result is a manifest that reads like a description of what you want rather than a set of provider-specific incantations.

### References

- [Alibaba Cloud RocketMQ 5.x Documentation](https://www.alibabacloud.com/help/en/apsaramq-for-rocketmq/)
- [Terraform alicloud_rocketmq_instance](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/rocketmq_instance)
- [Pulumi alicloud rocketmq](https://www.pulumi.com/registry/packages/alicloud/api-docs/rocketmq/)
- [Apache RocketMQ](https://rocketmq.apache.org/) — the open-source project behind Alibaba Cloud's managed service
