---
title: "Presets"
description: "Ready-to-deploy configuration presets for Block Volume"
type: "preset-list"
componentSlug: "block-volume"
componentTitle: "Block Volume"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-balanced-with-backup"
    rank: "01"
    title: "Balanced with Backup"
    excerpt: "This preset creates a 100 GB OCI Block Volume at the Balanced performance tier (10 VPUs/GB) with an assigned backup policy and a detached-volume autotune policy. This is the standard configuration..."
  - slug: "02-high-performance-encrypted"
    rank: "02"
    title: "High Performance Encrypted"
    excerpt: "This preset creates a 200 GB OCI Block Volume at the Higher Performance tier (20 VPUs/GB) with customer-managed KMS encryption, performance-based autotune that scales up to Ultra High Performance..."
---

# Block Volume Presets

Ready-to-deploy configuration presets for Block Volume. Each preset is a complete manifest you can copy, customize, and deploy.
