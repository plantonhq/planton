# AwsServerlessElasticache Examples

## 1. Minimal Redis Cache

The simplest possible serverless Redis cache. AWS manages all defaults.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: api-cache
  org: acme
  env: dev
  id: api-cache-dev
spec:
  region: us-west-2
  engine: redis
```

## 2. Minimal Memcached Cache

Serverless Memcached for simple volatile caching.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: session-cache
  org: acme
  env: dev
  id: session-cache-dev
spec:
  region: us-west-2
  engine: memcached
  majorEngineVersion: "1.6"
```

## 3. Redis with Scaling Limits

Control cost and performance by setting explicit scaling bounds.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: app-cache
  org: acme
  env: staging
  id: app-cache-staging
spec:
  region: us-west-2
  engine: redis
  majorEngineVersion: "7"
  description: Application cache with bounded scaling
  dataStorageMinGb: 1
  dataStorageMaxGb: 50
  ecpuMin: 1000
  ecpuMax: 50000
```

## 4. VPC-Placed Cache with KMS Encryption

Deploy inside a VPC with customer-managed encryption.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: secure-cache
  org: acme
  env: prod
  id: secure-cache-prod
spec:
  region: us-east-1
  engine: redis
  majorEngineVersion: "7"
  description: VPC-placed cache with CMK encryption
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
        name: cache-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-key
      fieldPath: status.outputs.key_arn
```

## 5. Production Redis with All Features

Full-featured Redis configuration with snapshots, encryption, access control, and scaling limits.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: prod-cache
  org: acme
  env: prod
  id: prod-cache-prod
spec:
  region: us-east-1
  engine: redis
  majorEngineVersion: "7"
  description: Production session and API response cache
  dataStorageMinGb: 5
  dataStorageMaxGb: 200
  ecpuMin: 5000
  ecpuMax: 500000
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        fieldPath: status.outputs.private_subnets.[0].id
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        fieldPath: status.outputs.private_subnets.[1].id
    - valueFrom:
        kind: AwsVpc
        name: prod-vpc
        fieldPath: status.outputs.private_subnets.[2].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: cache-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-key
      fieldPath: status.outputs.key_arn
  dailySnapshotTime: "03:00"
  snapshotRetentionLimit: 14
  userGroupId: app-redis-users
```

## 6. Valkey Cache

Open-source Redis-compatible engine with the same feature set.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: valkey-cache
  org: acme
  env: staging
  id: valkey-cache-staging
spec:
  region: us-west-2
  engine: valkey
  majorEngineVersion: "8"
  description: Valkey cache for OSS compatibility
  dataStorageMaxGb: 100
  ecpuMax: 25000
  dailySnapshotTime: "04:00"
  snapshotRetentionLimit: 7
```

## 7. Memcached with Scaling Limits (Infra Chart Pattern)

Memcached as part of a web application infra chart, with VPC wiring via `valueFrom`.

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsServerlessElasticache
metadata:
  name: web-cache
  org: acme
  env: prod
  id: web-cache-prod
spec:
  region: us-east-1
  engine: memcached
  description: Web response cache for the frontend tier
  dataStorageMinGb: 1
  dataStorageMaxGb: 20
  ecpuMin: 1000
  ecpuMax: 10000
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
        name: memcached-sg
        fieldPath: status.outputs.security_group_id
```
