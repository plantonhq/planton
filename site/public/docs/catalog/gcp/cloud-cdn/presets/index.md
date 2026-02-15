---
title: "Presets"
description: "Ready-to-deploy configuration presets for Cloud CDN"
type: "preset-list"
componentSlug: "cloud-cdn"
componentTitle: "Cloud CDN"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-gcs-static-site"
    rank: "01"
    title: "GCS Static Site CDN"
    excerpt: "This preset creates a Cloud CDN backed by a Google Cloud Storage bucket for serving static websites, single-page applications, or media assets. It caches all static content types automatically with a..."
  - slug: "02-cloud-run-backend"
    rank: "02"
    title: "Cloud Run Backend CDN"
    excerpt: "This preset creates a Cloud CDN backed by a Cloud Run service, with caching controlled by the origin's Cache-Control headers. It includes a custom domain with HTTPS redirect, making it suitable for..."
---

# Cloud CDN Presets

Ready-to-deploy configuration presets for Cloud CDN. Each preset is a complete manifest you can copy, customize, and deploy.
