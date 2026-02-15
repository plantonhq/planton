---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Record"
type: "preset-list"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-a-record"
    rank: "01"
    title: "A Record"
    excerpt: "This preset creates a standard DNS A record pointing a domain name to an IPv4 address. This is the most common DNS record type, used for mapping hostnames to IP addresses."
  - slug: "02-cname-record"
    rank: "02"
    title: "CNAME Record"
    excerpt: "This preset creates a DNS CNAME record that aliases one hostname to another. CNAME records are used when you want a subdomain to resolve to the same address as another domain without hardcoding an IP..."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
