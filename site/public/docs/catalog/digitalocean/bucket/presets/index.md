---
title: "Presets"
description: "Ready-to-deploy configuration presets for Bucket"
type: "preset-list"
componentSlug: "bucket"
componentTitle: "Bucket"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-private"
    rank: "01"
    title: "Private Bucket with Versioning"
    excerpt: "This preset creates a private DigitalOcean Spaces bucket with versioning enabled. Objects are not publicly accessible; use signed URLs or IAM for access. Versioning protects against accidental..."
  - slug: "02-public-static-website"
    rank: "02"
    title: "Public Static Website Bucket"
    excerpt: "This preset creates a public-read DigitalOcean Spaces bucket suitable for hosting static websites, CDN origins, or publicly served assets (JS, CSS, images). Objects are readable by anyone with the..."
---

# Bucket Presets

Ready-to-deploy configuration presets for Bucket. Each preset is a complete manifest you can copy, customize, and deploy.
