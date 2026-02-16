---
title: "Public Static Website Bucket"
description: "This preset creates a publicly accessible S3 bucket configured for static website hosting with CORS enabled for cross-origin asset loading. It allows GET and HEAD requests from any origin, making it..."
type: "preset"
rank: "02"
presetSlug: "02-public-static-website"
componentSlug: "s3-bucket"
componentTitle: "S3 Bucket"
provider: "aws"
icon: "package"
order: 2
---

# Public Static Website Bucket

This preset creates a publicly accessible S3 bucket configured for static website hosting with CORS enabled for cross-origin asset loading. It allows GET and HEAD requests from any origin, making it suitable for hosting HTML, CSS, JavaScript, images, and other static assets.

## When to Use

- Static website hosting (HTML, CSS, JavaScript)
- Public asset storage (images, fonts, downloadable files)
- Frontend application hosting (React, Vue, Angular builds) when not using CloudFront

## Key Configuration Choices

- **Public access** (`isPublic: true`) -- Block Public Access is disabled; objects are accessible from the internet
- **CORS enabled** -- Allows GET and HEAD from any origin with 1-hour cache; required when web pages on other domains load assets from this bucket
- **Versioning disabled** (`versioningEnabled: false`) -- Static websites rarely need versioning; enable if you need rollback capability
- **SSE-S3 encryption** -- Even public buckets benefit from server-side encryption at rest
- **Force destroy disabled** -- Prevents accidental bucket deletion

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region for the bucket (e.g., `us-east-1`) | Your deployment region |

## Related Presets

- **01-private-encrypted** -- Use instead for any bucket that should not be publicly accessible
