---
title: "Presets"
description: "Ready-to-deploy configuration presets for Subnet"
type: "preset-list"
componentSlug: "subnet"
componentTitle: "Subnet"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-general-purpose"
    rank: "01"
    title: "General-Purpose Subnet"
    excerpt: "This preset creates a general-purpose Azure Subnet with a /24 CIDR block (254 usable IPs) and common service endpoints for Storage, Key Vault, and SQL. This is the standard subnet configuration for..."
  - slug: "02-delegated-postgresql"
    rank: "02"
    title: "PostgreSQL Delegated Subnet"
    excerpt: "This preset creates an Azure Subnet delegated to PostgreSQL Flexible Server. Delegation grants the PostgreSQL service permission to inject server instances directly into the subnet, enabling..."
  - slug: "03-delegated-container-apps"
    rank: "03"
    title: "Delegated Container Apps Subnet"
    excerpt: "This preset creates a subnet delegated to Azure Container App Environments (`Microsoft.App/environments`). The /21 prefix provides 2,048 IP addresses — the minimum recommended for Container Apps..."
---

# Subnet Presets

Ready-to-deploy configuration presets for Subnet. Each preset is a complete manifest you can copy, customize, and deploy.
