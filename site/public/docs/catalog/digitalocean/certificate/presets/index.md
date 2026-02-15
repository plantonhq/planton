---
title: "Presets"
description: "Ready-to-deploy configuration presets for Certificate"
type: "preset-list"
componentSlug: "certificate"
componentTitle: "Certificate"
provider: "digitalocean"
icon: "package"
order: 200
presets:
  - slug: "01-lets-encrypt"
    rank: "01"
    title: "Let's Encrypt Certificate"
    excerpt: "This preset creates a free, auto-renewing SSL certificate from Let's Encrypt via DigitalOcean. Supports multiple domains and wildcards. DigitalOcean handles renewal automatically; use the certificate..."
  - slug: "02-custom"
    rank: "02"
    title: "Custom Certificate"
    excerpt: "This preset creates an SSL certificate from user-provided PEM content. Use when you have a certificate from an enterprise CA, a purchased certificate, or a certificate issued outside of Let's..."
---

# Certificate Presets

Ready-to-deploy configuration presets for Certificate. Each preset is a complete manifest you can copy, customize, and deploy.
