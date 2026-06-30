# General-Purpose ext4 Volume

This preset creates a DigitalOcean block storage volume pre-formatted with ext4, ready to attach and mount on a Droplet immediately. Suitable for application data, logs, media files, or any general-purpose storage need.

## When to Use

- Application data storage (uploads, logs, backups)
- Persistent storage for stateless Droplets
- Any workload that benefits from ext4's wide compatibility and journaling

## Key Configuration Choices

- **ext4 filesystem** (`filesystemType: ext4`) -- the most widely compatible Linux filesystem. Pre-formatting means the volume is immediately mountable after attachment, no manual `mkfs` required.
- **50 GiB** (`sizeGib: 50`) -- a reasonable starting size. DigitalOcean volumes can be resized up to 16 TiB without detaching.
- **Region** (`region: nyc1`) -- must match the Droplet's region. Volumes cannot cross regions.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `nyc1` | Target DigitalOcean region slug | Must match the Droplet's region |

## Related Presets

- **02-database-xfs** -- Use instead for database workloads (PostgreSQL, MySQL) where XFS provides better write performance
