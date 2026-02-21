---
title: "Redis Instance"
description: "Redis Instance deployment documentation"
icon: "package"
order: 100
componentName: "alicloudredisinstance"
---

# AliCloud Redis Instance

Deploys an Alibaba Cloud Redis (KVStore) instance for managed in-memory caching, session management, and real-time data processing. Supports both Redis and Memcache engines, with Redis as the default.

## What Gets Created

When you deploy an AliCloudRedisInstance resource, OpenMCF provisions:

- **KVStore Instance** -- an `alicloud_kvstore_instance` with the selected engine version, instance class, and network placement

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables or OpenMCF provider config
- **A VSwitch** -- the Redis instance is placed in a VSwitch (create one with AliCloudVswitch)
- The VSwitch's VPC and availability zone determine the instance's network placement

## Quick Start

Create a file `redis-instance.yaml`:

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudRedisInstance
metadata:
  name: my-redis
spec:
  region: cn-hangzhou
  instanceClass: redis.master.small.default
  password: "${REDIS_PASSWORD}"
  vswitchId:
    valueFrom:
      name: my-app-vswitch
```

Deploy:

```shell
openmcf apply -f redis-instance.yaml
```

This creates a Redis 7.0 instance with the default PostPaid billing and VPC password authentication.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | string | Alibaba Cloud region (e.g., `cn-hangzhou`) | Required; non-empty |
| `vswitchId` | StringValueOrRef | VSwitch ID. Can reference AliCloudVswitch via `valueFrom`. | Required |
| `instanceClass` | string | Instance specification (e.g., `redis.master.small.default`) | Required; non-empty |
| `password` | string | Authentication password (8-32 chars) | Required; 8-32 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `engineVersion` | string | `7.0` | Redis version: `2.8`, `4.0`, `5.0`, `6.0`, `7.0` |
| `instanceType` | string | `Redis` | Engine: `Redis` or `Memcache` |
| `dbInstanceName` | string | metadata.name | Instance display name (2-256 chars) |
| `zoneId` | string | | Primary availability zone |
| `secondaryZoneId` | string | | Standby AZ for multi-zone HA |
| `paymentType` | string | `PostPaid` | Billing: `PostPaid` or `PrePaid` |
| `securityIps` | list | | IP whitelist for access control |
| `securityGroupId` | string | | Security group ID |
| `resourceGroupId` | string | | Resource group for organizational grouping |
| `tags` | map | | Key-value tags |
| `shardCount` | int32 | | Data shards for cluster mode |
| `readOnlyCount` | int32 | | Read replicas in primary zone (1-9) |
| `sslEnable` | string | | SSL: `Enable`, `Disable`, `Update` |
| `tdeStatus` | string | | TDE encryption: `Enabled` |
| `encryptionKey` | string | | Custom KMS key for TDE |
| `vpcAuthMode` | string | `Open` | VPC auth: `Open` or `Close` |
| `config` | map | | Redis configuration parameters |
| `instanceReleaseProtection` | bool | `false` | Prevent accidental deletion |
| `maintainStartTime` | string | | Maintenance window start (e.g., `02:00Z`) |
| `maintainEndTime` | string | | Maintenance window end (e.g., `06:00Z`) |
| `backupPeriod` | list | | Backup days (e.g., `Monday`, `Wednesday`) |
| `backupTime` | string | | Backup time window (e.g., `02:00Z-03:00Z`) |
| `privateConnectionPrefix` | string | | Custom private connection prefix |
| `autoRenew` | bool | `false` | Auto-renewal for PrePaid |
| `autoRenewPeriod` | int32 | | Auto-renewal period in months (1-12) |
| `period` | string | | Subscription months: 1-9, 12, 24, 36 |

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | string | Redis instance ID (e.g., `r-xxxxx`) |
| `connection_domain` | string | Intranet (VPC-internal) connection domain |
| `private_connection_port` | string | Private connection port (default: 6379) |
| `private_ip` | string | Private IP address within the VSwitch |

## Related Components

- **AliCloudVswitch** -- VSwitch where the Redis instance is placed
- **AliCloudVpc** -- VPC that provides network isolation
- **AliCloudSecurityGroup** -- Network security rules for instance access
- **AliCloudKmsKey** -- Customer-managed key for TDE encryption
- **AliCloudPrivateDnsZone** -- Private DNS resolution for the instance endpoint
