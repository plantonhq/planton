---
title: "Private Encrypted Bucket"
description: "This preset creates a private S3 bucket with Block Public Access enabled, SSE-S3 encryption, and versioning turned on. This is the standard production bucket configuration that protects against..."
type: "preset"
rank: "01"
presetSlug: "01-private-encrypted"
componentSlug: "s3-bucket"
componentTitle: "S3 Bucket"
provider: "aws"
icon: "package"
order: 1
---

# Private Encrypted Bucket

This preset creates a private S3 bucket with Block Public Access enabled, SSE-S3 encryption, and versioning turned on. This is the standard production bucket configuration that protects against accidental public exposure, data loss, and unauthorized access.

## When to Use

- Application data storage (uploads, media, documents, backups)
- CI/CD artifact storage (build outputs, deployment packages)
- Any S3 bucket that should not be publicly accessible

## Key Configuration Choices

- **Private** (`isPublic: false`) -- S3 Block Public Access is enabled; no public access even if bucket policies are misconfigured
- **Versioning enabled** (`versioningEnabled: true`) -- Protects against accidental deletions and overwrites; all versions are retained
- **SSE-S3 encryption** (`encryptionType: ENCRYPTION_TYPE_SSE_S3`) -- Server-side encryption with AES-256; free and automatic
- **Force destroy disabled** (`forceDestroy: false`) -- Bucket cannot be deleted while it contains objects

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<aws-region>` | AWS region for the bucket (e.g., `us-east-1`) | Your deployment region |

## Related Presets

- **02-public-static-website** -- Use instead for static websites or public asset hosting
