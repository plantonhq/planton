---
title: "Object Bucket"
description: "Object Bucket deployment documentation"
icon: "package"
order: 100
componentName: "scalewayobjectbucket"
---

# Scaleway Object Bucket

Deploys a Scaleway Object Storage bucket with optional versioning, S3 Object Lock, lifecycle rules for automated object management, and CORS configuration for browser-based access. Bucket names are globally unique and derived from `metadata.name`.

## What Gets Created

When you deploy a ScalewayObjectBucket resource, OpenMCF provisions:

- **Object Storage Bucket** — an `object.Bucket` resource providing an S3-compatible storage container in the specified region, with tags derived from metadata labels
- **Versioning Configuration** — enabled inline on the bucket when `versioningEnabled` is `true`, retaining all previous versions of objects
- **Lifecycle Rules** — one or more lifecycle automation rules on the bucket when `lifecycleRules` is non-empty, supporting expiration, storage class transitions, and multipart upload cleanup
- **CORS Rules** — one or more Cross-Origin Resource Sharing rules on the bucket when `corsRules` is non-empty, allowing browser-based access from specified origins

## Prerequisites

- **Scaleway credentials** configured via environment variables or OpenMCF provider config
- **A globally unique bucket name** — `metadata.name` must be DNS-compatible and unique across all Scaleway Object Storage (similar to AWS S3 naming constraints)
- **Region selection** — one of `"fr-par"`, `"nl-ams"`, or `"pl-waw"`

## Quick Start

Create a file `object-bucket.yaml`:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: my-app-assets
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayObjectBucket.my-app-assets
spec:
  region: fr-par
```

Deploy:

```shell
openmcf apply -f object-bucket.yaml
```

This creates a single Object Storage bucket in Paris with no versioning, no lifecycle rules, and no CORS rules. Objects are accessible via the S3-compatible endpoint at `my-app-assets.s3.fr-par.scw.cloud`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | Scaleway region where the bucket will be created. Available regions: `"fr-par"`, `"nl-ams"`, `"pl-waw"`. Cannot be changed after creation. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `versioningEnabled` | `bool` | `false` | Enables S3-compatible object versioning. Every PUT creates a new version; DELETE inserts a delete marker. Once enabled, versioning can only be suspended, not fully disabled. |
| `objectLockEnabled` | `bool` | `false` | Enables S3 Object Lock (WORM protection). Can only be set at bucket creation time. Requires `versioningEnabled` to be `true`. |
| `forceDestroy` | `bool` | `false` | When `true`, all objects (including locked objects and versions) are deleted before the bucket is destroyed. Use `true` for dev/test, `false` for production. |
| `lifecycleRules` | `object[]` | `[]` | Lifecycle automation rules for object expiration, storage class transitions, and multipart upload cleanup. Rules are evaluated daily. |
| `lifecycleRules[].id` | `string` | — | Unique identifier for the rule. Required per rule. |
| `lifecycleRules[].enabled` | `bool` | `false` | Whether the rule is active. Disabled rules are retained but not evaluated. |
| `lifecycleRules[].prefix` | `string` | `""` | Object key prefix filter. Empty string applies the rule to all objects. |
| `lifecycleRules[].tags` | `map<string, string>` | `{}` | Tag-based filter. Rule applies only to objects matching all specified tags. |
| `lifecycleRules[].expirationDays` | `int32` | `0` | Days after creation to delete the object. `0` disables expiration. |
| `lifecycleRules[].transitions` | `object[]` | `[]` | Storage class transitions. Each entry moves matching objects to a cheaper storage class after a specified number of days. |
| `lifecycleRules[].transitions[].days` | `int32` | — | Days after creation to transition. Must be a positive integer. Required per transition. |
| `lifecycleRules[].transitions[].storageClass` | `string` | — | Target storage class: `"GLACIER"` (cold archival) or `"ONEZONE_IA"` (infrequent access, single-zone). Required per transition. |
| `lifecycleRules[].abortIncompleteMultipartUploadDays` | `int32` | `0` | Days after which incomplete multipart uploads are aborted. `0` disables cleanup. |
| `corsRules` | `object[]` | `[]` | CORS rules controlling which web origins can make cross-origin requests to the bucket's S3 endpoint. |
| `corsRules[].allowedMethods` | `string[]` | — | HTTP methods allowed for cross-origin requests (e.g., `"GET"`, `"PUT"`, `"POST"`, `"DELETE"`, `"HEAD"`). Required per rule, at least one. |
| `corsRules[].allowedOrigins` | `string[]` | — | Origins allowed to make cross-origin requests (e.g., `"https://app.example.com"`, `"*"`). Required per rule, at least one. |
| `corsRules[].allowedHeaders` | `string[]` | `[]` | Headers browsers may include in cross-origin requests (e.g., `"Content-Type"`, `"Authorization"`). |
| `corsRules[].exposeHeaders` | `string[]` | `[]` | Response headers browsers are allowed to read (e.g., `"ETag"`, `"x-amz-request-id"`). |
| `corsRules[].maxAgeSeconds` | `int32` | `0` | Seconds the browser caches the preflight response. `0` uses the browser default. |

## Examples

### Minimal Bucket for Development

A simple bucket in Paris with force-destroy enabled for clean teardown:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: dev-scratch-bucket
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.ScalewayObjectBucket.dev-scratch-bucket
spec:
  region: fr-par
  forceDestroy: true
```

### Versioned Bucket with Lifecycle Rules

A media storage bucket with versioning enabled and lifecycle rules that transition old objects to cold storage and expire them after one year:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: media-archive
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.ScalewayObjectBucket.media-archive
spec:
  region: nl-ams
  versioningEnabled: true
  lifecycleRules:
    - id: archive-old-media
      enabled: true
      prefix: "uploads/"
      expirationDays: 365
      transitions:
        - days: 30
          storageClass: ONEZONE_IA
        - days: 90
          storageClass: GLACIER
    - id: cleanup-temp-uploads
      enabled: true
      prefix: "tmp/"
      expirationDays: 7
      abortIncompleteMultipartUploadDays: 3
```

### Production Bucket with CORS and Object Lock

A production bucket hosting user-uploaded content for a web application, with CORS rules for browser uploads, versioning, Object Lock for compliance, and lifecycle cleanup:

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: prod-user-content
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.ScalewayObjectBucket.prod-user-content
spec:
  region: fr-par
  versioningEnabled: true
  objectLockEnabled: true
  forceDestroy: false
  corsRules:
    - allowedMethods:
        - GET
        - PUT
        - POST
        - DELETE
        - HEAD
      allowedOrigins:
        - https://app.example.com
        - https://admin.example.com
      allowedHeaders:
        - Content-Type
        - Authorization
        - x-amz-content-sha256
        - x-amz-date
      exposeHeaders:
        - ETag
        - x-amz-request-id
      maxAgeSeconds: 3600
    - allowedMethods:
        - GET
        - HEAD
      allowedOrigins:
        - https://cdn.example.com
      maxAgeSeconds: 86400
  lifecycleRules:
    - id: abort-stale-uploads
      enabled: true
      abortIncompleteMultipartUploadDays: 7
    - id: archive-old-content
      enabled: true
      prefix: "archive/"
      transitions:
        - days: 60
          storageClass: GLACIER
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucket_id` | `string` | Unique identifier of the bucket (format: `"{region}/{bucket-name}"`). Referenced by downstream resources. |
| `endpoint` | `string` | FQDN endpoint URL of the bucket (format: `"{bucket-name}.s3.{region}.scw.cloud"`). Used by S3-compatible clients and CDNs. |
| `api_endpoint` | `string` | S3 API endpoint URL for the bucket's region (format: `"https://s3.{region}.scw.cloud"`). Used with `--endpoint-url` in AWS CLI and similar tools. |
| `bucket_name` | `string` | Bucket name as it exists in Scaleway Object Storage. Matches `metadata.name`. |
| `region` | `string` | Region where the bucket is deployed. |

## Related Components

- [ScalewayKapsuleCluster](/docs/catalog/scaleway/kapsule-cluster) — deploys Kubernetes clusters whose workloads can read from and write to this bucket
- [ScalewayPrivateNetwork](/docs/catalog/scaleway/private-network) — provides private connectivity for workloads accessing the bucket
- [ScalewayRdbInstance](/docs/catalog/scaleway/rdb-instance) — deploys managed databases that may store references to objects in this bucket
