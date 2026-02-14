# Blank Data Volume

This preset creates an empty Cinder volume for application data storage. The volume is unformatted -- attach it to an instance via `OpenStackVolumeAttach`, then partition and format it from within the instance. This is the most common volume configuration for databases, logs, and application data.

## When to Use

- Database storage (PostgreSQL, MySQL, MongoDB data directories)
- Application data that must persist independently of instance lifecycle
- Log storage, media uploads, or any data requiring persistent block storage

## Key Configuration Choices

- **100 GB** (`size: 100`) -- adjust to match your storage needs
- **Blank source** -- no `snapshotId`, `sourceVolId`, or `imageId`; volume is empty
- **Default volume type** -- uses the Cinder default backend; add `volumeType` to specify a backend (e.g., SSD, NVMe, replicated)

## Placeholders to Replace

No placeholders -- this preset is deployable as-is after setting `metadata.name` and adjusting `size`.

## Related Presets

- **02-bootable-from-image** -- Use instead when creating a bootable root volume from a Glance image
