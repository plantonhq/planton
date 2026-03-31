---
title: "RocketMQ Instance"
description: "RocketMQ Instance deployment documentation"
icon: "package"
order: 100
componentName: "alicloudrocketmqinstance"
---

# AliCloud RocketMQ Instance

Deploys an Alibaba Cloud RocketMQ 5.x managed message broker with bundled topics and consumer groups. Supports edition-based throughput tiers (standard, professional, ultimate), deployment architectures (single node, cluster HA, serverless), and optional public internet access.

## What Gets Created

When you deploy an AliCloudRocketmqInstance resource, OpenMCF provisions:

- **RocketMQ 5.x Instance** — a managed message broker placed in the specified VPC with configurable edition, billing, and optional internet access
- **Topics** — one `alicloud_rocketmq_topic` per entry in `spec.topics[]`, each with a message type (NORMAL, FIFO, DELAY, or TRANSACTION) and optional throughput limits
- **Consumer Groups** — one `alicloud_rocketmq_consumer_group` per entry in `spec.consumerGroups[]`, each with delivery ordering, retry policy, and optional dead-letter topic routing

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **A VPC** where the instance will be deployed (referenced via `vpcId`)
- **A VSwitch** (optional but recommended) for placement in a specific availability zone
- **A security group** (optional) to control network-level access to the instance's VPC endpoint

## Quick Start

Create a file `rocketmq.yaml`:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRocketmqInstance
metadata:
  name: my-mq
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudRocketmqInstance.my-mq
spec:
  region: cn-hangzhou
  seriesCode: standard
  subSeriesCode: single_node
  vpcId:
    value: vpc-xxx
```

Deploy:

```shell
openmcf apply -f rocketmq.yaml
```

This creates a standard-edition single-node RocketMQ instance in the specified VPC with pay-as-you-go billing and no internet access.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region (e.g., `cn-hangzhou`, `us-west-1`). | Required; non-empty |
| `seriesCode` | `string` | Edition series: `standard`, `professional`, `ultimate`. Determines feature set and throughput ceiling. ForceNew. | Required; must be one of the three values |
| `subSeriesCode` | `string` | Deployment architecture: `cluster_ha`, `single_node`, `serverless`. ForceNew. | Required; must be one of the three values |
| `vpcId` | `StringValueOrRef` | VPC where the instance is deployed. ForceNew. Can reference an `AliCloudVpc` resource via `valueFrom`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `instanceName` | `string` | `metadata.name` | Human-readable instance name (2-64 characters). |
| `remark` | `string` | — | Instance description. |
| `paymentType` | `string` | `PayAsYouGo` | Billing method: `PayAsYouGo` or `Subscription`. |
| `period` | `int` | — | Subscription period length. Only for Subscription billing. |
| `periodUnit` | `string` | — | Subscription period unit: `Month` or `Year`. |
| `autoRenew` | `bool` | — | Enable auto-renewal for Subscription instances. |
| `autoRenewPeriod` | `int` | — | Auto-renewal period: 1, 2, 3, 6, or 12. |
| `vswitchId` | `StringValueOrRef` | — | VSwitch for VPC endpoint placement. ForceNew. Can reference an `AliCloudVswitch` resource via `valueFrom`. |
| `securityGroupId` | `string` | — | Security group for VPC endpoint access control. ForceNew. |
| `internetInfo.enabled` | `bool` | `false` | Enable public internet endpoint. ForceNew. |
| `internetInfo.flowOutType` | `string` | `payByTraffic` | Internet billing: `payByBandwidth` or `payByTraffic`. ForceNew. |
| `internetInfo.flowOutBandwidth` | `int` | — | Bandwidth in Mb/s (1-1000). Only for `payByBandwidth`. |
| `msgProcessSpec` | `string` | — | Throughput tier (e.g., `rmq.s1.micro`, `rmq.p2.4xlarge`, `rmq.u2.4xlarge`). Valid values depend on `seriesCode`. |
| `productInfo.messageRetentionTime` | `int` | — | Message retention in hours. Longer retention increases storage costs. |
| `productInfo.sendReceiveRatio` | `double` | — | Send/receive capacity ratio (0.2-0.5). |
| `productInfo.autoScaling` | `bool` | — | Enable throughput auto-scaling. |
| `productInfo.traceOn` | `bool` | — | Enable message trace for debugging and monitoring. |
| `productInfo.storageEncryption` | `bool` | — | Enable encryption at rest. ForceNew. |
| `productInfo.storageSecretKey` | `string` | — | KMS key for encryption at rest. Only when `storageEncryption` is true. ForceNew. |
| `ipWhitelists` | `string[]` | — | IP addresses or CIDR blocks allowed to access the instance. |
| `resourceGroupId` | `string` | — | Alibaba Cloud resource group ID. |
| `tags` | `map(string)` | — | Tags to apply to the instance. |
| `topics` | `list` | `[]` | Topics to create within the instance. See topic fields below. |
| `consumerGroups` | `list` | `[]` | Consumer groups to create within the instance. See consumer group fields below. |

### Topic Fields (`topics[]`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `topicName` | `string` | (required) | Topic name, unique within the instance. ForceNew. |
| `messageType` | `string` | `NORMAL` | Message type: `NORMAL`, `FIFO`, `DELAY`, `TRANSACTION`. ForceNew. |
| `remark` | `string` | — | Topic description. |
| `maxSendTps` | `int` | — | Maximum send transactions per second. |

### Consumer Group Fields (`consumerGroups[]`)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `consumerGroupId` | `string` | (required) | Consumer group ID, unique within the instance. ForceNew. |
| `deliveryOrderType` | `string` | — | Message delivery ordering: `Concurrently` (parallel) or `Orderly` (per message group). |
| `remark` | `string` | — | Consumer group description. |
| `maxReceiveTps` | `int` | — | Maximum receive transactions per second. |
| `consumeRetryPolicy.retryPolicy` | `string` | `DefaultRetryPolicy` | Retry strategy: `DefaultRetryPolicy` (exponential backoff) or `FixedRetryPolicy` (fixed interval). |
| `consumeRetryPolicy.maxRetryTimes` | `int` | `16` | Maximum retry attempts (0-1000). |
| `consumeRetryPolicy.deadLetterTargetTopic` | `string` | — | Topic for messages that exhaust all retries. |

## Examples

### Development Single-Node Instance

A minimal standard-edition instance for development and testing:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRocketmqInstance
metadata:
  name: dev-mq
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudRocketmqInstance.dev-mq
spec:
  region: cn-hangzhou
  seriesCode: standard
  subSeriesCode: single_node
  vpcId:
    value: vpc-abc123
```

### Production HA with Topics and Consumer Groups

A professional-edition cluster with FIFO and normal topics, consumer groups with custom retry policies, and VSwitch placement:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRocketmqInstance
metadata:
  name: prod-mq
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: messaging
    pulumi.openmcf.org/stack.name: prod.AliCloudRocketmqInstance.prod-mq
spec:
  region: cn-shanghai
  seriesCode: professional
  subSeriesCode: cluster_ha
  vpcId:
    valueFrom:
      name: prod-vpc
  vswitchId:
    valueFrom:
      name: prod-vswitch-a
  msgProcessSpec: rmq.p2.4xlarge
  productInfo:
    messageRetentionTime: 168
    traceOn: true
  ipWhitelists:
    - "10.0.0.0/8"
  tags:
    team: platform
  topics:
    - topicName: order-events
      messageType: NORMAL
      remark: Order lifecycle events
    - topicName: payment-events
      messageType: FIFO
      remark: Payment processing requiring strict ordering
    - topicName: delay-notifications
      messageType: DELAY
      remark: Delayed notification delivery
  consumerGroups:
    - consumerGroupId: GID_order_processor
      remark: Processes order lifecycle events
    - consumerGroupId: GID_payment_processor
      deliveryOrderType: Orderly
      consumeRetryPolicy:
        retryPolicy: FixedRetryPolicy
        maxRetryTimes: 5
        deadLetterTargetTopic: payment-dead-letter
    - consumerGroupId: GID_notification_sender
```

### Enterprise with Subscription, Encryption, and Internet Access

A mission-critical ultimate-edition instance with subscription billing, public internet access, encryption at rest, and auto-scaling:

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRocketmqInstance
metadata:
  name: enterprise-mq
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: fintech-corp
    pulumi.openmcf.org/project: messaging
    pulumi.openmcf.org/stack.name: prod.AliCloudRocketmqInstance.enterprise-mq
spec:
  region: cn-hangzhou
  seriesCode: ultimate
  subSeriesCode: cluster_ha
  vpcId:
    valueFrom:
      name: enterprise-vpc
  vswitchId:
    valueFrom:
      name: enterprise-vswitch-a
  securityGroupId: sg-mq-access
  paymentType: Subscription
  period: 12
  periodUnit: Month
  autoRenew: true
  autoRenewPeriod: 3
  msgProcessSpec: rmq.u2.4xlarge
  productInfo:
    messageRetentionTime: 336
    autoScaling: true
    traceOn: true
    storageEncryption: true
    storageSecretKey: kms-key-abc123
  internetInfo:
    enabled: true
    flowOutType: payByTraffic
  ipWhitelists:
    - "172.16.0.0/12"
    - "203.0.113.0/24"
  resourceGroupId: rg-production
  tags:
    compliance: soc2
    data-class: confidential
  topics:
    - topicName: transaction-events
      messageType: TRANSACTION
      remark: Two-phase commit transaction messages
    - topicName: audit-events
      messageType: NORMAL
      remark: Audit trail events
  consumerGroups:
    - consumerGroupId: GID_transaction_processor
      deliveryOrderType: Orderly
      consumeRetryPolicy:
        retryPolicy: FixedRetryPolicy
        maxRetryTimes: 10
        deadLetterTargetTopic: transaction-dead-letter
    - consumerGroupId: GID_audit_collector

```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | RocketMQ instance ID assigned by Alibaba Cloud |
| `tcp_endpoint` | `string` | VPC-internal TCP endpoint for producing and consuming messages |
| `internet_endpoint` | `string` | Public internet TCP endpoint (empty when internet access is disabled) |
| `topic_ids` | `map(string)` | Map of topic names to their Alibaba Cloud resource IDs |
| `consumer_group_ids` | `map(string)` | Map of consumer group IDs to their Alibaba Cloud resource IDs |

## Related Components

- [AliCloudVpc](/docs/catalog/alicloud/vpc) — provides the VPC for instance network placement
- [AliCloudVswitch](/docs/catalog/alicloud/vswitch) — provides the VSwitch for availability zone placement
- [AliCloudSecurityGroup](/docs/catalog/alicloud/security-group) — controls network-level access to the instance endpoint
- [AliCloudKmsKey](/docs/catalog/alicloud/kms-key) — provides the encryption key for storage encryption at rest
