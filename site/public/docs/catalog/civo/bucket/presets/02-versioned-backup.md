---
title: "Versioned Backup Bucket"
description: "This preset creates a Civo object storage bucket with versioning enabled, retaining all previous versions of every object. Suitable for backup repositories, compliance archives, and any scenario..."
type: "preset"
rank: "02"
presetSlug: "02-versioned-backup"
componentSlug: "bucket"
componentTitle: "Bucket"
provider: "civo"
icon: "package"
order: 2
---

# Versioned Backup Bucket

This preset creates a Civo object storage bucket with versioning enabled, retaining all previous versions of every object. Suitable for backup repositories, compliance archives, and any scenario where accidental overwrites or deletions must be recoverable.

## When to Use

- Database backup storage where point-in-time recovery is critical
- Compliance or audit archives requiring immutable history
- Configuration or artifact storage where rollback capability is needed
- Any use case where accidental deletion or overwrite must be reversible

## Key Configuration Choices

- **Versioning enabled** (`versioningEnabled: true`) -- every PUT creates a new version; previous versions are retained and retrievable via the S3-compatible API
- **Region** (`region: lon1`) -- choose the region closest to the backup source to minimize transfer latency

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `my-backup-bucket` | Globally unique, DNS-compatible bucket name (3-63 chars, lowercase) | Your naming convention |
| `lon1` | Target Civo region | Civo dashboard or `civo region ls` |

## Related Presets

- **01-standard** -- Use instead when versioning is unnecessary and storage cost predictability is preferred
