---
title: "GCS Bucket"
description: "GCS Bucket deployment documentation"
icon: "package"
order: 100
componentName: "gcpgcsbucket"
---

# GCP GCS Bucket

Deploys a Google Cloud Storage bucket with full control over storage class, access model, lifecycle management, encryption, CORS, static website hosting, and IAM bindings. The bucket is created in the specified GCP project and location with Uniform Bucket-Level Access enabled by default.

## What Gets Created

When you deploy a GcpGcsBucket resource, OpenMCF provisions:

- **GCS Bucket** — a Cloud Storage bucket in the specified project and location, with labels, storage class, and access settings applied
- **Versioning Configuration** — enabled on the bucket when `versioningEnabled` is `true`
- **Lifecycle Rules** — one lifecycle rule per entry in `lifecycleRules`, supporting automatic deletion and storage class transitions
- **Encryption Configuration** — Customer-Managed Encryption Key (CMEK) applied when `encryption.kmsKeyName` is set
- **CORS Rules** — cross-origin access rules applied when `corsRules` is non-empty
- **Website Configuration** — static website hosting settings applied when `website` is provided
- **Retention Policy** — WORM-compliant retention policy applied when `retentionPolicy` is provided
- **Logging Configuration** — legacy access logging directed to a target bucket when `logging` is provided
- **IAM Bindings** — one `BucketIAMBinding` per entry in `iamBindings`, granting specified roles to specified members on the bucket

## Prerequisites

- **GCP credentials** configured via environment variables or OpenMCF provider config
- **A GCP project** where the bucket will be created
- **A globally unique bucket name** — must be 3-63 characters, lowercase letters, numbers, hyphens, or dots
- **A Cloud KMS key** if using customer-managed encryption (optional)

## Quick Start

Create a file `gcs-bucket.yaml`:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGcsBucket
metadata:
  name: my-app-data
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpGcsBucket.my-app-data
spec:
  gcpProjectId: my-gcp-project-123
  bucketName: my-app-data-dev
  location: us-east1
```

Deploy:

```shell
openmcf apply -f gcs-bucket.yaml
```

This creates a standard-class GCS bucket with Uniform Bucket-Level Access in `us-east1`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `gcpProjectId` | `string` or `valueFrom` | The GCP project ID where the bucket is created. Can be a literal value or a reference to a GcpProject resource. | Required |
| `location` | `string` | Region, dual-region, or multi-region for the bucket (e.g., `us-east1`, `US`, `EU`). Immutable after creation. | Required |
| `bucketName` | `string` | Globally unique name for the GCS bucket. Lowercase letters, numbers, hyphens, dots. | Required, 3-63 chars, pattern `^[a-z0-9]([a-z0-9-._]*[a-z0-9])?$` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `uniformBucketLevelAccessEnabled` | `bool` | `false` | Enable Uniform Bucket-Level Access for IAM-only access control. Recommended `true`. |
| `storageClass` | `GcpGcsStorageClass` | `STANDARD` | Storage class for the bucket. One of: `STANDARD`, `NEARLINE`, `COLDLINE`, `ARCHIVE`. |
| `versioningEnabled` | `bool` | `false` | Enable object versioning to protect against accidental deletion or overwrite. |
| `lifecycleRules` | `GcpGcsLifecycleRule[]` | `[]` | Lifecycle rules for automatic object deletion or storage class transitions. |
| `lifecycleRules[].action.type` | `string` | — | Action type: `Delete` or `SetStorageClass`. |
| `lifecycleRules[].action.storageClass` | `GcpGcsStorageClass` | — | Target storage class (only for `SetStorageClass` action). |
| `lifecycleRules[].condition.ageDays` | `int32` | `0` | Age in days since object creation. |
| `lifecycleRules[].condition.createdBefore` | `string` | `""` | RFC 3339 date. Objects created before this date match. |
| `lifecycleRules[].condition.isLive` | `bool` | `false` | Match live (current version) objects. |
| `lifecycleRules[].condition.numNewerVersions` | `int32` | `0` | Number of newer versions to retain. |
| `lifecycleRules[].condition.matchesStorageClass` | `GcpGcsStorageClass[]` | `[]` | Match only objects in these storage classes. |
| `iamBindings` | `GcpGcsIamBinding[]` | `[]` | IAM bindings for bucket-level access control. |
| `iamBindings[].role` | `string` | — | IAM role to grant (e.g., `roles/storage.objectViewer`). |
| `iamBindings[].members` | `string[]` | — | Members to grant the role to. Format: `user:email`, `group:email`, `serviceAccount:email`, `allUsers`. |
| `iamBindings[].condition` | `string` | `""` | CEL condition expression for conditional access. |
| `encryption` | `GcpGcsEncryption` | `null` | CMEK encryption configuration. If omitted, Google-managed encryption is used. |
| `encryption.kmsKeyName` | `string` | — | Cloud KMS key name. Format: `projects/PROJECT/locations/LOC/keyRings/RING/cryptoKeys/KEY`. |
| `corsRules` | `GcpGcsCorsRule[]` | `[]` | CORS rules for cross-origin browser access. |
| `corsRules[].methods` | `string[]` | — | HTTP methods allowed (e.g., `GET`, `PUT`). |
| `corsRules[].origins` | `string[]` | — | Allowed origins (e.g., `https://example.com`). |
| `corsRules[].responseHeaders` | `string[]` | `[]` | Response headers browsers can access. |
| `corsRules[].maxAgeSeconds` | `int32` | `0` | Max time (seconds) browsers can cache preflight responses. |
| `website` | `GcpGcsWebsite` | `null` | Static website hosting configuration. |
| `website.mainPageSuffix` | `string` | `""` | Main page suffix (e.g., `index.html`). |
| `website.notFoundPage` | `string` | `""` | Custom 404 page (e.g., `404.html`). |
| `retentionPolicy` | `GcpGcsRetentionPolicy` | `null` | WORM-compliant retention policy. |
| `retentionPolicy.retentionPeriodSeconds` | `int64` | — | Minimum retention period in seconds. |
| `retentionPolicy.isLocked` | `bool` | `false` | Lock the retention policy (irreversible). |
| `requesterPays` | `bool` | `false` | Enable requester-pays mode. Requesters pay for data access and egress. |
| `logging` | `GcpGcsLogging` | `null` | Legacy access logging configuration. |
| `logging.logBucket` | `string` | — | Destination bucket for access logs. |
| `logging.logObjectPrefix` | `string` | `""` | Prefix for log object names. |
| `publicAccessPrevention` | `string` | `""` | Public access prevention policy. Values: `inherited` (default), `enforced`. |
| `gcpLabels` | `map<string, string>` | `{}` | Custom labels for cost tracking and governance. Merged with auto-generated labels. |

## Examples

### Private Bucket with Versioning and Lifecycle Cleanup

A bucket with versioning enabled and a lifecycle rule that deletes noncurrent object versions older than 30 days:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGcsBucket
metadata:
  name: app-backups
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.GcpGcsBucket.app-backups
spec:
  gcpProjectId: my-gcp-project-123
  bucketName: my-app-backups-dev
  location: us-central1
  uniformBucketLevelAccessEnabled: true
  publicAccessPrevention: enforced
  versioningEnabled: true
  lifecycleRules:
    - action:
        type: Delete
      condition:
        ageDays: 30
        isLive: false
```

### Static Website Hosting with CORS

A bucket configured to serve a static website with CORS rules for browser-based access:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGcsBucket
metadata:
  name: marketing-site
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGcsBucket.marketing-site
spec:
  gcpProjectId: my-gcp-project-123
  bucketName: marketing-site-prod
  location: US
  uniformBucketLevelAccessEnabled: true
  website:
    mainPageSuffix: index.html
    notFoundPage: 404.html
  corsRules:
    - methods:
        - GET
        - HEAD
      origins:
        - https://example.com
      responseHeaders:
        - Content-Type
      maxAgeSeconds: 3600
  iamBindings:
    - role: roles/storage.objectViewer
      members:
        - allUsers
```

### Full-Featured Bucket with Encryption, Lifecycle Tiers, and IAM

A production bucket with CMEK encryption, tiered lifecycle transitions, versioning, retention, custom labels, and service account access:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGcsBucket
metadata:
  name: prod-data-lake
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGcsBucket.prod-data-lake
spec:
  gcpProjectId: my-gcp-project-123
  bucketName: prod-data-lake
  location: us-east1
  storageClass: STANDARD
  uniformBucketLevelAccessEnabled: true
  publicAccessPrevention: enforced
  versioningEnabled: true
  encryption:
    kmsKeyName: projects/my-gcp-project-123/locations/us-east1/keyRings/data-ring/cryptoKeys/bucket-key
  retentionPolicy:
    retentionPeriodSeconds: 2592000
    isLocked: false
  lifecycleRules:
    - action:
        type: SetStorageClass
        storageClass: NEARLINE
      condition:
        ageDays: 30
        matchesStorageClass:
          - STANDARD
    - action:
        type: SetStorageClass
        storageClass: COLDLINE
      condition:
        ageDays: 90
        matchesStorageClass:
          - NEARLINE
    - action:
        type: Delete
      condition:
        ageDays: 365
        isLive: false
  iamBindings:
    - role: roles/storage.objectAdmin
      members:
        - serviceAccount:data-pipeline@my-gcp-project-123.iam.gserviceaccount.com
    - role: roles/storage.objectViewer
      members:
        - serviceAccount:analytics-reader@my-gcp-project-123.iam.gserviceaccount.com
        - group:data-team@example.com
  logging:
    logBucket: audit-logs-bucket
    logObjectPrefix: prod-data-lake/
  gcpLabels:
    team: data-engineering
    cost-center: analytics
```

### Using a Foreign Key Reference for Project ID

Reference an OpenMCF-managed GcpProject instead of hardcoding the project ID:

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpGcsBucket
metadata:
  name: shared-assets
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.GcpGcsBucket.shared-assets
spec:
  gcpProjectId:
    valueFrom:
      kind: GcpProject
      name: my-project
      fieldPath: status.outputs.project_id
  bucketName: shared-assets-prod
  location: us-east1
  uniformBucketLevelAccessEnabled: true
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucket_id` | `string` | The ID of the created GCS bucket |

## Related Components

- [GcpProject](/docs/catalog/gcp/project) — provides the GCP project where the bucket is created
- [GcpServiceAccount](/docs/catalog/gcp/service-account) — creates service accounts that can be added to `iamBindings` members
- [GcpGkeCluster](/docs/catalog/gcp/gke-cluster) — workloads running in GKE clusters commonly read from and write to GCS buckets
- [GcpVpc](/docs/catalog/gcp/vpc) — network configuration for private connectivity to GCS via Private Google Access
