# AliCloud OSS Bucket

Deploys an Alibaba Cloud Object Storage Service (OSS) bucket with configurable access control, storage class, zone redundancy, versioning, server-side encryption, lifecycle management, CORS rules, access logging, and automatic tag management. OSS is Alibaba Cloud's S3-compatible object storage service for unstructured data at any scale.

## What Gets Created

When you deploy an AliCloudStorageBucket resource, OpenMCF provisions:

- **OSS Bucket** -- an `alicloud_oss_bucket` resource (Pulumi: `oss.Bucket`) with the specified storage class, redundancy type, and access control
- **Versioning** -- optionally enabled to preserve all object versions for accidental deletion/overwrite recovery
- **Server-Side Encryption** -- optionally configured with AES256 (OSS-managed keys) or KMS (customer-managed keys)
- **Lifecycle Rules** -- automated object transitions between storage tiers and time-based expiration
- **CORS Rules** -- cross-origin resource sharing configuration for browser-based direct access
- **Access Logging** -- server access logs written to a target bucket for audit and debugging
- **Tags** -- system metadata tags (`resource`, `resource_name`, `resource_kind`, `organization`, `environment`) merged with user-defined `spec.tags`, with user values taking precedence on key conflict

## Prerequisites

- **Alibaba Cloud credentials** configured via environment variables (`ALICLOUD_ACCESS_KEY`, `ALICLOUD_SECRET_KEY`) or OpenMCF provider config
- **Globally unique bucket name** -- OSS bucket names must be unique across all Alibaba Cloud accounts worldwide (3-63 characters, lowercase letters, digits, and hyphens)
- **OpenMCF CLI** installed with either Pulumi or Terraform (OpenTofu) backend

## Quick Start

Create a file `oss-bucket.yaml`:

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudStorageBucket
metadata:
  name: my-bucket
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AliCloudStorageBucket.my-bucket
spec:
  region: cn-hangzhou
  bucketName: my-app-assets-bucket
```

Deploy:

```shell
openmcf apply -f oss-bucket.yaml
```

This creates a private Standard-tier OSS bucket with LRS redundancy in the `cn-hangzhou` region.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Alibaba Cloud region where the bucket will be created (e.g., `cn-hangzhou`, `cn-shanghai`, `us-west-1`). | Required; non-empty |
| `bucketName` | `string` | Globally unique bucket name. Lowercase letters, digits, and hyphens only. | Required; 3-63 characters |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `acl` | `string` | `"private"` | Access control: `private`, `public-read`, or `public-read-write`. |
| `storageClass` | `string` | `"Standard"` | Storage tier: `Standard`, `IA`, `Archive`, `ColdArchive`, `DeepColdArchive`. **Immutable after creation.** |
| `redundancyType` | `string` | `"LRS"` | Data redundancy: `LRS` (single-zone) or `ZRS` (cross-zone, ~1.5x cost). **Immutable after creation.** |
| `versioningEnabled` | `bool` | `false` | Enable object versioning for accidental deletion/overwrite recovery. |
| `serverSideEncryption` | `object` | -- | Encryption config with `sseAlgorithm` (`AES256` or `KMS`) and optional `kmsMasterKeyId`. |
| `lifecycleRules` | `list` | `[]` | Object lifecycle management rules (expiration, transitions, multipart cleanup). |
| `corsRules` | `list` | `[]` | Cross-origin resource sharing rules for browser-based access (max 10). |
| `logging` | `object` | -- | Access logging config with `targetBucket` and optional `targetPrefix`. |
| `forceDestroy` | `bool` | `false` | Delete all objects when destroying the bucket. Use with caution in production. |
| `resourceGroupId` | `string` | `""` | Resource group for organizational grouping and cost attribution. |
| `tags` | `map<string, string>` | `{}` | User-defined tags merged with system tags. |

## Examples

### Minimal Private Bucket

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudStorageBucket
metadata:
  name: dev-bucket
spec:
  region: cn-hangzhou
  bucketName: dev-assets-bucket
```

### Production Bucket with Versioning and Encryption

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudStorageBucket
metadata:
  name: prod-bucket
  org: my-org
  env: production
spec:
  region: cn-shanghai
  bucketName: prod-platform-data
  redundancyType: ZRS
  versioningEnabled: true
  serverSideEncryption:
    sseAlgorithm: AES256
  tags:
    team: platform
    costCenter: engineering
```

### Archive Bucket with Lifecycle Rules

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudStorageBucket
metadata:
  name: log-archive
  env: production
spec:
  region: cn-hangzhou
  bucketName: platform-log-archive
  versioningEnabled: true
  lifecycleRules:
    - prefix: ""
      enabled: true
      expirationDays: 365
      transitions:
        - days: 30
          storageClass: IA
        - days: 90
          storageClass: Archive
      abortMultipartUploadDays: 7
      noncurrentVersionExpirationDays: 30
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucket_name` | `string` | The bucket name (also serves as the bucket ID in OSS). |
| `extranet_endpoint` | `string` | Public internet endpoint (`{bucket}.oss-{region}.aliyuncs.com`) for external clients and CDN origins. |
| `intranet_endpoint` | `string` | VPC-internal endpoint (`{bucket}.oss-{region}-internal.aliyuncs.com`) for zero-cost, low-latency access from ECS, functions, and containers in the same region. |

## Related Components

- [AliCloudKmsKey](/docs/catalog/alicloud/alicloudkmskey) -- for customer-managed encryption keys when using KMS server-side encryption
- [AliCloudFcFunction](/docs/catalog/alicloud/alicloudfcfunction) -- uses OSS for function code storage
- [AliCloudCdnDomain](/docs/catalog/alicloud/alicloudcdndomain) -- uses OSS as an origin for CDN acceleration
