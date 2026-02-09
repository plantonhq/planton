# OpenStackVolume: Research & Design Documentation

## OpenStack Cinder Block Storage

Cinder is OpenStack's block storage service. It provides persistent block-level storage volumes for use with Nova compute instances. Cinder manages the creation, attaching, and detaching of the block devices to servers.

### Key Concepts

- **Volume**: A persistent block device, analogous to a physical hard drive, that can be attached to an instance
- **Volume Type**: Backend storage class (e.g., SSD vs HDD, Ceph vs LVM) -- determines performance and redundancy characteristics
- **Snapshot**: A point-in-time copy of a volume that can be used to create new volumes
- **Availability Zone**: Volumes must be in the same AZ as the instance they're attached to (in most deployments)

### Terraform Resource: `openstack_blockstorage_volume_v3`

The Terraform provider supports ~15 fields. We selected 9 for the 80/20 set:

| Included | Excluded | Reason for Exclusion |
|----------|----------|---------------------|
| size | enable_online_resize | TF operational behavior |
| description | volume_retype_policy | TF-specific behavior |
| volume_type | backup_id | Niche (backup restoration) |
| availability_zone | consistency_group_id | Admin-only CG operations |
| snapshot_id | source_replica | Niche replication |
| source_vol_id | scheduler_hints | Volume placement hints |
| image_id | | |
| metadata | | |
| region | | |

### Source Field Mutual Exclusion

The Terraform provider enforces `ConflictsWith` between `snapshot_id`, `source_vol_id`, `image_id`, and `backup_id`. We enforce the same constraint via a CEL validation on the spec message, excluding `backup_id` (not in our 80/20 set).

### Foreign Key: `image_id`

The `image_id` field uses `StringValueOrRef` with `default_kind = OpenStackImage` and `default_kind_field_path = "status.outputs.image_id"`. This enables InfraChart DAG wiring between an Image resource and a Volume, supporting the "create image then create bootable volume" pattern.

OpenStackImage (2514) is pre-registered in the enum but not yet implemented (Phase 4). The FK annotation is forward-looking -- literal UUIDs work immediately, and `value_from` references will work once the Image component is built.

### Pulumi Resource: `openstack.blockstorage.VolumeV3`

The Pulumi OpenStack SDK v5 provides `blockstorage.NewVolumeV3()` with a `VolumeV3Args` struct that maps 1:1 to the Terraform schema fields. All fields use `pulumi.StringPtr` for optional strings and `pulumi.Int` for size.

## Design Decisions

1. **`size` is required with `gt > 0`** -- Even for snapshot/clone sources where Cinder can infer the size, requiring an explicit size prevents accidental undersizing and makes the manifest self-documenting.

2. **`image_id` as StringValueOrRef (not plain string)** -- Unlike Instance where `image_id` was kept as a plain string (Session 10), Volume uses StringValueOrRef because the bootable-volume-from-image pattern is more naturally expressed as an InfraChart dependency. OpenStackImage (2514) enum is pre-registered for forward compatibility.

3. **`snapshot_id` and `source_vol_id` as plain strings** -- Snapshots are typically pre-existing operational artifacts, and volume cloning is a niche use case. FK annotations would add complexity without meaningful InfraChart benefit.

4. **No `tags` field** -- Unlike networking resources, the Cinder volume V3 Terraform resource does not have a `tags` schema field. Metadata serves a similar purpose.

5. **`metadata` as `map<string,string>`** -- Standard proto map type. Terraform's computed metadata is handled by the provider; we only set user-provided entries.
