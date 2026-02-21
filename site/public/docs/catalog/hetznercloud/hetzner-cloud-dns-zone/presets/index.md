---
title: "Presets"
description: "Ready-to-deploy configuration presets for Hetzner Cloud DNS Zone"
type: "preset-list"
componentSlug: "hetzner-cloud-dns-zone"
componentTitle: "Hetzner Cloud DNS Zone"
provider: "hetznercloud"
icon: "package"
order: 200
presets:
  - slug: "01-web-domain"
    rank: "01"
    title: "Web Domain"
    excerpt: "This preset creates a primary DNS zone on Hetzner Cloud's authoritative nameservers for a production website domain. It includes the standard record set a real website needs: an apex A record, a www..."
  - slug: "02-secondary-zone"
    rank: "02"
    title: "Secondary Zone"
    excerpt: "This preset creates a DNS zone in secondary mode where Hetzner Cloud acts as a secondary (slave) nameserver, synchronizing records from your external primary nameserver via zone transfer (AXFR/IXFR)...."
  - slug: "03-simple-zone"
    rank: "03"
    title: "Simple Zone"
    excerpt: "This preset creates a minimal primary DNS zone with just an apex A record and a www CNAME -- the bare minimum to make a domain resolve to a server. No email records, no security records, no delete..."
---

# Hetzner Cloud DNS Zone Presets

Ready-to-deploy configuration presets for Hetzner Cloud DNS Zone. Each preset is a complete manifest you can copy, customize, and deploy.
