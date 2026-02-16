---
title: "Presets"
description: "Ready-to-deploy configuration presets for DNS Record"
type: "preset-list"
componentSlug: "dns-record"
componentTitle: "DNS Record"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-proxied-a-record"
    rank: "01"
    title: "Proxied A Record"
    excerpt: "Creates an A record with Cloudflare proxy (orange cloud) enabled. Traffic flows through Cloudflare's CDN and DDoS protection, hiding your origin IP. Use for web-facing hostnames where you want..."
  - slug: "02-mx-email"
    rank: "02"
    title: "MX Record for Email"
    excerpt: "Creates an MX record for email delivery. Priority is required; MX records cannot be proxied. Use for configuring mail servers (Google Workspace, Microsoft 365, custom mail) for your domain."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
