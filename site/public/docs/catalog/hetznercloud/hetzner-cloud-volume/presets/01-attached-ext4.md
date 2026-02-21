---
title: "Attached ext4 Volume"
description: "This preset creates a general-purpose Hetzner Cloud block storage volume, formatted with ext4 and attached to an existing server with automount enabled. It provisions an `hcloud_volume` resource and..."
type: "preset"
rank: "01"
presetSlug: "01-attached-ext4"
componentSlug: "hetzner-cloud-volume"
componentTitle: "Hetzner Cloud Volume"
provider: "hetznercloud"
icon: "package"
order: 1
---

# Attached ext4 Volume

This preset creates a general-purpose Hetzner Cloud block storage volume, formatted with ext4 and attached to an existing server with automount enabled. It provisions an `hcloud_volume` resource and an `hcloud_volume_attachment` resource that connects the volume to the target server. After deployment, the volume is mounted and ready to use with no manual steps.

The 50 GB starting size is suitable for application data, logs, uploads, and media files. Hetzner Cloud allows increasing volume size after creation (the provider resizes the underlying block device online), but size can never be decreased -- plan accordingly.

## When to Use

- Adding persistent storage to a server for application data, logs, or file uploads
- Any workload where you need disk space beyond the server's local NVMe and want data to survive server replacement
- Development and staging environments where delete protection is not yet needed

## Key Configuration Choices

- **ext4 filesystem** (`format: ext4`) -- the most widely supported Linux filesystem; correct default for general-purpose workloads, supported by every Linux distribution Hetzner offers
- **50 GB** (`size: 50`) -- a reasonable starting point; increase via spec update at any time (the provider resizes online), but size can never be decreased
- **Automount enabled** (`automount: true`) -- Hetzner Cloud mounts the volume at a system-assigned path automatically after attachment, eliminating manual `mount` commands
- **Falkenstein location** (`location: fsn1`) -- Hetzner's largest datacenter; change to `nbg1`, `hel1`, `ash`, `hil`, or `sin` to match the target server's location (volume and server must be co-located)
- **No delete protection** -- keeps the preset minimal for getting started; see the `02-database-storage` preset for the production-hardened variant

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<server-id>` | Numeric ID of the Hetzner Cloud server to attach the volume to | The `status.outputs.server_id` of your HetznerCloudServer resource, or the Servers page in the Hetzner Cloud Console |

## Related Presets

- **02-database-storage** -- XFS-formatted volume with delete protection for database workloads
- **03-unattached-reserve** -- pre-provisioned volume not attached to any server, for infrastructure preparation
