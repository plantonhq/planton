---
title: "Versioned Bucket with Lifecycle Rules"
description: "This preset creates a Scaleway Object Storage bucket with versioning enabled and a lifecycle rule that transitions objects to Glacier cold storage after 90 days. Incomplete multipart uploads are..."
type: "preset"
rank: "02"
presetSlug: "02-versioned-lifecycle"
componentSlug: "object-bucket"
componentTitle: "Object Bucket"
provider: "scaleway"
icon: "package"
order: 2
---

# Versioned Bucket with Lifecycle Rules

This preset creates a Scaleway Object Storage bucket with versioning enabled and a lifecycle rule that transitions objects to Glacier cold storage after 90 days. Incomplete multipart uploads are automatically cleaned up after 7 days. This is the standard production configuration for data that needs version history and cost-optimized long-term storage.

## When to Use

- Production data requiring object version history for compliance or recovery
- Backups, logs, and archives that should be automatically moved to cold storage
- Applications where accidental overwrites or deletions need to be recoverable

## Key Configuration Choices

- **Versioning enabled** (`versioningEnabled: true`) -- every overwrite or delete creates a new version; previous versions can be restored
- **90-day Glacier transition** -- objects are automatically moved to Scaleway Glacier after 90 days, reducing storage costs for infrequently accessed data
- **7-day multipart cleanup** (`abortIncompleteMultipartUploadDays: 7`) -- prevents orphaned multipart uploads from accumulating and consuming storage
- **Force destroy disabled** (`forceDestroy: false`) -- bucket cannot be deleted while it contains objects or versions

## Placeholders to Replace

No placeholders -- this preset is ready to deploy as-is. Adjust the lifecycle `days` threshold and add `expirationDays` if objects should be permanently deleted after a certain period.

## Related Presets

- **01-private-bucket** -- Use instead for simple storage without versioning or lifecycle management
