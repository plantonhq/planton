# GCP GCS Bucket IAM Member

Deploys a single ADDITIVE IAM grant ON a Cloud Storage bucket (`google_storage_bucket_iam_member`) — one role, to one member, on one bucket. The grant merges into the bucket's IAM policy without touching any other member's bindings, and removal subtracts only this exact pair.

Use this kind for a grantee that DEPENDS ON the bucket. The canonical case is a `GcpLoggingSink` exporting into the bucket: the sink names the bucket as its destination, and its writer identity needs `roles/storage.objectCreator` on that same bucket. Declared in the bucket's own `iam_members`, that grant would make the bucket depend on the sink while the sink depends on the bucket. This kind depends on both, so the order is always bucket, then sink, then grant. For every other grantee, the bucket's own `iam_members` field is the simpler home.

## What Gets Created

When you deploy a GcpGcsBucketIamMember resource, Planton provisions:

- **Bucket IAM Member** — one (role, member[, condition]) entry merged into the target bucket's IAM policy

Nothing else in the policy is read as owned or modified — grants made by other charts, the bucket's own `iam_members`, or other tools are never clobbered.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing bucket** — referenced via `bucket` (a GcpGcsBucket's `bucket_id` output)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs
- **The member must exist** — a sink's writer identity, a service agent, or a GcpServiceAccount

## Quick Start

Create a file `sink-writer.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGcsBucketIamMember
metadata:
  name: audit-archive-writer
spec:
  bucket:
    value: acme-audit-archive
  role:
    value: roles/storage.objectCreator
  member:
    value: serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com
```

Deploy:

```shell
planton apply -f sink-writer.yaml
```

Or compose both sides by reference — the shape this kind exists for:

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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `bucket` | `StringValueOrRef` | The bucket whose IAM policy receives the grant, by its globally unique name (not a `gs://` URL). References a GcpGcsBucket's `bucket_id` output by default. Immutable. |
| `role` | `StringValueOrRef` | The role to grant on the bucket: a predefined role (`roles/storage.objectCreator`, `roles/storage.objectViewer`, `roles/storage.objectUser`, ...) or a custom role's full name. References a GcpIamCustomRole's `name` output by default. Immutable. |
| `member` | `StringValueOrRef` | The identity receiving the grant, in IAM member format. References a GcpServiceAccount's `member` output by default; a GcpLoggingSink's `writer_identity` is the other candidate. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `condition` | `object` | — | IAM Condition (`title`, `expression`, optional `description`) restricting when the grant applies — an expiry, or an object-name prefix. Requires uniform bucket-level access on the bucket. Part of the grant's identity. Immutable. |

There is no project field: bucket names are global, not project-scoped.

### Member Formats

| Format | Grants to |
|--------|-----------|
| `serviceAccount:<email>` | A service account, a sink's writer identity, or a Google service agent |
| `user:<email>` | A Google account |
| `group:<email>` | A Google group |
| `domain:<domain>` | Everyone in a Workspace or Cloud Identity domain |
| `principal://...` / `principalSet://...` | Workload identity federation principals |
| `allUsers` / `allAuthenticatedUsers` | Public access to the objects the role covers; refused by buckets with public access prevention enforced |

Grants to deleted principals (`deleted:...`) are refused by validation.

## This Kind vs the Bucket's `iam_members`

Both are additive and never fight. Use the bucket's own `iam_members` for grantees known when the bucket is declared — a workload's service account, a group, a service agent. Use this kind only when the grantee is created from the bucket (a sink exporting into it). Never declare the same (role, member) pair in both places: removing either one removes the grant.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `bucket` | `string` | The bucket whose policy received the grant (after reference resolution) |
| `role` | `string` | The granted role (after reference resolution) |
| `member` | `string` | The granted member (after reference resolution) |
| `etag` | `string` | The bucket IAM policy etag after the grant — useful for audit correlation |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Everything is immutable**: IAM grants have no update — changing bucket, role, member, or condition replaces the grant (destroy the old pair, create the new one), which mirrors the underlying API exactly.
- **Conditions are part of the grant's identity**: the same role granted with and without a condition are two independent grants. Conditions need uniform bucket-level access.
- **Only additive grants are modeled**: authoritative per-role bindings and whole-policy writes clobber every grant they do not list and are deliberately not modeled.
- **A sink exports nothing until this grant lands**: deploy the grant in the same chart as the sink so the first hourly batch is written.

## Related Components

- [GcpGcsBucket](/docs/catalog/gcp/gcpgcsbucket) — the bucket being granted on (its `bucket_id` output feeds this component; its own `iam_members` covers every other grantee)
- [GcpLoggingSink](/docs/catalog/gcp/gcploggingsink) — a sink exporting to the bucket (its `writer_identity` output feeds `member`)
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — a grantable identity (its `member` output feeds this component)
- [GcpPubSubTopicIamMember](/docs/catalog/gcp/gcppubsubtopiciammember) — the same pattern for a sink exporting to a topic

## Additional Resources

- [Cloud Storage IAM roles](https://cloud.google.com/storage/docs/access-control/iam-roles)
- [Route logs to supported destinations](https://cloud.google.com/logging/docs/export/configure_export_v2)
- [IAM Conditions](https://cloud.google.com/iam/docs/conditions-overview)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
