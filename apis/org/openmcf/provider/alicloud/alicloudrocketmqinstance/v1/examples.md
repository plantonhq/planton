# AliCloudRocketmqInstance Examples

## Minimal: Development Single-Node Instance

A basic standard-edition instance for development and testing with no topics or internet access.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRocketmqInstance
metadata:
  name: dev-mq
spec:
  region: cn-hangzhou
  seriesCode: standard
  subSeriesCode: single_node
  vpcId:
    value: vpc-abc123
```

## Production: Professional HA with Topics and Consumer Groups

A production-grade professional instance with FIFO and normal topics, consumer groups with custom retry policies, and VSwitch placement.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRocketmqInstance
metadata:
  name: prod-mq
  org: acme-corp
  env: production
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
    cost-center: messaging
  topics:
    - topicName: order-events
      messageType: NORMAL
      remark: Order lifecycle events
    - topicName: payment-events
      messageType: FIFO
      remark: Payment processing events requiring strict ordering
    - topicName: delay-notifications
      messageType: DELAY
      remark: Delayed notification delivery
  consumerGroups:
    - consumerGroupId: GID_order_processor
      remark: Processes order lifecycle events
    - consumerGroupId: GID_payment_processor
      deliveryOrderType: Orderly
      remark: Processes payments in order
      consumeRetryPolicy:
        retryPolicy: FixedRetryPolicy
        maxRetryTimes: 5
        deadLetterTargetTopic: payment-dead-letter
    - consumerGroupId: GID_notification_sender
      remark: Sends notifications from delayed messages
```

## Enterprise: Ultimate Edition with Subscription and Internet Access

A mission-critical ultimate-edition instance with subscription billing, internet access for external clients, encryption at rest, and auto-scaling.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRocketmqInstance
metadata:
  name: enterprise-mq
  org: fintech-corp
  env: production
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
