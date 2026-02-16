---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Zone"
type: "preset-list"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-primary-zone"
    rank: "01"
    title: "Primary DNS Zone"
    excerpt: "This preset creates a primary DNS zone in Designate. The zone is authoritative for the specified domain and can host A, AAAA, CNAME, MX, TXT, and other record types. Records can be added inline via..."
---

# DNS Zone Presets

Ready-to-deploy configuration presets for DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
