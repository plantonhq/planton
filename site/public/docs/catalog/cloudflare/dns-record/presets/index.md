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
  - slug: "03-srv-service"
    rank: "03"
    title: "SRV Record for a Service"
    excerpt: "Creates an SRV record that advertises the host and port of a service (SIP, XMPP, Minecraft, etc.). SRV records are structured: their priority, weight, port, and target are supplied through the..."
  - slug: "04-caa-certificate-authority"
    rank: "04"
    title: "CAA Record to Restrict Certificate Issuance"
    excerpt: "Creates a CAA record that controls which certificate authorities may issue certificates for your domain. CAA records are structured: their flags, tag, and value are supplied through the `data.caa`..."
---

# DNS Record Presets

Ready-to-deploy configuration presets for DNS Record. Each preset is a complete manifest you can copy, customize, and deploy.
