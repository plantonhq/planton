# AWS FSx OpenZFS File System

Deploys a fully managed NFS file system on Amazon FSx for OpenZFS with configurable deployment types (single-AZ or multi-AZ with automatic failover), ZSTD/LZ4 compression, snapshots, cloning, per-user/group quotas, and root volume NFS export configuration. Deployment type is the defining one-way door: it fixes the availability posture and the throughput ceiling, and only MULTI_AZ_1 supports elastic Intelligent-Tiering storage.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **FSx OpenZFS File System** -- an NFS file system with configurable deployment type (SINGLE_AZ_1, SINGLE_AZ_2, MULTI_AZ_1), SSD storage, throughput capacity, and optional disk IOPS configuration
- **Root Volume** -- automatically created with the file system; configurable compression (ZSTD, LZ4, or none), NFS export settings, record size, read-only mode, and per-user/group storage quotas
- **Network Interface(s)** -- one ENI for single-AZ deployments, two ENIs for multi-AZ with floating IP for failover
- **Disk IOPS Configuration** -- configured only when `diskIopsConfiguration` is provided; controls provisioned SSD IOPS in AUTOMATIC or USER_PROVISIONED mode
- **Automatic Backups** -- configured only when `automaticBackupRetentionDays` is greater than 0
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **One or two subnets** in the target VPC. Single-AZ deployments require exactly one subnet. Multi-AZ (MULTI_AZ_1) requires two subnets in different Availability Zones, plus a `preferredSubnetId`, `endpointIpAddressRange`, and `routeTableIds`. Provide subnet IDs directly or reference AwsSubnet Infra Components via ValueFromRef.
- **A security group** that allows NFS traffic between the file system and its clients: TCP port 111 (portmapper), TCP port 2049 (NFS), and TCP ports 20001-20003 (NFS mount). Provide the ID directly or reference an AwsSecurityGroup Infra Component.
- **A KMS key** (optional) -- for customer-managed encryption at rest instead of the default AWS-managed FSx key. Provide the ARN directly or reference an AwsKmsKey Infra Component.

## Deploy

### Console

Open the deployment store, find **AWS FSx OpenZFS File System**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single-AZ Development** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOpenzfsFileSystem
metadata:
  name: app-nfs-storage
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  deploymentType: SINGLE_AZ_2
  storageCapacityGib: 256
  throughputCapacity: 160
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
  securityGroupIds:
    - value: "sg-0a1b2c3d4e5f00001"
```

```shell
planton apply -f fsx-openzfs.yaml
```

This creates a single-AZ OpenZFS file system with 256 GiB SSD storage, 160 MB/s throughput, default root volume settings (no compression, 128 KiB record size), no backups, and automatic IOPS scaling. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the file system to a VPC and security group deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: private-subnet-a
        fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: openzfs-nfs-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the VPC, security group, and KMS key first, then provisions the OpenZFS file system with the resolved values.

## Key Configuration

These are the most important decisions when configuring an FSx OpenZFS file system. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Deployment type** -- SINGLE_AZ_2 is the recommended default for most workloads, supporting throughput up to 10,240 MB/s. SINGLE_AZ_HA_1 and SINGLE_AZ_HA_2 add a standby file server in the SAME Availability Zone for in-AZ failover. MULTI_AZ_1 provides automatic failover across two AZs with floating IP addresses for high availability -- and is the only type supporting Intelligent-Tiering. SINGLE_AZ_1 is the legacy generation with a lower throughput ceiling. This is a ForceNew attribute.

**Storage source and media** -- SSD is provisioned capacity (64-524,288 GiB), sized fresh or restored from an FSx backup via `backupId` (the restore inherits the backup's size). INTELLIGENT_TIERING (MULTI_AZ_1 only, ForceNew) is elastic -- no capacity dial at all; it requires a `readCacheConfiguration` so hot data is served at SSD latency.

**Throughput capacity** -- Absolute MB/s value from a fixed set of tiers. SINGLE_AZ_2 and MULTI_AZ_1 support 160-10,240 MB/s. Can be changed after creation to scale performance up or down without replacing the file system.

**Deletion behavior** -- By default, deletion FAILS while child volumes or snapshots exist (the safety interlock). Set `deleteOptions` to `DELETE_CHILD_VOLUMES_AND_SNAPSHOTS` to destroy them with the file system. When `skipFinalBackup` is off, `finalBackupTags` ride on the final backup taken at deletion.

**Root volume configuration** -- Configure `rootVolumeConfiguration` to set compression (`ZSTD` for best ratio, `LZ4` for lowest CPU overhead), NFS client access rules, ZFS record size (4-1024 KiB -- smaller for random I/O, larger for sequential), and per-user/group storage quotas. These settings apply to the root volume's NFS mount point.

**Single-AZ vs. Multi-AZ** -- Single-AZ offers lower cost and higher maximum throughput tiers. Multi-AZ adds automatic failover with floating IPs, requiring `endpointIpAddressRange` (a non-overlapping CIDR within your VPC) and `routeTableIds` for AWS-managed failover routing.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSubnet** (optional, multi-AZ) | `preferredSubnetId` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `file_system_id` | FSx file system identifier | EKS PersistentVolumes (OpenZFS CSI driver), ECS task definitions |
| `file_system_arn` | Amazon Resource Name of the file system | IAM policies for resource-level permissions |
| `dns_name` | NFS mount DNS name | Mount command: `mount -t nfs <dns_name>:/fsx /mnt/fsx` |
| `endpoint_ip_address` | Floating IP for multi-AZ failover, primary ENI IP for single-AZ | Static mount configurations, DNS record creation |
| `root_volume_id` | Root volume identifier | Parent reference when creating child OpenZFS volumes |
| `network_interface_ids` | ENI IDs for the file system | Security group debugging, network troubleshooting |
| `vpc_id` | VPC containing the file system | Security group rules, network placement verification |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single AZ development** -- SINGLE_AZ_2 with minimal storage and throughput. Cost-effective for development, testing, and CI/CD workloads that do not require cross-AZ redundancy. Start from the **Single-AZ Development** preset.

**Single AZ production** -- SINGLE_AZ_2 with higher throughput, ZSTD compression on the root volume, automatic backups, and customer-managed KMS encryption. Suitable for production NFS workloads including web serving, content management, and application data stores. Start from the **Single-AZ Production** preset.

**Multi AZ high availability** -- MULTI_AZ_1 with automatic failover across two AZs, endpoint IP address range, route table configuration, and automatic backups. Designed for business-critical NFS workloads requiring continuous availability. Start from the **Multi-AZ High Availability** preset.

**Multi-AZ Intelligent-Tiering** -- MULTI_AZ_1 on elastic INTELLIGENT_TIERING storage with a read cache: no capacity dial, storage grows and shrinks with the data. The shape for large datasets whose size is unpredictable. Start from the **Multi-AZ Intelligent-Tiering FSx OpenZFS** preset.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) -- provides subnets for file system network interface placement
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls NFS traffic (TCP 111, 2049, 20001-20003) access to the file system
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for encryption at rest
