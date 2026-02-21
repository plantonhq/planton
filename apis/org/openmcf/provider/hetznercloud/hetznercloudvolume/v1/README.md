# HetznerCloudVolume

The **HetznerCloudVolume** resource provisions a network-attached block storage volume in Hetzner Cloud. Volumes persist independently of any server — detaching or deleting the server does not affect the volume's data. This makes volumes the standard mechanism for persistent state in Hetzner Cloud: databases, application state, uploaded files, and any data that must survive server replacement.

## What It Represents

A [Hetzner Cloud Volume](https://docs.hetzner.cloud/#volumes) is an SSD-backed block storage device ranging from 10 GB to 10,240 GB (10 TB). The volume exists in a specific location (e.g., `fsn1`, `nbg1`, `hel1`) and can only be attached to servers in the same location. A volume can be attached to exactly one server at a time and appears as a standard block device in the server's OS.

Volumes can be pre-formatted with ext4 or xfs at creation time, or created raw for manual formatting. Size can be increased online after creation, but can never be decreased — the Hetzner Cloud API rejects size reductions.

## Bundled Resources

| Terraform Resource | Count | Created When | Purpose |
|---|---|---|---|
| `hcloud_volume` | 1 | Always | Provisions the block storage volume with the specified size, location, optional filesystem format, labels, and delete protection |
| `hcloud_volume_attachment` | 0 or 1 | When `serverId` is set | Attaches the volume to the specified server with optional automount |

The attachment is a separate resource because volumes outlive servers. Destroying the attachment (detaching) does not affect the volume, and destroying the volume does not require updating the server. This lifecycle separation is why the Hetzner Cloud API and both IaC providers model creation and attachment as distinct operations.

## Key Features

### Volume Size

The `size` field specifies the volume capacity in GB. The Hetzner Cloud range is 10–10,240 GB. Size can be increased after creation (the provider resizes the underlying block device online, without detaching), but it can never be decreased — the Hetzner Cloud API rejects reductions.

After increasing the volume size via the spec, the filesystem inside the volume must be resized separately from within the server's OS (`resize2fs` for ext4, `xfs_growfs` for xfs). The IaC module resizes the block device; the filesystem resize is an in-server operation.

### Location Affinity

The `location` field determines the physical datacenter where the volume is stored. The volume can only be attached to servers in the same location. Changing the location forces replacement of the volume (data loss).

### Filesystem Format

The `format` field selects a filesystem to create at volume creation time:

| Value | Behavior |
|-------|----------|
| `ext4` | General-purpose Linux filesystem. Recommended for most workloads. |
| `xfs` | High-performance filesystem suited for large files and high-throughput workloads. |
| (unspecified) | Raw block device with no filesystem. Must be formatted manually on the server. |

This is a create-time-only setting: the provider does not read the filesystem format back from the API after creation. Changing the format in the spec after the initial apply has no effect on the existing volume.

### Server Attachment

The `serverId` field accepts a literal Hetzner Cloud server ID (as a string) or a reference to a `HetznerCloudServer` resource's output via `valueFrom`. When set, an `hcloud_volume_attachment` resource is created. When omitted, the volume is created unattached and available for later attachment.

The referenced server must be in the same location as the volume.

### Automount

When `automount` is `true` and `serverId` is set, Hetzner Cloud automatically mounts the volume on the server after the initial attachment. This is a create-time-only setting — the provider does not read it back after creation.

Automount uses the volume's filesystem label to determine the mount point. For subsequent reattachments (to a different server), the volume is mounted at the last known mount point. If the mount point conflicts with an existing path, the mount may fail and manual mounting is required.

### Delete Protection

When `deleteProtection` is `true`, the volume cannot be deleted via the Hetzner Cloud API until protection is removed. This prevents accidental deletion of volumes containing critical data.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the volume from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence. The attachment resource does not support labels in the Hetzner Cloud API.

## Upstream Dependencies (What This Resource Needs)

| Dependency | Field | Required | Cardinality | Purpose |
|---|---|---|---|---|
| `HetznerCloudServer` | `spec.serverId` | No | 0..1 | Server to attach the volume to. Must be in the same location. |

The server dependency is optional. A volume can be created with no server reference.

## Downstream Dependents (What References This Resource)

No other OpenMCF components currently reference `HetznerCloudVolume` outputs. The `volume_id` output is available for external use (scripts, monitoring, manual operations) and for future components that may reference volumes.

## Stack Outputs

| Output | Description |
|---|---|
| `volume_id` | Hetzner Cloud numeric ID of the created volume (as string). Available for external reference. |
| `linux_device` | The Linux device path for the volume on the attached server (e.g., `/dev/disk/by-id/scsi-0HC_Volume_12345678`). Stable across reboots. Empty if the volume is not attached. Suitable for `/etc/fstab` entries. |

## References

- [Hetzner Cloud Volumes Documentation](https://docs.hetzner.cloud/#volumes)
- [Hetzner Cloud Volume Pricing](https://docs.hetzner.cloud/#pricing)
- [Terraform hcloud_volume Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/volume)
- [Terraform hcloud_volume_attachment Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/volume_attachment)
- [Pulumi hcloud.Volume Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/volume/)
- [Pulumi hcloud.VolumeAttachment Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/volumeattachment/)
