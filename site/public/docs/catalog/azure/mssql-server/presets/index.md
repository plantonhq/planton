---
title: "Presets"
description: "Ready-to-deploy configuration presets for MSSQL Server"
type: "preset-list"
componentSlug: "mssql-server"
componentTitle: "MSSQL Server"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Azure SQL Database"
    excerpt: "This preset creates an Azure SQL logical server with a Standard-tier (S0) database. The S0 SKU provides 10 DTUs of compute, 250 GB max storage, and geo-redundant backups -- a cost-effective entry..."
  - slug: "02-business-critical"
    rank: "02"
    title: "Business Critical Azure SQL Database"
    excerpt: "This preset creates an Azure SQL logical server with a Business Critical (BC_Gen5_2) vCore-based database. Business Critical tier provides local SSD storage for low-latency IO, zone-redundant..."
---

# MSSQL Server Presets

Ready-to-deploy configuration presets for MSSQL Server. Each preset is a complete manifest you can copy, customize, and deploy.
