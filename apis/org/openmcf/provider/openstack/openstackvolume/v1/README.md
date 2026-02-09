# OpenStackVolume

An OpenStack Cinder block storage volume managed through OpenMCF.

## Overview

`OpenStackVolume` provisions a persistent block storage volume on OpenStack's Cinder service. Volumes can be attached to compute instances via `OpenStackVolumeAttach` and persist independently of instance lifecycle -- making them essential for databases, application data, and any workload requiring durable storage.

## When to Use

- **Persistent data storage**: Databases, file systems, application state
- **Bootable volumes**: Create a volume from a Glance image and boot an instance from it
- **Volume cloning**: Clone an existing volume for testing or migration
- **Snapshot restoration**: Restore a volume from a point-in-time snapshot

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `description` | string | no | Human-readable description |
| `size` | int32 | **yes** | Volume size in GB (must be > 0) |
| `volume_type` | string | no | Cinder volume type (e.g., "SSD", "HDD") |
| `availability_zone` | string | no | AZ for the volume (ForceNew) |
| `snapshot_id` | string | no | Snapshot UUID to restore from (ForceNew) |
| `source_vol_id` | string | no | Volume UUID to clone from (ForceNew) |
| `image_id` | StringValueOrRef | no | Glance image ID for bootable volumes (ForceNew) |
| `metadata` | map | no | Key-value metadata |
| `region` | string | no | Region override |

## Outputs

| Output | Description |
|--------|-------------|
| `volume_id` | UUID of the created volume (FK target for VolumeAttach) |
| `name` | Volume name (from metadata.name) |
| `size` | Provisioned size in GB |
| `volume_type` | Cinder volume type |
| `availability_zone` | AZ where volume was created |
| `region` | OpenStack region |

## Foreign Key Relationships

- `image_id` -> `OpenStackImage.status.outputs.image_id` (optional, for bootable volumes)
- Referenced by: `OpenStackVolumeAttach.volume_id`

## Validations

- `size` must be greater than 0
- At most one of `snapshot_id`, `source_vol_id`, or `image_id` may be set (mutual exclusion)
