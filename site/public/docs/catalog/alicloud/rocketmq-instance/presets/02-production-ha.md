---
title: "Production HA RocketMQ"
description: "A production-grade RocketMQ 5.x instance using the professional edition with high-availability clustering. Includes example topics (NORMAL and FIFO) with matching consumer groups to demonstrate the..."
type: "preset"
rank: "02"
presetSlug: "02-production-ha"
componentSlug: "rocketmq-instance"
componentTitle: "RocketMQ Instance"
provider: "alicloud"
icon: "package"
order: 2
---

# Production HA RocketMQ

A production-grade RocketMQ 5.x instance using the professional edition with high-availability clustering. Includes example topics (NORMAL and FIFO) with matching consumer groups to demonstrate the two most common messaging patterns. Message tracing is enabled for debugging and monitoring, and messages are retained for 7 days (168 hours).

## When to Use

- Production workloads that require message durability and cluster-level high availability
- Event-driven microservices architectures using a mix of unordered and ordered messaging
- Teams graduating from a development single-node instance to a production deployment
- Applications that need message trace functionality for debugging distributed flows

## Key Configuration Choices

- **Professional edition** (`seriesCode: professional`) -- production-ready throughput with cluster HA support; balances cost and capability for most production workloads
- **Cluster HA** (`subSeriesCode: cluster_ha`) -- multi-node cluster with automatic failover; messages are replicated across nodes for durability
- **Message processing spec** (`msgProcessSpec`) -- determines throughput capacity; choose based on your expected TPS (e.g., `rmq.p2.4xlarge` for high-throughput workloads)
- **7-day message retention** (`messageRetentionTime: 168`) -- retains messages for 7 days for replay and debugging; adjust based on your recovery and audit requirements
- **Message tracing enabled** (`traceOn: true`) -- records the full lifecycle of each message (send, deliver, ack) for observability and debugging
- **VPC-only access** -- no internet access configured; the instance is reachable only within the VPC via IP whitelist
- **Two topic types** -- NORMAL for high-throughput unordered events, FIFO for strictly ordered processing (e.g., payment or state-machine events)
- **Matched consumer groups** -- each topic has a dedicated consumer group with the appropriate delivery order type (Concurrently vs Orderly)
- **Default retry policy** (`DefaultRetryPolicy`, 16 retries) -- exponential backoff retries before messages are dead-lettered; suitable for most transient failure scenarios

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<alibaba-cloud-region>` | Alibaba Cloud region (e.g., `cn-shanghai`) | Your deployment region |
| `<your-vpc-id>` | VPC ID for the instance | `AliCloudVpc` stack outputs |
| `<your-vswitch-id>` | VSwitch ID within the VPC | `AliCloudVswitch` stack outputs |
| `<your-security-group-id>` | Security group ID for network access control | `AliCloudSecurityGroup` stack outputs |
| `<msg-process-spec>` | Throughput tier (e.g., `rmq.p2.4xlarge`) | [RocketMQ pricing page](https://www.alibabacloud.com/product/mq) |
| `<your-vpc-cidr>` | VPC CIDR for IP whitelist (e.g., `10.0.0.0/8`) | Your VPC configuration |
| `<your-team>` | Team or business unit tag | Your organizational structure |
| `<your-normal-topic>` | Topic name for unordered messages (e.g., `order-events`) | Your application design |
| `<your-fifo-topic>` | Topic name for ordered messages (e.g., `payment-events`) | Your application design |
| `<your-normal-consumer-group>` | Consumer group ID for the normal topic (e.g., `GID_order_processor`) | Your application design |
| `<your-fifo-consumer-group>` | Consumer group ID for the FIFO topic (e.g., `GID_payment_processor`) | Your application design |

## Related Presets

- **01-development-single-node** -- Minimal single-node instance for development and testing
- **03-enterprise-encrypted** -- Ultimate edition with encryption, internet access, and subscription billing
