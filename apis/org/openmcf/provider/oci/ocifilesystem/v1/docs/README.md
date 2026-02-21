# OciFileSystem — Design Notes

## Design Rationale

OciFileSystem bundles the file system, mount target, export set configuration, and NFS exports into a single declarative resource. This matches the typical deployment pattern: a file system needs a mount target to be accessible, and at least one export to define the NFS path.

### Why bundle the mount target with the file system?

A file system without a mount target is not accessible over the network. While OCI allows a mount target to be shared across multiple file systems, the common case is a dedicated mount target per file system. Bundling them together means:

- One manifest creates a fully functional NFS endpoint — no multi-step orchestration needed.
- The mount target's lifecycle is tied to the file system — deletion removes both.
- IP address, hostname, subnet, and throughput configuration live alongside the file system they serve.

If sharing a mount target across file systems is needed, that requires managing mount targets as a separate component (deferred from v1).

### Why bundle exports with the file system?

Exports connect the file system to the mount target at specific paths. Without at least one export, the file system cannot be mounted. Including exports in the spec ensures that:

- The minimum-viable deployment (file system + mount target + export) is achieved in a single manifest.
- Export access control (per-CIDR permissions, identity squashing) is co-located with the resource it protects.
- Adding or removing exports is a simple list operation in the same manifest.

### Why configure the export set via the mount target?

OCI automatically creates an export set when a mount target is provisioned. The `maxFsStatBytes` and `maxFsStatFiles` settings control what NFS clients see for available capacity via statfs. These are properties of the export set, which is intrinsically tied to the mount target. Configuring them in `mountTarget` avoids exposing a separate export set resource.

The export set resource is only created when either `maxFsStatBytes` or `maxFsStatFiles` is set — otherwise, the auto-created export set is left with OCI defaults.

## Trade-offs

| Decision | Benefit | Cost |
|----------|---------|------|
| Dedicated mount target per file system | Simple lifecycle, predictable IP assignment | Uses one mount target per file system (OCI default limit: 2 per AD) |
| Exports as part of file system spec | One manifest creates a mountable file system | Cannot share exports across file systems without duplicating config |
| Export set config on mount target | No separate export set resource to manage | maxFsStatBytes/maxFsStatFiles only configurable if mount target is managed here |
| Export options per export | Fine-grained per-path access control | Repeated config if multiple exports share the same access rules |

## Resource Graph

```
OciFileSystem
├── oci_file_storage_file_system (always)
├── oci_file_storage_mount_target (always)
│   └── oci_file_storage_export_set (if maxFsStatBytes or maxFsStatFiles set)
└── oci_file_storage_export (1..N, one per exports[] entry)
    └── export_options (0..N inline, per-source access rules)
```

Exports declare `DependsOn` both the file system and the mount target. The export set declares `DependsOn` the mount target.

## Deferred from v1

The following are excluded from the initial version. Each has an independent lifecycle, low adoption, or requires separate orchestration:

- **oci_file_storage_snapshot** — operational concern with its own lifecycle (e.g. before-upgrade snapshots). Better managed via CLI or automation.
- **oci_file_storage_replication** — cross-region replication with independent lifecycle and its own target file system. Requires separate networking setup.
- **oci_file_storage_filesystem_snapshot_policy** — reusable across file systems. Referenced via `filesystemSnapshotPolicyId` but not created by this component.
- **oci_file_storage_file_system_quota_rule** — advanced admin feature with low adoption.
- **oci_file_storage_outbound_connector** — specialized LDAP integration for identity mapping.
- **Kerberos / LDAP ID mapping on mount target** — very low adoption; requires external directory infrastructure.
- **source_snapshot_id** — clone/restore scenario requiring separate orchestration.
- **Shared mount targets** — using one mount target across multiple file systems requires a separate mount target component.
- **defined_tags / system_tags** — managed by platform. `freeform_tags` are auto-populated from `metadata.labels`.
- **locks** — platform-managed resource locks.

## Freeform Tags

The module automatically populates freeform tags from metadata:

| Tag Key | Source |
|---------|--------|
| `resource` | `"true"` (constant) |
| `resource_kind` | `OciFileSystem` |
| `resource_id` | `metadata.id` |
| `organization` | `metadata.org` (if set) |
| `environment` | `metadata.env` (if set) |
| All `metadata.labels` | Copied as-is |
