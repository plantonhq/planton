# AWS S3 Object Set

Deploys one or more objects into an existing S3 bucket, supporting inline text content, base64-encoded binary content, and server-side copies of existing S3 objects. Objects are managed declaratively alongside the infrastructure that consumes them, each with its own content headers, storage class, encryption override, Object Lock retention, and — for copies — preconditions that guard artifact promotion against a source that changed.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **S3 Object (one per inline-content entry)** -- an S3 object for each `objects` item carrying `content` or `contentBase64`, uploaded to the target bucket with the specified key, content, content headers (type, caching, encoding, disposition, language), user metadata, website redirect, storage class, per-object encryption override (SSE mode, KMS key, bucket key), integrity checksum, Object Lock retention and legal hold, canned ACL, and tags
- **S3 Object Copy (one per copy entry)** -- for each `objects` item carrying `copyFrom`, a server-side copy of the named source object into the target bucket, with the same per-object destination surface plus copy-time preconditions, Requester Pays acknowledgment, and metadata preserve-or-replace control
- **Merged Tags** -- each object receives tags merged from three sources in increasing precedence: the resource-identity tags, set-level `tags`, and per-object `tags`

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An S3 bucket** -- the target bucket must exist. Provide the bucket name directly or reference an AwsS3Bucket Infra Component via ValueFromRef.
- **Bucket region match** -- the `region` field must match the region of the target bucket.

## Deploy

### Console

Open the deployment store, find **AWS S3 Object Set**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Configuration Files** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsS3ObjectSet
metadata:
  name: app-config
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  bucket:
    value: "my-app-bucket"
  objects:
    - key: config/app.json
      content: '{"env": "prod", "logLevel": "info"}'
      contentType: application/json
```

```shell
planton apply -f s3-objects.yaml
```

This uploads a single JSON configuration file to the `config/app.json` key in the target bucket. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the object set to an S3 bucket deployed in the same InfraPipeline:

```yaml
spec:
  bucket:
    valueFrom:
      kind: AwsS3Bucket
      name: app-assets
      fieldPath: status.outputs.bucket_id
```

The InfraPipeline resolves the dependency graph, deploys the S3 bucket first, then uploads the objects with the resolved bucket name.

## Key Configuration

These are the most important decisions when configuring an S3 object set. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Object source** -- Each object requires exactly one of `content` (inline UTF-8 text), `contentBase64` (base64-encoded binary), or `copyFrom` (server-side copy of an existing S3 object). Use `content` for configuration files, JSON, YAML, and HTML; `contentBase64` for images, compiled assets, or pre-compressed data; and `copyFrom` for content of any size that already lives in S3 -- build artifacts promoted between buckets, environments seeded from a golden bucket.

**Copy behavior** -- A copy takes the source's current version at deploy time and by default preserves the source's metadata and content headers; set `copyFrom.replaceMetadata` to write this entry's own headers and metadata instead. Copy-time preconditions (`copyIfMatch`, `copyIfNoneMatch`, `copyIfModifiedSince`, `copyIfUnmodifiedSince`) fail the deploy if the source no longer matches -- the guard for promotions. The set's identity tags always reach the copy. Copies are ordered after the set's inline-content objects, so a copy may duplicate an object declared in the same set.

**Content type** -- Defaults to `application/octet-stream`. Set appropriate MIME types (`application/json`, `text/html`, `image/png`) for downstream consumers and CDN serving. Incorrect content types can cause browsers to download files instead of rendering them.

**Object keys** -- The `key` field determines the object's path within the bucket (e.g., `config/app.json`, `assets/logo.png`). Organize keys with prefix-based directory structures for clean bucket navigation and prefix-filtered lifecycle rules.

**Caching and encoding** -- Set `cacheControl` for CDN and browser caching behavior (e.g., `max-age=86400` for immutable assets, `no-cache` for configuration). Set `contentEncoding` (e.g., `gzip`) when uploading pre-compressed content.

**Per-object security overrides** -- Uniform posture belongs on the bucket: default encryption, versioning, and ownership controls are AwsS3Bucket settings that S3 applies to every upload automatically. The per-object `serverSideEncryption`, `kmsKey`, `bucketKeyEnabled`, and `acl` fields are overrides for individual objects that must diverge -- if every object carries the same override, move the setting to the bucket.

**Object Lock and integrity** -- `objectLockMode` plus `objectLockRetainUntilDate` (always set together) make an object version undeletable until the date passes; COMPLIANCE mode cannot be shortened by anyone, and destroy fails until then. Both require the bucket to have been created with Object Lock enabled -- as does `forceDestroy`, which bypasses GOVERNANCE retention during delete. `checksumAlgorithm` stores an additional object checksum retrievable via GetObjectAttributes.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsS3Bucket** | `bucket` | `status.outputs.bucket_id` |
| **AwsS3Bucket** | `objects[].copyFrom.sourceBucket` | `status.outputs.bucket_id` |
| **AwsKmsKey** | `objects[].kmsKey` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `bucket_id` | The bucket the objects were uploaded to | Downstream references, E2E verification |
| `object_arns` | Map of object key to ARN (`arn:aws:s3:::bucket/key`) | IAM policy Resource lists granting access to exactly these objects |
| `object_etags` | Map of object key to ETag (content hash) | Cache invalidation, change detection |
| `object_version_ids` | Map of object key to version ID (versioned buckets only) | Version-specific object access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Configuration files** -- Upload application configuration files (JSON, YAML) to an S3 bucket for applications that read config from S3 at startup. Proper MIME content types set for downstream consumers. Start from the **Configuration Files** preset.

**Static website assets** -- HTML pages with `no-cache`, fingerprinted assets with a year-long immutable cache, and `websiteRedirect` marker objects for moved pages — the per-asset Cache-Control split that drives correct CDN behavior. Public access comes from the bucket's policy and website configuration, never per-object ACLs. Start from the **Static Website Assets** preset.

**Encrypted compliance drop** -- an audit artifact under WORM retention: per-object KMS encryption, a SHA256 upload checksum retrievable via GetObjectAttributes, and COMPLIANCE-mode Object Lock nobody can shorten. Start from the **Encrypted Compliance Drop** preset.

**Artifact promotion** -- Copy a released build artifact from a golden artifacts bucket into an environment's bucket with ETag preconditions guarding against a source that changed since release. Start from the **Promote Golden Artifacts** preset.

## Works With

- [**AWS S3 Bucket**](/infra-catalog/aws-s3-bucket) -- provides the target bucket for object uploads
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- customer-managed key for per-object SSE-KMS encryption overrides