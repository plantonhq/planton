---
title: "Presets"
description: "Ready-to-deploy configuration presets for FSx for OpenZFS"
type: "preset-list"
componentSlug: "fsx-for-openzfs"
componentTitle: "FSx for OpenZFS"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-single-az-development"
    rank: "01"
    title: "Preset: Single-AZ Development"
    excerpt: "**Use case**: Development and testing environments where cost efficiency matters more than high availability or performance."
  - slug: "02-single-az-production"
    rank: "02"
    title: "Preset: Single-AZ Production"
    excerpt: "**Use case**: Production workloads in a single availability zone where NFS performance, data compression, encryption, and daily backups are required."
  - slug: "03-multi-az-high-availability"
    rank: "03"
    title: "Preset: Multi-AZ High Availability"
    excerpt: "**Use case**: Mission-critical production workloads requiring automatic failover across availability zones, provisioned IOPS, storage quotas, and extended backup retention."
---

# FSx for OpenZFS Presets

Ready-to-deploy configuration presets for FSx for OpenZFS. Each preset is a complete manifest you can copy, customize, and deploy.
