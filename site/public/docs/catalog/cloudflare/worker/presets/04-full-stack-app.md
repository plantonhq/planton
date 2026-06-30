---
title: "Preset: Full-Stack App (script + Static Assets)"
description: "A full-stack application: a Worker script serves dynamic API routes while static assets serve the front end — all from one deployable at the edge. This pairs Cloudflare Workers Static Assets with..."
type: "preset"
rank: "04"
presetSlug: "04-full-stack-app"
componentSlug: "worker"
componentTitle: "Worker"
provider: "cloudflare"
icon: "package"
order: 4
---

# Preset: Full-Stack App (script + Static Assets)

A full-stack application: a Worker script serves dynamic API routes while static
assets serve the front end — all from one deployable at the edge. This pairs
Cloudflare Workers Static Assets with Worker Functions.

## When to use

- An app with both a built front end (SPA/static) and server-side logic (API,
  auth, SSR) on the Workers runtime.
- You want one resource to own the whole app and its backing bindings (KV, D1,
  R2, Queues, …).

## Key choices

- `content` (or `r2Bundle`): the Worker script handling dynamic routes.
- `assets.directory`: the built front-end output.
- `assets.bindingName: ASSETS`: exposes the asset namespace to the script as
  `env.ASSETS`, so code can serve assets explicitly (`env.ASSETS.fetch(request)`).
- `assets.config.runWorkerFirstRules`: routes only matching paths (e.g.
  `/api/*`) through the script; everything else is served as a static asset.
  Use `runWorkerFirst: true` instead to run the script on every request.
- `kvNamespaces` (and other binding groups): wire the app to its backing
  resources via `valueFrom` references.

## Placeholders

| Placeholder | Description |
|---|---|
| `<cloudflare-account-id>` | 32-character Cloudflare account ID |
