# GCP GCS Bucket

Deploys a Google Cloud Storage bucket — the durable object store behind static sites, data lakes, build artifacts, backups, and every GCP service that stages data. The spec covers the full bucket surface: the four placement shapes (region, multi-region, predefined and custom dual-regions with turbo replication), the modern IAM-only access model with additive per-bucket grants, Autoclass and lifecycle management, WORM retention with the irreversible lock, soft delete, CMEK, static website serving, CORS, access logging, network-layer IP filtering, and the bucket's structural companions — folders (real directories on hierarchical-namespace buckets), managed folders (prefix-scoped IAM anchors), and Pub/Sub notification configs (object-change events for event-driven pipelines). Placement, hierarchical namespace, and object retention are create-time decisions — changing them replaces the bucket and its data.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Cloud Storage Bucket** -- a GCS bucket in the specified project and location, with the configured placement shape, storage class or Autoclass, access model, data-protection posture (versioning, retention, soft delete, holds), and optional features (website, CORS, logging, IP filter)
- **IAM Grants** -- created only when `iamMembers` are provided; each entry additively grants one role to one member at the bucket level (optionally scoped by an IAM condition) and composes safely with grants made elsewhere
- **Lifecycle Rules** -- created only when `lifecycleRules` are provided; automate object deletion, storage class transitions, and multipart-upload cleanup based on age, version count, prefixes, and more
- **Folders** -- created only when `folders` are provided (hierarchical-namespace buckets); real directories created parents-first, each with its own force-destroy posture
- **Managed Folders** -- created only when `managedFolders` are provided; prefix-scoped IAM anchor points for grants like "read `reports/` only"
- **Notification Configs** -- created only when `notifications` are provided; object-change event feeds into Pub/Sub topics (the GCS service agent's `roles/pubsub.publisher` grant on the topic is a composed prerequisite)
- **Storage API enablement** -- `storage.googleapis.com` enabled in the target project (never disabled on destroy)
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) plus any custom labels from `labels`, applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the bucket will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **A globally unique bucket name** -- GCS bucket names must be unique across all GCP projects worldwide. Must be 3-63 characters, lowercase letters, numbers, hyphens, or dots.
- **For CMEK**: a Cloud KMS key whose key the GCS service agent can use (`roles/cloudkms.cryptoKeyEncrypterDecrypter`).

## Deploy

### Console

Open the deployment store, find **GCP GCS Bucket**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private Standard Bucket** preset in the [Presets](#presets) tab to pre-populate a secure private bucket configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGcsBucket
metadata:
  name: app-data
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  bucketName: acme-prod-app-data
  location: us-central1
  uniformBucketLevelAccessEnabled: true
  publicAccessPrevention: enforced
```

```shell
planton apply -f gcs-bucket.yaml
```

This creates a STANDARD storage bucket in `us-central1` with the modern security posture: IAM-only access control and public access impossible. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the bucket to a GCP project and KMS key deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  kmsKeyName:
    valueFrom:
      kind: GcpKmsKey
      name: bucket-key
      fieldPath: status.outputs.key_id
```

The InfraPipeline resolves the dependency graph, deploys the project and key first, then provisions the bucket with the resolved values.

## Key Configuration

These are the most important decisions when configuring a GCS bucket. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Placement** -- `location` takes a region (`us-east1`), a multi-region (`US`, `EU`, `ASIA`), or a predefined dual-region (`NAM4`). For a custom dual-region, set a multi-region plus `customPlacementConfig.dataLocations` naming exactly two regions; add `rpo: ASYNC_TURBO` for a 15-minute replication SLA on dual-region buckets. Placement is immutable — changing it replaces the bucket and its data.

**Access model** -- Set `uniformBucketLevelAccessEnabled: true` (recommended) for IAM-only access control, and `publicAccessPrevention: "enforced"` so no grant — present or future — can ever expose the bucket. Grant access through additive `iamMembers` entries; reference a GcpServiceAccount resource for workload identities (its `member` output is exactly the value this field wants), and scope grants with IAM conditions for prefix-limited or expiring access.

**Storage class and Autoclass** -- `storageClass: STANDARD` for frequently accessed data; NEARLINE/COLDLINE/ARCHIVE trade cheaper storage against retrieval cost and minimum duration. When access patterns are uncertain, enable `autoclass` instead — GCS moves each object to the cheapest justified class automatically (the spec rejects combining Autoclass with SetStorageClass lifecycle rules).

**Data protection** -- Enable `versioningEnabled: true` to keep noncurrent copies on every overwrite and delete, and always pair it with a lifecycle rule bounding version history. `retentionPolicy` adds WORM compliance (a LOCKED policy is irreversible); `softDeletePolicy` tunes GCP's 7-day deleted-object recovery window (0 disables it for scratch buckets). `forceDestroy` stays false by default so destroying a non-empty bucket fails instead of erasing data, and `deletionPolicy: PREVENT` fails the destroy outright for buckets automation must never remove. Lifecycle conditions also match by object size (`sizeAboveBytes`/`sizeBelowBytes`) — e.g. archive only large artifacts.

**Encryption** -- Google-managed encryption is used by default. Set `kmsKeyName` to a Cloud KMS key path — or reference a GcpKmsKey resource — for customer-managed encryption (CMEK) when compliance requires key control. To make the posture mandatory rather than advisory, `encryptionEnforcement` restricts which encryption types NEW objects may use — the CMEK-only compliance shape sets `googleManagedRestrictionMode` and `customerSuppliedRestrictionMode` to `FullyRestricted` (existing objects keep their encryption).

**Static website hosting** -- Configure `website.mainPageSuffix` and `website.notFoundPage` for direct GCS hosting (HTTP-only). For production HTTPS sites, front the bucket with the L7 load-balancer family: a GcpBackendBucket (CDN lives there) routed by a GcpUrlMap behind a GcpTargetHttpsProxy.

**Network-layer security** -- `ipFilter` restricts FROM WHERE the bucket may be reached before IAM is even evaluated: public CIDR ranges and/or VPC networks (reference GcpVpcNetwork resources). Defense-in-depth for data-exfiltration control.

**Structure and events** -- On hierarchical-namespace buckets, `folders` creates real directories (list every ancestor; the API never auto-creates parents). `managedFolders` needs only uniform bucket-level access and anchors prefix-scoped IAM — grant on `reports/` through the managed-folder IAM surface without granting the bucket. `notifications` streams object-change events (create, delete, metadata update, archive) into a Pub/Sub topic — reference a GcpPubSubTopic and grant the project's GCS service agent `roles/pubsub.publisher` on it first (the agent's email derives from this kind's `project_number` output). Notification configs are immutable: any change replaces the config, with a brief un-replayed gap in event delivery.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** | `kmsKeyName` | `status.outputs.key_id` |
| **GcpServiceAccount** | `iamMembers[].member` | `status.outputs.member` |
| **GcpGcsBucket** | `logging.logBucket` | `status.outputs.bucket_id` |
| **GcpVpcNetwork** | `ipFilter.vpcNetworkSources[].network` | `status.outputs.network_id` |
| **GcpPubSubTopic** | `notifications[].topic` | `status.outputs.topic_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `bucket_id` | Bucket ID (equals the globally unique name) | Backend buckets, function sources, Dataproc staging, Pub/Sub storage sinks |
| `bucket_name` | Bucket name (identical to `bucket_id`) | Whichever key reads naturally in configuration |
| `url` | Base URI in the form `gs://<bucket_name>` | SDKs, `gcloud storage`, data pipelines |
| `self_link` | API self link (`https://www.googleapis.com/storage/v1/b/...`) | REST tooling and audits |
| `location` | Location as reported by GCS (upper-cased) | Placement-aware automation |
| `project_number` | Numeric project number of the owning project | Service-agent IAM bindings |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private standard bucket** -- Uniform bucket-level access with public access prevention enforced, versioning enabled, a lifecycle rule bounding version history, and an additive grant for the workload's service account. The recommended default for application data, backups, and internal file storage. Start from the **Private Standard Bucket** preset.

**Static website** -- Public read via a single additive `allUsers` `objectViewer` grant (with prevention `inherited`), CORS opened for the application origin, and website routing with `index.html` and `404.html`. Start from the **Static Website Bucket** preset.

**Dual-region data lake with Autoclass** -- A custom dual-region pinned to your analytics regions, Autoclass to ARCHIVE, multipart-upload hygiene and `tmp/` TTL rules, and prefix-scoped reader grants via IAM conditions. Start from the **Dual-Region Data Lake with Autoclass** preset.

**Event-driven pipeline bucket** -- A private bucket that publishes object-change events to a Pub/Sub topic: `OBJECT_FINALIZE` under an `uploads/` prefix triggers downstream processing (Cloud Run, Cloud Functions, Eventarc all consume the topic). Composes with a GcpPubSubTopic and the GCS service agent's publisher grant. Start from the **Event-Driven Pipeline Bucket** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the bucket is created
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- customer-managed encryption for objects at rest
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- workload identities receiving bucket IAM grants
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- VPC sources for the bucket IP filter
- [**GCP Pub/Sub Topic**](/infra-catalog/gcp-pub-sub-topic) -- destination for bucket notification events
- [**GCP Backend Bucket**](/infra-catalog/gcp-backend-bucket) -- serves this bucket through the HTTPS load-balancer chain with CDN
