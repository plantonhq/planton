---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Record"
type: "preset-list"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "civo"
icon: "package"
order: 200
presets:
  - slug: "01-a-record"
    rank: "01"
    title: "A Record"
    excerpt: "This preset creates a standalone A record pointing a subdomain to an IPv4 address. This is the most common DNS record type, used for mapping hostnames to servers. Use standalone DNS records when..."
  - slug: "02-cname-record"
    rank: "02"
    title: "CNAME Record"
    excerpt: "This preset creates a standalone CNAME record aliasing a subdomain to another hostname. CNAME records are the standard way to point subdomains to load balancers, CDNs, or other services that expose a..."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
