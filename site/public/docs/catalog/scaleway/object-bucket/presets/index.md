---
title: "Presets"
description: "Ready-to-deploy configuration presets for Object Bucket"
type: "preset-list"
componentSlug: "object-bucket"
componentTitle: "Object Bucket"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-private-bucket"
    rank: "01"
    title: "Private Object Bucket"
    excerpt: "This preset creates a minimal Scaleway Object Storage bucket with default settings. The bucket is private (no public access), has no versioning, and cannot be destroyed while it contains objects...."
  - slug: "02-versioned-lifecycle"
    rank: "02"
    title: "Versioned Bucket with Lifecycle Rules"
    excerpt: "This preset creates a Scaleway Object Storage bucket with versioning enabled and a lifecycle rule that transitions objects to Glacier cold storage after 90 days. Incomplete multipart uploads are..."
---

# Object Bucket Presets

Ready-to-deploy configuration presets for Object Bucket. Each preset is a complete manifest you can copy, customize, and deploy.
