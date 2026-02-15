---
title: "Presets"
description: "Ready-to-deploy configuration presets for Storage Account"
type: "preset-list"
componentSlug: "storage-account"
componentTitle: "Storage Account"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-general-purpose-v2"
    rank: "01"
    title: "General-Purpose v2 Storage Account"
    excerpt: "This preset creates an Azure Storage Account with General-Purpose v2 (StorageV2), Standard tier, and locally redundant storage. It includes a default-deny network posture, blob versioning, 7-day soft..."
  - slug: "02-production-geo-redundant"
    rank: "02"
    title: "Production Geo-Redundant Storage Account"
    excerpt: "This preset creates an Azure Storage Account with geo-redundant storage (GRS), 30-day soft delete retention, IP-restricted network access, and blob versioning. GRS replicates data to Azure's paired..."
---

# Storage Account Presets

Ready-to-deploy configuration presets for Storage Account. Each preset is a complete manifest you can copy, customize, and deploy.
