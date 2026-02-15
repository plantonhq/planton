---
title: "Presets"
description: "Ready-to-deploy configuration presets for AzureRedisCache: Deployment Landscape & Design Research"
type: "preset-list"
componentSlug: "azurerediscache-deployment-landscape-design-research"
componentTitle: "AzureRedisCache: Deployment Landscape & Design Research"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Redis Cache"
    excerpt: "This preset creates an Azure Cache for Redis with Standard tier (primary + replica), 1 GB cache size (C1), Redis 6, and TLS 1.2 enforcement. The Standard tier provides a replicated two-node cache..."
  - slug: "02-premium-vnet"
    rank: "02"
    title: "Premium Redis Cache with VNet Injection"
    excerpt: "This preset creates an Azure Cache for Redis with Premium tier injected into a virtual network subnet. VNet injection provides private IP addressing and network isolation -- the cache is not..."
---

# AzureRedisCache: Deployment Landscape & Design Research Presets

Ready-to-deploy configuration presets for AzureRedisCache: Deployment Landscape & Design Research. Each preset is a complete manifest you can copy, customize, and deploy.
