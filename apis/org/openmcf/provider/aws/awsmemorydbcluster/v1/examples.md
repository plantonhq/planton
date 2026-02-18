# Examples

## Minimal Development Cluster

A single-shard cluster with no replicas, ideal for development and testing.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemorydbCluster
metadata:
  name: dev-memorydb
spec:
  region: us-east-1
  engine: redis
  engineVersion: "7.1"
  nodeType: db.t4g.small
  numShards: 1
  numReplicasPerShard: 0
  aclName: open-access
  tlsEnabled: true
```

## Production HA Cluster

A multi-shard cluster with replicas, custom ACL, snapshots, and VPC placement via cross-resource references.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemorydbCluster
metadata:
  name: session-store
spec:
  region: us-east-1
  engine: redis
  engineVersion: "7.1"
  description: Production session store with HA
  nodeType: db.r7g.large
  numShards: 2
  numReplicasPerShard: 2
  aclName: my-prod-acl
  tlsEnabled: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: memorydb-key
      fieldPath: status.outputs.key_arn
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: main-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: main-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: memorydb-sg
        fieldPath: status.outputs.security_group_id
  snapshotRetentionLimit: 7
  snapshotWindow: "03:00-04:00"
  maintenanceWindow: "sun:05:00-sun:06:00"
  parameterGroupFamily: memorydb_redis7
  parameters:
    - name: activedefrag
      value: "yes"
    - name: maxmemory-policy
      value: volatile-lru
  snsTopicArn:
    valueFrom:
      kind: AwsSnsTopic
      name: infra-alerts
      fieldPath: status.outputs.topic_arn
```

## High-Throughput with Data Tiering

A large-scale cluster using data tiering for cost-efficient handling of cold data.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemorydbCluster
metadata:
  name: analytics-store
spec:
  region: us-east-1
  engine: redis
  engineVersion: "7.1"
  description: High-throughput analytics store with data tiering
  nodeType: db.r6gd.xlarge
  numShards: 4
  numReplicasPerShard: 2
  aclName: analytics-acl
  dataTiering: true
  tlsEnabled: true
  kmsKeyId:
    value: arn:aws:kms:us-east-1:123456789012:key/mrk-abc123
  subnetIds:
    - value: subnet-0a1b2c3d4e5f00001
    - value: subnet-0a1b2c3d4e5f00002
    - value: subnet-0a1b2c3d4e5f00003
  securityGroupIds:
    - value: sg-0123456789abcdef0
  snapshotRetentionLimit: 14
  snapshotWindow: "02:00-03:00"
  maintenanceWindow: "wed:04:00-wed:05:00"
  parameterGroupFamily: memorydb_redis7
  parameters:
    - name: activedefrag
      value: "yes"
    - name: maxmemory-policy
      value: volatile-lru
    - name: timeout
      value: "300"
  snsTopicArn:
    value: arn:aws:sns:us-east-1:123456789012:infra-alerts
  autoMinorVersionUpgrade: true
```

## Infra Chart Reference Pattern

Using OpenMCF cross-resource references to wire MemoryDB into a larger infrastructure chart.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsMemorydbCluster
metadata:
  name: app-memorydb
spec:
  region: us-east-1
  engine: redis
  engineVersion: "7.1"
  nodeType: db.r7g.large
  numShards: 2
  numReplicasPerShard: 1
  aclName: app-acl
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: app-vpc
        fieldPath: status.outputs.private_subnets.[1].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: app-redis-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: app-encryption-key
      fieldPath: status.outputs.key_arn
  snapshotRetentionLimit: 7
```
