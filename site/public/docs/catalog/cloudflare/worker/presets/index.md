---
title: "Presets"
description: "Ready-to-deploy configuration presets for Worker"
type: "preset-list"
componentSlug: "worker"
componentTitle: "Worker"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-api-with-custom-domain"
    rank: "01"
    title: "API Worker with Custom Domain"
    excerpt: "Full-featured Worker with KV bindings, custom domain DNS routing, and environment variables. Use for production APIs that need storage, a custom hostname, and config. Script bundle must be pre-built..."
  - slug: "02-minimal"
    rank: "02"
    title: "Minimal Worker"
    excerpt: "Bare minimum Cloudflare Worker with only the script bundle. No KV bindings, DNS routes, or env vars. Use when deploying a Worker that will be attached to routes or configured elsewhere (e.g., via..."
---

# Worker Presets

Ready-to-deploy configuration presets for Worker. Each preset is a complete manifest you can copy, customize, and deploy.
