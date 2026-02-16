---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Zone"
type: "preset-list"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-public-zone"
    rank: "01"
    title: "Public DNS Zone"
    excerpt: "This preset creates a Cloud DNS managed zone with IAM permissions granted to service accounts for cert-manager and external-dns. DNS records are managed separately via `GcpDnsRecord` resources,..."
---

# DNS Zone Presets

Ready-to-deploy configuration presets for DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
