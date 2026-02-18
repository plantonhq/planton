# AWS S3 Bucket

Deploys an S3 bucket with encryption, versioning, public access controls, lifecycle rules, server access logging, and CORS configuration. Encryption is always enabled — defaulting to SSE-S3 (AES-256) when no encryption type is specified.

## What Gets Created

When you deploy an AwsS3Bucket resource, OpenMCF provisions:

- **S3 Bucket** — the storage bucket itself, named from `metadata.name`
- **Server-Side Encryption** — always configured; SSE-S3 (AES-256) by default, or SSE-KMS with a specified key
- **Public Access Block** — blocks all public access by default; relaxed only when `isPublic` is `true`
- **Ownership Controls** — set to `BucketOwnerEnforced` (ACLs disabled)
- **Versioning** — enabled when `versioningEnabled` is `true`
- **Lifecycle Rules** — storage class transitions, object expiration, and multipart upload cleanup (when configured)
- **Access Logging** — logs to a target bucket (when configured)
- **CORS Configuration** — cross-origin rules for web applications (when configured)

Note: Replication is defined in the spec but not yet implemented in the Pulumi module. A warning is logged if replication is configured. Use the Terraform module for replication support.

## Prerequisites

- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A KMS key ARN** if using SSE-KMS encryption
- **A target S3 bucket** in the same region if enabling access logging
- **A destination bucket with versioning** if configuring replication (Terraform only)

## Quick Start

Create a file `bucket.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: my-bucket
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsS3Bucket.my-bucket
spec:
  region: us-east-1
```

Deploy:

```shell
openmcf apply -f bucket.yaml
```

This creates a private S3 bucket in `us-east-1` with SSE-S3 encryption, public access blocked, and ACLs disabled.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the bucket will be created (e.g., `us-west-2`, `eu-west-1`). | Required; non-empty |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `isPublic` | `bool` | `false` | When `true`, removes the public access block. When `false`, all public access is blocked. |
| `versioningEnabled` | `bool` | `false` | Enable versioning to keep all object versions. Recommended for production. |
| `encryptionType` | `enum` | `ENCRYPTION_TYPE_SSE_S3` | `ENCRYPTION_TYPE_SSE_S3` (AES-256, free) or `ENCRYPTION_TYPE_SSE_KMS` (KMS-managed keys with CloudTrail audit). |
| `kmsKeyId` | `string` | — | KMS key ID or ARN. Required when `encryptionType` is `ENCRYPTION_TYPE_SSE_KMS`. |
| `tags` | `map<string, string>` | `{}` | Key-value tags for cost allocation and governance. AWS allows up to 50 tags. |
| `forceDestroy` | `bool` | `false` | Delete all objects before destroying the bucket. Use with caution. |
| `lifecycleRules` | `LifecycleRule[]` | `[]` | Automate storage transitions and object expiration. See Lifecycle Rules below. |
| `logging.enabled` | `bool` | `false` | Enable server access logging. |
| `logging.targetBucket` | `string` | — | Target bucket for access logs. Must be in the same region. Required when logging is enabled. |
| `logging.targetPrefix` | `string` | `""` | Prefix for log object keys. |
| `cors.corsRules` | `CorsRule[]` | `[]` | CORS rules for web applications. Each rule requires `allowedMethods` and `allowedOrigins`. |

#### Lifecycle Rule Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Unique rule identifier. |
| `enabled` | `bool` | Whether the rule is active. |
| `prefix` | `string` | Object key prefix filter (e.g., `"logs/"`). Empty applies to all objects. |
| `transitionDays` | `int` | Days after creation to transition to `transitionStorageClass`. |
| `transitionStorageClass` | `enum` | Target storage class: `STORAGE_CLASS_STANDARD_IA`, `STORAGE_CLASS_GLACIER_FLEXIBLE_RETRIEVAL`, `STORAGE_CLASS_GLACIER_DEEP_ARCHIVE`, etc. |
| `expirationDays` | `int` | Days after creation to delete objects. `0` means no expiration. |
| `noncurrentVersionExpirationDays` | `int` | Days to expire old versions (requires versioning). |
| `abortIncompleteMultipartUploadDays` | `int` | Days to abort incomplete multipart uploads. |

## Examples

### Versioned Bucket with Tags

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: app-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsS3Bucket.app-data
spec:
  region: us-east-1
  versioningEnabled: true
  tags:
    Environment: production
    Project: my-app
    CostCenter: engineering
```

### Bucket with Lifecycle Rules

Transition logs to cheaper storage, then archive, then delete:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: app-logs
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsS3Bucket.app-logs
spec:
  region: us-west-2
  versioningEnabled: true
  lifecycleRules:
    - id: transition-to-ia
      enabled: true
      prefix: "logs/"
      transitionDays: 30
      transitionStorageClass: STORAGE_CLASS_STANDARD_IA
    - id: archive-old-logs
      enabled: true
      prefix: "logs/"
      transitionDays: 90
      transitionStorageClass: STORAGE_CLASS_GLACIER_FLEXIBLE_RETRIEVAL
      expirationDays: 365
      noncurrentVersionExpirationDays: 30
    - id: cleanup-multipart
      enabled: true
      abortIncompleteMultipartUploadDays: 7
```

### Full-Featured with KMS, Logging, and CORS

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsS3Bucket
metadata:
  name: secure-bucket
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsS3Bucket.secure-bucket
spec:
  region: us-east-1
  versioningEnabled: true
  encryptionType: ENCRYPTION_TYPE_SSE_KMS
  kmsKeyId: arn:aws:kms:us-east-1:123456789012:key/abcdef01-2345-6789-abcd-ef0123456789
  tags:
    Environment: production
    Compliance: hipaa
  logging:
    enabled: true
    targetBucket: access-logs-bucket
    targetPrefix: "logs/secure-bucket/"
  cors:
    corsRules:
      - allowedMethods:
          - GET
          - PUT
        allowedOrigins:
          - "https://app.example.com"
        allowedHeaders:
          - "*"
        maxAgeSeconds: 3600
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucket_id` | `string` | Name/ID of the S3 bucket |
| `bucket_arn` | `string` | ARN of the bucket (e.g., `arn:aws:s3:::my-bucket`) |
| `region` | `string` | AWS region where the bucket was created |
| `bucket_regional_domain_name` | `string` | Regional domain name (e.g., `my-bucket.s3.us-east-1.amazonaws.com`) |
| `hosted_zone_id` | `string` | Route53 hosted zone ID for the bucket's region, used for alias records |

## Related Components

- [AwsKmsKey](/docs/catalog/aws/awskmskey) — create a KMS key for SSE-KMS encryption
- [AwsRoute53Zone](/docs/catalog/aws/awsroute53zone) — DNS zone for hosting a static website bucket
