---
title: "Presets"
description: "Ready-to-deploy configuration presets for Volume"
type: "preset-list"
componentSlug: "volume"
componentTitle: "Volume"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-general-purpose-ext4"
    rank: "01"
    title: "General-Purpose ext4 Volume"
    excerpt: "This preset creates a DigitalOcean block storage volume pre-formatted with ext4, ready to attach and mount on a Droplet immediately. Suitable for application data, logs, media files, or any..."
  - slug: "02-database-xfs"
    rank: "02"
    title: "Database XFS Volume"
    excerpt: "This preset creates a DigitalOcean block storage volume pre-formatted with XFS, optimized for database workloads. XFS provides superior write performance for the sequential and random I/O patterns..."
---

# Volume Presets

Ready-to-deploy configuration presets for Volume. Each preset is a complete manifest you can copy, customize, and deploy.
