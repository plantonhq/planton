---
title: "Presets"
description: "Ready-to-deploy configuration presets for Email Routing (Zone)"
type: "preset-list"
componentSlug: "email-routing-zone"
componentTitle: "Email Routing (Zone)"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-forward-catch-all"
    rank: "01"
    title: "Preset: Forward-All Email Routing"
    excerpt: "Enable Email Routing on a zone and forward all otherwise-unmatched mail to a verified destination address. The simplest \"catch everything → my inbox\" setup."
  - slug: "02-drop-catch-all"
    rank: "02"
    title: "Preset: Drop Catch-All Email Routing"
    excerpt: "Enable Email Routing and drop all mail that no explicit rule matched. Use with per-address `CloudflareEmailRoutingRule`s that forward only the addresses you care about; everything else is silently..."
---

# Email Routing (Zone) Presets

Ready-to-deploy configuration presets for Email Routing (Zone). Each preset is a complete manifest you can copy, customize, and deploy.
