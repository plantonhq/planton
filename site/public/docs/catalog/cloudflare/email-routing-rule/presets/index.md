---
title: "Presets"
description: "Ready-to-deploy configuration presets for Email Routing Rule"
type: "preset-list"
componentSlug: "email-routing-rule"
componentTitle: "Email Routing Rule"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-forward-address"
    rank: "01"
    title: "Preset: Forward an Address"
    excerpt: "Forward mail sent to a specific recipient (e.g. `support@`) to one or more verified destination mailboxes."
  - slug: "02-worker"
    rank: "02"
    title: "Preset: Route to an Email Worker"
    excerpt: "Hand mail matching a recipient to an Email Worker for custom processing (parsing, webhooks, storage, auto-responses)."
---

# Email Routing Rule Presets

Ready-to-deploy configuration presets for Email Routing Rule. Each preset is a complete manifest you can copy, customize, and deploy.
