---
title: "Presets"
description: "Ready-to-deploy configuration presets for Filestore Instance"
type: "preset-list"
componentSlug: "filestore-instance"
componentTitle: "Filestore Instance"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-dev-basic"
    rank: "01"
    title: "Preset: Dev Basic"
    excerpt: "**Tier**: BASIC_SSD (single-zone SSD) **Use case**: Development, testing, CI/CD pipelines"
  - slug: "02-production-enterprise"
    rank: "02"
    title: "Preset: Production Enterprise"
    excerpt: "**Tier**: ENTERPRISE (regional HA) **Use case**: Production workloads requiring high availability and security"
  - slug: "03-high-performance-zonal"
    rank: "03"
    title: "Preset: High Performance Zonal"
    excerpt: "**Tier**: ZONAL (modern SSD with IOPS tuning) **Use case**: Performance-sensitive workloads requiring high throughput"
---

# Filestore Instance Presets

Ready-to-deploy configuration presets for Filestore Instance. Each preset is a complete manifest you can copy, customize, and deploy.
