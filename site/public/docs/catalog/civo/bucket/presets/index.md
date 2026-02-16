---
title: "Presets"
description: "Ready-to-deploy configuration presets for Bucket"
type: "preset-list"
componentSlug: "bucket"
componentTitle: "Bucket"
provider: "civo"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Object Storage Bucket"
    excerpt: "This preset creates a Civo object storage bucket with versioning disabled. This is the most common configuration for application assets, static files, and general-purpose storage where previous..."
  - slug: "02-versioned-backup"
    rank: "02"
    title: "Versioned Backup Bucket"
    excerpt: "This preset creates a Civo object storage bucket with versioning enabled, retaining all previous versions of every object. Suitable for backup repositories, compliance archives, and any scenario..."
---

# Bucket Presets

Ready-to-deploy configuration presets for Bucket. Each preset is a complete manifest you can copy, customize, and deploy.
