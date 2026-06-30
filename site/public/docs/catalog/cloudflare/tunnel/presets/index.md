---
title: "Presets"
description: "Ready-to-deploy configuration presets for Tunnel"
type: "preset-list"
componentSlug: "tunnel"
componentTitle: "Tunnel"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-public-hostname"
    rank: "01"
    title: "Preset: Publish a private app on a public hostname"
    excerpt: "Expose a single private web app at a public hostname through the tunnel — the most common Cloudflare Tunnel setup."
  - slug: "02-access-protected"
    rank: "02"
    title: "Preset: Access-protected admin hostname"
    excerpt: "Publish an internal admin UI through the tunnel and require Cloudflare Access on every request, wired to an Access application you manage."
  - slug: "03-private-network-connector"
    rank: "03"
    title: "Preset: Private-network connector (for WARP access)"
    excerpt: "A tunnel with no public hostnames, used purely to make private IP ranges reachable to WARP clients via routes."
---

# Tunnel Presets

Ready-to-deploy configuration presets for Tunnel. Each preset is a complete manifest you can copy, customize, and deploy.
