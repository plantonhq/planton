---
title: "GCS Static Site CDN"
description: "This preset creates a Cloud CDN backed by a Google Cloud Storage bucket for serving static websites, single-page applications, or media assets. It caches all static content types automatically with a..."
type: "preset"
rank: "01"
presetSlug: "01-gcs-static-site"
componentSlug: "cloud-cdn"
componentTitle: "Cloud CDN"
provider: "gcp"
icon: "package"
order: 1
---

# GCS Static Site CDN

This preset creates a Cloud CDN backed by a Google Cloud Storage bucket for serving static websites, single-page applications, or media assets. It caches all static content types automatically with a 1-hour default TTL and 1-day maximum.

## When to Use

- Static websites (HTML, CSS, JS, images) hosted in GCS
- Single-page applications (React, Vue, Angular) with a GCS backend
- Media or file download sites where content changes infrequently

## Key Configuration Choices

- **GCS bucket backend** (`gcsBucket`) -- serves content directly from a Cloud Storage bucket
- **Cache all static** (`cacheMode: CACHE_ALL_STATIC`) -- automatically caches common static file types and respects Cache-Control headers for dynamic content
- **1-hour default TTL** (`defaultTtlSeconds: 3600`) -- for content without explicit Cache-Control headers
- **1-day max TTL** (`maxTtlSeconds: 86400`) -- hard ceiling that overrides origin Cache-Control
- **Negative caching** (`enableNegativeCaching: true`) -- caches 404 responses to reduce origin load during missing-file requests

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID | `GcpProject` outputs |
| `<gcs-bucket-name>` | Name of the GCS bucket serving static content | `GcpGcsBucket` outputs or GCS console |

## Related Presets

- **02-cloud-run-backend** -- Use when the origin is a Cloud Run service instead of a GCS bucket
