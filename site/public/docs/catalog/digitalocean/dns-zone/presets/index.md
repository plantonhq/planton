---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Zone"
type: "preset-list"
componentSlug: "dns-zone"
componentTitle: "DNS Zone"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-simple-website"
    rank: "01"
    title: "Simple Website Zone"
    excerpt: "This preset creates a DNS zone with a single A record pointing the root domain to your web server or load balancer IP. Minimal configuration for getting a domain live quickly. TTL is set to 1 hour..."
  - slug: "02-production-with-email"
    rank: "02"
    title: "Production Zone with Email Records"
    excerpt: "This preset creates a DNS zone with A, MX, and TXT (SPF) records for a production website that also receives email. The A record points the domain to your web server; MX directs mail to your mail..."
---

# DNS Zone Presets

Ready-to-deploy configuration presets for DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
