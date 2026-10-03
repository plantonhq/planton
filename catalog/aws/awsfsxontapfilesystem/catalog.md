# AWS FSx ONTAP File System

Deploys a fully managed NetApp ONTAP file system on Amazon FSx with multi-protocol access (NFS, SMB, iSCSI), configurable deployment types (single-AZ or multi-AZ with automatic failover), scale-out HA pairs, and built-in data services including snapshots, compression, deduplication, and SnapMirror replication. The file system alone serves no data: clients mount through Storage Virtual Machines and volumes, which are their own kinds (AwsFsxOntapStorageVirtualMachine, AwsFsxOntapVolume) that attach by `file_system_id`.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **FSx ONTAP File System** -- an enterprise-grade file system with configurable deployment type (SINGLE_AZ_1, SINGLE_AZ_2, MULTI_AZ_1, MULTI_AZ_2), SSD primary storage (the only media ONTAP supports -- cost tiering happens per volume), throughput sized per HA pair or for the whole file system, and optional scale-out with up to 12 HA pairs on SINGLE_AZ_2
- **Management Endpoint** -- provides SSH (ONTAP CLI) and REST API access for advanced administration including LIF management, SnapMirror configuration, and aggregate monitoring
- **Intercluster Endpoint** -- enables NetApp SnapMirror replication between FSx ONTAP file systems for cross-region disaster recovery
- **Disk IOPS Configuration** -- configured only when `diskIopsConfiguration` is provided; controls provisioned SSD IOPS in AUTOMATIC or USER_PROVISIONED mode
- **Automatic Backups** -- configured only when `automaticBackupRetentionDays` is greater than 0
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **One or two subnets** in the target VPC. Single-AZ deployments require exactly one subnet. Multi-AZ deployments require two subnets in different Availability Zones, plus a `preferredSubnetId` designating the active file server's AZ. Provide subnet IDs directly or reference an AwsVpc Infra Component via ValueFromRef.
- **A security group** that allows traffic for NFS (TCP 2049), SMB (TCP 445), iSCSI (TCP 3260), portmapper (TCP 111), mountd (TCP 635), NFS lock/status (TCP 4045-4046), and ONTAP REST API (TCP 443). Provide the ID directly or reference an AwsSecurityGroup Infra Component.
- **A KMS key** (optional) -- for customer-managed encryption at rest instead of the default AWS-managed FSx key. Provide the ARN directly or reference an AwsKmsKey Infra Component.
- **An endpoint IP address range** (multi-AZ only) -- a CIDR block within the VPC that does not overlap with existing subnets, used for floating IPs during failover.

## Deploy

### Console

Open the deployment store, find **AWS FSx ONTAP File System**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Single-AZ Development FSx ONTAP** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapFileSystem
metadata:
  name: enterprise-nas
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  deploymentType: SINGLE_AZ_2
  storageCapacityGib: 1024
  throughputCapacityPerHaPair: 384
  subnetIds:
    - value: "subnet-0a1b2c3d4e5f00001"
  securityGroupIds:
    - value: "sg-0a1b2c3d4e5f00001"
```

```shell
planton apply -f fsx-ontap.yaml
```

This creates a single-AZ ONTAP file system with 1 TiB SSD storage, 384 MB/s throughput (the smallest tier AWS accepts on SINGLE_AZ_2), one HA pair, no backups, and automatic IOPS scaling. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the file system to a VPC and security group deployed in the same InfraPipeline:

```yaml
spec:
  subnetIds:
    - valueFrom:
        kind: AwsVpc
        name: production-vpc
        fieldPath: status.outputs.private_subnets.[0].id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: ontap-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the VPC, security group, and KMS key first, then provisions the ONTAP file system with the resolved values.

## Key Configuration

These are the most important decisions when configuring an FSx ONTAP file system. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Deployment type** -- SINGLE_AZ_2 is the recommended default for most workloads, supporting scale-out HA pairs (1-12) that can be increased without replacement. MULTI_AZ_2 provides automatic failover across two AZs for high availability, fixed at 1 HA pair. This is a ForceNew attribute.

**Throughput sizing** -- exactly one of two arms carries the value. `throughputCapacityPerHaPair` (the current-generation arm, recommended) sets the throughput for each HA pair: 384/768/1536/3072/6144 MB/s on SINGLE_AZ_2 and MULTI_AZ_2 (1536/3072/6144 with multiple pairs), or 128-4096 MB/s on the first-generation types. `throughputCapacity` (the first-generation arm) sizes the whole file system at 128-4096 MB/s. Total throughput equals the per-pair tier multiplied by `haPairs`; only SINGLE_AZ_2 scales out (1-12 pairs, added in place). Gen-2 types scale throughput in place; gen-1 types replace the file system.

**Storage type** -- ONTAP supports only SSD primary storage; AWS rejects HDD and INTELLIGENT_TIERING for this file system type at create time. Cost tiering is achieved per volume via tiering policies on AwsFsxOntapVolume, which move cold data to the built-in elastic capacity pool. This is a ForceNew attribute.

**ONTAP administration** -- Set `fsxAdminPassword` to enable SSH and REST API access to the ONTAP CLI for advanced operations like SnapMirror configuration, LIF management, and aggregate monitoring. Omit if only NFS/SMB/iSCSI data access is needed.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |
| **AwsSubnet** (multi-AZ only) | `preferredSubnetId` | `status.outputs.subnet_id` |
| **AwsSubnet** (optional, multi-AZ) | `routeTableIds` | `status.outputs.route_table_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId` | `status.outputs.key_arn` |

AwsVpc subnet outputs (e.g. `status.outputs.private_subnets.[*].id`) also satisfy the subnet fields when the whole network rides one AwsVpc resource.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `file_system_id` | FSx file system identifier | Storage Virtual Machine creation, IAM policy resource references |
| `file_system_arn` | Amazon Resource Name of the file system | IAM policies for resource-level permissions |
| `management_dns_name` | ONTAP CLI and REST API endpoint | SSH administration: `ssh fsxadmin@<endpoint>` |
| `management_ip_addresses` | Management endpoint IPs | Direct IP access when DNS is unavailable |
| `intercluster_dns_name` | SnapMirror replication endpoint | Cross-region disaster recovery peering |
| `intercluster_ip_addresses` | Intercluster endpoint IPs | SnapMirror peering when DNS is unavailable |
| `network_interface_ids` | ENI IDs for the file system | Security group debugging, network troubleshooting |
| `vpc_id` | VPC containing the file system | Security group rules, network placement verification |
| `owner_id` | AWS account ID owning the file system | Cross-account IAM conditions, audit tooling |

Data-access endpoints (NFS/SMB/iSCSI DNS names) are deliberately not outputs of the file system -- they live on the Storage Virtual Machine (AwsFsxOntapStorageVirtualMachine), which attaches by `file_system_id`.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Single AZ development** -- SINGLE_AZ_2 with 1 TiB SSD, 384 MB/s throughput (the smallest gen-2 tier), 1 HA pair, no backups. Cost-effective for development, testing, and workloads that do not require cross-AZ redundancy. Start from the **Single-AZ Development FSx ONTAP** preset.

**Single AZ production** -- SINGLE_AZ_2 with SSD storage, higher throughput, automatic backups, and customer-managed KMS encryption. Suitable for production NAS workloads, database storage, and VMware Cloud on AWS environments. Start from the **Single-AZ Production FSx ONTAP** preset.

**Multi AZ high availability** -- MULTI_AZ_2 with automatic failover across two AZs, SSD storage, backups, and endpoint IP range configuration. Designed for business-critical workloads requiring sub-second failover and zero data loss. Start from the **Multi-AZ High Availability FSx ONTAP** preset.

**Scale-out HA pairs** -- SINGLE_AZ_2 with multiple HA pairs multiplying throughput and IOPS; pairs can be added in place as the workload grows. The shape for the largest single-namespace performance requirements. Start from the **Scale-Out FSx ONTAP (Multiple HA Pairs)** preset.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) -- provides subnets for file system network interface placement
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls multi-protocol traffic (NFS, SMB, iSCSI, ONTAP API) access to the file system
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for encryption at rest
- [**AWS FSx ONTAP Storage Virtual Machine**](/infra-catalog/aws-fsx-ontap-storage-virtual-machine) -- the data-serving layer that attaches by `file_system_id`; clients mount through its endpoints
