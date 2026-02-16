---
title: "Presets"
description: "Ready-to-deploy configuration presets for Container Registry"
type: "preset-list"
componentSlug: "container-registry"
componentTitle: "Container Registry"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Container Registry"
    excerpt: "This preset creates an Azure Container Registry with Standard SKU and admin user disabled. Standard tier provides 100 GB storage, enhanced throughput for image pulls, and webhook support --..."
  - slug: "02-premium-geo-replicated"
    rank: "02"
    title: "Premium Container Registry with Geo-Replication"
    excerpt: "This preset creates an Azure Container Registry with Premium SKU and geo-replication to a secondary region. Premium tier provides 500 GB storage, geo-replication for multi-region image distribution,..."
---

# Container Registry Presets

Ready-to-deploy configuration presets for Container Registry. Each preset is a complete manifest you can copy, customize, and deploy.
