---
title: "Presets"
description: "Ready-to-deploy configuration presets for Certificate Pack"
type: "preset-list"
componentSlug: "certificate-pack"
componentTitle: "Certificate Pack"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-advanced-txt"
    rank: "01"
    title: "Preset: Advanced Certificate (TXT validation)"
    excerpt: "The recommended default: an advanced certificate pack covering the zone apex and a wildcard, issued by Google Trust Services and validated via TXT. For a zone on Cloudflare's nameservers, TXT..."
  - slug: "02-lets-encrypt-annual"
    rank: "02"
    title: "Preset: Let's Encrypt, Apex-Only, Annual"
    excerpt: "A single-hostname (apex) certificate issued by Let's Encrypt with the longest validity (365 days). Useful when you only need to cover the bare domain and prefer a yearly rotation cadence."
---

# Certificate Pack Presets

Ready-to-deploy configuration presets for Certificate Pack. Each preset is a complete manifest you can copy, customize, and deploy.
