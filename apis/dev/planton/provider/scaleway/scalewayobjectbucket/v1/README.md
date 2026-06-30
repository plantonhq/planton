# Scaleway Object Storage Bucket

## Overview

The **ScalewayObjectBucket** resource kind provides a declarative interface for creating and managing S3-compatible object storage buckets on Scaleway. It covers the essential configuration that most teams need: bucket creation, versioning, lifecycle automation, and CORS policies for web applications.

Scaleway Object Storage implements the S3 API, which means you can use any S3-compatible tool (AWS CLI, s3cmd, rclone, Boto3, AWS SDKs) by pointing the endpoint to `s3.<region>.scw.cloud`. This makes it a drop-in replacement for AWS S3 in most workflows.

This is a **standalone resource** wrapping a single `scaleway_object_bucket` Terraform resource. Specialized features like bucket policies (JSON IAM), object lock retention rules, and static website hosting are handled by separate Terraform resources and can be added in a future version.

## Key Features

### S3-Compatible API

Scaleway Object Storage supports the core S3 API surface:

- **Object CRUD**: GET, PUT, DELETE, HEAD, LIST operations
- **Multipart uploads**: Large file uploads in parallel parts
- **Versioning**: Full S3-compatible object versioning
- **Lifecycle management**: Automated expiration and storage class transitions
- **CORS**: Cross-Origin Resource Sharing for browser-based access
- **Object Lock**: WORM (Write Once Read Many) protection for compliance

### Storage Classes

Scaleway provides three storage tiers for lifecycle transitions:

| Class | Use Case | Retrieval | Cost |
|-------|----------|-----------|------|
| **Standard** | Frequently accessed data (default) | Instant | Highest |
| **ONEZONE_IA** | Infrequently accessed, single-zone | Instant | Medium |
| **GLACIER** | Archival data, long-term retention | Minutes to hours | Lowest |

### Lifecycle Automation

Lifecycle rules automate data management to control costs:

- **Expire old objects**: Delete logs, temp files, or stale data after N days
- **Transition to cold storage**: Move infrequent data to ONEZONE_IA or GLACIER
- **Abort stale multipart uploads**: Reclaim storage from incomplete uploads

### CORS for Web Applications

CORS rules enable browser-based JavaScript to interact with the bucket directly:

- Direct file uploads from web applications
- Serving images, media, or assets to frontend apps
- Pre-signed URL workflows for controlled access

## Architecture

### What This Resource Creates

A single Scaleway Object Storage bucket with inline configuration:

```
ScalewayObjectBucket
└── scaleway_object_bucket
    ├── Versioning (optional)
    ├── Lifecycle Rules (optional, repeated)
    │   ├── Expiration
    │   ├── Transitions (GLACIER, ONEZONE_IA)
    │   └── Abort Incomplete Multipart Upload
    ├── CORS Rules (optional, repeated)
    └── Object Lock (optional, creation-time only)
```

### What This Resource Does NOT Create

These are deferred to future versions or separate resource kinds:

- **Bucket ACL** (`scaleway_object_bucket_acl`) -- Deprecated on the main resource
- **Bucket Policy** (`scaleway_object_bucket_policy`) -- JSON IAM policies
- **Lock Configuration** (`scaleway_object_bucket_lock_configuration`) -- Retention rules
- **Website Configuration** (`scaleway_object_bucket_website_configuration`) -- Static hosting

### Dependency Position

ScalewayObjectBucket is a **leaf resource** with no upstream dependencies:

```
No upstream dependencies
    │
    ▼
ScalewayObjectBucket (outputs: bucket_id, endpoint, api_endpoint)
    │
    ▼
Downstream consumers:
  ├── ScalewayServerlessFunction (references bucket endpoint)
  ├── ScalewayServerlessContainer (reads/writes to bucket)
  ├── Applications (S3 client configuration)
  └── ScalewayDnsRecord (CNAME to bucket endpoint for custom domain)
```

### Tag Model Difference

**Important**: Scaleway Object Storage uses **key-value map tags** (`{"key": "value"}`), unlike other Scaleway resources that use flat string tags (`["key=value"]`). This is because Object Storage uses the S3-compatible API which supports structured tags natively. The IaC modules handle this difference transparently.

## Available Regions

| Region | Location | Code |
|--------|----------|------|
| Paris | Paris, France | `fr-par` |
| Amsterdam | Amsterdam, Netherlands | `nl-ams` |
| Warsaw | Warsaw, Poland | `pl-waw` |

Choose the region closest to your application for lowest latency. Consider data residency requirements (EU data in EU regions). Bucket names must be globally unique across all regions.

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `region` | string | Yes | Scaleway region (`fr-par`, `nl-ams`, `pl-waw`) |
| `versioning_enabled` | bool | No | Enable S3-compatible versioning (default: false) |
| `object_lock_enabled` | bool | No | Enable Object Lock / WORM (requires versioning, default: false) |
| `lifecycle_rules` | list | No | Lifecycle automation rules |
| `cors_rules` | list | No | CORS rules for web applications |
| `force_destroy` | bool | No | Allow deletion with objects inside (default: false) |

### Lifecycle Rule Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique rule identifier |
| `enabled` | bool | Whether rule is active |
| `prefix` | string | Object key prefix filter |
| `tags` | map | Tag-based filter |
| `expiration_days` | int | Days to expire objects (0 = disabled) |
| `transitions` | list | Storage class transitions |
| `abort_incomplete_multipart_upload_days` | int | Days to abort stale uploads |

### CORS Rule Fields

| Field | Type | Description |
|-------|------|-------------|
| `allowed_methods` | list | HTTP methods (GET, PUT, POST, DELETE, HEAD) |
| `allowed_origins` | list | Origins to allow (e.g., `https://app.example.com`) |
| `allowed_headers` | list | Allowed request headers |
| `expose_headers` | list | Response headers exposed to browser |
| `max_age_seconds` | int | Preflight cache duration |

## Stack Outputs

| Output | Description |
|--------|-------------|
| `bucket_id` | Unique bucket identifier (format: `region/bucket-name`) |
| `endpoint` | FQDN endpoint URL (e.g., `my-bucket.s3.fr-par.scw.cloud`) |
| `api_endpoint` | S3 API endpoint (e.g., `https://s3.fr-par.scw.cloud`) |
| `bucket_name` | Bucket name for S3 client configuration |
| `region` | Region where the bucket is deployed |

## Versioning

When `versioning_enabled: true`:

- Every PUT creates a new version of the object
- DELETE inserts a delete marker (object is not physically removed)
- Previous versions can be retrieved by version ID
- **Cannot be fully disabled once enabled** (only suspended)
- Increases storage costs as versions accumulate
- Use lifecycle rules with `noncurrent_version_expiration_days` to control costs

## Object Lock

When `object_lock_enabled: true`:

- Enables WORM (Write Once Read Many) protection
- **Requires versioning** (enforced by proto validation)
- **Can only be enabled at creation time** (cannot add to existing bucket)
- **Cannot be removed** once enabled
- Default retention rules configured via separate Lock Configuration resource

## S3 Compatibility

Access the bucket using any S3-compatible tool:

```bash
# AWS CLI
aws --endpoint-url https://s3.fr-par.scw.cloud s3 ls s3://my-bucket/

# Upload a file
aws --endpoint-url https://s3.fr-par.scw.cloud s3 cp myfile.txt s3://my-bucket/

# rclone
rclone copy myfile.txt scaleway:my-bucket/
```

Configure your S3 client with:
- **Endpoint**: `https://s3.<region>.scw.cloud`
- **Access Key**: From Scaleway IAM API keys
- **Secret Key**: From Scaleway IAM API keys
- **Region**: Same as bucket region

## Lifecycle Constraints

- **Bucket name**: Globally unique, DNS-compatible, 3-63 characters
- **Region**: Cannot be changed after creation
- **Object Lock**: Cannot be added to existing bucket or removed
- **Versioning**: Cannot be fully disabled once enabled (only suspended)
- **Force destroy**: Required to delete a bucket that contains objects

## Examples

See [examples.md](./examples.md) for comprehensive YAML manifest examples.

## Infrastructure as Code

### Pulumi

The Pulumi module uses the new `scaleway/object` subpackage (`object.NewBucket()`), which replaces the deprecated top-level `ObjectBucket` resource.

### Terraform

The Terraform module wraps the `scaleway_object_bucket` resource with dynamic blocks for lifecycle rules and CORS configuration.

## References

- [Scaleway Object Storage Documentation](https://www.scaleway.com/en/docs/object-storage/)
- [Scaleway S3 API Compatibility](https://www.scaleway.com/en/docs/object-storage/api-cli/using-api-call-list/)
- [Terraform: scaleway_object_bucket](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/object)
- [Pulumi: scaleway.object.Bucket](https://www.pulumi.com/registry/packages/scaleway/api-docs/objectbucket/)
