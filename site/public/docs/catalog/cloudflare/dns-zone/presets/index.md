---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Zone"
type: "preset-list"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-free-plan"
    rank: "01"
    title: "Free Plan Zone"
    excerpt: "Creates a Cloudflare DNS zone with no inline records, using the free plan. Ideal when you want to manage DNS records separately via CloudflareDnsRecord resources. Requires zone_name and account_id."
---

# DNS Zone Presets

Ready-to-deploy configuration presets for DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
