---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Record"
type: "preset-list"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "oci"
icon: "package"
order: 200
presets:
  - slug: "01-a-record"
    rank: "01"
    title: "A Record"
    excerpt: "This preset creates a DNS A record that maps a fully qualified domain name to an IPv4 address. A records are the most fundamental DNS record type and the starting point for pointing a domain at any..."
  - slug: "02-cname-alias"
    rank: "02"
    title: "CNAME Alias"
    excerpt: "This preset creates a DNS CNAME record that aliases one domain name to another. CNAME records are the standard mechanism for pointing subdomains at canonical hostnames, external services, or CDN..."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
