---
title: "Presets"
description: "Ready-to-deploy configuration presets for Custom Hostname (Cloudflare for SaaS)"
type: "preset-list"
componentSlug: "custom-hostname-cloudflare-for-saas"
componentTitle: "Custom Hostname (Cloudflare for SaaS)"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-saas-vanity-domain"
    rank: "01"
    title: "Preset: SaaS Vanity Domain (recommended)"
    excerpt: "The recommended default: onboard a customer's hostname with a Cloudflare-issued DV certificate validated over TXT. The customer adds a CNAME and the ownership TXT record (from the stack outputs) and..."
  - slug: "02-byo-certificate"
    rank: "02"
    title: "Preset: Bring Your Own Certificate (Enterprise)"
    excerpt: "For Enterprise accounts that upload their own certificate and key for the custom hostname instead of using a Cloudflare-issued DV certificate, and route it to a specific origin."
---

# Custom Hostname (Cloudflare for SaaS) Presets

Ready-to-deploy configuration presets for Custom Hostname (Cloudflare for SaaS). Each preset is a complete manifest you can copy, customize, and deploy.
