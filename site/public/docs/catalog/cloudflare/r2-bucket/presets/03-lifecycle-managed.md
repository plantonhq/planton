---
title: "Private R2 Bucket with Lifecycle and Retention"
description: "A private bucket that manages its own data over time: it tiers objects to Infrequent Access storage, expires them after a year, cleans up stalled multipart uploads, and locks audit objects for a..."
type: "preset"
rank: "03"
presetSlug: "03-lifecycle-managed"
componentSlug: "r2-bucket"
componentTitle: "R2 Bucket"
provider: "cloudflare"
icon: "package"
order: 3
---

# Private R2 Bucket with Lifecycle and Retention

A private bucket that manages its own data over time: it tiers objects to Infrequent Access storage, expires them after a year, cleans up stalled multipart uploads, and locks audit objects for a compliance retention period.

## When to Use

- Log or event archives that should get cheaper with age, then be deleted
- Compliance data that must be retained (write-once) for a fixed period
- Any bucket where storage cost and retention policy matter

## Key Configuration Choices

- **lifecycle.rules** (`lifecycle.rules`) -- Each rule selects objects by `conditions.prefix` (empty = all) and applies transitions.
- **storageClassTransitions** -- Moves objects to Infrequent Access by age or date (the only supported target class).
- **deleteObjectsTransition** -- Expires (deletes) objects by age or on a date.
- **abortMultipartUploadsTransition** -- Cleans up incomplete multipart uploads (always age-based).
- **lock.rules** -- Retains matching objects (write-once) for an age, until a date, or indefinitely.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<bucket-name>` | Unique bucket name | Choose DNS-safe name (e.g., app-logs-prod) |
| `<cloudflare-account-id>` | Cloudflare account ID | Dashboard → Overview → Account ID |

## Related Presets

- **01-private** -- A plain private bucket with no lifecycle management
- **02-public-cdn** -- A public bucket served over a custom domain
