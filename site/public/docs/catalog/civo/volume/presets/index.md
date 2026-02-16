---
title: "Presets"
description: "Ready-to-deploy configuration presets for Volume"
type: "preset-list"
componentSlug: "volume"
componentTitle: "Volume"
provider: "civo"
icon: "package"
order: 200
presets:
  - slug: "01-ext4-general"
    rank: "01"
    title: "General-Purpose ext4 Volume"
    excerpt: "This preset creates a 50 GiB block storage volume pre-formatted with ext4. ext4 is the most widely supported Linux filesystem and the best default for general application data, logs, and file storage."
  - slug: "02-xfs-database"
    rank: "02"
    title: "XFS Database Volume"
    excerpt: "This preset creates a 100 GiB block storage volume pre-formatted with XFS, optimized for database workloads. XFS excels at large sequential writes and parallel I/O, making it the preferred filesystem..."
---

# Volume Presets

Ready-to-deploy configuration presets for Volume. Each preset is a complete manifest you can copy, customize, and deploy.
