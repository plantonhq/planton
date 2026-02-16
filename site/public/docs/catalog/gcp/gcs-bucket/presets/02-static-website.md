---
title: "Static Website Bucket"
description: "This preset creates a GCS bucket configured for static website hosting with public read access, CORS rules for browser access, and website routing (index.html / 404.html). For production websites,..."
type: "preset"
rank: "02"
presetSlug: "02-static-website"
componentSlug: "gcs-bucket"
componentTitle: "GCS Bucket"
provider: "gcp"
icon: "package"
order: 2
---

# Static Website Bucket

This preset creates a GCS bucket configured for static website hosting with public read access, CORS rules for browser access, and website routing (index.html / 404.html). For production websites, pair this with a Cloud CDN preset for HTTPS and global distribution.

## When to Use

- Static websites (HTML, CSS, JS) served directly from GCS
- Development or internal sites where direct GCS hosting is sufficient
- Staging environments for static site generators (Hugo, Gatsby, Next.js export)

## Key Configuration Choices

- **Website configuration** -- `index.html` as main page, `404.html` for not-found responses
- **Public read access** (`allUsers` with `objectViewer` role) -- all objects are publicly readable
- **CORS enabled** -- allows GET/HEAD from any origin (adjust `origins` for production)
- **Uniform bucket-level access** -- IAM-only, consistent with platform convention
- **No versioning** -- static site content is typically regenerated, not version-tracked

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<your-bucket-name>` | Globally unique bucket name | Choose a unique name |
| `<gcp-region>` | Bucket location (e.g., `us-central1`) | Your deployment region |

## Related Presets

- **01-private-standard** -- Use for private data storage with versioning and public access prevention
