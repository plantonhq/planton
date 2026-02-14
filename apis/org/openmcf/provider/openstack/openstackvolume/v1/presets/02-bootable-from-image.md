# Bootable Volume from Image

This preset creates a Cinder volume pre-populated with a Glance image. The resulting volume is bootable and can be used as a root disk for instances. Create the volume first, then reference it in an `OpenStackInstance` block device mapping or attach it via `OpenStackVolumeAttach`.

## When to Use

- Pre-provisioning boot volumes before instance creation
- Workflows where volume creation and instance creation are separate DAG nodes
- Creating multiple instances from the same image with persistent root disks

## Key Configuration Choices

- **50 GB** (`size: 50`) -- must be at least as large as the image's `minDiskGb` requirement
- **Image source** (`imageId`) -- Cinder copies the image contents into the volume at creation time
- **ForceNew** -- `imageId` is immutable; changing the source image requires recreating the volume

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<image-id>` | UUID of the Glance image to populate the volume with | `openstack image list` or `OpenStackImage` status outputs |

## Related Presets

- **01-blank-data** -- Use instead for empty data volumes (not bootable)
