---
title: "Standard Object Storage Bucket"
description: "This preset creates a Civo object storage bucket with versioning disabled. This is the most common configuration for application assets, static files, and general-purpose storage where previous..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "bucket"
componentTitle: "Bucket"
provider: "civo"
icon: "package"
order: 1
---

# Standard Object Storage Bucket

This preset creates a Civo object storage bucket with versioning disabled. This is the most common configuration for application assets, static files, and general-purpose storage where previous versions of objects are not needed.

## When to Use

- Static asset hosting (images, CSS, JavaScript)
- Application file uploads and media storage
- Data exports and staging areas
- Any storage use case where you don't need object version history

## Key Configuration Choices

- **Versioning disabled** (default) -- no previous versions retained; overwrites replace the object in place, keeping storage costs predictable
- **Region** (`region: lon1`) -- Civo object storage is region-bound; choose the region closest to your compute workloads

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `my-bucket` | Globally unique, DNS-compatible bucket name (3-63 chars, lowercase) | Your naming convention |
| `lon1` | Target Civo region | Civo dashboard or `civo region ls` |

## Related Presets

- **02-versioned-backup** -- Use instead when you need object versioning for backup/compliance scenarios
