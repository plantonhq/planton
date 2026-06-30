---
title: "Preset: Static Site (Workers Static Assets)"
description: "A pure static site or single-page app served from Cloudflare's edge — no server-side code. This is the build-and-upload hosting model (the converged successor to Cloudflare Pages): build your site..."
type: "preset"
rank: "03"
presetSlug: "03-static-site"
componentSlug: "worker"
componentTitle: "Worker"
provider: "cloudflare"
icon: "package"
order: 3
---

# Preset: Static Site (Workers Static Assets)

A pure static site or single-page app served from Cloudflare's edge — no
server-side code. This is the build-and-upload hosting model (the converged
successor to Cloudflare Pages): build your site locally or in CI, and the module
uploads the output directory.

## When to use

- A marketing site, docs site, or SPA (React/Vue/Svelte/Angular build output).
- You build the artifact yourself and want to deploy it as desired state — a new
  deploy with a changed `directory` ships a new version.

## Key choices

- `assets.directory`: your build output directory, uploaded at deploy.
- `assets.config.notFoundHandling: single-page-application`: serves `/index.html`
  for unmatched routes (drop this for a multi-page static site, or use `404-page`
  to serve `/404.html`).
- No script source is set — this Worker is assets-only.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
