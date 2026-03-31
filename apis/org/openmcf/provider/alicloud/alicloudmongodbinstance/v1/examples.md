# AliCloudMongodbInstance Examples

## Minimal: Development Instance

A basic MongoDB 7.0 replica-set instance for development and testing.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudMongodbInstance
metadata:
  name: dev-mongodb
spec:
  region: cn-hangzhou
  engineVersion: "7.0"
  dbInstanceClass: dds.mongo.mid
  dbInstanceStorage: 20
  accountPassword: "${MONGODB_PASSWORD}"
  vswitchId:
    value: vsw-abc123
```

## Production: Multi-Zone HA with Read Replicas

A production MongoDB 6.0 replica set deployed across three availability zones for maximum fault tolerance, with read replicas for read scaling and daily backups.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudMongodbInstance
metadata:
  name: prod-mongodb
  org: acme-corp
  env: production
spec:
  region: cn-shanghai
  engineVersion: "6.0"
  dbInstanceClass: mongo.x8.large
  dbInstanceStorage: 200
  accountPassword: "${MONGODB_PASSWORD}"
  vswitchId:
    valueFrom:
      name: prod-app-vswitch-a
  zoneId: cn-shanghai-a
  secondaryZoneId: cn-shanghai-b
  hiddenZoneId: cn-shanghai-c
  replicationFactor: 5
  readonlyReplicas: 2
  storageType: cloud_essd2
  provisionedIops: 3000
  securityIpList:
    - "10.0.0.0/8"
  securityGroupId: sg-app-tier
  dbInstanceReleaseProtection: true
  maintainStartTime: "02:00Z"
  maintainEndTime: "06:00Z"
  backupPeriod:
    - Monday
    - Wednesday
    - Friday
  backupTime: "03:00Z-04:00Z"
  parameters:
    operationProfiling.slowOpThresholdMs: "200"
  resourceGroupId: rg-production
  tags:
    team: platform
    cost-center: infrastructure
```

## Enterprise: TDE Encryption with Prepaid Billing

A security-hardened MongoDB instance with TDE encryption at rest, SSL for in-transit encryption, and subscription billing for cost optimization.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AliCloudMongodbInstance
metadata:
  name: secure-mongodb
  org: fintech-corp
  env: production
spec:
  region: cn-hangzhou
  engineVersion: "7.0"
  dbInstanceClass: mongo.x8.xlarge
  dbInstanceStorage: 500
  accountPassword: "${MONGODB_PASSWORD}"
  vswitchId:
    valueFrom:
      name: secure-vswitch
  zoneId: cn-hangzhou-a
  secondaryZoneId: cn-hangzhou-b
  hiddenZoneId: cn-hangzhou-c
  replicationFactor: 3
  storageEngine: WiredTiger
  storageType: cloud_essd3
  instanceChargeType: PrePaid
  period: 12
  autoRenew: true
  autoRenewDuration: 3
  sslAction: Open
  tdeStatus: enabled
  encryptionKey: kms-key-abc123
  securityIpList:
    - "172.16.0.0/12"
  dbInstanceReleaseProtection: true
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
