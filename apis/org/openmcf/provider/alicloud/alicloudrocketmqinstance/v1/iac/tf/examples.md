# AlicloudRocketmqInstance Terraform Examples

Apply any of the manifests below with the OpenMCF CLI:

```shell
openmcf tofu init --manifest hack/manifest.yaml
openmcf tofu plan --manifest hack/manifest.yaml
openmcf tofu apply --manifest hack/manifest.yaml --auto-approve
```

---

## Minimal: Development Single-Node Instance

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudRocketmqInstance
metadata:
  name: dev-mq
spec:
  region: cn-hangzhou
  seriesCode: standard
  subSeriesCode: single_node
  vpcId:
    value: vpc-abc123
```

```shell
openmcf tofu apply --manifest dev-mq.yaml --auto-approve
```

---

## Production: Professional HA with Topics

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudRocketmqInstance
metadata:
  name: prod-mq
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  seriesCode: professional
  subSeriesCode: cluster_ha
  vpcId:
    value: vpc-prod-shanghai
  vswitchId:
    value: vsw-prod-az-a
  msgProcessSpec: rmq.p2.4xlarge
  productInfo:
    messageRetentionTime: 168
    traceOn: true
  topics:
    - topicName: order-events
      messageType: NORMAL
    - topicName: payment-events
      messageType: FIFO
  consumerGroups:
    - consumerGroupId: GID_order_processor
    - consumerGroupId: GID_payment_processor
      deliveryOrderType: Orderly
      consumeRetryPolicy:
        retryPolicy: FixedRetryPolicy
        maxRetryTimes: 5
        deadLetterTargetTopic: payment-dead-letter
```

```shell
openmcf tofu apply --manifest prod-mq.yaml --auto-approve
```

---

## Enterprise: Subscription with Internet Access

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudRocketmqInstance
metadata:
  name: enterprise-mq
  org: fintech-corp
  env: production
spec:
  region: cn-hangzhou
  seriesCode: ultimate
  subSeriesCode: cluster_ha
  vpcId:
    value: vpc-enterprise
  vswitchId:
    value: vsw-enterprise-a
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
  topics:
    - topicName: transaction-events
      messageType: TRANSACTION
    - topicName: audit-events
      messageType: NORMAL
  consumerGroups:
    - consumerGroupId: GID_transaction_processor
      deliveryOrderType: Orderly
      consumeRetryPolicy:
        retryPolicy: FixedRetryPolicy
        maxRetryTimes: 10
        deadLetterTargetTopic: transaction-dead-letter
    - consumerGroupId: GID_audit_collector
```

```shell
openmcf tofu apply --manifest enterprise-mq.yaml --auto-approve
```

---

## After Deploying

Verify the instance status:

```shell
openmcf tofu output instance_id
openmcf tofu output tcp_endpoint
```

To tear down:

```shell
openmcf tofu destroy --manifest <manifest>.yaml --auto-approve
```
