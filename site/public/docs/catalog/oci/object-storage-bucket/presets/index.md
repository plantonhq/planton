---
title: "Presets"
description: "Ready-to-deploy configuration presets for Object Storage Bucket"
type: "preset-list"
componentSlug: "object-storage-bucket"
componentTitle: "Object Storage Bucket"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-private-versioned"
    rank: "01"
    title: "Private Versioned"
    excerpt: "This preset creates a private Object Storage bucket with versioning enabled, KMS encryption, auto-tiering for cost optimization, and lifecycle rules to archive old object versions and clean up..."
  - slug: "02-archive-storage"
    rank: "02"
    title: "Archive Storage"
    excerpt: "This preset creates an archive-tier Object Storage bucket with a 7-year retention rule for compliance data. Archive storage offers the lowest per-GB cost in OCI Object Storage, suitable for data that..."
  - slug: "03-public-read"
    rank: "03"
    title: "Public Read"
    excerpt: "This preset creates a public-read Object Storage bucket for serving static assets directly to end users. Individual objects can be accessed via URL without authentication, but the bucket contents..."
---

# Object Storage Bucket Presets

Ready-to-deploy configuration presets for Object Storage Bucket. Each preset is a complete manifest you can copy, customize, and deploy.
