---
title: "Presets"
description: "Ready-to-deploy configuration presets for NAS File System"
type: "preset-list"
componentSlug: "nas-file-system"
componentTitle: "NAS File System"
provider: "alicloud"
icon: "package"
order: 200
presets:
  - slug: "01-standard-nfs"
    rank: "01"
    title: "Standard NFS File System"
    excerpt: "This preset creates a minimal standard NAS file system with NFS protocol, Performance storage, and default VPC-wide access. The file system auto-scales capacity as data is written, with no..."
  - slug: "02-production-encrypted"
    rank: "02"
    title: "Production Encrypted NFS"
    excerpt: "This preset creates a production-grade NAS file system with NAS-managed encryption at rest and a custom access group restricting mount access to a specific application subnet. Root squashing is..."
---

# NAS File System Presets

Ready-to-deploy configuration presets for NAS File System. Each preset is a complete manifest you can copy, customize, and deploy.
