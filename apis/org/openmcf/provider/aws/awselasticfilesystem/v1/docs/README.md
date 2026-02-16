# AWS Elastic File System: Architecture Reference

This document provides a deep technical reference for Amazon EFS as deployed via the AwsElasticFileSystem API. It covers protocol fundamentals, storage and throughput models, security, cost, and integration patterns.

---

## 1. NFS Protocol Basics

### How EFS Works

Amazon EFS is a fully managed implementation of the Network File System (NFS) protocol. It exposes a POSIX-compliant file system over the network, allowing multiple clients to read and write files concurrently with standard file semantics (open, read, write, close, lock, etc.).

**Protocol stack:** EFS supports NFSv4.0 and NFSv4.1. Clients mount the file system using standard NFS mount commands or the AWS-provided `amazon-efs-utils` mount helper, which adds TLS encryption for data in transit.

**Architecture:**

1. **File system** — A logical namespace (a namespace of files and directories) stored in AWS. It has a unique ID (`fs-xxxxxxxx`) and is regional.
2. **Mount targets** — Network endpoints per Availability Zone. Each mount target is an ENI in a subnet. Clients connect to a mount target to mount the file system.
3. **Access points** — Optional application-specific entry points that enforce a POSIX identity and root directory.

**Data flow:** Client → NFS request → Mount target (ENI) → EFS backend → Durable storage (replicated across AZs for Standard storage class).

**Consistency:** EFS provides strong read-after-write consistency. Writes are durable once acknowledged. Multiple clients see the same view of the file system. File locking (flock, POSIX locks) is supported for coordination across processes.

**NFSv4 vs NFSv3:** EFS supports only NFSv4.0 and NFSv4.1. NFSv4 uses a single port (2049) and is stateful; the client maintains a session with the server. NFSv3 used multiple RPC ports and was stateless; EFS does not support NFSv3. Ensure clients use NFSv4-capable mount options.

### NFS vs Other Storage Protocols

| Protocol | Storage Type | Typical Use |
|----------|--------------|-------------|
| EFS (NFS) | File system | Shared POSIX files, concurrent access |
| EBS | Block | Databases, boot volumes, single-instance |
| S3 | Object | Blobs, backups, static assets |
| FSx Lustre | Parallel file | HPC, ML training |

EFS fills the niche of shared, mutable file storage that requires POSIX semantics. Applications that expect a traditional file system (e.g., `open()`, `read()`, `write()`, file locking) work with EFS without modification. Object storage (S3) requires SDK-style APIs and does not support random file access or locking.

---

## 2. Mount Targets

### One Per AZ, Placed in Subnets

A mount target is a network interface in a specific subnet. EFS allows **at most one mount target per Availability Zone** per file system. If you provide two subnets in the same AZ, the AWS API returns an error at deploy time.

**Placement rules:**

- Each subnet must be in a different AZ.
- For regional (multi-AZ) file systems, provide one subnet per AZ where you need access.
- For One Zone file systems, provide exactly one subnet in the AZ specified by `availability_zone_name`.
- Subnets are typically private; mount targets do not require public IPs.

**Network requirements:**

- **NFS TCP port 2049** — Security groups must allow inbound traffic on port 2049 from the clients that will mount the file system.
- **NFS over TLS** — When using the `amazon-efs-utils` mount helper with `tls` option, the same port is used; TLS is negotiated at the application layer.

**DNS names:**

- **Regional DNS name** (`fs-xxx.efs.region.amazonaws.com`) — Resolves to mount targets in any AZ. Use for general-purpose mounts.
- **Per-AZ DNS name** (`az.fs-xxx.efs.region.amazonaws.com`) — Resolves to the mount target in that AZ, avoiding cross-AZ traffic. Use when you want to pin a client to a specific AZ for latency or cost.

**Outputs:** `mount_target_ids`, `mount_target_ips`, and `mount_target_dns_names` map subnet IDs to their respective mount target identifiers and endpoints.

### Cross-AZ Traffic and Cost

When a client in AZ-A mounts using the regional DNS name, NFS traffic may be routed to a mount target in AZ-B if the client's AZ mount target is unavailable or if the resolver returns a different target. Cross-AZ traffic incurs data transfer charges. To avoid this:

- Use per-AZ DNS names (`mount_target_dns_names`) when you know the client's AZ.
- Co-locate clients and mount targets in the same AZ when possible.

### Mount Target Limits

- Maximum 1 mount target per AZ per file system.
- Mount targets cannot be moved between subnets; delete and recreate if you need to change placement.
- Each mount target consumes an ENI; ensure your VPC has sufficient ENI capacity for the number of mount targets across all EFS file systems.

### Mount Target Troubleshooting

**Connection refused or timeout:** Verify security groups allow inbound TCP 2049 from the client's security group. Ensure the client is in a subnet that can reach the mount target (same VPC, or VPC peering with appropriate route tables). For Lambda, the function must be in a VPC with at least one subnet that has a mount target.

**Stale file handle:** The NFS client lost its session. Remount the file system. For long-running processes, implement reconnection logic or use the `amazon-efs-utils` mount helper with `_netdev` and `nofail` for resilient boot-time mounts.

**Cross-AZ latency:** If clients experience high latency, use per-AZ DNS names from `mount_target_dns_names` to pin traffic to the mount target in the client's AZ.

---

## 3. Storage Classes

EFS offers three storage classes with different cost and durability characteristics:

| Storage Class | Durability | Cost | Use Case |
|---------------|------------|------|----------|
| **Standard** | Multi-AZ (regional) | Base rate | Frequently accessed data, hot tier |
| **Infrequent Access (IA)** | Multi-AZ | ~92% cheaper than Standard | Data accessed less than once per month |
| **Archive** | Multi-AZ | ~96% cheaper than Standard | Long-term archival, rarely accessed |

**Lifecycle management:** Files do not move automatically. You configure lifecycle policies that transition files:

- **Standard → IA** — After a period of no access (e.g., 30, 60, 90 days).
- **IA → Archive** — After a period of no access (requires IA transition first).
- **IA/Archive → Standard** — After access (e.g., `AFTER_1_ACCESS`).

**Access fees:** IA and Archive charge per GB of data retrieved. Standard does not. Choose the right tier based on access patterns to optimize cost.

**One Zone storage:** For One Zone file systems, the equivalent tiers are One Zone, One Zone-IA, and One Zone-Archive. Data is stored in a single AZ; cost is ~47% lower than Standard.

### Storage Class Selection Heuristics

- **Hot data** (accessed daily): Keep in Standard.
- **Warm data** (accessed weekly/monthly): Transition to IA after 30–90 days.
- **Cold data** (accessed rarely or never): Transition to Archive after 90–365 days.
- **Unpredictable access:** Use `transition_to_primary_storage_class: AFTER_1_ACCESS` so frequently accessed files automatically move back to Standard.

### Lifecycle Transition Mechanics

AWS tracks "last access time" for lifecycle decisions. Read, write, and metadata operations (stat, listdir) update the access time. Lifecycle evaluation runs asynchronously; transitions are not instantaneous. A file may remain in Standard for up to 24–48 hours after the transition period elapses before moving to IA. Archive transitions follow a similar delay after IA. Plan retention and access patterns accordingly.

---

## 4. Throughput Modes

Throughput determines how much read/write bandwidth the file system can sustain. Three modes are available:

### Bursting (Default)

- **Mechanism:** Throughput scales with file system size. 50 MiB/s per TiB of Standard storage, with burst credits up to 100 MiB/s.
- **Use case:** Small to medium file systems with variable access. Good for dev/test and light production.
- **Limitation:** Burst credits can be exhausted under sustained heavy load; throughput then drops to baseline.

### Provisioned

- **Mechanism:** Fixed throughput regardless of storage size. Set `provisioned_throughput_in_mibps` (1.0–3414.0 for generalPurpose, 1.0–1024.0 for maxIO).
- **Use case:** Predictable, high-throughput workloads (e.g., media processing, batch jobs).
- **Cost:** You pay for provisioned throughput even when idle; use only when you need guaranteed capacity.

### Elastic

- **Mechanism:** Throughput scales automatically with workload. No provisioning or burst credits.
- **Use case:** Unpredictable or spiky access patterns. Recommended for most production workloads.
- **Requirement:** Requires `generalPurpose` performance mode (default).
- **Cost:** Pay for throughput used; no idle capacity charges.

**Recommendation:** Prefer elastic for most workloads. Use provisioned only when you have a known, sustained throughput requirement.

### Bursting Credit Mechanics (Bursting Mode Only)

- **Baseline:** 50 MiB/s per TiB of Standard storage. A 1 TiB file system has 50 MiB/s baseline.
- **Burst:** Up to 100 MiB/s when credits are available.
- **Credits:** Accumulate when usage is below baseline; deplete when usage exceeds baseline.
- **Exhaustion:** When credits are depleted, throughput drops to baseline. Sustained heavy load can exhaust credits quickly on small file systems.

For production workloads with variable or unpredictable traffic, elastic mode avoids credit management entirely.

### Elastic Throughput Deep Dive

Elastic mode scales throughput based on actual I/O patterns. There is no pre-provisioning; you pay only for throughput consumed. Throughput scales up within seconds when demand increases and scales down when idle. There is a minimum charge for the first 1 MiB/s; beyond that, billing is per MiB/s per hour. For workloads with unpredictable spikes (e.g., batch jobs, ML inference, web apps with variable traffic), elastic avoids both burst credit exhaustion and over-provisioning costs.

### Provisioned Throughput Scaling

When using provisioned mode, you can increase `provisioned_throughput_in_mibps` at any time. Decreases may require a cooldown period (typically 24 hours) before taking effect. Plan capacity for peak load; under-provisioning causes throttling.

---

## 5. Performance Modes

### generalPurpose (Default)

- **Characteristics:** Lowest per-operation latency, suitable for most workloads.
- **Throughput:** Supports bursting, provisioned, and elastic.
- **Use case:** Web applications, ECS tasks, Lambda, EKS pods, general-purpose file sharing.

### maxIO

- **Characteristics:** Higher aggregate throughput for highly parallelized workloads (thousands of EC2 instances). Slightly higher per-operation latency.
- **Throughput:** Supports bursting and provisioned only. **Does not support elastic.**
- **Status:** AWS recommends `generalPurpose` + `elastic` throughput as the replacement for maxIO. maxIO is effectively deprecated for new deployments.

**ForceNew:** Performance mode cannot be changed after creation. Choose `generalPurpose` unless you have a specific maxIO requirement.

---

## 6. One Zone vs Regional Storage

### Regional (Standard)

- **Durability:** Data replicated across multiple AZs.

- **Availability:** Survives AZ-level failure.

- **Cost:** Higher storage cost than One Zone.

- **Use case:** Production workloads requiring high availability and durability.

### One Zone

- **Durability:** Data stored in a single AZ.

- **Availability:** If the AZ fails, the file system is unavailable until AWS recovers.

- **Cost:** ~47% cheaper than Standard.

- **Use case:** Dev/test, non-critical data, or workloads that tolerate AZ-level failure.

**ForceNew:** `availability_zone_name` cannot be changed after creation. You cannot convert between One Zone and regional storage in place.

---

## 7. Access Points

### POSIX Identity Enforcement

An access point defines a POSIX user (UID) and group (GID) that are applied to all file operations through that access point. The NFS client's claimed identity is overridden — regardless of what UID/GID the client sends, EFS uses the access point's identity.

**Use case:** ECS tasks and Lambda functions run as specific UIDs. Without an access point, the application must manage POSIX permissions on the shared file system. With an access point, you enforce a fixed identity and restrict the visible root directory.

### Root Directory Restriction

Each access point can expose a specific directory as the root (`/`). When a client mounts via the access point, it sees only that subtree. Paths outside the root are inaccessible.

**Creation info:** If the root directory path does not exist, you can provide `creation_info` (owner UID/GID, permissions) so EFS creates it automatically when the access point is created.

**Secondary GIDs:** The `posix_user` can include `secondary_gids` — additional group IDs used for group permission checks. Useful when an application needs access to files owned by multiple groups (e.g., a shared analytics user that belongs to both `data-team` and `eng-team` groups).

**Path depth limit:** The access point root directory path is limited to 4 subdirectories (e.g., `/a/b/c/d`). Paths deeper than 4 levels cannot be used as the root. The path must be absolute and start with `/`.

### Use with ECS

- **Task definition:** Add an EFS volume with `file_system_id` and `transit_encryption: ENABLED` (or `NONE` for internal-only).
- **Access point:** Use `access_point_id` from `status.outputs.access_point_ids.<name>`.
- **Container:** Mount the volume at a path (e.g., `/mnt/efs`). The container sees the access point's root directory as `/`.

### Use with Lambda

- **File system config:** Lambda requires an access point **ARN** (not ID). Use `status.outputs.access_point_arns.<name>`.
- **Mount path:** Lambda mounts at a configurable path (e.g., `/mnt/efs`). The function reads/writes using the access point's POSIX identity.
- **Limitation:** Lambda supports only one EFS mount per function. Use a single access point that exposes the needed subtree.

---

## 8. Encryption

### Encryption at Rest (KMS)

- **Default:** When `encrypted` is true (recommended), EFS encrypts all data and metadata at rest.
- **AWS-managed key:** Omit `kms_key_id` to use `aws/elasticfilesystem`. No key management required.
- **Customer-managed key:** Set `kms_key_id` to a KMS key ARN for audit trails, key rotation, and compliance.
- **ForceNew:** Cannot enable encryption or change the KMS key after creation.

### Encryption in Transit

- **NFS over TLS:** Use the `amazon-efs-utils` mount helper with the `tls` option:

  ```bash
  sudo mount -t efs -o tls fs-xxx:/ /mnt/efs
  ```

- **Resource policy:** Enforce encryption in transit by adding a resource policy that denies requests when `aws:SecureTransport` is false:

  ```json
  {
    "Effect": "Deny",
    "Principal": "*",
    "Action": "*",
    "Resource": "*",
    "Condition": {
      "Bool": { "aws:SecureTransport": "false" }
    }
  }
  ```

  Clients that do not use TLS will be denied. This is recommended for production.

### Mount Options for TLS and Performance

The `amazon-efs-utils` mount helper supports several options. Use `tls` for encryption in transit. Use `iam` for IAM-based authentication (required when the resource policy restricts access). Use `_netdev` in `/etc/fstab` so the mount waits for network availability at boot. Use `nofail` to prevent boot failure if EFS is temporarily unavailable.

Example `/etc/fstab` entry:

```
fs-0123456789abcdef0.efs.us-east-1.amazonaws.com:/ /mnt/efs efs _netdev,tls,iam 0 0
```

---

## 9. Lifecycle Policies

Lifecycle policies automatically move files between storage classes to reduce cost.

### Transition to IA

- **Trigger:** Files not accessed for the specified period.

- **Values:** `AFTER_1_DAY`, `AFTER_7_DAYS`, `AFTER_14_DAYS`, `AFTER_30_DAYS`, `AFTER_60_DAYS`, `AFTER_90_DAYS`, `AFTER_180_DAYS`, `AFTER_270_DAYS`, `AFTER_365_DAYS`.

- **Use case:** Logs, backups, cold data that is rarely read.

### Transition to Archive

- **Trigger:** Files in IA not accessed for the specified period.

- **Requirement:** `transition_to_ia` must be set first. Files must pass through IA before Archive.

- **Use case:** Long-term retention, compliance archives.

### Transition to Primary Storage Class

- **Trigger:** Files in IA or Archive are accessed.

- **Value:** `AFTER_1_ACCESS` — move back to Standard after one read.

- **Use case:** "Warm" frequently accessed files automatically; keep cold files in cheaper tiers.

### Cost Trade-off

- **Standard:** Higher storage cost, no retrieval fee.
- **IA:** Lower storage cost, per-GB retrieval fee.
- **Archive:** Lowest storage cost, highest retrieval fee.

Choose transition periods based on expected access patterns. Aggressive transitions (e.g., 1 day to IA) save storage cost but increase retrieval costs if files are accessed often.

---

## 10. Backup Integration

### AWS Backup

When `backup_enabled` is true, EFS creates an automatic daily backup policy. AWS Backup can create snapshots of the file system for point-in-time recovery.

**Configuration:** The AwsElasticFileSystem API enables the EFS backup policy. Full AWS Backup vault and plan configuration is done separately (e.g., via AWS Backup console, Terraform, or CloudFormation).

**Use case:** Disaster recovery, compliance, audit requirements.

**Mutable:** Backup can be enabled or disabled at any time without replacing the file system.

---

## 11. Cost Model Overview

### Storage Class Pricing

- **Standard:** Higher $/GB-month.
- **IA:** ~92% cheaper than Standard; per-GB retrieval fee.
- **Archive:** ~96% cheaper than Standard; higher per-GB retrieval fee.

### Throughput Pricing

- **Bursting:** No separate throughput charge; included in storage pricing. Burst credits may incur overage if exhausted.
- **Provisioned:** $/MiB/s-month. Pay for provisioned capacity even when idle.
- **Elastic:** $/MiB/s for throughput used. No idle charge.

### One Zone Discount

One Zone storage classes are ~47% cheaper than their regional equivalents.

### Optimization Tips

1. Use lifecycle policies to move cold data to IA/Archive.
2. Prefer elastic throughput for variable workloads.
3. Use One Zone for dev/test.
4. Consider `transition_to_primary_storage_class` to avoid unnecessary retrieval costs for frequently accessed files.

---

## 12. Security

### Resource Policies

- **Purpose:** IAM resource policies attached to the file system control who can access it and under what conditions.
- **Common uses:**
  - Enforce encryption in transit (deny unencrypted NFS).
  - Restrict access to specific IAM principals or VPCs.
  - Prevent root access from NFS clients.
- **Format:** JSON policy document. Provide via the `policy` spec field.

### Security Groups

- **Mount targets:** Security groups attached to mount targets must allow inbound NFS (TCP 2049) from client security groups.
- **Client:** EC2 instances, ECS tasks, Lambda ENIs, and EKS nodes must have security groups that are allowed by the mount target security groups.
- **Best practice:** Create a dedicated security group for EFS clients and reference it in `security_group_ids`.

### IAM

- **File system operations:** Creating, updating, and deleting file systems require IAM permissions (`elasticfilesystem:CreateFileSystem`, etc.).
- **Resource-level:** Use `file_system_arn` in IAM policies for fine-grained access control.
- **KMS:** If using a customer-managed key, ensure the EFS service role and KMS key policy allow EFS to use the key.

### Access Points and Least Privilege

- Access points enforce least-privilege by restricting the visible root directory and POSIX identity.
- Prefer access points over direct file system mounts for ECS and Lambda.

---

## 13. Common Integration Patterns

### EKS CSI Driver

1. Deploy the EFS CSI driver (or use the AWS-provided add-on).
2. Create a StorageClass that provisions PersistentVolumes backed by EFS.
3. Reference `file_system_id` from `status.outputs.file_system_id` in the StorageClass or provisioner config.
4. Pods mount the PersistentVolume; data is shared across nodes.

**Example (conceptual):**

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: efs-sc
provisioner: efs.csi.aws.com
parameters:
  fileSystemId: <from status.outputs.file_system_id>
  directoryPerms: "700"
```

### ECS Task Volumes

1. In the task definition, add an EFS volume:

   ```json
   "volumes": [
     {
       "name": "app-data",
       "efsVolumeConfiguration": {
         "fileSystemId": "<file_system_id>",
         "transitEncryption": "ENABLED",
         "authorizationConfig": {
           "accessPointId": "<access_point_id>",
           "iam": "ENABLED"
         }
       }
     }
   ]
   ```

2. Reference `file_system_id` and `access_point_id` from AwsElasticFileSystem outputs.
3. Mount the volume in the container definition at `/mnt/efs` (or desired path).

### Lambda File System Config

1. Create an EFS file system with an access point.
2. In the Lambda function configuration, add `fileSystemConfig`:

   ```json
   "fileSystemConfig": {
     "arn": "<access_point_arn>",
     "localMountPath": "/mnt/efs"
   }
   ```

3. Use `valueFrom` to reference `status.outputs.access_point_arns.<name>`.
4. Ensure the Lambda function is in a VPC with connectivity to the EFS mount targets (same VPC or VPC peering, security groups allowing NFS).

### EC2 Direct NFS Mount

1. Launch an EC2 instance in a subnet with a mount target (or in a subnet that can reach the mount target).
2. Attach a security group that allows outbound NFS (TCP 2049) and is allowed by the mount target security group.
3. Install `amazon-efs-utils` and mount:

   ```bash
   sudo yum install -y amazon-efs-utils
   sudo mount -t efs -o tls fs-xxx:/ /mnt/efs
   ```

4. Use `dns_name` or `mount_target_dns_names` for the mount target.

### Batch Jobs (EC2 / ECS)

For batch workloads that process large files (e.g., video transcoding, data transformation), EFS provides shared scratch space. Create an access point with a dedicated root directory (e.g., `/batch-jobs`) and POSIX identity matching the batch job user. Each job writes to a unique subdirectory; the access point restricts visibility to the batch root. Use elastic throughput for spiky batch traffic.

### ML Training and Inference

Lambda functions loading ML models from EFS benefit from access points that expose a read-only model directory. Mount the access point at `/mnt/models`; the function reads weights without managing permissions. For multi-worker training (e.g., SageMaker with shared checkpoint storage), use EFS with elastic throughput. Workers in different AZs access the same file system; use per-AZ DNS names to minimize cross-AZ traffic.

---

## 14. Resource Policy Examples

Beyond encryption-in-transit enforcement, resource policies can restrict access by VPC or principal.

### Restrict to Specific VPC

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": [
        "elasticfilesystem:ClientMount",
        "elasticfilesystem:ClientWrite",
        "elasticfilesystem:ClientRootAccess"
      ],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "elasticfilesystem:AccessPointArn": "arn:aws:elasticfilesystem:region:account:access-point/fsap-xxx"
        }
      }
    }
  ]
}
```

### Deny Root Access

To prevent NFS clients from accessing the file system as root (UID 0), combine access points (which override client identity) with a policy that restricts `ClientRootAccess`. Access points are the primary mechanism; the policy adds defense in depth.

---

## 15. Limits and Quotas

| Limit | Value |
|-------|-------|
| File systems per account per region | 1,000 (soft; request increase if needed) |
| Mount targets per file system | 1 per AZ |
| Access points per file system | 1,000 |
| Max file size | 47.9 TiB |
| Max directory depth | Unlimited (practical limit ~1,000 levels) |
| Access point root path depth | Up to 4 subdirectories |
| Throughput (generalPurpose) | Up to 3,414 MiB/s provisioned |
| Throughput (maxIO) | Up to 1,024 MiB/s provisioned |

**Connection limits:** A single mount target can handle thousands of concurrent NFS connections. Throughput, not connection count, is typically the bottleneck. Monitor CloudWatch metrics (`ClientConnections`, `DataReadIOBytes`, `DataWriteIOBytes`) for capacity planning.

---

## 16. KMS Key Policy Requirements

When using a customer-managed KMS key (`kms_key_id`), the key policy must allow the EFS service to use it. AWS automatically adds the required permissions when you create an encrypted file system via the console; when using IaC, ensure the key policy includes:

```json
{
  "Sid": "Allow EFS",
  "Effect": "Allow",
  "Principal": {
    "Service": "elasticfilesystem.amazonaws.com"
  },
  "Action": [
    "kms:Decrypt",
    "kms:CreateGrant"
  ],
  "Resource": "*",
  "Condition": {
    "StringEquals": {
      "kms:CallerAccount": "<your-account-id>"
    }
  }
}
```

The EFS service uses `CreateGrant` to obtain decrypt permissions for the key. Without this, file system creation fails with a KMS access denied error.

---

## 18. CloudWatch Metrics and Monitoring

EFS publishes metrics to CloudWatch at the file system level. Key metrics:

| Metric | Description |
|--------|-------------|
| `ClientConnections` | Number of NFS client connections to the file system. |
| `DataReadIOBytes` | Bytes read from the file system. Use for throughput analysis. |
| `DataWriteIOBytes` | Bytes written to the file system. |
| `MetadataIOBytes` | Metadata operations (stat, listdir, etc.). |
| `PermittedThroughput` | Throughput permitted for the file system (bursting baseline or provisioned). |
| `BurstCreditBalance` | Remaining burst credits (bursting mode only). Exhaustion causes throttling. |

**Alarms:** Set alarms on `BurstCreditBalance` approaching zero for bursting file systems. For provisioned mode, monitor `DataReadIOBytes` and `DataWriteIOBytes` against your provisioned throughput to detect under-provisioning. For elastic mode, monitor throughput utilization for cost analysis.

---

## 19. Summary

| Topic | Key Takeaway |
|-------|---------------|
| **NFS** | EFS is NFSv4; clients mount via mount targets. |
| **Mount targets** | One per AZ, in subnets; NFS TCP 2049 required. |
| **Storage classes** | Standard, IA, Archive; lifecycle policies automate tiering. |
| **Throughput** | Prefer elastic; use provisioned for fixed high throughput. |
| **Performance mode** | Use generalPurpose; maxIO is deprecated. |
| **One Zone** | ~47% cheaper; single AZ only. |
| **Access points** | Enforce POSIX identity and root; use with ECS/Lambda. |
| **Encryption** | At rest (KMS); in transit (TLS + resource policy). |
| **Backup** | Enable for production; AWS Backup integration. |
| **Security** | Resource policies, security groups, IAM, access points. |

For API reference and examples, see the parent [README.md](../README.md) and [examples.md](../examples.md).

---

## Appendix: Quick Reference

| Spec Field | Default | ForceNew |
|------------|---------|----------|
| encrypted | true | Yes |
| performance_mode | generalPurpose | Yes |
| throughput_mode | bursting | No |
| availability_zone_name | (none) | Yes |
| kms_key_id | aws/elasticfilesystem | Yes |
