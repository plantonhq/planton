# AWS Elastic File System

Deploys a fully managed NFS file system on Amazon EFS with configurable encryption, throughput modes, lifecycle tiering, per-AZ mount targets with optional static IPv4/IPv6 addressing, an IAM resource policy, and cross-region or cross-AZ disaster-recovery replication. Automatic daily backups via AWS Backup and storage-class lifecycle transitions are switched on from the same spec. Application-level entry points live on the separate [AWS EFS Access Point](/infra-catalog/aws-efs-access-point) resource, which references this file system.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **EFS File System** -- an elastic NFS file system with configurable encryption at rest, performance mode (generalPurpose or maxIO), throughput mode (bursting, provisioned, or elastic), and optional One Zone storage
- **Mount Targets** -- one per entry in `mountTargets`, each placed in its subnet's Availability Zone; a row can pin a static IPv4/IPv6 address and choose its address family (IPv4-only, IPv6-only, or dual-stack)
- **Backup Policy** -- created only when `backupEnabled` is `true`; enables automatic daily backups via AWS Backup
- **File System Policy** -- created only when `policy` is provided; an IAM resource policy for enforcing encryption in transit, restricting principals, or preventing root access
- **Lifecycle Policies** -- configured only when `transitionToIa`, `transitionToArchive`, or `transitionToPrimaryStorageClass` are set; automates storage class transitions to reduce costs
- **Replication Configuration** -- created only when `replication` is provided; keeps a read-only replica in sync in another region and/or Availability Zone
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **At least one subnet** in the target VPC. For regional (multi-AZ) file systems, declare one mount target per AZ for maximum availability. For One Zone file systems, declare exactly one mount target in a subnet of the pinned AZ. Subnet IDs can be provided directly or referenced from an AwsSubnet Infra Component via ValueFromRef.
- **A security group** that allows inbound NFS traffic (TCP port 2049) from the clients that will mount the file system. Provide the ID directly or reference an AwsSecurityGroup Infra Component. Empty attaches the VPC's default security group.
- **A KMS key** (optional) -- required only when using a customer-managed encryption key instead of the default AWS-managed `aws/elasticfilesystem` key. Provide the ARN directly or reference an AwsKmsKey Infra Component.

## Deploy

### Console

Open the deployment store, find **AWS Elastic File System**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **General Purpose Regional EFS** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticFileSystem
metadata:
  name: app-shared-storage
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  encrypted: true
  throughputMode: bursting
  mountTargets:
    - subnetId:
        value: "subnet-0a1b2c3d4e5f00001"
    - subnetId:
        value: "subnet-0a1b2c3d4e5f00002"
  securityGroupIds:
    - value: "sg-0a1b2c3d4e5f00001"
```

```shell
planton apply -f efs.yaml
```

This creates an encrypted regional EFS file system with bursting throughput and mount targets in two AZs, with no lifecycle policies, resource policy, replication, or backup. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the file system to subnets and a security group deployed in the same InfraPipeline:

```yaml
spec:
  mountTargets:
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: private-subnet-a
          fieldPath: status.outputs.subnet_id
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: private-subnet-b
          fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: efs-nfs-sg
        fieldPath: status.outputs.security_group_id
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: data-encryption-key
      fieldPath: status.outputs.key_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security group, and KMS key first, then provisions the EFS file system with the resolved values.

## Key Configuration

These are the most important decisions when configuring an EFS file system. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Performance mode** -- Defaults to `generalPurpose`, which provides lowest per-operation latency for most workloads. Choose `maxIO` only for highly parallelized workloads with thousands of concurrent clients -- and note AWS rejects maxIO combined with elastic throughput or One Zone storage. This is a ForceNew attribute -- changing it requires replacing the file system.

**Throughput mode** -- `bursting` scales throughput with file system size (50 MiB/s per TiB). `provisioned` gives fixed throughput independent of size (set `provisionedThroughputInMibps`; AWS enforces a 24-hour cooldown between changes). `elastic` auto-scales throughput based on workload demand and is recommended for unpredictable access patterns. Elastic requires generalPurpose performance mode.

**Regional vs. One Zone** -- Leave `availabilityZoneName` empty for multi-AZ redundancy (Standard storage). Set it to a specific AZ (e.g., `us-east-1a`) for One Zone storage, which costs less precisely because data lives in a single AZ. One Zone allows exactly one mount target, in a subnet of that AZ. Suitable for dev/test or workloads that tolerate AZ-level failure. This is a ForceNew attribute.

**Lifecycle tiering** -- Configure `transitionToIa` to move files to Infrequent Access storage after a period of no access, and add `transitionToArchive` for a still-lower storage class (requires the IA transition) -- both trade per-GiB storage cost against per-access charges, so they pay off on data that goes cold. Set `transitionToPrimaryStorageClass: AFTER_1_ACCESS` to automatically warm frequently accessed files back to Standard.

**File system policy** -- Provide `policy` (a JSON IAM policy document) to enforce non-negotiable guardrails: deny unencrypted NFS connections (`aws:SecureTransport`), prevent root access from clients, or require IAM authentication for all mounts. `bypassPolicyLockoutSafetyCheck` requires a policy and should stay off unless deploying a deliberate lockout posture -- a locked-out policy is recoverable only by the account root.

**Replication** -- Provide `replication` with a destination region and/or Availability Zone to keep a read-only DR replica in sync. Setting the AZ makes the replica a cheaper One Zone file system; same-region-different-AZ is a valid shape. Replicas are always encrypted. To replicate into an existing file system, reference it in `destinationFileSystemId` -- that file system must have set its own `replicationOverwriteProtection` to `DISABLED` first.

**Access points** -- Application-level entry points (POSIX identity enforcement, root-directory pinning) are the separate [AWS EFS Access Point](/infra-catalog/aws-efs-access-point) resource. Each access point references this file system's `file_system_id` output, and Lambda functions and ECS task definitions reference the access point's own outputs.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `mountTargets[].subnetId` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsKmsKey** (optional) | `kmsKeyId`, `replication.destinationKmsKeyId` | `status.outputs.key_arn` |
| **AwsElasticFileSystem** (optional) | `replication.destinationFileSystemId` | `status.outputs.file_system_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `file_system_id` | EFS file system identifier | AwsEfsAccessPoint `fileSystemId`, EKS PersistentVolumes (EFS CSI driver), ECS task definition EFS volumes |
| `file_system_arn` | Amazon Resource Name of the file system | IAM policies for resource-level permissions |
| `dns_name` | Regional NFS mount DNS name | EC2 direct NFS mount, EKS CSI driver configuration |
| `mount_target_ids` | Map of subnet ID to mount target ID | Monitoring and troubleshooting mount target health |
| `mount_target_ips` | Map of subnet ID to mount target IPv4 address | Static NFS mount configurations, network debugging |
| `mount_target_ipv6_addresses` | Map of subnet ID to mount target IPv6 address | IPv6-only and dual-stack NFS clients |
| `mount_target_dns_names` | Map of subnet ID to AZ-specific mount target DNS | AZ-local NFS mounts to avoid cross-AZ traffic |
| `replication_destination_file_system_id` | File system ID of the replication destination | Failover automation, DR runbooks |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**General purpose regional** -- Encrypted multi-AZ file system with bursting throughput and automatic backups. The standard production configuration for shared storage across EKS pods, ECS tasks, or EC2 instances. Start from the **General Purpose Regional EFS** preset.

**One Zone dev** -- Encrypted single-AZ file system with bursting throughput. The low-cost shape for development and testing environments where AZ-level redundancy is not required. Start from the **One Zone Dev EFS** preset.

**Production elastic tiered** -- Encrypted multi-AZ file system with elastic throughput and lifecycle tiering (IA after 30 days, Archive after 90 days, warm on access). Pair it with [AWS EFS Access Point](/infra-catalog/aws-efs-access-point) resources for per-application isolation. Start from the **Production Elastic EFS with Lifecycle Tiering and DR Replication** preset.

## Works With

- [**AWS EFS Access Point**](/infra-catalog/aws-efs-access-point) -- application-specific entry points that enforce POSIX identity and root directory, referenced by Lambda and ECS task definitions
- [**AWS Subnet**](/infra-catalog/aws-subnet) -- provides the subnets for mount target placement across Availability Zones
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- controls NFS traffic (TCP 2049) access to mount targets
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides a customer-managed key for encryption at rest (and for the replica's encryption)
