---
title: "Presets"
description: "Ready-to-deploy configuration presets for R2 Bucket"
type: "preset-list"
componentSlug: "r2-bucket"
componentTitle: "R2 Bucket"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-private"
    rank: "01"
    title: "Private R2 Bucket"
    excerpt: "Creates a private R2 bucket with no public access. Data is accessible only via Workers, API tokens, or the Cloudflare Dashboard. Use for backups, logs, or any object storage that must stay private."
  - slug: "02-public-cdn"
    rank: "02"
    title: "Public R2 Bucket with Custom Domain"
    excerpt: "Creates a public R2 bucket served via a custom domain (e.g., media.example.com). Combines public access with a branded CDN URL. Requires a Cloudflare DNS zone for the domain."
  - slug: "03-lifecycle-managed"
    rank: "03"
    title: "Private R2 Bucket with Lifecycle and Retention"
    excerpt: "A private bucket that manages its own data over time: it tiers objects to Infrequent Access storage, expires them after a year, cleans up stalled multipart uploads, and locks audit objects for a..."
---

# R2 Bucket Presets

Ready-to-deploy configuration presets for R2 Bucket. Each preset is a complete manifest you can copy, customize, and deploy.
