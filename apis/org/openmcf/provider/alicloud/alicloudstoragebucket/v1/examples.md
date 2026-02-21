# Examples

## Minimal Private Bucket

Creates a private OSS bucket with default settings -- Standard storage class, LRS redundancy, no versioning, no encryption.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudStorageBucket
metadata:
  name: my-bucket
spec:
  region: cn-hangzhou
  bucketName: my-app-assets-bucket
```

## Production Bucket with Versioning and Encryption

A production-ready bucket with ZRS redundancy for cross-zone durability, versioning enabled for accidental deletion recovery, and AES256 server-side encryption at rest.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudStorageBucket
metadata:
  name: prod-data-bucket
  org: my-org
  env: production
spec:
  region: cn-shanghai
  bucketName: prod-platform-data
  acl: private
  storageClass: Standard
  redundancyType: ZRS
  versioningEnabled: true
  serverSideEncryption:
    sseAlgorithm: AES256
  forceDestroy: false
  resourceGroupId: rg-prod-123
  tags:
    team: platform
    costCenter: engineering
```

## Archive Bucket with Lifecycle Rules

A cost-optimized bucket that transitions objects to cheaper storage tiers over time: IA after 30 days, Archive after 90 days, and expires objects after 365 days. Incomplete multipart uploads are cleaned up after 7 days. Noncurrent versions (from versioning) are expired after 30 days.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudStorageBucket
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
  tags:
    purpose: log-archive
    retention: 1-year
```

## Bucket with CORS for Browser Access

An OSS bucket configured for direct browser uploads from a web application, with CORS rules allowing cross-origin PUT and POST requests.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudStorageBucket
metadata:
  name: upload-bucket
spec:
  region: ap-southeast-1
  bucketName: user-uploads-bucket
  acl: private
  serverSideEncryption:
    sseAlgorithm: AES256
  corsRules:
    - allowedOrigins:
        - "https://app.example.com"
        - "https://staging.example.com"
      allowedMethods:
        - GET
        - PUT
        - POST
      allowedHeaders:
        - "*"
      exposeHeaders:
        - ETag
        - x-oss-request-id
      maxAgeSeconds: 3600
  lifecycleRules:
    - prefix: "tmp/"
      enabled: true
      expirationDays: 7
      abortMultipartUploadDays: 1
```

## Bucket with KMS Encryption and Access Logging

A security-hardened bucket using a customer-managed KMS key for encryption, with access logs written to a separate bucket.

```yaml
apiVersion: alicloud.openmcf.org/v1
kind: AlicloudStorageBucket
metadata:
  name: secure-bucket
  org: compliance-org
  env: production
spec:
  region: cn-beijing
  bucketName: secure-compliance-data
  redundancyType: ZRS
  versioningEnabled: true
  serverSideEncryption:
    sseAlgorithm: KMS
    kmsMasterKeyId: "cmk-abc123def456"
  logging:
    targetBucket: audit-logs-bucket
    targetPrefix: "oss-access-logs/secure-compliance-data/"
  tags:
    compliance: required
    dataClassification: confidential
```
