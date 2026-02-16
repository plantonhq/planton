---
title: "Presets"
description: "Ready-to-deploy configuration presets for Service Plan"
type: "preset-list"
componentSlug: "service-plan"
componentTitle: "Service Plan"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-linux-standard"
    rank: "01"
    title: "Linux Standard Plan"
    excerpt: "This preset creates an Azure App Service Plan on the Standard S1 tier with a single Linux worker. Standard tier provides auto-scaling up to 10 instances, staging slots, daily backups, and a 99.95%..."
  - slug: "02-linux-premium"
    rank: "02"
    title: "Linux Premium Plan with Zone Redundancy"
    excerpt: "This preset creates an Azure App Service Plan on the Premium v3 P1v3 tier with 3 Linux workers distributed across availability zones. Premium v3 provides faster processors, SSD storage, double the..."
---

# Service Plan Presets

Ready-to-deploy configuration presets for Service Plan. Each preset is a complete manifest you can copy, customize, and deploy.
