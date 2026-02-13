# Scaleway Object Storage Bucket Examples

This document provides YAML manifest examples for creating and managing Scaleway Object Storage buckets using OpenMCF.

## Table of Contents

- [Minimal Private Bucket](#minimal-private-bucket)
- [Versioned Bucket for Backups](#versioned-bucket-for-backups)
- [Bucket with Lifecycle Rules](#bucket-with-lifecycle-rules)
- [Bucket with CORS for Web App](#bucket-with-cors-for-web-app)
- [Production Configuration](#production-configuration)
- [Infra Chart Composition](#infra-chart-composition)

---

## Minimal Private Bucket

The simplest configuration with only required fields.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: app-uploads
spec:
  region: fr-par
```

**What you get:**
- Private bucket (no public access by default)
- No versioning (cheaper for temporary data)
- No lifecycle rules (manual management)
- `force_destroy: false` (cannot delete if objects exist)

**Access the bucket:**
```bash
aws --endpoint-url https://s3.fr-par.scw.cloud s3 ls s3://app-uploads/
aws --endpoint-url https://s3.fr-par.scw.cloud s3 cp myfile.txt s3://app-uploads/
```

---

## Versioned Bucket for Backups

Enable versioning to protect critical data from accidental deletion.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: prod-db-backups
  labels:
    environment: production
    data-classification: critical
spec:
  region: nl-ams
  versioning_enabled: true
```

**Key characteristics:**
- Versioning protects against accidental overwrites and deletes
- Every PUT creates a new version, DELETE inserts a delete marker
- Previous versions retrievable by version ID
- **Cannot be fully disabled once enabled** (only suspended)
- Increases storage costs as versions accumulate

**Backup workflow:**
```bash
# Daily backup (creates new version if key exists)
pg_dump mydb | gzip | \
  aws --endpoint-url https://s3.nl-ams.scw.cloud s3 cp - \
  s3://prod-db-backups/daily/latest.sql.gz

# List versions
aws --endpoint-url https://s3.nl-ams.scw.cloud s3api list-object-versions \
  --bucket prod-db-backups --prefix daily/
```

---

## Bucket with Lifecycle Rules

Automate storage management for cost optimization.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: app-logs
  labels:
    environment: production
    purpose: logging
spec:
  region: fr-par
  versioning_enabled: true
  lifecycle_rules:
    # Rule 1: Archive old logs to cold storage, then delete
    - id: archive-and-expire-logs
      enabled: true
      prefix: "logs/"
      transitions:
        - days: 30
          storage_class: ONEZONE_IA
        - days: 90
          storage_class: GLACIER
      expiration_days: 365

    # Rule 2: Abort incomplete multipart uploads after 7 days
    - id: cleanup-multipart
      enabled: true
      abort_incomplete_multipart_upload_days: 7

    # Rule 4: Expire staging data tagged for auto-cleanup
    - id: expire-staging-data
      enabled: true
      tags:
        lifecycle: auto-expire
      expiration_days: 7
```

**Cost optimization flow:**
```
Day 0:   Object created in STANDARD
Day 30:  Object transitions to ONEZONE_IA (cheaper)
Day 90:  Object transitions to GLACIER (cheapest)
Day 365: Object expired (deleted)
```

---

## Bucket with CORS for Web App

Configure CORS to allow a web application to upload files directly to the bucket.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: user-uploads
  labels:
    environment: production
    purpose: user-content
spec:
  region: fr-par
  cors_rules:
    # Allow the production web app to upload and read files
    - allowed_methods:
        - GET
        - PUT
        - POST
        - DELETE
        - HEAD
      allowed_origins:
        - "https://app.example.com"
        - "https://www.example.com"
      allowed_headers:
        - "*"
      expose_headers:
        - ETag
        - x-amz-request-id
      max_age_seconds: 3600

    # Allow any origin for GET (public content)
    - allowed_methods:
        - GET
        - HEAD
      allowed_origins:
        - "*"
      max_age_seconds: 86400
```

**Frontend upload example:**
```javascript
const endpoint = 'https://s3.fr-par.scw.cloud';
const s3 = new AWS.S3({ endpoint, region: 'fr-par' });

await s3.upload({
  Bucket: 'user-uploads',
  Key: `users/${userId}/profile.jpg`,
  Body: imageBuffer,
  ContentType: 'image/jpeg'
}).promise();
```

---

## Production Configuration

A comprehensive production-ready configuration combining all features.

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: prod-media-assets
  labels:
    environment: production
    team: platform
    cost-center: engineering
spec:
  region: fr-par
  versioning_enabled: true
  force_destroy: false

  lifecycle_rules:
    # Tier old media to cheaper storage
    - id: tier-old-media
      enabled: true
      prefix: "media/"
      transitions:
        - days: 60
          storage_class: ONEZONE_IA
        - days: 180
          storage_class: GLACIER

    # Expire temporary upload staging area
    - id: expire-temp-uploads
      enabled: true
      prefix: "tmp/"
      expiration_days: 1

    # Abort stale multipart uploads
    - id: multipart-cleanup
      enabled: true
      abort_incomplete_multipart_upload_days: 7

  cors_rules:
    - allowed_methods:
        - GET
        - PUT
        - POST
        - HEAD
      allowed_origins:
        - "https://app.example.com"
      allowed_headers:
        - "*"
      expose_headers:
        - ETag
      max_age_seconds: 3600
```

---

## Development vs Production

### Development

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: dev-app-uploads
  labels:
    environment: development
spec:
  region: fr-par
  force_destroy: true  # Easy teardown
```

### Production

```yaml
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayObjectBucket
metadata:
  name: prod-app-uploads
  labels:
    environment: production
spec:
  region: fr-par
  versioning_enabled: true
  force_destroy: false  # Prevent accidental data loss
  lifecycle_rules:
    - id: multipart-cleanup
      enabled: true
      abort_incomplete_multipart_upload_days: 7
```

---

## Infra Chart Composition

When composing ScalewayObjectBucket into infra charts, downstream resources reference the bucket's outputs using `valueFrom`:

```yaml
# In an infra chart template:
# A serverless function that reads from the bucket
apiVersion: scaleway.openmcf.org/v1
kind: ScalewayServerlessFunction
metadata:
  name: "{{ values.env }}-image-processor"
  relationships:
    - kind: ScalewayObjectBucket
      name: "{{ values.env }}-media-assets"
      type: uses
spec:
  # ... function config ...
  environment_variables:
    S3_BUCKET_NAME: "{{ values.env }}-media-assets"
    S3_ENDPOINT:
      valueFrom:
        kind: ScalewayObjectBucket
        name: "{{ values.env }}-media-assets"
        fieldPath: status.outputs.api_endpoint
```

The bucket's outputs enable downstream resources to:
- **`endpoint`**: Configure S3 client with bucket-specific URL
- **`api_endpoint`**: Configure S3 client with regional API URL
- **`bucket_name`**: Reference the bucket by name in environment variables
- **`region`**: Use for region-aware downstream configuration

---

## Next Steps

- Review [README.md](./README.md) for detailed field descriptions and architecture
- See [iac/pulumi/](./iac/pulumi/) for Pulumi deployment
- See [iac/tf/](./iac/tf/) for Terraform deployment
- Explore [Scaleway Object Storage docs](https://www.scaleway.com/en/docs/object-storage/)
