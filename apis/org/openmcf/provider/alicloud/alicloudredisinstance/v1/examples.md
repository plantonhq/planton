# AliCloudRedisInstance Examples

## Minimal: Development Cache

A basic Redis 7.0 instance for development and testing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRedisInstance
metadata:
  name: dev-redis
spec:
  region: cn-hangzhou
  instanceClass: redis.master.small.default
  password: "${REDIS_PASSWORD}"
  vswitchId:
    value: vsw-abc123
```

## Production: Multi-Zone HA with Cluster Sharding

A production Redis cluster with cross-AZ failover, multiple data shards for throughput, and read replicas for read scaling.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRedisInstance
metadata:
  name: prod-cache
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  instanceClass: redis.sharding.mid.default
  password: "${REDIS_PASSWORD}"
  vswitchId:
    valueFrom:
      name: prod-app-vswitch-a
  engineVersion: "7.0"
  zoneId: cn-shanghai-a
  secondaryZoneId: cn-shanghai-b
  shardCount: 4
  readOnlyCount: 2
  securityIps:
    - "10.0.0.0/8"
  securityGroupId: sg-app-tier
  instanceReleaseProtection: true
  maintainStartTime: "02:00Z"
  maintainEndTime: "06:00Z"
  backupPeriod:
    - Monday
    - Wednesday
    - Friday
  backupTime: "03:00Z-04:00Z"
  config:
    maxmemory-policy: allkeys-lru
    timeout: "300"
  resourceGroupId: rg-production
  tags:
    team: platform
    cost-center: infrastructure
```

## Enterprise: SSL + TDE Encryption with Prepaid Billing

A security-hardened Redis instance with TDE encryption at rest, SSL for in-transit encryption, and subscription billing for cost optimization.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudRedisInstance
metadata:
  name: secure-redis
  org: fintech-corp
  env: production
spec:
  region: cn-hangzhou
  instanceClass: redis.master.large.default
  password: "${REDIS_PASSWORD}"
  vswitchId:
    valueFrom:
      name: secure-app-vswitch
  engineVersion: "7.0"
  zoneId: cn-hangzhou-a
  secondaryZoneId: cn-hangzhou-b
  paymentType: PrePaid
  period: "12"
  autoRenew: true
  autoRenewPeriod: 3
  sslEnable: Enable
  tdeStatus: Enabled
  encryptionKey: kms-key-abc123
  securityIps:
    - "172.16.0.0/12"
  instanceReleaseProtection: true
  backupPeriod:
    - Monday
    - Tuesday
    - Wednesday
    - Thursday
    - Friday
    - Saturday
    - Sunday
  backupTime: "02:00Z-03:00Z"
  tags:
    compliance: pci-dss
    data-class: confidential
```
