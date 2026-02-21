---
title: "MongodbInstance"
description: "MongodbInstance deployment documentation"
icon: "package"
order: 100
componentName: "alicloudmongodbinstance"
---

# AliCloud MongodbInstance

Deploy and manage Alibaba Cloud ApsaraDB for MongoDB replica-set instances with configurable replication, multi-zone HA, encryption, and backup policies.

## Overview

AliCloudMongodbInstance provisions a managed MongoDB replica-set instance on Alibaba Cloud. It supports configurable replication factors (1, 3, 5, or 7 nodes), read-only replicas for read scaling, multi-zone high availability across three AZs, and both TDE and cloud disk encryption at rest.

This component wraps a single `alicloud_mongodb_instance` Terraform resource (replica-set mode). Sharding deployments require a separate component.

## Prerequisites

- An Alibaba Cloud VPC with at least one VSwitch in the target region
- For multi-zone HA: VSwitch(es) available in each target AZ
- For TDE encryption: A KMS key in the same region
- Alibaba Cloud credentials with permissions for ApsaraDB for MongoDB

## Quick Start

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudMongodbInstance
metadata:
  name: my-mongodb
spec:
  region: cn-hangzhou
  engineVersion: "7.0"
  dbInstanceClass: dds.mongo.mid
  dbInstanceStorage: 20
  accountPassword: "${MONGODB_PASSWORD}"
  vswitchId:
    value: vsw-abc123
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., "cn-hangzhou") |
| `vswitchId` | StringValueOrRef | VSwitch ID for network placement |
| `engineVersion` | string | MongoDB version: 4.0, 4.2, 4.4, 5.0, 6.0, 7.0 |
| `dbInstanceClass` | string | Instance specification (e.g., "dds.mongo.mid") |
| `dbInstanceStorage` | int | Storage in GB |
| `accountPassword` | string | Root account password (8-32 chars) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dbInstanceName` | string | metadata.name | Display name |
| `zoneId` | string | auto | Primary availability zone |
| `secondaryZoneId` | string | - | Standby node AZ |
| `hiddenZoneId` | string | - | Hidden node AZ (3-zone HA) |
| `replicationFactor` | int | 3 | Replica set size: 1, 3, 5, 7 |
| `readonlyReplicas` | int | - | Read replicas: 0-5 |
| `storageEngine` | string | WiredTiger | WiredTiger or RocksDB |
| `storageType` | string | - | cloud_essd1/2/3, cloud_auto, local_ssd |
| `provisionedIops` | int | - | IOPS for cloud storage |
| `instanceChargeType` | string | PostPaid | PostPaid or PrePaid |
| `securityIpList` | list | [127.0.0.1] | Allowed IP addresses |
| `securityGroupId` | string | - | ECS security group ID |
| `resourceGroupId` | string | - | Resource group for organization |
| `tags` | map | - | Key-value tags |
| `sslAction` | string | - | Open, Close, or Update |
| `tdeStatus` | string | - | "enabled" for TDE encryption |
| `encryptionKey` | string | - | KMS key ID for TDE |
| `encrypted` | bool | false | Cloud disk encryption |
| `cloudDiskEncryptionKey` | string | - | KMS key for disk encryption |
| `maintainStartTime` | string | - | Maintenance window start (UTC) |
| `maintainEndTime` | string | - | Maintenance window end (UTC) |
| `backupTime` | string | - | Backup window (e.g., "02:00Z-03:00Z") |
| `backupPeriod` | list | - | Backup days (e.g., ["Monday"]) |
| `parameters` | map | - | MongoDB engine parameters |
| `dbInstanceReleaseProtection` | bool | false | Prevent accidental deletion |
| `period` | int | - | Subscription months (PrePaid) |
| `autoRenew` | bool | false | Auto-renewal (PrePaid) |
| `autoRenewDuration` | int | - | Auto-renewal months (1-12) |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `instance_id` | MongoDB instance ID (e.g., "dds-xxxxx") |
| `replica_set_name` | Replica set name for connection strings |

## Related Components

- **AliCloudVpc** -- VPC for network isolation
- **AliCloudVswitch** -- VSwitch for subnet placement
- **AliCloudSecurityGroup** -- Network access control
- **AliCloudKmsKey** -- Encryption key for TDE
- **AliCloudPrivateDnsZone** -- Private DNS for internal resolution
