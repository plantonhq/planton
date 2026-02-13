# Scaleway Block Storage Volume

## Overview

The **ScalewayBlockVolume** resource kind provides a declarative interface for creating and managing Block Storage volumes on Scaleway. Block volumes are network-attached SSD storage that persist independently of Instance lifecycle, making them ideal for data that must survive Instance stops, restarts, and replacements.

## Key Features

- **Two performance tiers** -- SBS 5K (5,000 IOPS) for general workloads and SBS 15K (15,000 IOPS) for databases and latency-sensitive applications. Both use modern NVMe disks.
- **Hot resize** -- Volume size can be increased in-place without detaching. Shrinking is not supported.
- **In-place tier changes** -- Performance tier can be upgraded or downgraded without recreating the volume.
- **Snapshot restore** -- Create a volume from an existing Block Storage snapshot for cloning or disaster recovery.
- **Automatic tagging** -- Standard OpenMCF metadata labels are applied as Scaleway tags for resource identification and cost allocation.

## Scaleway Terraform Resource Mapping

| OpenMCF Kind | Terraform Resource | Relationship |
|---|---|---|
| ScalewayBlockVolume | `scaleway_block_volume` | 1:1 standalone |

This is a standalone (non-composite) resource wrapping a single Terraform resource.

## Spec Fields

| Field | Type | Required | Description |
|---|---|---|---|
| `zone` | string | Yes | Availability Zone (e.g., "fr-par-1"). Cannot be changed after creation. |
| `size_gb` | uint32 | Yes | Volume size in GB (5-10240). Can be increased, cannot be shrunk. |
| `performance_tier` | enum | Yes | `sbs_5k` (5,000 IOPS) or `sbs_15k` (15,000 IOPS). Can be changed in-place. |
| `snapshot_id` | string | No | Block Storage snapshot ID to clone from. |

### Performance Tiers

| Tier | IOPS | Use Case |
|---|---|---|
| `sbs_5k` | 5,000 (combined R+W) | Web servers, application data, dev environments, CI/CD artifacts |
| `sbs_15k` | 15,000 (combined R+W) | Databases, message queues, real-time analytics, latency-sensitive apps |

> **Note:** For `sbs_15k` volumes, the attached Instance must have at least 3 GiB/s of block bandwidth to fully utilize the IOPS capacity.

## Stack Outputs

| Output | Description |
|---|---|
| `volume_id` | Zoned volume ID (`{zone}/{uuid}`). Primary cross-resource reference. |
| `volume_name` | Volume name as it exists in Scaleway Block Storage. |
| `zone` | Availability Zone where the volume is deployed. |

## Dependencies

**Upstream:** None. ScalewayBlockVolume is a standalone leaf resource with no `StringValueOrRef` inputs.

**Downstream:** The `volume_id` output can be referenced by ScalewayInstance to attach block storage. A volume must be in the same Availability Zone as the Instance it attaches to.

## Important Constraints

### Raw Block Device
Scaleway block volumes are raw block devices. There is no pre-formatting option in the API. After attaching a volume to an Instance, you must:
1. Format the volume: `mkfs.ext4 /dev/sdX` or `mkfs.xfs /dev/sdX`
2. Mount the volume: `mount /dev/sdX /mnt/data`
3. Add to `/etc/fstab` for persistence across reboots

### Size Operations
- **Increase:** Supported in-place (hot resize from the API side). After increasing, grow the partition and filesystem inside the OS (`growpart`, `resize2fs` for ext4, `xfs_growfs` for XFS).
- **Decrease:** Not supported. The Scaleway provider rejects any plan that shrinks `size_gb`.

### Zone Co-location
A block volume can **only** be attached to an Instance in the **same Availability Zone**. Plan your zone strategy before creating volumes.

### Attachment Limits
- Maximum **15 volumes** per Instance
- Maximum **10 TB** per volume
- A volume must be in the "available" state (detached) to attach to a different Instance

### What's Not Included (Deferred)
- **Block snapshots** (`scaleway_block_snapshot`) -- separate lifecycle; can be a future kind.
- **Legacy volume migration** (`instance_volume_id`) -- one-time operational migration from the Instance API's `b_ssd` volumes. Use Terraform directly for this.
