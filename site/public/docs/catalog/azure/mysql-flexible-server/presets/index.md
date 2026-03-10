---
title: "Presets"
description: "Ready-to-deploy configuration presets for MySQL Flexible Server"
type: "preset-list"
componentSlug: "mysql-flexible-server"
componentTitle: "MySQL Flexible Server"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-production-public"
    rank: "01"
    title: "Production MySQL with Public Access"
    excerpt: "This preset creates an Azure Database for MySQL Flexible Server with General Purpose compute, 32 GB storage with auto-grow enabled, and public network access controlled by firewall rules. A starter..."
  - slug: "02-production-vnet"
    rank: "02"
    title: "Production MySQL with VNet Integration"
    excerpt: "This preset creates an Azure Database for MySQL Flexible Server injected into a virtual network subnet. Public network access is automatically disabled when `delegatedSubnetId` is set, ensuring the..."
  - slug: "03-development"
    rank: "03"
    title: "Development MySQL Flexible Server"
    excerpt: "This preset creates a minimal Azure Database for MySQL Flexible Server using the Burstable B1ms SKU (~$12/month) with 20 GB storage, no high availability, no geo-redundant backup, and public access..."
  - slug: "04-ha-zone-redundant"
    rank: "04"
    title: "HA Zone-Redundant MySQL Flexible Server"
    excerpt: "This preset creates an Azure Database for MySQL Flexible Server with Zone-Redundant high availability, geo-redundant backup, and a General Purpose D4ds_v4 SKU (4 vCPU, 16 GiB RAM). The primary runs..."
---

# MySQL Flexible Server Presets

Ready-to-deploy configuration presets for MySQL Flexible Server. Each preset is a complete manifest you can copy, customize, and deploy.
