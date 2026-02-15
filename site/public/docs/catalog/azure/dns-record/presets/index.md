---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Record"
type: "preset-list"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-a-record"
    rank: "01"
    title: "A Record"
    excerpt: "This preset creates a DNS A record mapping a subdomain to an IPv4 address within an Azure DNS Zone. A records are the most fundamental DNS record type, used to point domain names to the IP addresses..."
  - slug: "02-cname-record"
    rank: "02"
    title: "CNAME Record"
    excerpt: "This preset creates a DNS CNAME record that aliases a subdomain to another domain name. CNAME records are used when you want a domain name to resolve to the same IP as another hostname, commonly for..."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
