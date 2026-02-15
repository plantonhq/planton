---
title: "Presets"
description: "Ready-to-deploy configuration presets for Private Endpoint"
type: "preset-list"
componentSlug: "private-endpoint"
componentTitle: "Private Endpoint"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-sql-server"
    rank: "01"
    title: "Private Endpoint for Azure SQL Database"
    excerpt: "This preset creates an Azure Private Endpoint that connects an Azure SQL Database server to a VNet subnet via Private Link. It includes a DNS zone group registration so that the SQL server's FQDN..."
  - slug: "02-storage-account"
    rank: "02"
    title: "Private Endpoint for Azure Blob Storage"
    excerpt: "This preset creates an Azure Private Endpoint that connects an Azure Storage Account's blob service to a VNet subnet via Private Link. It includes a DNS zone group registration so that the storage..."
---

# Private Endpoint Presets

Ready-to-deploy configuration presets for Private Endpoint. Each preset is a complete manifest you can copy, customize, and deploy.
