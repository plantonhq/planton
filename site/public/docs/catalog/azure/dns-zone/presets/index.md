---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Zone"
type: "preset-list"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-public-zone"
    rank: "01"
    title: "Public DNS Zone"
    excerpt: "This preset creates an Azure DNS Zone for hosting public DNS records for a domain. The zone is created empty -- DNS records are managed separately via `AzureDnsRecord` resources or added inline via..."
---

# DNS Zone Presets

Ready-to-deploy configuration presets for DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
