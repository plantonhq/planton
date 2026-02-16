---
title: "Database XFS Volume"
description: "This preset creates a DigitalOcean block storage volume pre-formatted with XFS, optimized for database workloads. XFS provides superior write performance for the sequential and random I/O patterns..."
type: "preset"
rank: "02"
presetSlug: "02-database-xfs"
componentSlug: "volume"
componentTitle: "Volume"
provider: "digitalocean"
icon: "package"
order: 2
---

# Database XFS Volume

This preset creates a DigitalOcean block storage volume pre-formatted with XFS, optimized for database workloads. XFS provides superior write performance for the sequential and random I/O patterns typical of PostgreSQL, MySQL, and other databases.

## When to Use

- Self-managed PostgreSQL, MySQL, or other database engines on Droplets
- Any workload with heavy sequential writes (WAL files, transaction logs)
- Large data volumes where XFS's allocation group parallelism provides a performance advantage

## Key Configuration Choices

- **XFS filesystem** (`filesystemType: xfs`) -- optimized for large files and parallel I/O. PostgreSQL and MySQL both benefit from XFS for WAL and data files.
- **100 GiB** (`sizeGib: 100`) -- larger starting size appropriate for database workloads. Resize as data grows (up to 16 TiB).
- **Database tag** -- enables tag-based firewall rules restricting access to database-tier resources.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `nyc1` | Target DigitalOcean region slug | Must match the Droplet's region |

## Related Presets

- **01-general-purpose-ext4** -- Use instead for general application data where ext4 compatibility is preferred
