---
title: "Presets"
description: "Ready-to-deploy configuration presets for S3 Bucket"
type: "preset-list"
componentSlug: "s3-bucket"
componentTitle: "S3 Bucket"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-private-encrypted"
    rank: "01"
    title: "Private Encrypted Bucket"
    excerpt: "This preset creates a private S3 bucket with Block Public Access enabled, SSE-S3 encryption, and versioning turned on. This is the standard production bucket configuration that protects against..."
  - slug: "02-public-static-website"
    rank: "02"
    title: "Public Static Website Bucket"
    excerpt: "This preset creates a publicly accessible S3 bucket configured for static website hosting with CORS enabled for cross-origin asset loading. It allows GET and HEAD requests from any origin, making it..."
---

# S3 Bucket Presets

Ready-to-deploy configuration presets for S3 Bucket. Each preset is a complete manifest you can copy, customize, and deploy.
