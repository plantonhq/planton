# GCP GCS Bucket IAM Member

Grants one role, to one identity, on ONE Cloud Storage bucket — for grantees that depend on the bucket, like a logging sink exporting into it. Additive: it merges into the bucket's IAM policy without touching any other member's bindings, and removal subtracts only this exact (role, member) pair. The bucket's own `iam_members` stays the home for every other grantee; this kind exists because a sink names the bucket, so a grant to the sink's identity declared on the bucket would be a dependency cycle.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Bucket IAM Member Binding** -- a `google_storage_bucket_iam_member` merging the (role, member) pair into the target bucket's IAM policy, with an optional IAM Condition attached

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials permitted to set IAM policy on the target bucket (e.g. `roles/storage.admin` on the bucket or its project). Map it as the default for your environment, or specify it explicitly.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A Cloud Storage bucket** whose IAM policy receives the grant. Provide its name directly or reference a GcpGcsBucket Infra Component via ValueFromRef.
- **The identity** receiving the grant must already exist — a sink's writer identity exists once the sink is created.
- **Uniform bucket-level access** on the bucket if the grant carries a condition.

## Deploy

### Console

Open the deployment store, find **GCP GCS Bucket IAM Member**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the grant definition. Start from the **Logging Sink Writer** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGcsBucketIamMember
metadata:
  name: audit-archive-writer
  org: acme-corp
  env: prod
spec:
  bucket:
    value: acme-audit-archive
  role:
    value: roles/storage.objectCreator
  member:
    value: serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com
```

```shell
planton apply -f gcp-gcs-bucket-iam-member.yaml
```

This merges one binding into the bucket's policy. An Infra Job tracks the provisioning in real time.

### InfraChart

The composed form is where this kind shines — bucket, sink, and grant wired in one InfraPipeline without a cycle:

```yaml
spec:
  bucket:
    valueFrom:
      kind: GcpGcsBucket
      name: audit-archive
      fieldPath: status.outputs.bucket_id
  role:
    value: roles/storage.objectCreator
  member:
    valueFrom:
      kind: GcpLoggingSink
      name: audit-export
      fieldPath: status.outputs.writer_identity
```

The InfraPipeline deploys the bucket, then the sink that names it, then this grant with both values resolved.

## Key Configuration

These are the most important decisions when configuring a grant. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**This kind or the bucket's `iam_members`** -- use this kind only for a grantee created from the bucket (a sink exporting into it). Everyone else belongs in the bucket's own `iam_members`, where the grants travel with the bucket. Never declare the same (role, member) pair in both: removing either removes the grant.

**The sink pattern** -- a logging sink writes hourly batches as its writer identity. Grant that identity `roles/storage.objectCreator` — write only, never read — by referencing the sink's `writer_identity` output.

**Role** -- `roles/storage.objectCreator` for write-only uploaders, `roles/storage.objectViewer` to read, `roles/storage.objectUser` to read, write, and delete, `roles/storage.legacyBucketReader` to list, or a custom role's full name.

**IAM Condition** -- an optional CEL expression scoping when or where the grant applies: an expiry, or an object-name prefix such as `resource.name.startsWith("projects/_/buckets/<bucket>/objects/logs/")`. Conditions require uniform bucket-level access. The condition is part of the grant's identity.

**Everything replaces atomically** -- an IAM grant has no update. Changing bucket, role, member, or condition replaces the grant, and for the moment between delete and create the member cannot write.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpGcsBucket** | `bucket` | `status.outputs.bucket_id` |
| **GcpIamCustomRole** (optional) | `role` | `status.outputs.name` |
| **GcpServiceAccount** (optional) | `member` | `status.outputs.member` |
| **GcpLoggingSink** (optional) | `member` | `status.outputs.writer_identity` |

### What This Kind Provides

This kind has no outputs a downstream Infra Component would consume: `status.outputs` records the grant's post-resolution facts — the (`bucket`, `role`, `member`) triple after any references were resolved, plus the bucket IAM policy `etag` at the moment this grant merged. They exist for audit and drift review, not for ValueFromRef wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Logging sink writer** -- a sink's writer identity gets write-only access to its destination bucket; the grant every bucket log export needs. Start from the **Logging Sink Writer** preset.

**Prefix-scoped writer** -- an uploader may create objects only under one prefix, enforced by an IAM condition. Start from the **Prefix-Scoped Object Writer** preset.

## Works With

- [**GCP GCS Bucket**](/infra-catalog/gcp-gcs-bucket) -- its `bucket_id` output feeds the bucket field; its own `iam_members` covers grantees that do not depend on it
- [**GCP Logging Sink**](/infra-catalog/gcp-logging-sink) -- its `writer_identity` output feeds the member field for bucket log exports
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- its `member` output feeds the member field for workload access
- [**GCP IAM Custom Role**](/infra-catalog/gcp-iam-custom-role) -- its `name` output feeds the role field for curated permission bundles
