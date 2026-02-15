---
title: "Presets"
description: "Ready-to-deploy configuration presets for PostgreSQL Flexible Server"
type: "preset-list"
componentSlug: "postgresql-flexible-server"
componentTitle: "PostgreSQL Flexible Server"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-production-public"
    rank: "01"
    title: "Production PostgreSQL with Public Access"
    excerpt: "This preset creates an Azure Database for PostgreSQL Flexible Server with General Purpose compute, 32 GB storage, auto-grow disabled, and public network access controlled by firewall rules. A starter..."
  - slug: "02-production-vnet"
    rank: "02"
    title: "Production PostgreSQL with VNet Integration"
    excerpt: "This preset creates an Azure Database for PostgreSQL Flexible Server injected into a virtual network subnet. Public network access is automatically disabled when `delegatedSubnetId` is set, ensuring..."
---

# PostgreSQL Flexible Server Presets

Ready-to-deploy configuration presets for PostgreSQL Flexible Server. Each preset is a complete manifest you can copy, customize, and deploy.
