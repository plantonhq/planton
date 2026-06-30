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
    title: "Preset: Edge API with Custom Domain"
    excerpt: "A production-shaped Worker: deployed from a CI-built bundle in R2, wired to KV and D1 by reference, exposed on a managed custom domain, with observability on."
  - slug: "02-minimal"
    rank: "02"
    title: "Preset: Minimal Worker"
    excerpt: "The smallest deployable Worker: an inline script exposed on a workers.dev subdomain. No external bindings or custom domains."
  - slug: "03-static-site"
    rank: "03"
    title: "Preset: Static Site (Workers Static Assets)"
    excerpt: "A pure static site or single-page app served from Cloudflare's edge — no server-side code. This is the build-and-upload hosting model (the converged successor to Cloudflare Pages): build your site..."
  - slug: "04-full-stack-app"
    rank: "04"
    title: "Preset: Full-Stack App (script + Static Assets)"
    excerpt: "A full-stack application: a Worker script serves dynamic API routes while static assets serve the front end — all from one deployable at the edge. This pairs Cloudflare Workers Static Assets with..."
---

# Worker Presets

Ready-to-deploy configuration presets for Worker. Each preset is a complete manifest you can copy, customize, and deploy.
