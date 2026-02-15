---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Zone"
type: "preset-list"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-root-zone"
    rank: "01"
    title: "Root DNS Zone"
    excerpt: "This preset creates a Scaleway DNS zone for a root domain with no inline records. Records are managed separately as standalone `ScalewayDnsRecord` resources, which is the recommended pattern for..."
---

# DNS Zone Presets

Ready-to-deploy configuration presets for DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
