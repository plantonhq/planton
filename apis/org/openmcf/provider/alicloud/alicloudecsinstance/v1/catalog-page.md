# AlicloudEcsInstance

Deploy and manage Alibaba Cloud ECS compute instances with configurable instance types, disk encryption, data disks, public IP, spot pricing, and IAM role attachment.

## Overview

AlicloudEcsInstance provisions a managed ECS virtual machine on Alibaba Cloud. It supports the full range of ECS instance families (general purpose, compute-optimized, memory-optimized, GPU, etc.), multiple disk categories (cloud_essd, cloud_ssd, cloud_efficiency), SSH key or password authentication, and flexible billing (PostPaid, PrePaid, Spot).

This component wraps a single `alicloud_instance` Terraform resource. Data disks are created inline using the resource's built-in `data_disks` block, keeping their lifecycle coupled to the instance.

## Prerequisites

- An Alibaba Cloud VPC with at least one VSwitch in the target region
- At least one security group in the same VPC
- An SSH key pair (recommended) or password for instance access
- For encrypted disks: a KMS key in the same region
- Alibaba Cloud credentials with ECS permissions

## Quick Start

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudEcsInstance
metadata:
  name: my-ecs
spec:
  region: cn-hangzhou
  instanceType: ecs.g7.large
  imageId: ubuntu_22_04_x64_20G_alibase_20230515.vhd
  vswitchId:
    value: vsw-abc123
  securityGroupIds:
    - value: sg-abc123
  keyName: my-keypair
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., "cn-hangzhou") |
| `vswitchId` | StringValueOrRef | VSwitch ID for network placement |
| `securityGroupIds` | list of StringValueOrRef | Security group IDs (at least one) |
| `instanceType` | string | ECS instance type (e.g., "ecs.g7.large") |
| `imageId` | string | OS image ID |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `instanceName` | string | metadata.name | Display name (2-128 chars) |
| `hostName` | string | auto | OS-level hostname |
| `description` | string | - | Instance description (2-256 chars) |
| `systemDisk.category` | string | cloud_essd | Disk type |
| `systemDisk.size` | int | 40 | System disk size in GB |
| `systemDisk.performanceLevel` | string | - | PL0/PL1/PL2/PL3 (ESSD only) |
| `systemDisk.encrypted` | bool | false | Enable disk encryption |
| `systemDisk.kmsKeyId` | string | - | KMS key for encryption |
| `dataDisks[].size` | int | required | Data disk size in GB |
| `dataDisks[].category` | string | cloud_essd | Disk type |
| `dataDisks[].name` | string | - | Disk display name |
| `dataDisks[].performanceLevel` | string | - | PL0/PL1/PL2/PL3 |
| `dataDisks[].encrypted` | bool | false | Enable encryption |
| `dataDisks[].kmsKeyId` | string | - | KMS key for encryption |
| `dataDisks[].snapshotId` | string | - | Create from snapshot |
| `dataDisks[].deleteWithInstance` | bool | true | Delete when instance released |
| `keyName` | string | - | SSH key pair name |
| `password` | string | - | Login password (8-30 chars) |
| `internetMaxBandwidthOut` | int | 0 | Outbound bandwidth in Mbps (>0 allocates public IP) |
| `internetChargeType` | string | - | PayByTraffic or PayByBandwidth |
| `instanceChargeType` | string | PostPaid | PostPaid or PrePaid |
| `period` | int | - | Subscription months (PrePaid) |
| `periodUnit` | string | - | Week or Month |
| `spotStrategy` | string | - | NoSpot, SpotAsPriceGo, SpotWithPriceLimit |
| `spotPriceLimit` | double | - | Max hourly price (spot only) |
| `userData` | string | - | Cloud-init script (base64) |
| `roleName` | string | - | RAM role for instance profile |
| `deletionProtection` | bool | false | Prevent accidental deletion |
| `securityEnhancementStrategy` | string | - | Active or Deactive |
| `resourceGroupId` | string | - | Resource group for organization |
| `tags` | map | - | Key-value tags |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `instance_id` | ECS instance ID (e.g., "i-bp1xxxxx") |
| `private_ip` | Primary private IP address |
| `public_ip` | Public IP (empty if no public IP) |

## Related Components

- **AlicloudVpc** -- VPC for network isolation
- **AlicloudVswitch** -- VSwitch for subnet placement
- **AlicloudSecurityGroup** -- Network access control rules
- **AlicloudEipAddress** -- Elastic IP (alternative to auto-allocated public IP)
- **AlicloudKmsKey** -- Encryption key for disk encryption
- **AlicloudRamRole** -- IAM role for instance profile
