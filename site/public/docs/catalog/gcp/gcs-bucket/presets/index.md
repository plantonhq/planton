---
title: "Presets"
description: "Ready-to-deploy configuration presets for GCS Bucket"
type: "preset-list"
componentSlug: "gcs-bucket"
componentTitle: "GCS Bucket"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-private-standard"
    rank: "01"
    title: "Private Standard Bucket"
    excerpt: "This preset creates a private GCS bucket with uniform bucket-level access, versioning, public access prevention, and lifecycle rules to control version sprawl. It represents the standard..."
  - slug: "02-static-website"
    rank: "02"
    title: "Static Website Bucket"
    excerpt: "This preset creates a GCS bucket configured for static website hosting with public read access, CORS rules for browser access, and website routing (index.html / 404.html). For production websites,..."
---

# GCS Bucket Presets

Ready-to-deploy configuration presets for GCS Bucket. Each preset is a complete manifest you can copy, customize, and deploy.
