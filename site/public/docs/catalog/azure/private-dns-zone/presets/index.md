---
title: "Presets"
description: "Ready-to-deploy configuration presets for Private DNS Zone"
type: "preset-list"
componentSlug: "private-dns-zone"
componentTitle: "Private DNS Zone"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard"
    rank: "01"
    title: "Standard Private DNS Zone"
    excerpt: "This preset creates an Azure Private DNS Zone for PostgreSQL Flexible Server Private Link resolution, linked to a Virtual Network. When a private endpoint is created for a PostgreSQL server, its..."
  - slug: "02-mysql"
    rank: "02"
    title: "MySQL Private DNS Zone"
    excerpt: "This preset creates a Private DNS Zone for Azure Database for MySQL Flexible Server Private Link connectivity. The zone name `privatelink.mysql.database.azure.com` is required by Azure for DNS..."
  - slug: "03-sql-server"
    rank: "03"
    title: "SQL Server Private DNS Zone"
    excerpt: "This preset creates a Private DNS Zone for Azure SQL Server (MSSQL) Private Endpoint connectivity. The zone name `privatelink.database.windows.net` is required by Azure for DNS resolution of SQL..."
  - slug: "04-storage"
    rank: "04"
    title: "Storage Account Private DNS Zone (Blob)"
    excerpt: "This preset creates a Private DNS Zone for Azure Storage Account Blob Private Endpoint connectivity. The zone name `privatelink.blob.core.windows.net` is required by Azure for DNS resolution of..."
---

# Private DNS Zone Presets

Ready-to-deploy configuration presets for Private DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
