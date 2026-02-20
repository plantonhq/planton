# OciObjectStorageBucket — Design Notes

## Design Rationale

OciObjectStorageBucket bundles the Object Storage bucket, its lifecycle policy, and replication policies into a single declarative resource. This matches how operators think about bucket management: a bucket and its data governance rules are one logical unit.

### Why bundle lifecycle rules with the bucket?

OCI models lifecycle as a single `oci_objectstorage_object_lifecycle_policy` resource containing all rules for a bucket. There is exactly one lifecycle policy per bucket — it cannot be split across multiple Pulumi resources. Bundling lifecycle rules in the bucket spec reflects this 1:1 relationship and prevents conflicts that would arise from separate resources competing to own the same policy.

### Why bundle replication policies with the bucket?

Replication policies are tightly coupled to a specific source bucket. Each policy is an immutable OCI resource that replicates objects from the source bucket to a destination bucket. Managing them alongside the source bucket keeps the disaster recovery configuration co-located with the data it protects. Since all replication fields are immutable, updates require delete-and-recreate, which is simpler to reason about when the policy lives in the same manifest as the bucket.

### Why bundle retention rules with the bucket?

OCI retention rules are inline properties of the bucket resource (not separate sub-resources). The Pulumi OCI provider manages them as part of `objectstorage.BucketArgs.RetentionRules`. Separating them into a standalone component would add complexity without any benefit — they cannot be shared across buckets and they are part of the bucket's API surface.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Single lifecycle policy resource | Matches OCI's 1:1 model; no conflict risk | Adding/removing one rule reapplies all rules |
| Replication policies in same manifest | DR config co-located with source bucket | Changing destination requires delete + recreate of the policy entry |
| Retention rules inline on bucket | Matches Pulumi provider model; fewer resources to manage | All retention rules are updated together on any change |
| Enum values use proto names | YAML values match generated code exactly | Less readable than OCI Console names (e.g. `lifecycle_archive` vs `ARCHIVE`) |

## Resource Graph

```
OciObjectStorageBucket
├── oci_objectstorage_bucket (always)
│   └── retention_rules (inline, 0..100)
├── oci_objectstorage_object_lifecycle_policy (if lifecycleRules non-empty)
│   └── rules (1..N)
└── oci_objectstorage_replication_policy (0..N, one per entry)
```

The lifecycle policy and each replication policy declare `DependsOn` the bucket to ensure correct creation order.

## Deferred from v1

The following are excluded from the initial version. Each is either an operational concern with an independent lifecycle, a networking feature managed separately, or too low-adoption to include without demand:

- **Pre-authenticated requests** — operational access tokens with their own expiry lifecycle. Better managed via CLI or a separate component.
- **Private endpoints** — networking concern tied to VCN/subnet topology, managed independently.
- **Individual objects** — application-level concern; uploading objects is outside infrastructure provisioning scope.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.
- **Object Storage archival (restore)** — read-time operation, not a provisioning concern.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciObjectStorageBucket` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
