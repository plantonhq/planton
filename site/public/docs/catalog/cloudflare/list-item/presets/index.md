---
title: "Presets"
description: "Ready-to-deploy configuration presets for List Item"
type: "preset-list"
componentSlug: "list-item"
componentTitle: "List Item"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-ip-entry"
    rank: "01"
    title: "Preset: IP List Entry"
    excerpt: "Add a single IP/CIDR to an `ip`-kind `CloudflareList`, wired to the list by reference so it composes in an infra chart."
  - slug: "02-redirect-entry"
    rank: "02"
    title: "Preset: Bulk Redirect Entry"
    excerpt: "Add a single source→target redirect to a `redirect`-kind `CloudflareList`. A redirect ruleset resolves the list with `from_list`."
---

# List Item Presets

Ready-to-deploy configuration presets for List Item. Each preset is a complete manifest you can copy, customize, and deploy.
