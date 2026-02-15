---
title: "Presets"
description: "Ready-to-deploy configuration presets for Key Vault"
type: "preset-list"
componentSlug: "key-vault"
componentTitle: "Key Vault"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard-rbac"
    rank: "01"
    title: "Standard Key Vault with RBAC"
    excerpt: "This preset creates an Azure Key Vault with Standard SKU, Azure RBAC authorization, purge protection, and 90-day soft delete retention. This is the recommended configuration for most production..."
  - slug: "02-premium-network-restricted"
    rank: "02"
    title: "Premium Key Vault with Network Restrictions"
    excerpt: "This preset creates an Azure Key Vault with Premium SKU (HSM-backed keys), Azure RBAC, and network access control. Network ACLs restrict access to specific IP ranges and VNet subnets, with a..."
---

# Key Vault Presets

Ready-to-deploy configuration presets for Key Vault. Each preset is a complete manifest you can copy, customize, and deploy.
