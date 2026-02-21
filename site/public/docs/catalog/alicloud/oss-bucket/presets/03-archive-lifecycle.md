---
title: "Archive Bucket with Lifecycle Rules"
description: "This preset creates a cost-optimized OSS bucket that automatically transitions objects through progressively cheaper storage tiers and expires them after one year. Versioning and encryption are..."
type: "preset"
rank: "03"
presetSlug: "03-archive-lifecycle"
componentSlug: "oss-bucket"
componentTitle: "OSS Bucket"
provider: "alicloud"
icon: "package"
order: 3
---

# Archive Bucket with Lifecycle Rules

This preset creates a cost-optimized OSS bucket that automatically transitions objects through progressively cheaper storage tiers and expires them after one year. Versioning and encryption are enabled, and old versions are cleaned up after 30 days. Designed for log archives, audit trails, and data retention workflows.

## When to Use

- Log aggregation and archival (application logs, access logs, audit trails)
- Data retention policies requiring automatic expiration after a defined period
- Cost optimization for large datasets that are frequently written but rarely accessed after the initial period
- Compliance workflows where data must be retained for a minimum duration, then securely deleted

## Key Configuration Choices

- **Lifecycle transitions** -- objects transition through storage tiers as they age:
  - **0-29 days**: Standard tier (frequent access)
  - **30-89 days**: IA (Infrequent Access) tier -- lower cost, retrieval fee per read
  - **90-364 days**: Archive tier -- significantly lower cost, minutes to restore
  - **365 days**: Permanent deletion

  This tiering strategy reduces storage costs by 60-80% compared to keeping all data in Standard.

- **Abort multipart upload** (`abortMultipartUploadDays: 7`) -- automatically cleans up incomplete multipart uploads after 7 days. Without this, orphaned upload parts accumulate silently and consume storage.

- **Noncurrent version expiration** (`noncurrentVersionExpirationDays: 30`) -- old object versions are expired after 30 days. This balances recovery capability (30-day window to restore previous versions) with storage cost control.

- **Versioning + encryption** -- standard production practices carried over from preset 02.

- **Empty prefix** (`prefix: ""`) -- the rule applies to all objects in the bucket. To scope lifecycle rules to specific paths, set the prefix (e.g., `"logs/"`, `"backups/"`).

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<alibaba-cloud-region>` | Alibaba Cloud region code | Your deployment region strategy |
| `<globally-unique-bucket-name>` | Bucket name (3-63 chars, globally unique) | Choose a name (e.g., `myorg-platform-log-archive`) |
| `<bucket-purpose>` | What this bucket stores (e.g., `log-archive`, `audit-trail`, `backup`) | Your data classification |

## Customization

- **Longer retention**: Change `expirationDays` to 730 (2 years), 1095 (3 years), or remove it entirely for indefinite retention.
- **Deep cold archival**: Add a third transition `{ days: 180, storageClass: ColdArchive }` for data that is truly never accessed.
- **Prefix scoping**: Set `prefix: "logs/"` to apply lifecycle rules only to objects under the `logs/` key prefix.

## Related Presets

- **01-private-standard** -- use instead for development buckets without lifecycle management
- **02-versioned-encrypted** -- use instead for production data without automatic expiration
