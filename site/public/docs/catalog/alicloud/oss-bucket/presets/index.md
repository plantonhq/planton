---
title: "Presets"
description: "Ready-to-deploy configuration presets for OSS Bucket"
type: "preset-list"
componentSlug: "oss-bucket"
componentTitle: "OSS Bucket"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-private-standard"
    rank: "01"
    title: "Private Standard Bucket"
    excerpt: "This preset creates a minimal private OSS bucket with default settings: Standard storage class, LRS redundancy, no versioning, no encryption, and no lifecycle rules. Ideal for getting started quickly..."
  - slug: "02-versioned-encrypted"
    rank: "02"
    title: "Versioned Encrypted Bucket"
    excerpt: "This preset creates a production-grade OSS bucket with zone-redundant storage (ZRS), object versioning, and AES256 server-side encryption at rest. Designed for workloads where data durability,..."
  - slug: "03-archive-lifecycle"
    rank: "03"
    title: "Archive Bucket with Lifecycle Rules"
    excerpt: "This preset creates a cost-optimized OSS bucket that automatically transitions objects through progressively cheaper storage tiers and expires them after one year. Versioning and encryption are..."
---

# OSS Bucket Presets

Ready-to-deploy configuration presets for OSS Bucket. Each preset is a complete manifest you can copy, customize, and deploy.
