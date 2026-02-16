---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Record"
type: "preset-list"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-a-record"
    rank: "01"
    title: "A Record"
    excerpt: "This preset creates a standalone DNS A record in Designate, mapping a hostname to an IPv4 address. Use standalone records (instead of inline records in `OpenStackDnsZone`) when individual records..."
  - slug: "02-cname-record"
    rank: "02"
    title: "CNAME Record"
    excerpt: "This preset creates a standalone DNS CNAME record in Designate, aliasing one hostname to another. The target can be within the same zone or an external domain. CNAME records are commonly used for..."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
