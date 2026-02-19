# AlicloudOssBucket

Manages an Alibaba Cloud Object Storage Service (OSS) bucket.

## Overview

OSS is Alibaba Cloud's S3-compatible object storage service. A bucket is the top-level container for objects with globally unique naming. This component provisions a single OSS bucket with configurable versioning, encryption, lifecycle management, CORS rules, and access logging.

### What Gets Created

- **OSS Bucket** -- an `alicloud_oss_bucket` resource (Pulumi: `oss.Bucket`) with the specified storage class, redundancy type, and ACL
- **Versioning** -- optionally enabled to preserve all object versions
- **Server-Side Encryption** -- optionally configured with AES256 or KMS
- **Lifecycle Rules** -- automated object transitions and expiration policies
- **CORS Rules** -- cross-origin access configuration for browser-based clients
- **Access Logging** -- server access logs written to a target bucket
- **Tags** -- system metadata tags merged with user-defined tags

### Important: Immutable Settings

`storage_class` and `redundancy_type` are **immutable after creation**. Changing either value requires destroying and recreating the bucket. Choose carefully during initial provisioning.

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `region` | string | Alibaba Cloud region (e.g., "cn-hangzhou") |
| `bucketName` | string | Globally unique bucket name (3-63 chars) |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `acl` | string | `"private"` | Access control: `private`, `public-read`, `public-read-write` |
| `storageClass` | string | `"Standard"` | Storage tier (immutable): `Standard`, `IA`, `Archive`, `ColdArchive`, `DeepColdArchive` |
| `redundancyType` | string | `"LRS"` | Data redundancy (immutable): `LRS` (single-zone), `ZRS` (cross-zone) |
| `versioningEnabled` | bool | `false` | Enable object versioning |
| `serverSideEncryption` | object | -- | Encryption config: `sseAlgorithm` (`AES256` or `KMS`), optional `kmsMasterKeyId` |
| `lifecycleRules` | list | `[]` | Object lifecycle management rules |
| `corsRules` | list | `[]` | Cross-origin resource sharing rules (max 10) |
| `logging` | object | -- | Access logging: `targetBucket` (required), `targetPrefix` |
| `forceDestroy` | bool | `false` | Delete all objects on bucket destroy |
| `resourceGroupId` | string | `""` | Resource group for organizational grouping |
| `tags` | map | `{}` | Key-value tags |

### Lifecycle Rule Fields

| Field | Type | Description |
|-------|------|-------------|
| `prefix` | string | Object name prefix filter (empty = all objects) |
| `enabled` | bool | Whether this rule is active |
| `expirationDays` | int | Days after creation to expire objects |
| `transitions` | list | Storage class transitions: `days` + `storageClass` |
| `abortMultipartUploadDays` | int | Days to clean up incomplete multipart uploads |
| `noncurrentVersionExpirationDays` | int | Days to expire old object versions |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `bucket_name` | The bucket name (also serves as the bucket ID) |
| `extranet_endpoint` | Public internet endpoint |
| `intranet_endpoint` | VPC-internal endpoint (zero-cost intra-region access) |

## Dependencies

None. OSS buckets are standalone resources with no upstream dependencies.

## Related Components

- **AlicloudKmsKey** -- for customer-managed encryption keys (KMS SSE)
- **AlicloudFcFunction** -- uses OSS for function code storage
- **AlicloudCdnDomain** -- uses OSS as an origin for CDN acceleration
