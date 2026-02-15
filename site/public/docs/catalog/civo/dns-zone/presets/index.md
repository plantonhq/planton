---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Zone"
type: "preset-list"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "civo"
icon: "package"
order: 200
presets:
  - slug: "01-simple-website"
    rank: "01"
    title: "Simple Website DNS Zone"
    excerpt: "This preset creates a DNS zone with an apex A record pointing to a server IP and a `www` CNAME aliasing to the apex. This is the most common DNS configuration for a website: visitors reach the site..."
  - slug: "02-with-email"
    rank: "02"
    title: "Website + Email DNS Zone"
    excerpt: "This preset creates a DNS zone for a domain that serves both a website and receives email. Includes an apex A record, www CNAME, MX record for mail routing, and a TXT record for SPF email..."
---

# DNS Zone Presets

Ready-to-deploy configuration presets for DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
