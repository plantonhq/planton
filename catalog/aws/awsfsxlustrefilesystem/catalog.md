# AWS FSx Lustre File System

Deploys a high-performance parallel file system on Amazon FSx for Lustre with configurable deployment types (scratch or persistent), storage media, throughput tiers, optional S3 data repository integration, and CloudWatch audit logging. Lustre file systems are single-AZ and single-subnet, and the deployment type is the defining one-way door: scratch for ephemeral speed at lowest cost, persistent for durability and backups -- with PERSISTENT_2 moving S3 integration to per-directory Data Repository Associations entirely.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **FSx Lustre File System** -- a high-performance parallel file system placed in a single subnet, with configurable deployment type (SCRATCH_1, SCRATCH_2, PERSISTENT_1, PERSISTENT_2), storage media (SSD, HDD, or elastic Intelligent-Tiering), throughput tiers, and optional LZ4 data compression -- sized fresh or restored from an FSx backup via `backupId`
- **Network Interface** -- an ENI in the specified subnet for Lustre client connectivity over TCP port 988 and data channels 1018-1023
- **Log Configuration** -- created only when `logConfiguration` is provided; sends Lustre audit events (file access, creation, deletion) to a CloudWatch Logs log group
- **Metadata Configuration** -- configured only when `metadataConfiguration` is provided on PERSISTENT_2 deployments; controls metadata IOPS for file creation and listing operations
- **Automatic Backups** -- configured only on PERSISTENT deployments when `automaticBackupRetentionDays` is greater than 0
- **S3 Data Repository** -- configured when `importPath` is provided (every deployment type except PERSISTENT_2); enables transparent access to S3 objects as files on the file system, with optional auto-import policies and export
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **One subnet** in the target VPC. Lustre file systems are single-AZ -- exactly one subnet is supported. All compute resources mounting this file system must have network connectivity to this subnet. Provide the subnet ID directly or reference an AwsSubnet Infra Component via ValueFromRef.
- **A security group** that allows Lustre traffic between the file system and its clients: TCP port 988 (Lustre protocol) and TCP ports 1018-1023 (data channels). Provide the ID directly or reference an AwsSecurityGroup Infra Component.
- **A KMS key** (optional) -- for customer-managed encryption at rest instead of the default AWS-managed FSx key. Provide the ARN directly or reference an AwsKmsKey Infra Component.
- **A CloudWatch log group** (optional) -- required only when enabling audit logging. The log group must have a resource policy allowing FSx to write to it. Reference an AwsCloudwatchLogGroup Infra Component via ValueFromRef.

## Deploy

### Console

Open the deployment store, find **AWS FSx Lustre File System**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Scratch Development FSx Lustre** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxLustreFileSystem
metadata:
  name: ml-training-scratch
  org: acme-corp
  env: dev
spec:
  region: us-west-2
  deploymentType: SCRATCH_2
  storageCapacityGib: 1200
  storageType: SSD
  subnetId:
    value: "subnet-0a1b2c3d4e5f00001"
  securityGroupIds:
    - value: "sg-0a1b2c3d4e5f00001"
```

```shell
planton apply -f fsx-lustre.yaml
```

This creates a SCRATCH_2 Lustre file system with 1.2 TiB of SSD storage, no backups, no S3 integration, and no audit logging. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the file system to a VPC and security group deployed in the same InfraPipeline:

```yaml
spec:
  subnetId:
    valueFrom:
      kind: AwsSubnet
      name: private-subnet-a
      fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: lustre-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the VPC, security group, and KMS key first, then provisions the FSx Lustre file system with the resolved values.

## Key Configuration

These are the most important decisions when configuring an FSx Lustre file system. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Deployment type** -- SCRATCH_2 is the default for ephemeral processing jobs (no replication, no backups, lowest cost). PERSISTENT_2 is recommended for production workloads requiring data durability, automatic backups, and higher throughput tiers. PERSISTENT_1 supports HDD storage for cost-optimized capacity workloads. This is a ForceNew attribute.

**Storage capacity and type** -- Minimum 1200 GiB. SSD provides sub-millisecond latency for most workloads. HDD is only available with PERSISTENT_1 (with a READ/NONE `driveCacheType` arm) and offers lower cost at higher latency. INTELLIGENT_TIERING (PERSISTENT_2 only) is elastic -- no provisioned capacity at all; instead it requires `throughputCapacity` in multiples of 4000 MB/s, a `dataReadCacheConfiguration`, and a `metadataConfiguration`. Provisioned storage can be increased after creation but never decreased. Alternatively, restore from an FSx backup with `backupId` -- the file system inherits the backup's size, so capacity is not specified on a restore.

**Per-unit storage throughput** -- Applies to PERSISTENT deployments on SSD/HDD. Controls throughput per TiB of storage. PERSISTENT_2 supports up to 1000 MB/s/TiB for maximum performance. PERSISTENT_1 with HDD supports 12 or 40 MB/s/TiB for capacity-optimized workloads.

**Root squash and EFA** -- `rootSquashConfiguration` maps root users on client instances to an unprivileged UID:GID (with an exempt-NID list) -- the POSIX guardrail for shared multi-team clusters; it updates in place. `efaEnabled` (PERSISTENT_2 + metadata configuration required, ForceNew) enables OS-bypass networking and GPUDirect Storage for the most latency-sensitive HPC/ML fleets.

**S3 data repository** -- Two generations exist and cannot be mixed on one file system. The legacy in-spec `importPath`/`exportPath` arm works on SCRATCH_1/SCRATCH_2/PERSISTENT_1 and is immutable. The modern generation is the [AWS FSx Data Repository Association](/infra-catalog/aws-fsx-data-repository-association) kind -- per-directory links with independent lifecycles and per-event sync policies, and the ONLY option on PERSISTENT_2.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetId` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |
| **AwsCloudwatchLogGroup** (optional) | `logConfiguration.destination` | `status.outputs.log_group_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `file_system_id` | FSx file system identifier | EKS PersistentVolumes (FSx CSI driver), ECS task definitions, AWS Batch compute environments |
| `file_system_arn` | Amazon Resource Name of the file system | IAM policies, data repository associations |
| `dns_name` | FSx DNS name for mount commands | Lustre mount command: `mount -t lustre <dns_name>@tcp:/<mount_name> /mnt/fsx` |
| `mount_name` | Lustre mount name (auto-generated) | Combined with `dns_name` to construct the full mount path |
| `network_interface_ids` | ENI IDs created for the file system | Security group debugging, network troubleshooting |
| `vpc_id` | VPC containing the file system | Security group rules, network placement verification |
| `file_system_type_version` | Deployed Lustre version | Workload compatibility validation |
| `owner_id` | AWS account ID of the file system owner | Cross-account access configuration |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Scratch development** -- SCRATCH_2 with 1.2 TiB SSD storage. No data replication or backups. Ideal for short-lived ML training jobs, HPC simulations, and batch processing where data is sourced from S3 and results are written back. Start from the **Scratch Development FSx Lustre** preset.

**Persistent high throughput** -- PERSISTENT_2 with 2.4 TiB SSD, 1000 MB/s/TiB throughput, LZ4 compression, automatic backups, and automatic metadata IOPS. Designed for production ML training pipelines and HPC workloads that need durable, high-performance storage. Start from the **Persistent High Throughput FSx Lustre** preset.

**Persistent capacity data lake** -- PERSISTENT_1 with 6 TiB HDD, 12 MB/s/TiB throughput, LZ4 compression, and 14-day backup retention. Cost-optimized for large datasets accessed sequentially, such as data lake processing and genomics pipelines. Start from the **Persistent Capacity Data Lake FSx Lustre** preset.

**Intelligent-Tiering elastic** -- PERSISTENT_2 on INTELLIGENT_TIERING storage: no provisioned capacity at all, throughput in 4000 MB/s multiples, a read cache, and metadata IOPS. The shape for datasets whose size is unknown or highly variable. Start from the **Intelligent-Tiering Elastic FSx Lustre** preset.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) -- provides the subnet for the file system's network interface placement
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls Lustre traffic (TCP 988, 1018-1023) access to the file system
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for encryption at rest
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- receives Lustre audit events for compliance and debugging
- [**AWS FSx Data Repository Association**](/infra-catalog/aws-fsx-data-repository-association) -- links directories to S3 buckets with bidirectional sync (consumes `file_system_id`)
