# OciBlockVolume — Design Notes

## Design Rationale

OciBlockVolume wraps the block volume and an optional backup policy assignment into a single declarative resource. This matches how operators think about block storage: a volume and its backup schedule are one logical unit.

### Why is backup policy assignment bundled?

A backup policy assignment is a simple 1:1 binding between a volume and a policy. It has no independent lifecycle — removing the volume makes the assignment meaningless. Bundling it avoids a separate manifest and keeps backup configuration co-located with the volume it protects.

### Why are autotune policies inline on the volume?

OCI models autotune policies as inline properties of the volume resource, not as separate sub-resources. The Pulumi OCI provider manages them as part of `core.VolumeArgs.AutotunePolicies`. Separating them would add a resource boundary that doesn't exist in the API.

### Why are cross-region replicas inline on the volume?

Like autotune policies, replicas are inline on the volume in the OCI API. Each replica specifies a target AD and optional encryption key. Managing them separately would create a misleading abstraction — replicas cannot exist without the source volume and are always configured as part of the volume lifecycle.

### Why is `sizeInGbs` not marked as optional despite having an OCI default?

OCI defaults to 1 TB (1024 GB) when size is omitted. This is a dangerous default for development and testing — accidentally creating a 1 TB volume is expensive. By requiring explicit size, the manifest author must make a conscious decision about capacity.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Backup policy as sub-resource | Co-located backup config; single manifest | Adding/removing policy requires re-apply of the volume stack |
| Autotune policies inline | Matches OCI API model; no extra resource boundary | All policies updated together on any change |
| Cross-region replicas inline | Matches OCI API model; DR config co-located | Adding a replica triggers volume update, not an independent create |
| Explicit `sizeInGbs` | Prevents accidental 1 TB default | Slightly more verbose minimal manifest |
| VPUs/GB as integer | Matches OCI API (string-encoded int) | Must know valid values (0, 10, 20, 30-120) |

## Resource Graph

```
OciBlockVolume
├── oci_core_volume (always)
│   ├── autotune_policies (inline, 0..N)
│   └── block_volume_replicas (inline, 0..N)
└── oci_core_volume_backup_policy_assignment (if backupPolicyId is set)
```

The backup policy assignment declares the volume as its `AssetId`, ensuring correct creation order.

## Deferred from v1

The following are excluded from the initial version:

- **source_details** — clone/restore from existing volume, backup, or replica. These are operational workflows, not initial provisioning.
- **cluster_placement_group_id** — very niche placement constraint for HPC workloads.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.
- **is_auto_tune_enabled** — deprecated by OCI in favor of `autotune_policies`.
- **size_in_mbs** — deprecated by OCI in favor of `size_in_gbs`.
- **volume_backup_id** — deprecated by OCI in favor of `source_details`.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciBlockVolume` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
