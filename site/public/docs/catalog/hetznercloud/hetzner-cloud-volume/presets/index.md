---
title: "Presets"
description: "Ready-to-deploy configuration presets for Hetzner Cloud Volume"
type: "preset-list"
componentSlug: "hetzner-cloud-volume"
componentTitle: "Hetzner Cloud Volume"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-attached-ext4"
    rank: "01"
    title: "Attached ext4 Volume"
    excerpt: "This preset creates a general-purpose Hetzner Cloud block storage volume, formatted with ext4 and attached to an existing server with automount enabled. It provisions an `hcloud_volume` resource and..."
  - slug: "02-database-storage"
    rank: "02"
    title: "Database Storage Volume"
    excerpt: "This preset creates a production-grade Hetzner Cloud block storage volume optimized for database workloads. It uses the XFS filesystem for high-throughput sequential writes, enables delete protection..."
  - slug: "03-unattached-reserve"
    rank: "03"
    title: "Unattached Reserve Volume"
    excerpt: "This preset creates a Hetzner Cloud block storage volume that is formatted and ready to use but not attached to any server. It provisions a single `hcloud_volume` resource with delete protection..."
---

# Hetzner Cloud Volume Presets

Ready-to-deploy configuration presets for Hetzner Cloud Volume. Each preset is a complete manifest you can copy, customize, and deploy.
