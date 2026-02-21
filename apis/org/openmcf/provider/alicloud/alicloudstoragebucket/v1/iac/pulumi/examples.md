# AliCloudStorageBucket Pulumi Examples

Create a YAML manifest using one of the examples below, then deploy with the OpenMCF CLI:

```bash
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes
```

---

## Minimal Private Bucket

A bucket with only the required fields, suitable for development or quick testing.

```yaml
apiVersion: ali-cloud.openmcf.org/v1
kind: AliCloudStorageBucket
metadata:
  name: dev-bucket
spec:
  region: cn-hangzhou
  bucketName: dev-assets-bucket
```

This creates a private OSS bucket with Standard storage class and LRS redundancy.

---

## Production Bucket with Versioning and Encryption

A production-ready bucket with ZRS redundancy, versioning, and AES256 encryption.

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

- ZRS provides cross-zone durability for production workloads
- Versioning enables recovery from accidental deletions
- AES256 encryption uses OSS-managed keys with zero additional configuration

---

## Archive Bucket with Lifecycle Rules

A cost-optimized bucket that transitions objects through storage tiers over time.

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

- Objects transition to IA after 30 days, Archive after 90 days, and are deleted after 365 days
- Incomplete multipart uploads are cleaned up after 7 days
- Old object versions are expired after 30 days to control storage costs
