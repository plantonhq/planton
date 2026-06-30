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
  - slug: "01-basic-zone"
    rank: "01"
    title: "Basic Zone"
    excerpt: "Creates a Cloudflare DNS zone with no inline records. Ideal when you want to manage DNS records separately via CloudflareDnsRecord resources. Requires only `zoneName` and `accountId`; the zone..."
  - slug: "02-dnssec-signed"
    rank: "02"
    title: "DNSSEC-Signed Zone"
    excerpt: "Creates a zone with DNSSEC enabled. Cloudflare signs the zone, and the DS record material (digest, key tag, algorithm, and the full DS record) is published as stack outputs for you to enter at your..."
---

# DNS Zone Presets

Ready-to-deploy configuration presets for DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
