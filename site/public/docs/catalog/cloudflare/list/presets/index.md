---
title: "Presets"
description: "Ready-to-deploy configuration presets for List"
type: "preset-list"
componentSlug: "list"
componentTitle: "List"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-ip-allowlist"
    rank: "01"
    title: "Preset: IP Allowlist"
    excerpt: "An `ip`-kind list to collect trusted IPs/CIDRs that WAF or custom rules reference with `ip.src in $office_allowlist`."
  - slug: "02-bulk-redirect"
    rank: "02"
    title: "Preset: Bulk Redirect List"
    excerpt: "A `redirect`-kind list holding source→target URL rules. A redirect ruleset (`CloudflareRuleset`, http_request_redirect phase) resolves these with `from_list`, enabling large-scale URL redirects..."
---

# List Presets

Ready-to-deploy configuration presets for List. Each preset is a complete manifest you can copy, customize, and deploy.
