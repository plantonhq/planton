---
title: "Presets"
description: "Ready-to-deploy configuration presets for User Assigned Identity"
type: "preset-list"
componentSlug: "user-assigned-identity"
componentTitle: "User Assigned Identity"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Managed Identity"
    excerpt: "This preset creates an Azure User-Assigned Managed Identity with a single RBAC role assignment. This is the most common pattern -- a single identity with one targeted permission grant, used for..."
  - slug: "02-multi-role"
    rank: "02"
    title: "Multi-Role Application Identity"
    excerpt: "This preset creates an Azure User-Assigned Managed Identity with role assignments for Key Vault, Storage, and Container Registry access. This is the standard pattern for production application..."
---

# User Assigned Identity Presets

Ready-to-deploy configuration presets for User Assigned Identity. Each preset is a complete manifest you can copy, customize, and deploy.
