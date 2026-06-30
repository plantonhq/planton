# HetznerCloudSnapshot

The **HetznerCloudSnapshot** resource creates a point-in-time disk image from a Hetzner Cloud server. The snapshot is stored as a Hetzner Cloud Image (type `snapshot`) and can be used as a boot source when creating new servers. Snapshots persist independently of the source server — deleting or replacing the server does not affect existing snapshots.

## What It Represents

A [Hetzner Cloud Snapshot](https://docs.hetzner.cloud/#server-actions-create-image) captures the complete disk state of a server at the moment the snapshot is taken. The resulting image includes the OS, installed software, configuration, and all data on the server's local disk. Snapshots are the standard mechanism for golden images, pre-upgrade rollback points, and server cloning in Hetzner Cloud.

Snapshots are billed based on the disk size of the source server, not the amount of data actually used. A snapshot of a CX22 server (40 GB local disk) costs the same regardless of whether 2 GB or 38 GB of that disk is in use.

## Bundled Resources

| Terraform Resource | Count | Created When | Purpose |
|---|---|---|---|
| `hcloud_snapshot` | 1 | Always | Creates a server snapshot stored as a Hetzner Cloud Image. The snapshot captures the full disk of the source server. |

This is a single-resource component with no conditional resources.

## Key Features

### Server Reference

The `serverId` field identifies the server to snapshot. It accepts a literal Hetzner Cloud server ID (as a string) or a reference to a `HetznerCloudServer` resource's output via `valueFrom`. The `valueFrom` shorthand resolves the server's numeric ID automatically, avoiding manual lookups.

Changing `serverId` after creation forces replacement of the snapshot — the existing snapshot is destroyed and a new one is created from the new server. This is because the Hetzner Cloud provider marks `server_id` as `ForceNew`.

### Description

The `description` field provides a human-readable label for the snapshot. It is optional and can be updated after creation without replacing the snapshot. Useful for tagging purpose (e.g., "pre-upgrade baseline 2026-02-19", "golden image v2.1").

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the snapshot from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence on key conflicts.

### Snapshot as a Boot Source

The `snapshot_id` output is a Hetzner Cloud image ID. This ID can be passed as the `image` field when creating a new `HetznerCloudServer`, booting the new server from the captured disk state. This enables the golden image workflow: configure a server once, snapshot it, and stamp out identical servers from the snapshot.

## Upstream Dependencies (What This Resource Needs)

| Dependency | Field | Required | Cardinality | Purpose |
|---|---|---|---|---|
| `HetznerCloudServer` | `spec.serverId` | Yes | 1 | The server whose disk is captured as a snapshot. |

The server must exist before the snapshot can be created. A `valueFrom` reference establishes this dependency edge automatically.

## Downstream Dependents (What References This Resource)

| Dependent | Field | Purpose |
|---|---|---|
| `HetznerCloudServer` | `spec.image` | Boot a new server from this snapshot's image ID. |

The `snapshot_id` output is the Hetzner Cloud image ID that can be used as a server's `image` parameter.

## Stack Outputs

| Output | Description |
|---|---|
| `snapshot_id` | The Hetzner Cloud image ID of the created snapshot (as a string). Usable as the `image` parameter when creating new servers. |

## References

- [Hetzner Cloud Server Actions — Create Image](https://docs.hetzner.cloud/#server-actions-create-image)
- [Hetzner Cloud Images Documentation](https://docs.hetzner.cloud/#images)
- [Terraform hcloud_snapshot Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/snapshot)
- [Pulumi hcloud.Snapshot Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/snapshot/)
