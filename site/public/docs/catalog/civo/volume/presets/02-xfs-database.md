---
title: "XFS Database Volume"
description: "This preset creates a 100 GiB block storage volume pre-formatted with XFS, optimized for database workloads. XFS excels at large sequential writes and parallel I/O, making it the preferred filesystem..."
type: "preset"
rank: "02"
presetSlug: "02-xfs-database"
componentSlug: "volume"
componentTitle: "Volume"
provider: "civo"
icon: "package"
order: 2
---

# XFS Database Volume

This preset creates a 100 GiB block storage volume pre-formatted with XFS, optimized for database workloads. XFS excels at large sequential writes and parallel I/O, making it the preferred filesystem for PostgreSQL, MySQL, and other database engines.

## When to Use

- Database storage (PostgreSQL, MySQL, MongoDB data directories)
- Write-heavy workloads with large sequential I/O patterns
- Workloads that benefit from XFS's parallel allocation and delayed logging

## Key Configuration Choices

- **XFS filesystem** (`filesystemType: xfs`) -- superior sequential write throughput and parallel I/O handling; preferred by PostgreSQL and MySQL documentation
- **100 GiB** (`sizeGib: 100`) -- typical database starting size; scale up to 16,000 GiB as data grows
- **Region** (`region: lon1`) -- must match the database instance region

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `my-database-volume` | Descriptive volume name | Your naming convention |
| `lon1` | Target Civo region (must match instance) | Civo dashboard or `civo region ls` |
| `100` | Volume size in GiB (1-16,000) | Your storage requirements |

## Related Presets

- **01-ext4-general** -- Use instead for general-purpose storage where maximum filesystem compatibility is preferred
