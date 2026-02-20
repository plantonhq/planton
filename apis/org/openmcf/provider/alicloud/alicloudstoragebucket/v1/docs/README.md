# AlicloudStorageBucket -- Research Documentation

## Alibaba Cloud OSS Overview

Alibaba Cloud Object Storage Service (OSS) is a cloud-native, S3-compatible object storage platform. It provides scalable, durable, and highly available storage for unstructured data -- files, images, backups, logs, and static assets.

### Key Characteristics

- **Global namespace**: Bucket names are unique across all Alibaba Cloud accounts worldwide.
- **Regional**: Each bucket is created in a specific region. Cross-region access incurs data transfer charges; intra-region VPC access is free.
- **Immutable settings**: `storage_class` and `redundancy_type` cannot be changed after bucket creation.
- **Flat namespace**: Objects are stored in a flat key-value structure. "/" in keys creates a virtual directory hierarchy.

## Provider Resources

### Terraform

- **Primary**: `alicloud_oss_bucket` -- creates the bucket with inline versioning, encryption, lifecycle, CORS, and logging blocks.
- **Companion resources** (not managed by this component):
  - `alicloud_oss_bucket_acl` (since 1.220.0, replaces inline `acl`)
  - `alicloud_oss_bucket_policy` (since 1.220.0, replaces inline `policy`)
  - `alicloud_oss_bucket_referer` (since 1.220.0, replaces inline `referer_config`)
  - `alicloud_oss_bucket_object` -- manage individual objects
  - `alicloud_oss_bucket_replication` -- cross-region replication

### Pulumi

- **Primary**: `oss.Bucket` (token: `alicloud:oss/bucket:Bucket`)
- **SDK import**: `github.com/pulumi/pulumi-alicloud/sdk/v3/go/alicloud/oss`
- **Constructor**: `oss.NewBucket(ctx, name, &oss.BucketArgs{...})`

## Storage Classes

| Class | Access Pattern | Restore Time | Min Billing | Use Case |
|-------|---------------|--------------|-------------|----------|
| Standard | Frequent | Instant | None | Active data, websites, mobile apps |
| IA (Infrequent Access) | Monthly | Instant | 30 days | Backups, disaster recovery |
| Archive | Quarterly | Minutes | 60 days | Compliance archives, media archives |
| ColdArchive | Yearly | Hours | 180 days | Long-term retention |
| DeepColdArchive | Rarely | 12-48 hours | 180 days | Regulatory archives |

## Redundancy Types

| Type | Description | Durability | Availability |
|------|-------------|------------|--------------|
| LRS | Locally Redundant Storage | 99.999999999% (11 nines) | 99.99% |
| ZRS | Zone-Redundant Storage | 99.999999999999% (14 nines) | 99.995% |

ZRS replicates data across three availability zones within a region. It costs approximately 1.5x more than LRS but provides higher availability and can survive a full AZ failure.

## Lifecycle Management

Lifecycle rules automate object management:

1. **Transitions**: Move objects to cheaper tiers (Standard -> IA -> Archive -> ColdArchive -> DeepColdArchive)
2. **Expiration**: Permanently delete objects after a retention period
3. **Abort multipart uploads**: Clean up incomplete uploads to prevent storage waste
4. **Noncurrent version expiration**: Expire old versions in versioned buckets

Rules are scoped by prefix and evaluated independently. Up to 1000 rules per bucket.

## Versioning

When enabled, OSS keeps all versions of every object. Delete operations create a "delete marker" rather than removing data. Benefits:

- Accidental deletion recovery
- Accidental overwrite recovery
- Audit trail of changes

Combine with `noncurrent_version_expiration_days` in lifecycle rules to control storage costs.

## Server-Side Encryption

Two algorithms available:

- **AES256**: OSS-managed keys. Zero configuration, suitable for most workloads.
- **KMS**: Alibaba Cloud Key Management Service. Supports customer-managed keys (CMK) for regulated workloads. When `kms_master_key_id` is omitted, OSS uses a default service key.

## CORS Configuration

Required when web browsers need to access OSS directly (e.g., JavaScript file uploads, serving static assets from a different origin). Maximum 10 rules per bucket. Rules are evaluated in order; the first match wins.

## Scope Decisions

### Included in v1

- Bucket creation with storage class, redundancy, and ACL
- Versioning (boolean toggle)
- Server-side encryption (AES256 / KMS)
- Lifecycle rules (prefix-scoped, days-based transitions and expiration)
- CORS rules
- Access logging
- Force destroy flag
- Resource group assignment
- Tags

### Excluded from v1 (available via companion resources)

- **Static website hosting** -- separate concern (`alicloud_oss_bucket_website`)
- **Transfer acceleration** -- niche performance feature
- **Access monitor** -- operational concern
- **Bucket policy** -- use `alicloud_oss_bucket_policy`
- **Referer config** -- use `alicloud_oss_bucket_referer`
- **Cross-region replication** -- use `alicloud_oss_bucket_replication`
- **Advanced lifecycle filters** (object size, tag exclusions) -- covered by the provider directly if needed

## Endpoint Patterns

- **Extranet**: `{bucket}.oss-{region}.aliyuncs.com`
- **Intranet**: `{bucket}.oss-{region}-internal.aliyuncs.com`
- **Accelerated**: `{bucket}.oss-accelerate.aliyuncs.com` (requires transfer acceleration)

## References

- [OSS documentation](https://www.alibabacloud.com/help/en/oss/)
- [Terraform alicloud_oss_bucket](https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/oss_bucket)
- [Pulumi oss.Bucket](https://www.pulumi.com/registry/packages/alicloud/api-docs/oss/bucket/)
