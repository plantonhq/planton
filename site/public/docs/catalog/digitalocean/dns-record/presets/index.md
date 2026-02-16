---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Record"
type: "preset-list"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-a-record"
    rank: "01"
    title: "A Record"
    excerpt: "This preset creates a standard A record that points a hostname to an IPv4 address. Use for the root domain or any subdomain that should resolve to a Droplet, load balancer, or other IP. TTL is set to..."
  - slug: "02-cname-record"
    rank: "02"
    title: "CNAME Record"
    excerpt: "This preset creates a CNAME record that aliases a subdomain to another hostname. Common for `www` pointing to the root domain or a CDN, or for subdomains pointing to external services. The target..."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
