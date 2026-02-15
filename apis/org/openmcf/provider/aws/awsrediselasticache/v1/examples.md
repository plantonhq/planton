## Minimal single-node Redis (YAML)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: my-redis
spec:
  engine: redis
  engineVersion: "7.1"
  description: Development Redis cache
  nodeType: cache.t3.micro
  numCacheClusters: 1
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
```

## HA non-clustered with VPC references (YAML)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: session-cache
spec:
  engine: redis
  engineVersion: "7.1"
  description: Session cache with automatic failover
  nodeType: cache.r7g.large
  numCacheClusters: 3
  automaticFailoverEnabled: true
  multiAzEnabled: true
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
  transitEncryptionMode: required
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: my-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: redis-sg
        fieldPath: status.outputs.security_group_id
  snapshotRetentionLimit: 7
  snapshotWindow: "03:00-04:00"
  maintenanceWindow: "sun:05:00-sun:06:00"
```

## Clustered (sharded) production with KMS and logging (YAML)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: product-catalog
spec:
  engine: redis
  engineVersion: "7.1"
  description: Product catalog cache with horizontal sharding
  nodeType: cache.r7g.xlarge
  numNodeGroups: 3
  replicasPerNodeGroup: 2
  automaticFailoverEnabled: true
  multiAzEnabled: true
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: redis-key
      fieldPath: status.outputs.key_arn
  subnetIds:
    - value: subnet-aaa
    - value: subnet-bbb
    - value: subnet-ccc
  securityGroupIds:
    - value: sg-123
  parameterGroupFamily: redis7
  parameters:
    - name: maxmemory-policy
      value: volatile-lru
    - name: timeout
      value: "300"
  logDeliveryConfigurations:
    - destinationType: cloudwatch-logs
      destination:
        value: /aws/elasticache/product-catalog
      logFormat: json
      logType: slow-log
    - destinationType: cloudwatch-logs
      destination:
        value: /aws/elasticache/product-catalog-engine
      logFormat: json
      logType: engine-log
  notificationTopicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: infra-alerts
      fieldPath: status.outputs.topic_arn
  snapshotRetentionLimit: 14
  snapshotWindow: "02:00-03:00"
  maintenanceWindow: "wed:04:00-wed:05:00"
```

## Valkey engine (YAML)

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsRedisElasticache
metadata:
  name: valkey-cache
spec:
  engine: valkey
  engineVersion: "7.2"
  description: Valkey in-memory cache
  nodeType: cache.t3.medium
  numCacheClusters: 2
  automaticFailoverEnabled: true
  atRestEncryptionEnabled: true
  transitEncryptionEnabled: true
```

## CLI flows

Validate:

```bash
openmcf validate --manifest redis.yaml
```

Pulumi deploy:

```bash
openmcf pulumi update --manifest redis.yaml --stack org/project/stack --module-dir apis/org/openmcf/provider/aws/awsrediselasticache/v1/iac/pulumi
```

Terraform deploy:

```bash
openmcf tofu apply --manifest redis.yaml --auto-approve
```
