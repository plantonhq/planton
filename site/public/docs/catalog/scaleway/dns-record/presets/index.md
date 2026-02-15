---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Record"
type: "preset-list"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "scaleway"
icon: "package"
order: 200
presets:
  - slug: "01-a-record"
    rank: "01"
    title: "A Record"
    excerpt: "This preset creates a DNS A record that maps a hostname to an IPv4 address. This is the most common DNS record type, used to point domain names at servers, load balancers, or other IP-based endpoints."
  - slug: "02-cname-record"
    rank: "02"
    title: "CNAME Record"
    excerpt: "This preset creates a DNS CNAME record that maps a hostname to another hostname. CNAMEs are used to create aliases -- for example, pointing `www.example.com` to the canonical hostname of an..."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
