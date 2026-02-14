# General-Purpose ext4 Volume

This preset creates a 50 GiB block storage volume pre-formatted with ext4. ext4 is the most widely supported Linux filesystem and the best default for general application data, logs, and file storage.

## When to Use

- Application data storage (uploads, media, caches)
- Log aggregation and archival
- Any workload that benefits from a proven, general-purpose filesystem

## Key Configuration Choices

- **ext4 filesystem** (`filesystemType: ext4`) -- most compatible Linux filesystem; journaled, widely supported by all Linux distributions
- **50 GiB** (`sizeGib: 50`) -- reasonable starting size; scale up to 16,000 GiB as needed
- **Region** (`region: lon1`) -- must match the instance region for attachment

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `my-data-volume` | Descriptive volume name | Your naming convention |
| `lon1` | Target Civo region (must match instance) | Civo dashboard or `civo region ls` |
| `50` | Volume size in GiB (1-16,000) | Your storage requirements |

## Related Presets

- **02-xfs-database** -- Use instead for database workloads where XFS's superior sequential write performance is beneficial
